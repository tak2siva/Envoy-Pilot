package storage

import (
	consul "github.com/hashicorp/consul/api"
)

type XdsConfigDao struct {
	consulHandle *consul.KV
}

func (dao *XdsConfigDao) GetLatestVersion() string {
	pair, _, err := dao.consulHandle.Get("foo", nil)
	if err != nil {
		panic(err)
	}
	return string(pair.Value)
}
