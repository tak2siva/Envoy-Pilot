package storage

import (
	"Envoy-xDS/cmd/server/model"
)

var once *XdsConfigDao

type XdsConfigDao struct {
	consulWrapper ConsulWrapper
}

func (dao *XdsConfigDao) GetLatestVersion(model.EnvoySubscriber) string {
	// dao.consulWrapper.Get()
	return ""
}

func (dao *XdsConfigDao) RegisterSubscriber(sub *model.EnvoySubscriber) {
	id := dao.consulWrapper.GetUniqId()
	sub.Id = id
	dao.consulWrapper.Set(sub.BuildInstanceKey()+"/meta", sub.ToJSON())
}

func GetXdsConfigDao() *XdsConfigDao {
	if once == nil {
		once = &XdsConfigDao{consulWrapper: GetConsulWrapper()}
	}
	return once
}
