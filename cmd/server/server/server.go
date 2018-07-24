package server

import (
	"Envoy-xDS/cmd/server/service"
	"Envoy-xDS/cmd/server/storage"
	"strings"
)

const envoySubscriberKey = "envoySubscriber"

var clusterService *service.ClusterService
var xdsConfigDao *storage.XdsConfigDao

func init() {
	clusterService = service.GetClusterService()
	xdsConfigDao = storage.GetXdsConfigDao()
}

func getReqVersion(version string) string {
	if len(version) != 0 {
		return strings.Trim(version, `"'`)
	}
	return ""
}

type Server struct{}
