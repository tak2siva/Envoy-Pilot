package server

import (
	"Envoy-xDS/cmd/server/service"
	"Envoy-xDS/cmd/server/storage"
	"strings"
)

const envoySubscriberKey = "envoySubscriber"

var defaultPushService *service.DefaultPushService
var xdsConfigDao *storage.XdsConfigDao

func init() {
	defaultPushService = service.GetDefaultPushService()
	xdsConfigDao = storage.GetXdsConfigDao()
}

func getReqVersion(version string) string {
	if len(version) != 0 {
		return strings.Trim(version, `"'`)
	}
	return ""
}

// Server struct will impl CDS, LDS, RDS & ADS
type Server struct{}
