package mapper

import (
	"testing"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_core1 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/go-test/deep"
)

var clusterMapper ClusterMapper

func init() {
	clusterMapper = ClusterMapper{}
}
func TestClusterMapper_GetCluster(t *testing.T) {
	json := `
	{
		"name": "bifrost",
		"connect_timeout": "5s",
		"type": "STRICT_DNS",
		"lb_policy": "ROUND_ROBIN",
		"hosts": [{
			"socket_address": {
				"address": "127.0.0.1",
				"port_value": 1234
			}
		}]
	}
	`

	expectedObj := v2.Cluster{
		Name:           "bifrost",
		ConnectTimeout: 5 * time.Second,
		Type:           v2.Cluster_STRICT_DNS,
		LbPolicy:       v2.Cluster_ROUND_ROBIN,
		Hosts: []*envoy_api_v2_core1.Address{
			{&envoy_api_v2_core1.Address_SocketAddress{
				SocketAddress: &envoy_api_v2_core1.SocketAddress{
					Address:       "127.0.0.1",
					PortSpecifier: &envoy_api_v2_core1.SocketAddress_PortValue{PortValue: 1234},
				},
			}},
		},
	}
	actualObj, _ := clusterMapper.GetCluster(json)

	if diff := deep.Equal(actualObj, &expectedObj); diff != nil {
		t.Error(diff)
	}
}

func TestClusterMapper_BuildDuration(t *testing.T) {
	res := BuildDuration("5s")
	expected := 5 * time.Second
	if res != expected {
		t.Errorf("incorrect time parse: %d, want: %d.", res, expected)
	}
}
