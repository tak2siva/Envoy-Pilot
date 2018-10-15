package storage

import (
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/util"
)

var once *ConsulConfigDao

type ConsulConfigDao struct {
	consulWrapper ConsulWrapper
}

func (dao *ConsulConfigDao) GetLatestVersion(sub *model.EnvoySubscriber) string {
	return util.TrimVersion(dao.consulWrapper.GetString(sub.BuildRootKey() + "version"))
}

func (dao *ConsulConfigDao) GetLatestVersionFor(subscriberKey string) string {
	return util.TrimVersion(dao.consulWrapper.GetString(subscriberKey + "version"))
}

func (dao *ConsulConfigDao) IsRepoPresent(sub *model.EnvoySubscriber) bool {
	if dao.consulWrapper.Get(sub.BuildRootKey()+"version") == nil || dao.consulWrapper.Get(sub.BuildRootKey()+"config") == nil {
		return false
	}
	return true
}

func (dao *ConsulConfigDao) GetConfigJson(sub *model.EnvoySubscriber) (string, string) {
	return dao.consulWrapper.GetString(sub.BuildRootKey() + "config"), dao.GetLatestVersion(sub)
}

func GetConsulConfigDao() *ConsulConfigDao {
	if once == nil {
		once = &ConsulConfigDao{consulWrapper: GetConsulWrapper()}
	}
	return once
}
