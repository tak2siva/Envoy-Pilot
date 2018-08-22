package mapper

import (
	"Envoy-Pilot/cmd/server/util"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"

	envoy_api_v2_core1 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

type ClusterMapper struct{}

func buildDnsType(rawObj map[string]interface{}) v2.Cluster_DiscoveryType {
	typeStr := getString(rawObj, "type")
	typeStr = strings.ToUpper(typeStr)
	typeID := v2.Cluster_DiscoveryType_value[typeStr]
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
	return res
}

func buildHttp2ProtocolOptions(rawObj interface{}) *envoy_api_v2_core1.Http2ProtocolOptions {
	if rawObj == nil {
		return &envoy_api_v2_core1.Http2ProtocolOptions{}
	}
	objMap := toMap(rawObj)
	http2 := envoy_api_v2_core1.Http2ProtocolOptions{}

	if keyExists(objMap, "hpack_table_size") {
		val, err := getUIntValue(objMap, "hpack_table_size")
		util.CheckAndPanic(err)
		http2.HpackTableSize = &val
	}

	if keyExists(objMap, "max_concurrent_streams") {
		val, err := getUIntValue(objMap, "max_concurrent_streams")
		util.CheckAndPanic(err)
		http2.MaxConcurrentStreams = &val
	}

	if keyExists(objMap, "initial_stream_window_size") {
		val, err := getUIntValue(objMap, "initial_stream_window_size")
		util.CheckAndPanic(err)
		http2.InitialStreamWindowSize = &val
	}

	if keyExists(objMap, "initial_connection_window_size") {
		val, err := getUIntValue(objMap, "initial_connection_window_size")
		util.CheckAndPanic(err)
		http2.InitialConnectionWindowSize = &val
	}

	return &http2
}

func (c *ClusterMapper) GetCluster(rawObj interface{}) (retCluster v2.Cluster, retErr error) {
	var rawObjMap map[string]interface{}
	if rawObj != nil {
		rawObjMap = toMap(rawObj)
	} else {
		return v2.Cluster{}, nil
	}

	var clusterObj = v2.Cluster{}

	clusterObj.Name = rawObjMap["name"].(string)
	cxTimeout := BuildDuration(getString(rawObjMap, "connect_timeout"))
	clusterObj.ConnectTimeout = cxTimeout
	clusterObj.Type = buildDnsType(rawObjMap)
	clusterObj.LbPolicy = buildLbPolicy(rawObjMap)
	clusterObj.Http2ProtocolOptions = buildHttp2ProtocolOptions(rawObjMap["http2_protocol_options"])

	hosts, err := buildHosts(rawObjMap)
	if err != nil {
		log.Panic(err)
	}
	clusterObj.Hosts = hosts
	return clusterObj, nil
}

func (c *ClusterMapper) GetClusters(clusterJson string) (retCluster []*v2.Cluster, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("*********************************")
			log.Printf("Recovered %s from %s: %s\n", "GetClusters", r, debug.Stack())
			log.Println("*********************************")
			retErr = errors.New(fmt.Sprintf("%s", r))
		}
	}()

	var rawArr []interface{}
	err := json.Unmarshal([]byte(clusterJson), &rawArr)
	if err != nil {
		panic(err)
	}

	var clusters = make([]*v2.Cluster, len(rawArr))
	for i, rawCluster := range rawArr {
		val, err := c.GetCluster(rawCluster)
		if err != nil {
			panic(err)
		}
		clusters[i] = &val
	}
	return clusters, nil
}

func (c *ClusterMapper) GetResources(configJson string) ([]types.Any, error) {
	typeUrl := cache.ClusterType

	clusters, err := c.GetClusters(configJson)
	if err != nil {
		log.Printf("Error parsing cluster config")
		return nil, err
	}

	resources := make([]types.Any, len(clusters))

	for i, cluster := range clusters {
		data, err := proto.Marshal(cluster)
		if err != nil {
			log.Printf("Error marshalling cluster...\n")
			log.Panic(err)
		}
		resources[i] = types.Any{
			Value:   data,
			TypeUrl: typeUrl,
		}
	}

	return resources, err
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
