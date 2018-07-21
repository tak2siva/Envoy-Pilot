package storage

import (
	"Envoy-xDS/cmd/server/model"
	"log"
)

var once *XdsConfigDao

type XdsConfigDao struct {
	consulWrapper ConsulWrapper
}

func (dao *XdsConfigDao) GetLatestVersion(sub *model.EnvoySubscriber) string {
	return dao.consulWrapper.GetString(sub.BuildRootKey() + "version")
}

func (dao *XdsConfigDao) SetClusterACK(sub *model.EnvoySubscriber, ack string) {
	dao.consulWrapper.Set(sub.BuildInstanceKey()+"/clusterACK/"+ack, "true")
}

func (dao *XdsConfigDao) IsACKPresent(sub *model.EnvoySubscriber, ack string) bool {
	return dao.consulWrapper.Get(sub.BuildInstanceKey()+"/clusterACK/"+ack) != nil
}

func (dao *XdsConfigDao) RegisterSubscriber(sub *model.EnvoySubscriber) {
	id := dao.consulWrapper.GetUniqId()
	sub.Id = id
	dao.consulWrapper.Set(sub.BuildInstanceKey()+"/meta", sub.ToJSON())
}

func (dao *XdsConfigDao) IsRepoPresent(sub *model.EnvoySubscriber) bool {
	if dao.consulWrapper.Get(sub.BuildRootKey()+"version") == nil || dao.consulWrapper.Get(sub.BuildRootKey()+"config") == nil {
		log.Printf("No repo found for %s instance %d\n", sub.BuildRootKey(), sub.Id)
		return false
	}
	return true
}

func GetXdsConfigDao() *XdsConfigDao {
	if once == nil {
		once = &XdsConfigDao{consulWrapper: GetConsulWrapper()}
	}
	return once
}
