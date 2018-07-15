package xdscluster

import (
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"

	envoy_api_v2_core1 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

func GetCluster() *v2.Cluster {
	return &v2.Cluster{
		Name:           "bifrost",
		ConnectTimeout: 250 * time.Microsecond,
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
}

func GetResources(typeUrl string) []types.Any {
	resources := make([]types.Any, 1)
	data, err := proto.Marshal(GetCluster())
	check(err)

	resources[0] = types.Any{
		Value:   data,
		TypeUrl: typeUrl,
	}

	return resources
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
