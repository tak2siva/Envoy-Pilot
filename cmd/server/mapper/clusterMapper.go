package mapper

import (
	"Envoy-xDS/cmd/server/util"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"

	envoy_api_v2_core1 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
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
		res, err := buildHost(rowMap)
		if err != nil {
			return nil, err
		}
		list[i] = &res
	}
	return list, nil
}

func buildHost(rowMap map[string]interface{}) (envoy_api_v2_core1.Address, error) {
	addrMap := rowMap["socket_address"].(map[string]interface{})

	port, err := getUInt(addrMap, "port_value")
	if err != nil {
		return envoy_api_v2_core1.Address{}, err
	}

	res := envoy_api_v2_core1.Address{
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
	return res, nil
}

func BuildDuration(str string) time.Duration {
	res, err := time.ParseDuration(str)
	if err != nil {
		log.Printf("Error parsing string to duration %s\n", str)
		log.Println(err)
		panic("Error parsing duration")
	}
	// log.Printf("time json: %s\n", str)
	// log.Printf("time parsed: %d\n", res)
	return res
}

func (c *ClusterMapper) GetCluster(clusterJson string) (retCluster *v2.Cluster, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("*********************************")
			log.Printf("Recovered %s from %s: %s\n", "GetClusters", r, debug.Stack())
			log.Println("*********************************")
			retErr = errors.New(fmt.Sprintf("%s", r))
		}
	}()

	var rawObj = make(map[string]interface{})
	var clusterObj = &v2.Cluster{}

	err := json.Unmarshal([]byte(clusterJson), &rawObj)
	util.Check(err)

	clusterObj.Name = rawObj["name"].(string)
	cxTimeout := BuildDuration(getString(rawObj, "connect_timeout"))
	clusterObj.ConnectTimeout = cxTimeout
	clusterObj.Type = buildDnsType(rawObj)
	clusterObj.LbPolicy = buildLbPolicy(rawObj)

	hosts, err := buildHosts(rawObj)
	if err != nil {
		return nil, err
	}
	clusterObj.Hosts = hosts
	return clusterObj, nil
}

func (c *ClusterMapper) GetResources(configJson string) ([]types.Any, error) {
	typeUrl := cache.ClusterType
	resources := make([]types.Any, 1)

	protoVal, err := c.GetCluster(configJson)
	if err != nil {
		log.Printf("Error parsing cluster config")
		return nil, err
	}
	data, err := proto.Marshal(protoVal)
	if err != nil {
		log.Printf("Error building cluster resource...\n")
		return nil, err
	}

	resources[0] = types.Any{
		Value:   data,
		TypeUrl: typeUrl,
	}

	return resources, err
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
