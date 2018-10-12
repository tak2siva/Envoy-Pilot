package storage

import (
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/util"
	"fmt"
)

var once *XdsConfigDao

type XdsConfigDao struct {
	consulWrapper ConsulWrapper
}

func (dao *XdsConfigDao) GetLatestVersion(sub *model.EnvoySubscriber) string {
	return util.TrimVersion(dao.consulWrapper.GetString(sub.BuildRootKey() + "version"))
}

func (dao *XdsConfigDao) GetLatestVersionFor(subscriberKey string) string {
	return util.TrimVersion(dao.consulWrapper.GetString(subscriberKey + "version"))
}

func (dao *XdsConfigDao) IsRepoPresent(sub *model.EnvoySubscriber) bool {
	if dao.consulWrapper.Get(sub.BuildRootKey()+"version") == nil || dao.consulWrapper.Get(sub.BuildRootKey()+"config") == nil {
		return false
	}
	return true
}

func (dao *XdsConfigDao) GetConfigJson(sub *model.EnvoySubscriber) (string, string) {
	return dao.consulWrapper.GetString(sub.BuildRootKey() + "config"), dao.GetLatestVersion(sub)
}

// func (dao *XdsConfigDao) UpdateEnvoySubscriber(sub *model.EnvoySubscriber) {
// 	log.Printf("Updating envoy subscriber %+v\n", sub)
// 	dao.consulWrapper.Set(metaKey(sub), sub.ToJSON())
// }

func nonceStreamKey(sub *model.EnvoySubscriber, nonce string) string {
	return fmt.Sprintf("%s/Nonce/Stream/%s", sub.BuildInstanceKey(), nonce)
}

func metaKey(sub *model.EnvoySubscriber) string {
	return sub.BuildInstanceKey() + "/meta"
}

func GetXdsConfigDao() *XdsConfigDao {
	if once == nil {
		once = &XdsConfigDao{consulWrapper: GetConsulWrapper()}
	}
	return once
}
