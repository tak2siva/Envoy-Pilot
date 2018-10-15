package service

import (
	"Envoy-Pilot/cmd/server/constant"
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/storage"
	"context"
	"log"

	"github.com/rs/xid"
)

var singletonRegisterService *RegisterService
var pollTopics = make(map[string]*model.ConfigMeta)

// RegisterService  a service class for cluster specific functionalities
type RegisterService struct {
	xdsConfigDao    storage.XdsConfigDao
	subscriberDao   *storage.SubscriberDao
	dispatchService *DispatchService
	watchService    *WatchService
}

// GetRegisterService get a singleton instance
func GetRegisterService() *RegisterService {
	if singletonRegisterService == nil {
		singletonRegisterService = &RegisterService{
			xdsConfigDao:    storage.GetXdsConfigDao(),
			subscriberDao:   storage.GetSubscriberDao(),
			dispatchService: GetDispatchService(),
			watchService:    GetWatchService(),
		}
	}
	return singletonRegisterService
}

// RegisterEnvoy register & subscribe new envoy instance
func (c *RegisterService) RegisterEnvoy(ctx context.Context,
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

func (c *RegisterService) RegisterEnvoyADS(ctx context.Context,
	stream XDSStreamServer,
	subscriber *model.EnvoySubscriber, dispatchChannel chan model.ConfigMeta) {
	subscriber.Guid = xid.New().String()
	c.watchService.register(subscriber)
	c.watchService.firstTimeCheck(subscriber, dispatchChannel)
}

// RemoveSubscriber Delete entry
func (c *RegisterService) DeleteSubscriber(subscriber *model.EnvoySubscriber) {
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
