package mapper

import (
	"encoding/json"
	"log"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"

	envoy_api_v2_core1 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/jsonpb"
)

type ClusterMapper struct{}

func buildDnsType(rawObj map[string]interface{}) v2.Cluster_DiscoveryType {
	typeID := v2.Cluster_DiscoveryType_value[getString(rawObj, "type")]
	return v2.Cluster_DiscoveryType(typeID)
}

func buildLbPolicy(rawObj map[string]interface{}) v2.Cluster_LbPolicy {
	lbID := v2.Cluster_LbPolicy_value[getString(rawObj, "lb_policy")]
	return v2.Cluster_LbPolicy(lbID)
}

func buildHosts(rawObj map[string]interface{}) ([]*envoy_api_v2_core1.Address, error) {
	rawHosts := rawObj["hosts"].([]interface{})
	list := make([]*envoy_api_v2_core1.Address, len(rawHosts))

	for i, row := range rawHosts {
		rowMap := row.(map[string]interface{})
		addrMap := rowMap["socket_address"].(map[string]interface{})
		log.Printf("**** %+v\n", addrMap)

		port, err := getUInt(addrMap, "port_value")
		if err != nil {
			return nil, err
		}

		list[i] = &envoy_api_v2_core1.Address{
			Address: &envoy_api_v2_core1.Address_SocketAddress{
				SocketAddress: &envoy_api_v2_core1.SocketAddress{
					Protocol: envoy_api_v2_core1.TCP,
					Address:  getString(addrMap, "address"),
					PortSpecifier: &envoy_api_v2_core1.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		}
	}
	return list, nil
}

func (c *ClusterMapper) GetCluster(clusterJson string) (*v2.Cluster, error) {
	var rawObj = make(map[string]interface{})
	var clusterObj = &v2.Cluster{}

	json.Unmarshal([]byte(clusterJson), &rawObj)

	clusterObj.Name = rawObj["name"].(string)
	// TODO add duration unmarshal
	cxTimeout, err := getInt(rawObj, "connect_timeout")
	if err != nil {
		return nil, err
	}
	clusterObj.ConnectTimeout = time.Duration(cxTimeout) * time.Millisecond
	clusterObj.Type = buildDnsType(rawObj)
	clusterObj.LbPolicy = buildLbPolicy(rawObj)

	hosts, err := buildHosts(rawObj)
	if err != nil {
		return nil, err
	}
	clusterObj.Hosts = hosts
	return clusterObj, nil
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
		log.Printf("error: %s", err.Error())
	}

	protoVal, _ := c.GetCluster("")
	_, err := m.MarshalToString(protoVal)
	if err != nil {
		log.Printf("\n marshal error: %s", err.Error())
	}

	data, err := proto.Marshal(protoVal)
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
