package manager

import (
	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

var (
	nonceMap           = make(map[string]*v2.DiscoveryResponse)
	versionMap         = make(map[string]bool)
	lastUpdatedVersion string
)

func IsACK(req *v2.DiscoveryRequest) bool {
	if _, ok := nonceMap[req.ResponseNonce]; ok {
		return true
	}
	return false
}

func IsOutDated(versionInfo string) bool {
	if _, ok := versionMap[versionInfo]; ok {
		return false
	}
	return true
}

func UpdateMap(resp *v2.DiscoveryResponse) {
	nonceMap[resp.Nonce] = resp
	lastUpdatedVersion = resp.VersionInfo
}
