package service

import (
	"Envoy-xDS/cmd/server/model"
	"Envoy-xDS/cmd/server/storage"
)

var singletonClusterService *ClusterService

type ClusterService struct {
	xdsConfigDao *storage.XdsConfigDao
}

func (c *ClusterService) IsOutdated(en *model.EnvoySubscriber) bool {
	return c.xdsConfigDao.GetLatestVersion(en) != en.LastUpdatedVersion
}

func GetClusterService() *ClusterService {
	if singletonClusterService == nil {
		singletonClusterService = &ClusterService{xdsConfigDao: storage.GetXdsConfigDao()}
	}
	return singletonClusterService
}
