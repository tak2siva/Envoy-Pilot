package mapper

import (
	"Envoy-Pilot/cmd/server/util"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"

	envoy_api_v2_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_api_v2_cluster "github.com/envoyproxy/go-control-plane/envoy/api/v2/cluster"
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

func buildHosts(rawObj interface{}) ([]*envoy_api_v2_core1.Address, error) {
	if rawObj == nil {
		return make([]*envoy_api_v2_core1.Address, 0), nil
	}

	rawHosts := toArray(rawObj)
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
		return nil
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

func GetConfigSourceType(typeVal string) envoy_api_v2_core1.ApiConfigSource_ApiType {
	id := envoy_api_v2_core1.ApiConfigSource_ApiType_value[typeVal]
	return envoy_api_v2_core1.ApiConfigSource_ApiType(id)
}

func buildEdsClusterConfig(rawObj interface{}) *v2.Cluster_EdsClusterConfig {
	if rawObj == nil {
		return nil
	}

	eds_cluster_config := toMap(rawObj)
	eds_config := toMap(eds_cluster_config["eds_config"])
	api_config_source := toMap(eds_config["api_config_source"])
	grpc_services := toArray(api_config_source["grpc_services"])

	resGrpcServices := make([]*envoy_api_v2_core1.GrpcService, len(grpc_services))

	for i, grpc_service := range grpc_services {
		objMap := toMap(grpc_service)
		envoy_grpc := toMap(objMap["envoy_grpc"])
		resGrpcServices[i] = &envoy_api_v2_core1.GrpcService{
			TargetSpecifier: &envoy_api_v2_core1.GrpcService_EnvoyGrpc_{
				EnvoyGrpc: &envoy_api_v2_core1.GrpcService_EnvoyGrpc{
					ClusterName: getString(envoy_grpc, "cluster_name"),
				},
			},
		}
	}

	return &v2.Cluster_EdsClusterConfig{
		EdsConfig: &envoy_api_v2_core1.ConfigSource{
			ConfigSourceSpecifier: &envoy_api_v2_core1.ConfigSource_ApiConfigSource{
				ApiConfigSource: &envoy_api_v2_core1.ApiConfigSource{
					ApiType:      GetConfigSourceType(getString(api_config_source, "api_type")),
					GrpcServices: resGrpcServices,
				},
			},
		},
	}
}

func buildThresholds(rawObj interface{}) []*envoy_api_v2_cluster.CircuitBreakers_Thresholds {
	if rawObj == nil {
		return nil
	}
	objArr := toArray(rawObj)
	thresholds := make([]*envoy_api_v2_cluster.CircuitBreakers_Thresholds, len(objArr))

	for i, rawThreshold := range objArr {
		thresholdMap := toMap(rawThreshold)
		priorityId := envoy_api_v2_core1.RoutingPriority_value[getString(thresholdMap, "priority")]

		threshold := envoy_api_v2_cluster.CircuitBreakers_Thresholds{
			Priority: envoy_api_v2_core1.RoutingPriority(priorityId),
		}
		if thresholdMap["max_connections"] != nil {
			val, err := getUIntValue(thresholdMap, "max_connections")
			util.CheckAndPanic(err)
			threshold.MaxConnections = &val
		}
		if thresholdMap["max_pending_requests"] != nil {
			val, err := getUIntValue(thresholdMap, "max_pending_requests")
			util.CheckAndPanic(err)
			threshold.MaxPendingRequests = &val
		}
		if thresholdMap["max_requests"] != nil {
			val, err := getUIntValue(thresholdMap, "max_requests")
			util.CheckAndPanic(err)
			threshold.MaxRequests = &val
		}
		if thresholdMap["max_retries"] != nil {
			val, err := getUIntValue(thresholdMap, "max_retries")
			util.CheckAndPanic(err)
			threshold.MaxRetries = &val
		}
		thresholds[i] = &threshold
	}
	return thresholds
}

func buildCircuitBreakers(rawObj interface{}) *envoy_api_v2_cluster.CircuitBreakers {
	if rawObj == nil {
		return nil
	}
	objMap := toMap(rawObj)
	return &envoy_api_v2_cluster.CircuitBreakers{
		Thresholds: buildThresholds(objMap["thresholds"]),
	}
}

func buildClusterTLSContext(rawObj interface{}) *envoy_api_v2_auth.UpstreamTlsContext {
	if rawObj == nil {
		return nil
	}
	tlsCtxMap := toMap(rawObj)
	commonTlsCtxMap := toMap(tlsCtxMap["common_tls_context"])
	upstreamTlsCtx := &envoy_api_v2_auth.UpstreamTlsContext{
		CommonTlsContext: &envoy_api_v2_auth.CommonTlsContext{
			TlsCertificates: buildTlsCerts(commonTlsCtxMap["tls_certificates"]),
			AlpnProtocols:   buildAlpnProtocol(commonTlsCtxMap["alpn_protocols"]),
		},
	}
	if keyExists(tlsCtxMap, "sni") {
		upstreamTlsCtx.Sni = getString(tlsCtxMap, "sni")
	}
	return upstreamTlsCtx
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
	clusterObj.EdsClusterConfig = buildEdsClusterConfig(rawObjMap["eds_cluster_config"])
	clusterObj.CircuitBreakers = buildCircuitBreakers(rawObjMap["circuit_breakers"])
	clusterObj.TlsContext = buildClusterTLSContext(rawObjMap["tls_context"])

	hosts, err := buildHosts(rawObjMap["hosts"])
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
	// err := json.Unmarshal([]byte(clusterJson), &rawArr)
	// if err != nil {
	// 	panic(err)
	// }
	rawArr = util.ImportJsonOrYaml(clusterJson)

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
