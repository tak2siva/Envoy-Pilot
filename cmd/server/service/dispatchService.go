package service

import (
	"Envoy-Pilot/cmd/server/constant"
	"Envoy-Pilot/cmd/server/mapper"
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/storage"
	"Envoy-Pilot/cmd/server/util"
	"context"
	"log"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/google/uuid"
)

var singletonDispatchService *DispatchService

type DispatchService struct {
	xdsConfigDao    storage.XdsConfigDao
	subscriberDao   *storage.SubscriberDao
	clusterMapper   mapper.ClusterMapper
	listenerMapper  mapper.ListenerMapper
	v2HelperService *V2HelperService
}

func GetDispatchService() *DispatchService {
	if singletonDispatchService == nil {
		singletonDispatchService = &DispatchService{
			xdsConfigDao:    storage.GetXdsConfigDao(),
			subscriberDao:   storage.GetSubscriberDao(),
			clusterMapper:   mapper.ClusterMapper{},
			v2HelperService: &V2HelperService{},
		}
	}
	return singletonDispatchService
}

// XDSStreamServer common data type for xDS stream
type XDSStreamServer interface {
	Send(*v2.DiscoveryResponse) error
	Recv() (*v2.DiscoveryRequest, error)
	Context() context.Context
}

func (c *DispatchService) buildDiscoveryResponseFor(subscriber *model.EnvoySubscriber) (*v2.DiscoveryResponse, error) {
	mapper := mapper.GetMapperFor(subscriber.SubscribedTo)
	configJson, version := c.xdsConfigDao.GetConfigJson(subscriber)
	clusterObj, err := mapper.GetResources(configJson)

	if err != nil {
		log.Printf("Unable to build discovery response for %s\n", subscriber.BuildInstanceKey2())
		log.Println(err)
		return nil, err
	}

	responseUUID := uuid.New().String()
	response := &v2.DiscoveryResponse{
		VersionInfo: version,
		Resources:   clusterObj,
		TypeUrl:     c.v2HelperService.GetTypeUrlFor(subscriber.SubscribedTo),
		Nonce:       responseUUID,
	}
	return response, nil
}

func (c *DispatchService) dispatchData(ctx context.Context, stream XDSStreamServer,
	dispatchChannel chan model.ConfigMeta) {
	for updateInfo := range dispatchChannel {
		select {
		case <-ctx.Done():
			return
		default:
		}

		subscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
		// var currentSubscriber *model.EnvoySubscriber
		if subscriber.IsADS() {
			subscriber = subscriber.GetAdsSubscriber(updateInfo.Topic)
			util.CheckNil(subscriber)
		}
		response, err := c.buildDiscoveryResponseFor(subscriber)
		if err != nil {
			log.Panicf("Unable to dispatch for %s\n", subscriber.BuildInstanceKey2())
			continue
		}

		// TODO add log level
		// log.Printf("%+v\n", response)
		// log.Printf("Sending config to %s \n %+v \n", subscriber.BuildInstanceKey2(), response)

		c.subscriberDao.SaveNonce(subscriber, response.Nonce)
		err = stream.Send(response)
		if err != nil {
			log.Println("error sending to client")
			log.Println(err)
			c.subscriberDao.RemoveNonce(subscriber, response.Nonce)
		} else {
			log.Printf("Successfully Sent config to %s \n", subscriber.BuildInstanceKey2())
		}
	}
}

// HandleACK check if the response is an ACK
// if not ignore
// if yes update the last updated version
func (c *DispatchService) HandleACK(subscriber *model.EnvoySubscriber, req *v2.DiscoveryRequest) {
	log.Printf("Received ACK %s from %s", req.ResponseNonce, subscriber.BuildInstanceKey2())
	c.subscriberDao.RemoveNonce(subscriber, req.ResponseNonce)
	subscriber.LastUpdatedVersion = req.VersionInfo
}
