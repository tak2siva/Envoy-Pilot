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
	"github.com/rs/xid"
)

var singletonDefaultPushService *DefaultPushService
var versionChangeChannel = make(chan model.ConfigMeta)
var pollTopics = make(map[string]*model.ConfigMeta)

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
		log.Printf("[Global] Found update %s --> %s for  %s\n", actual, latest, subscribedTopic)
	}
	return res
}

// RegisterEnvoy register & subscribe new envoy instance
func (c *DefaultPushService) RegisterEnvoy(ctx context.Context,
	stream XDSStreamServer,
	subscriber *model.EnvoySubscriber, dispatchChannel chan model.ConfigMeta) {
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

func (c *DefaultPushService) RegisterEnvoyADS(ctx context.Context,
	stream XDSStreamServer,
	subscriber *model.EnvoySubscriber, dispatchChannel chan model.ConfigMeta) {
	subscriber.Guid = xid.New().String()
	c.register(subscriber)
	c.firstTimeCheck(subscriber, dispatchChannel)
}

// RemoveSubscriber Delete entry
func (c *DefaultPushService) DeleteSubscriber(subscriber *model.EnvoySubscriber) {
	c.subscriberDao.DeleteSubscriber(subscriber)
	log.Printf("Deleting subscriber %s\n", subscriber.BuildInstanceKey2())
	if subscriber.IsADS() {
		for _, topic := range constant.SUPPORTED_TYPES {
			sub := subscriber.AdsList[topic]
			if sub != nil {
				delete(pollTopics, sub.BuildRootKey())
			}
		}
	} else {
		delete(pollTopics, subscriber.BuildRootKey())
	}
}

func (c *DefaultPushService) firstTimeCheck(subscriber *model.EnvoySubscriber, dispatchChannel chan model.ConfigMeta) {
	if !c.xdsConfigDao.IsRepoPresent(subscriber) {
		log.Printf("No repo found for subscriber %s\n", subscriber.BuildRootKey())
	} else {
		latestVersion := c.xdsConfigDao.GetLatestVersion(subscriber)
		if subscriber.IsOutdated(latestVersion) {
			log.Printf("Found update %s --> %s dispatching for %s\n", subscriber.LastUpdatedVersion, latestVersion, subscriber.BuildInstanceKey2())
			dispatchChannel <- model.ConfigMeta{Key: subscriber.BuildRootKey(), Topic: subscriber.SubscribedTo, Version: latestVersion}
		} else {
			log.Printf("Already Upto date %s\n", subscriber.BuildInstanceKey2())
		}
	}
}

func (c *DefaultPushService) listenForUpdates(ctx context.Context, dispatchChannel chan model.ConfigMeta) {
	subscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	util.CheckNil(subscriber)
	c.registerPollTopic(ctx)
	c.firstTimeCheck(subscriber, dispatchChannel)

	for message := range versionChangeChannel {
		if message.Key == subscriber.BuildRootKey() {
			if subscriber.IsOutdated(message.Version) {
				log.Printf("Found update %s --> %s dispatching for %s\n", subscriber.LastUpdatedVersion, message.Version, subscriber.BuildInstanceKey2())
				dispatchChannel <- message
			}
		}
	}
}

func (c *DefaultPushService) listenForUpdatesADS(ctx context.Context, dispatchChannel chan model.ConfigMeta) {
	adsSubscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	util.CheckNil(adsSubscriber)
	c.registerPollTopicADS(ctx)

	for _, subscriber := range adsSubscriber.AdsList {
		c.firstTimeCheck(subscriber, dispatchChannel)
	}

	for message := range versionChangeChannel {
		for _, subscriber := range adsSubscriber.AdsList {
			if message.Key == subscriber.BuildRootKey() {
				if subscriber.IsOutdated(message.Version) {
					log.Printf("Found update %s --> %s dispatching for %s\n", subscriber.LastUpdatedVersion, message.Version, subscriber.BuildInstanceKey2())
					dispatchChannel <- message
				}
			}
		}
	}
}

func (c *DefaultPushService) registerPollTopic(ctx context.Context) {
	subscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	c.register(subscriber)
}

func (c *DefaultPushService) registerPollTopicADS(ctx context.Context) {
	adsSubscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	for _, topic := range constant.SUPPORTED_TYPES {
		subscriber := adsSubscriber.AdsList[topic]
		if subscriber != nil {
			c.register(subscriber)
		}
	}
}

func (c *DefaultPushService) register(subscriber *model.EnvoySubscriber) {
	if _, ok := pollTopics[subscriber.BuildRootKey()]; !ok {
		pollTopics[subscriber.BuildRootKey()] = &model.ConfigMeta{Key: subscriber.BuildRootKey(), Topic: subscriber.SubscribedTo, Version: subscriber.LastUpdatedVersion}
	}
}

func ConsulPollLoop() {
	pushService := GetDefaultPushService()
	log.Printf("Starting Consul Poll Loop..\n")
	for {
		time.Sleep(10 * time.Second)
		for configKey, configMeta := range pollTopics {
			latestVersion := pushService.xdsConfigDao.GetLatestVersionFor(configKey)
			meta := model.ConfigMeta{Key: configKey, Topic: configMeta.Topic, Version: latestVersion}
			versionChangeChannel <- meta
			pollTopics[configKey] = &meta
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
	c.subscriberDao.RemoveNonce(subscriber, req.ResponseNonce)
	subscriber.LastUpdatedVersion = req.VersionInfo
}
