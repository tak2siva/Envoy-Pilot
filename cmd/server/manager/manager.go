package manager

import (
	"fmt"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

var (
	nonceMap   = make(map[string]*v2.DiscoveryResponse)
	versionMap = make(map[string]bool)
)

func IsACK(req *v2.DiscoveryRequest) bool {
	if _, ok := nonceMap[req.ResponseNonce]; ok {
		return true
	}
	return false
}

func IsOutDated(req *v2.DiscoveryRequest) bool {
	if _, ok := versionMap[req.VersionInfo]; ok {
		return false
	}
	return true
}

func UpdateMap(nonceChannel chan *v2.DiscoveryResponse) {
	for {
		record, isOpen := <-nonceChannel
		if isOpen {
			nonceMap[record.Nonce] = record
		} else {
			fmt.Println("Closing nonce map channel")
		}
	}
}
