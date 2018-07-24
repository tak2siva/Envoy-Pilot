package service

import (
	"Envoy-xDS/cmd/server/constant"
	"Envoy-xDS/cmd/server/mapper"
	"Envoy-xDS/cmd/server/model"
	"Envoy-xDS/cmd/server/storage"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/gogo/protobuf/types"
	"github.com/google/uuid"
)

var singletonClusterService *ClusterService

const envoySubscriberKey = "envoySubscriber"

// ClusterService  a service class for cluster specific functionalities
type ClusterService struct {
	xdsConfigDao   *storage.XdsConfigDao
	clusterMapper  mapper.ClusterMapper
	listenerMapper mapper.ListenerMapper
}

// IsOutdated check if the last dispatched config is outdated
func (c *ClusterService) IsOutdated(en *model.EnvoySubscriber) bool {
	log.Printf("latestVersion: %s --- actualVersion: %s", c.xdsConfigDao.GetLatestVersion(en), en.LastUpdatedVersion)
	return c.xdsConfigDao.GetLatestVersion(en) != en.LastUpdatedVersion
}

func (c *ClusterService) RegisterEnvoy(ctx context.Context,
	stream xDSStreamServer,
	subscriber *model.EnvoySubscriber, dispatchChannel chan bool) {
	c.xdsConfigDao.RegisterSubscriber(subscriber)
	go c.consulPoll(ctx, dispatchChannel)
	go c.dispatchCluster(ctx, stream, dispatchChannel)
}

func (c *ClusterService) consulPoll(ctx context.Context, dispatchChannel chan bool) {
	for {
		time.Sleep(10 * time.Second)
		select {
		case <-ctx.Done():
			return
		default:
		}
		subscriber := ctx.Value(envoySubscriberKey).(*model.EnvoySubscriber)
		log.Printf("Checking consul for %s..\n", subscriber.BuildInstanceKey())
		if !c.xdsConfigDao.IsRepoPresent(subscriber) {
			log.Println("No repo found for subscriber %s\n", subscriber.BuildInstanceKey())
			continue
		}
		if c.IsOutdated(subscriber) {
			log.Printf("Found update dispatching for %s\n", subscriber.BuildInstanceKey())
			dispatchChannel <- true
		}
	}
}

type Mapper interface {
	GetResources(configJson string) ([]types.Any, error)
}

func (c *ClusterService) getMapperFor(topic string) Mapper {
	switch topic {
	case constant.SUBSCRIBE_CDS:
		return &mapper.ClusterMapper{}
	case constant.SUBSCRIBE_LDS:
		return &mapper.ListenerMapper{}
	case constant.SUBSCRIBE_RDS:
		return &mapper.RouteMapper{}
	default:
		panic(fmt.Sprintf("No mapper found for type %s\n", topic))
	}
}

func (c *ClusterService) getTypeUrlFor(topic string) string {
	switch topic {
	case constant.SUBSCRIBE_CDS:
		return cache.ClusterType
	case constant.SUBSCRIBE_LDS:
		return cache.ListenerType
	case constant.SUBSCRIBE_RDS:
		return cache.RouteType
	default:
		panic(fmt.Sprintf("No TypeUrl found for type %s\n", topic))
	}
}

func (c *ClusterService) buildDiscoveryResponseFor(subscriber *model.EnvoySubscriber) (*v2.DiscoveryResponse, error) {
	mapper := c.getMapperFor(subscriber.SubscribedTo)
	configJson, version := c.xdsConfigDao.GetConfigJson(subscriber)
	clusterObj, err := mapper.GetResources(configJson)

	if err != nil {
		log.Printf("Unable to build discovery response for %s\n", subscriber.BuildInstanceKey())
		log.Println(err)
		return nil, err
	}

	responseUUID := uuid.New().String()
	response := &v2.DiscoveryResponse{
		VersionInfo: version,
		Resources:   clusterObj,
		TypeUrl:     c.getTypeUrlFor(subscriber.SubscribedTo),
		Nonce:       responseUUID,
	}
	return response, nil
}

type xDSStreamServer interface {
	Send(*v2.DiscoveryResponse) error
}

func (c *ClusterService) dispatchCluster(ctx context.Context, stream xDSStreamServer,
	dispatchChannel chan bool) {
	for range dispatchChannel {
		select {
		case <-ctx.Done():
			return
		default:
		}

		subscriber := ctx.Value(envoySubscriberKey).(*model.EnvoySubscriber)
		response, err := c.buildDiscoveryResponseFor(subscriber)
		if err != nil {
			log.Panicf("Unable to dispatch for %s\n", subscriber.BuildInstanceKey())
			continue
		}

		log.Printf("%+v\n", response)
		log.Printf("Sending config to %s \n %+v \n", subscriber.BuildInstanceKey(), response)

		c.xdsConfigDao.SaveNonce(subscriber, response.Nonce)
		err = stream.Send(response)
		if err != nil {
			log.Println("error sending to client")
			log.Println(err)
			c.xdsConfigDao.RemoveNonce(subscriber, response.Nonce)
		} else {
			log.Printf("Successfully Sent config to %s \n", subscriber.BuildInstanceKey())
		}
	}
}

func (c *ClusterService) HandleACK(subscriber *model.EnvoySubscriber, req *v2.DiscoveryRequest) {
	log.Printf("Received ACK %s from %s", req.ResponseNonce, subscriber.BuildInstanceKey())
	c.xdsConfigDao.RemoveNonce(subscriber, req.ResponseNonce)
	subscriber.LastUpdatedVersion = req.VersionInfo
	c.xdsConfigDao.UpdateEnvoySubscriber(subscriber)
}

// GetClusterService get a singleton instance
func GetClusterService() *ClusterService {
	if singletonClusterService == nil {
		singletonClusterService = &ClusterService{
			xdsConfigDao:  storage.GetXdsConfigDao(),
			clusterMapper: mapper.ClusterMapper{},
		}
	}
	return singletonClusterService
}
