package storage

import (
	"Envoy-Pilot/cmd/server/cache"
	"Envoy-Pilot/cmd/server/model"
	"fmt"
	"log"

	"github.com/rs/xid"
)

type SubscriberDao struct {
	consulWrapper ConsulWrapper
}

var once *SubscriberDao

func GetSubscriberDao() *SubscriberDao {
	if once == nil {
		once = &SubscriberDao{consulWrapper: GetConsulWrapper()}
	}
	return once
}

func (dao *SubscriberDao) RegisterSubscriber(sub *model.EnvoySubscriber) {
	guid := xid.New().String()
	if len(guid) > 0 {
		log.Fatal(fmt.Sprintf("Subscrber %+v registered already", sub))
	}
	sub.Guid = guid
	cache.SUBSCRIBER_CACHE[guid] = sub
}

func (dao *SubscriberDao) DeleteSubscriber(sub *model.EnvoySubscriber) {
	delete(cache.SUBSCRIBER_CACHE, sub.Guid)
}
