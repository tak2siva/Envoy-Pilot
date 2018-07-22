package service

import (
	"Envoy-xDS/cmd/server/model"
	"Envoy-xDS/cmd/server/storage"
)

var singletonClusterService *ClusterService

// ClusterService  a service class for cluster specific functionalities
type ClusterService struct {
	xdsConfigDao *storage.XdsConfigDao
}

// IsOutdated check if the last dispatched config is outdated
func (c *ClusterService) IsOutdated(en *model.EnvoySubscriber) bool {
	return c.xdsConfigDao.GetLatestVersion(en) != en.LastUpdatedVersion
}

// GetClusterService get a singleton instance
func GetClusterService() *ClusterService {
	if singletonClusterService == nil {
		singletonClusterService = &ClusterService{xdsConfigDao: storage.GetXdsConfigDao()}
	}
	return singletonClusterService
}
