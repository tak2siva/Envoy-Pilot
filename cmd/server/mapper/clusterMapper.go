package mapper

import (
	"fmt"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"

	envoy_api_v2_core1 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/jsonpb"
)

type ClusterMapper struct{}

func (c *ClusterMapper) GetCluster() *v2.Cluster {
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

func (c *ClusterMapper) GetResources(typeUrl string) []types.Any {
	resources := make([]types.Any, 1)
	json := `
		{
			"name": "bifrost",
			"connect_timeout": "250",
			"type": 1,
			"lb_policy": 0,
			"hosts": [{
				"socket_address": {
					"address": "127.0.0.1",
					"portValue": 1234
				}
			}]
		}
	`

	obj := v2.Cluster{}
	m := &jsonpb.Marshaler{}
	if err := jsonpb.UnmarshalString(json, &obj); err != nil {
		// if err := types.UnmarshalAny(json, obj); err != nil {
		fmt.Printf("error: %s", err.Error())
	} else {
		fmt.Println("Worked \\o/ %+v", obj)
	}
	fmt.Println("\n*************************")

	protoVal := c.GetCluster()
	strVal, err := m.MarshalToString(protoVal)
	if err != nil {
		fmt.Printf("\n marshal error: %s", err.Error())
	} else {
		fmt.Println("\n Marshal Worked \\o/ %+v", strVal)
	}

	data, err := proto.Marshal(c.GetCluster())
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
