package server

import (
	"testing"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

func TestServer_IsValidSubscriber(t *testing.T) {
	validReq := v2.DiscoveryRequest{
		Node: &envoy_api_v2_core.Node{
			Id:      "replica-1",
			Cluster: "rider-service",
		},
	}
	inValidReq1 := v2.DiscoveryRequest{
		Node: &envoy_api_v2_core.Node{
			Cluster: "rider-service",
		},
	}
	inValidReq2 := v2.DiscoveryRequest{
		Node: &envoy_api_v2_core.Node{
			Id: "replica-1",
		},
	}

	if !IsValidSubscriber(&validReq) {
		t.Errorf("%+v \n is valid subscriber", validReq)
	}

	if IsValidSubscriber(&inValidReq1) {
		t.Errorf("%+v \n is NOT valid subscriber", inValidReq1)
	}

	if IsValidSubscriber(&inValidReq2) {
		t.Errorf("%+v \n is NOT valid subscriber", inValidReq2)
	}

}
