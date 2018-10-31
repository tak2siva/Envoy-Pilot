package service

import (
	"Envoy-Pilot/cmd/server/constant"
	"Envoy-Pilot/cmd/server/metrics"
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/storage"
	"Envoy-Pilot/cmd/server/util"
	"context"
	"log"
	"time"
)

var singletonWatchService *WatchService
var versionChangeChannel = make(chan model.ConfigMeta)

type WatchService struct {
	xdsConfigDao    storage.XdsConfigDao
	subscriberDao   *storage.SubscriberDao
	dispatchService *DispatchService
}

// WatchService get a singleton instance
func GetWatchService() *WatchService {
	if singletonWatchService == nil {
		singletonWatchService = &WatchService{
			xdsConfigDao:    storage.GetXdsConfigDao(),
			subscriberDao:   storage.GetSubscriberDao(),
			dispatchService: GetDispatchService(),
		}
	}
	return singletonWatchService
}

func (c *WatchService) firstTimeCheck(subscriber *model.EnvoySubscriber, dispatchChannel chan model.ConfigMeta) {
	if !c.xdsConfigDao.IsRepoPresent(subscriber) {
		log.Printf("No repo found for subscriber %s\n", subscriber.BuildRootKey())
	} else {
		latestVersion := c.xdsConfigDao.GetLatestVersion(subscriber)
		if subscriber.IsOutdated(latestVersion) {
			log.Printf("Found update %s --> %s dispatching for %s\n", subscriber.LastUpdatedVersion, latestVersion, subscriber.BuildInstanceKey2())
			dispatchChannel <- model.ConfigMeta{Key: subscriber.BuildRootKey(), Topic: subscriber.SubscribedTo, Version: latestVersion}
			metrics.IncXdsUpdateCounter(subscriber)
		} else {
			log.Printf("Already Upto date %s[%s:%s]\n", subscriber.BuildInstanceKey2(), subscriber.LastUpdatedVersion, latestVersion)
		}
	}
}

func (c *WatchService) listenForUpdates(ctx context.Context, dispatchChannel chan model.ConfigMeta) {
	subscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	util.CheckNil(subscriber)
	c.registerPollTopic(ctx)
	c.firstTimeCheck(subscriber, dispatchChannel)

	for message := range versionChangeChannel {
		if message.Key == subscriber.BuildRootKey() {
			if subscriber.IsOutdated(message.Version) {
				log.Printf("Found update %s --> %s dispatching for %s\n", subscriber.LastUpdatedVersion, message.Version, subscriber.BuildInstanceKey2())
				dispatchChannel <- message
				metrics.IncXdsUpdateCounter(subscriber)
			}
		}
	}
}

func (c *WatchService) listenForUpdatesADS(ctx context.Context, dispatchChannel chan model.ConfigMeta) {
	adsSubscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	util.CheckNil(adsSubscriber)
	c.registerPollTopicADS(ctx)

	for _, subscriber := range adsSubscriber.AdsList {
		c.firstTimeCheck(subscriber, dispatchChannel)
	}

	for message := range versionChangeChannel {
		for _, subscriber := range adsSubscriber.AdsList {
			if message.Key == subscriber.BuildRootKey() {
				if subscriber.IsOutdated(message.Version) && subscriber.SubscribedTo == message.Topic {
					log.Printf("Found update %s --> %s dispatching for %s\n", subscriber.LastUpdatedVersion, message.Version, subscriber.BuildInstanceKey2())
					dispatchChannel <- message
					metrics.IncXdsUpdateCounter(subscriber)
				}
			}
		}
	}
}

func (c *WatchService) register(subscriber *model.EnvoySubscriber) {
	if _, ok := pollTopics[subscriber.BuildRootKey()]; !ok {
		pollTopics[subscriber.BuildRootKey()] = &model.ConfigMeta{Key: subscriber.BuildRootKey(), Topic: subscriber.SubscribedTo, Version: subscriber.LastUpdatedVersion}
	}
}

func (c *WatchService) registerPollTopic(ctx context.Context) {
	subscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	c.register(subscriber)
}

func (c *WatchService) registerPollTopicADS(ctx context.Context) {
	adsSubscriber := ctx.Value(constant.ENVOY_SUBSCRIBER_KEY).(*model.EnvoySubscriber)
	for _, topic := range constant.SUPPORTED_TYPES {
		subscriber := adsSubscriber.AdsList[topic]
		if subscriber != nil {
			c.register(subscriber)
		}
	}
}

func ConsulPollLoop() {
	pushService := GetRegisterService()
	log.Printf("Starting Poll Loop..\n")
	for {
		time.Sleep(constant.POLL_INTERVAL)
		for configKey, configMeta := range pollTopics {
			latestVersion := pushService.xdsConfigDao.GetLatestVersionFor(configKey)
			if pushService.xdsConfigDao.IsRepoPresentFor(configKey) {
				meta := model.ConfigMeta{Key: configKey, Topic: configMeta.Topic, Version: latestVersion}
				versionChangeChannel <- meta
				pollTopics[configKey] = &meta
			}
		}
	}
}
