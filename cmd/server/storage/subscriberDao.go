package storage

import (
	"Envoy-Pilot/cmd/server/cache"
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/util"
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
	if util.SyncMapExists(&cache.SUBSCRIBER_CACHE, guid) {
		log.Printf("---%s---\n", sub.Guid)
		log.Fatal(fmt.Sprintf("Subscrber %+v registered already", sub))
	}
	sub.Guid = guid
	util.SyncMapSet(&cache.SUBSCRIBER_CACHE, guid, sub)
}

func (dao *SubscriberDao) DeleteSubscriber(sub *model.EnvoySubscriber) {
	util.SyncMapDelete(&cache.SUBSCRIBER_CACHE, sub.Guid)
}

func (dao *SubscriberDao) SaveNonce(sub *model.EnvoySubscriber, nonce string) {
	log.Printf("Writing ACK %s\n", nonceStreamKey(sub, nonce))
	util.SyncMapSet(&cache.NONCE_CACHE, nonceStreamKey(sub, nonce), true)
}

func (dao *SubscriberDao) IsACK(sub *model.EnvoySubscriber, ack string) bool {
	return util.SyncMapExists(&cache.NONCE_CACHE, nonceStreamKey(sub, ack))
}

func (dao *SubscriberDao) RemoveNonce(sub *model.EnvoySubscriber, nonce string) {
	util.SyncMapDelete(&cache.NONCE_CACHE, nonceStreamKey(sub, nonce))
}
