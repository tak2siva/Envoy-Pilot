package storage

import (
	"Envoy-xDS/cmd/server/model"
	"fmt"
	"log"
)

var once *XdsConfigDao

type XdsConfigDao struct {
	consulWrapper ConsulWrapper
}

func (dao *XdsConfigDao) GetLatestVersion(sub *model.EnvoySubscriber) string {
	return dao.consulWrapper.GetString(sub.BuildRootKey() + "version")
}

func (dao *XdsConfigDao) RegisterSubscriber(sub *model.EnvoySubscriber) {
	id := dao.consulWrapper.GetUniqId()
	sub.Id = id
	dao.consulWrapper.Set(metaKey(sub), sub.ToJSON())
	log.Printf("Registered new subscriber %s", sub.BuildInstanceKey())
}

func (dao *XdsConfigDao) IsRepoPresent(sub *model.EnvoySubscriber) bool {
	if dao.consulWrapper.Get(sub.BuildRootKey()+"version") == nil || dao.consulWrapper.Get(sub.BuildRootKey()+"config") == nil {
		log.Printf("No repo found for %s instance %d\n", sub.BuildRootKey(), sub.Id)
		return false
	}
	return true
}

func (dao *XdsConfigDao) GetConfigJson(sub *model.EnvoySubscriber) (string, string) {
	return dao.consulWrapper.GetString(sub.BuildRootKey() + "config"), dao.GetLatestVersion(sub)
}

func (dao *XdsConfigDao) SaveNonce(sub *model.EnvoySubscriber, nonce string) {
	dao.consulWrapper.Set(nonceStreamKey(sub, nonce), "true")
	log.Printf("Writing ACK %s\n", nonceStreamKey(sub, nonce))
}

func (dao *XdsConfigDao) IsACK(sub *model.EnvoySubscriber, ack string) bool {
	return dao.consulWrapper.Get(nonceStreamKey(sub, ack)) != nil
}

func (dao *XdsConfigDao) IsListenerACK(sub *model.EnvoySubscriber, ack string) bool {
	return dao.consulWrapper.Get(nonceStreamKey(sub, ack)) != nil
}

func (dao *XdsConfigDao) RemoveNonce(sub *model.EnvoySubscriber, nonce string) {
	err := dao.consulWrapper.Delete(nonceStreamKey(sub, nonce))
	if err != nil {
		log.Printf("Error deleting nonce %s\n", nonceStreamKey(sub, nonce))
	}
}

func (dao *XdsConfigDao) RemoveListenerNonce(sub *model.EnvoySubscriber, nonce string) {
	err := dao.consulWrapper.Delete(nonceStreamKey(sub, nonce))
	if err != nil {
		log.Printf("Error deleting nonce %s\n", nonceStreamKey(sub, nonce))
	}
}

func (dao *XdsConfigDao) UpdateEnvoySubscriber(sub *model.EnvoySubscriber) {
	log.Printf("Updating envoy subscriber %+v\n", sub)
	dao.consulWrapper.Set(metaKey(sub), sub.ToJSON())
}

func nonceStreamKey(sub *model.EnvoySubscriber, nonce string) string {
	return fmt.Sprintf("%s/Nonce/Stream/%s", sub.BuildInstanceKey(), nonce)
}

// func nonceListenerKey(sub *model.EnvoySubscriber, nonce string) string {
// 	return sub.BuildInstanceKey() + "/Nonce/StreamListeners/" + nonce
// }

func metaKey(sub *model.EnvoySubscriber) string {
	return sub.BuildInstanceKey() + "/meta"
}

func GetXdsConfigDao() *XdsConfigDao {
	if once == nil {
		once = &XdsConfigDao{consulWrapper: GetConsulWrapper()}
	}
	return once
}
