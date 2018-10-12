package service

import (
	"Envoy-Pilot/cmd/server/constant"
	"Envoy-Pilot/cmd/server/mapper"
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/storage"
	"Envoy-Pilot/cmd/server/util"
	"context"
	"log"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/google/uuid"
)

var singletonDefaultPushService *DefaultPushService
var versionChangeChannel = make(chan string)
var pollTopics = make(map[string]string)

// DefaultPushService  a service class for cluster specific functionalities
type DefaultPushService struct {
	xdsConfigDao    *storage.XdsConfigDao
	subscriberDao   *storage.SubscriberDao
	clusterMapper   mapper.ClusterMapper
	listenerMapper  mapper.ListenerMapper
	v2HelperService *V2HelperService
}

// GetDefaultPushService get a singleton instance
func GetDefaultPushService() *DefaultPushService {
	if singletonDefaultPushService == nil {
		singletonDefaultPushService = &DefaultPushService{
			xdsConfigDao:    storage.GetXdsConfigDao(),
			subscriberDao:   storage.GetSubscriberDao(),
			clusterMapper:   mapper.ClusterMapper{},
			v2HelperService: &V2HelperService{},
		}
	}
	return singletonDefaultPushService
}

// IsOutdated check if the last dispatched config is outdated
func (c *DefaultPushService) IsOutdated(en *model.EnvoySubscriber) bool {
	latest := util.TrimVersion(c.xdsConfigDao.GetLatestVersion(en))
	actual := util.TrimVersion(en.LastUpdatedVersion)
	res := latest != actual
	if res {
		log.Printf("Found update actual: %s --- latest: %s for  %s\n", actual, latest, en.BuildInstanceKey())
	}
	return res
}

func (c *DefaultPushService) IsOutdated2(subscribedTopic string, lastVersion string) bool {
	latest := util.TrimVersion(c.xdsConfigDao.GetLatestVersionFor(subscribedTopic))
	actual := util.TrimVersion(lastVersion)
	res := latest != actual
	if res {
		log.Printf("Found update actual: %s --- latest: %s for  %s\n", actual, latest, subscribedTopic)
	}
	return res
}

// RegisterEnvoy register & subscribe new envoy instance
func (c *DefaultPushService) RegisterEnvoy(ctx context.Context,
	stream XDSStreamServer,
	subscriber *model.EnvoySubscriber, dispatchChannel chan string) {
	if subscriber.IsADS() {
		c.subscriberDao.RegisterSubscriber(subscriber)
		go c.listenForUpdatesADS(ctx, dispatchChannel)
		go c.dispatchData(ctx, stream, dispatchChannel)
	} else {
		c.subscriberDao.RegisterSubscriber(subscriber)
		go c.listenForUpdates(ctx, dispatchChannel)
		go c.dispatchData(ctx, stream, dispatchChannel)
	}
}

// RemoveSubscriber Delete entry
func (c *DefaultPushService) DeleteSubscriber(subscriber *model.EnvoySubscriber) {
	c.subscriberDao.DeleteSubscriber(subscriber)
}

func (c *DefaultPushService) listenForUpdates(ctx context.Context, dispatchChannel chan string) {
	i := 0
	subscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	c.registerPollTopic(ctx, dispatchChannel)

	if !c.xdsConfigDao.IsRepoPresent(subscriber) {
		log.Printf("No repo found for subscriber %s\n", subscriber.BuildInstanceKey())
	}
	for message := range versionChangeChannel {
		if message == subscriber.BuildRootKey() {
			if i == 0 {
				// Move Up
				log.Printf("[%s] latestVersion: %s --- actualVersion: %s\n", subscriber.BuildInstanceKey(), c.xdsConfigDao.GetLatestVersion(subscriber), subscriber.LastUpdatedVersion)
				i++
			}

			// In complete
			// Update version cache after
			// Use new isOutdated
			// init updateLoop

			if c.IsOutdated(subscriber) {
				log.Printf("Found update dispatching for %s\n", subscriber.BuildInstanceKey())
				dispatchChannel <- ""
			}
		}
	}
}

func (c *DefaultPushService) listenForUpdatesADS(ctx context.Context, dispatchChannel chan string) {
	i := 0
	subscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	c.registerPollTopicADS(ctx, dispatchChannel)

	if !c.xdsConfigDao.IsRepoPresent(subscriber) {
		log.Printf("No repo found for subscriber %s\n", subscriber.BuildInstanceKey())
	}
	for message := range versionChangeChannel {
		// Match any of ads members
		// Refactor
		hasUpdate := false
		for _, sub := range subscriber.AdsList {
			hasUpdate = message == sub.BuildRootKey()
		}
		if hasUpdate {
			if i == 0 {
				// Move Up
				log.Printf("[%s] latestVersion: %s --- actualVersion: %s\n", subscriber.BuildInstanceKey(), c.xdsConfigDao.GetLatestVersion(subscriber), subscriber.LastUpdatedVersion)
				i++
			}

			// In complete

			if c.IsOutdated(subscriber) {
				log.Printf("Found update dispatching for %s\n", subscriber.BuildInstanceKey())
				dispatchChannel <- ""
			}
		}
	}
}

func (c *DefaultPushService) registerPollTopic(ctx context.Context, dispatchChannel chan string) {
	subscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	pollTopics[subscriber.BuildRootKey()] = "nil"
}

func (c *DefaultPushService) registerPollTopicADS(ctx context.Context, dispatchChannel chan string) {
	adsSubscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	for _, topic := range constant.SUPPORTED_TYPES {
		subscriber := adsSubscriber.AdsList[topic]
		pollTopics[subscriber.BuildRootKey()] = "nil"
	}
}

func (c *DefaultPushService) consulPollLoop(ctx context.Context, dispatchChannel chan string) {
	for {
		time.Sleep(10 * time.Second)
		// Check for updates
		// Update
		for topic, version := range pollTopics {
			if c.IsOutdated2(topic, version) {
				versionChangeChannel <- topic
			}
		}
	}
}

func (c *DefaultPushService) buildDiscoveryResponseFor(subscriber *model.EnvoySubscriber) (*v2.DiscoveryResponse, error) {
	mapper := mapper.GetMapperFor(subscriber.SubscribedTo)
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
		TypeUrl:     c.v2HelperService.GetTypeUrlFor(subscriber.SubscribedTo),
		Nonce:       responseUUID,
	}
	return response, nil
}

// XDSStreamServer common data type for xDS stream
type XDSStreamServer interface {
	Send(*v2.DiscoveryResponse) error
	Recv() (*v2.DiscoveryRequest, error)
	Context() context.Context
}

func (c *DefaultPushService) dispatchData(ctx context.Context, stream XDSStreamServer,
	dispatchChannel chan string) {
	for topic := range dispatchChannel {
		select {
		case <-ctx.Done():
			return
		default:
		}

		subscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
		// var currentSubscriber *model.EnvoySubscriber
		if subscriber.IsADS() {
			subscriber = subscriber.GetAdsSubscriber(topic)
		}
		response, err := c.buildDiscoveryResponseFor(subscriber)
		if err != nil {
			log.Panicf("Unable to dispatch for %s\n", subscriber.BuildInstanceKey())
			continue
		}

		// TODO add log level
		// log.Printf("%+v\n", response)
		// log.Printf("Sending config to %s \n %+v \n", subscriber.BuildInstanceKey(), response)

		c.subscriberDao.SaveNonce(subscriber, response.Nonce)
		err = stream.Send(response)
		if err != nil {
			log.Println("error sending to client")
			log.Println(err)
			c.subscriberDao.RemoveNonce(subscriber, response.Nonce)
		} else {
			log.Printf("Successfully Sent config to %s \n", subscriber.BuildInstanceKey())
		}
	}
}

// HandleACK check if the response is an ACK
// if not ignore
// if yes update the last updated version
func (c *DefaultPushService) HandleACK(subscriber *model.EnvoySubscriber, req *v2.DiscoveryRequest) {
	log.Printf("Received ACK %s from %s", req.ResponseNonce, subscriber.BuildInstanceKey())
	// c.xdsConfigDao.RemoveNonce(subscriber, req.ResponseNonce)
	c.subscriberDao.RemoveNonce(subscriber, req.ResponseNonce)
	subscriber.LastUpdatedVersion = req.VersionInfo
	// c.xdsConfigDao.UpdateEnvoySubscriber(subscriber)
}
