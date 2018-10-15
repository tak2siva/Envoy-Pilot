package service

import (
	"Envoy-Pilot/cmd/server/constant"
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/storage"
	"context"
	"log"

	"github.com/rs/xid"
)

var singletonDefaultPushService *DefaultPushService
var versionChangeChannel = make(chan model.ConfigMeta)
var pollTopics = make(map[string]*model.ConfigMeta)

// DefaultPushService  a service class for cluster specific functionalities
type DefaultPushService struct {
	xdsConfigDao    *storage.XdsConfigDao
	subscriberDao   *storage.SubscriberDao
	dispatchService *DispatchService
	watchService    *WatchService
}

// GetDefaultPushService get a singleton instance
func GetDefaultPushService() *DefaultPushService {
	if singletonDefaultPushService == nil {
		singletonDefaultPushService = &DefaultPushService{
			xdsConfigDao:    storage.GetXdsConfigDao(),
			subscriberDao:   storage.GetSubscriberDao(),
			dispatchService: GetDispatchService(),
			watchService:    GetWatchService(),
		}
	}
	return singletonDefaultPushService
}

// RegisterEnvoy register & subscribe new envoy instance
func (c *DefaultPushService) RegisterEnvoy(ctx context.Context,
	stream XDSStreamServer,
	subscriber *model.EnvoySubscriber, dispatchChannel chan model.ConfigMeta) {
	if subscriber.IsADS() {
		c.subscriberDao.RegisterSubscriber(subscriber)
		go c.watchService.listenForUpdatesADS(ctx, dispatchChannel)
		go c.dispatchService.dispatchData(ctx, stream, dispatchChannel)
	} else {
		c.subscriberDao.RegisterSubscriber(subscriber)
		go c.watchService.listenForUpdates(ctx, dispatchChannel)
		go c.dispatchService.dispatchData(ctx, stream, dispatchChannel)
	}
}

func (c *DefaultPushService) RegisterEnvoyADS(ctx context.Context,
	stream XDSStreamServer,
	subscriber *model.EnvoySubscriber, dispatchChannel chan model.ConfigMeta) {
	subscriber.Guid = xid.New().String()
	c.watchService.register(subscriber)
	c.watchService.firstTimeCheck(subscriber, dispatchChannel)
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
