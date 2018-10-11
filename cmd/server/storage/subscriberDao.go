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

var onceSubscriberDao *SubscriberDao

func GetSubscriberDao() *SubscriberDao {
	if onceSubscriberDao == nil {
		onceSubscriberDao = &SubscriberDao{consulWrapper: GetConsulWrapper()}
	}
	return onceSubscriberDao
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

func (dao *SubscriberDao) SaveNonce(sub *model.EnvoySubscriber, nonce string) {
	log.Printf("Writing ACK %s\n", nonceStreamKey(sub, nonce))
	cache.NONCE_CACHE[nonceStreamKey(sub, nonce)] = true
}

func (dao *SubscriberDao) IsACK(sub *model.EnvoySubscriber, ack string) bool {
	_, ok := cache.NONCE_CACHE[nonceStreamKey(sub, ack)]
	return ok
}

func (dao *SubscriberDao) RemoveNonce(sub *model.EnvoySubscriber, nonce string) {
	delete(cache.NONCE_CACHE, nonceStreamKey(sub, nonce))
}
