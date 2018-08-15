package mapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/util"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_api_v2_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	als "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	alf "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	hc "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/health_check/v2"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

type ListenerMapper struct{}

func buildRds(rawConfigObj interface{}) hcm.HttpConnectionManager_Rds {
	var rawConfig map[string]interface{}
	if rawConfig != nil {
		rawConfig = rawConfigObj.(map[string]interface{})
	} else {
		return hcm.HttpConnectionManager_Rds{}
	}
	rdsSource := core.ConfigSource{}
	configSource := rawConfig["config_source"].(map[string]interface{})
	sourceMap := configSource["api_config_source"].(map[string]interface{})
	// grpcService := sourceMap["grpc_services"].([]interface{})
	grpcService := sourceMap["grpc_services"].(map[string]interface{})
	envoyGrpc := grpcService["envoy_grpc"].(map[string]interface{})

	rdsSource.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			ApiType: core.ApiConfigSource_GRPC,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: getString(envoyGrpc, "cluster_name")},
				},
			}},
		},
	}

	return hcm.HttpConnectionManager_Rds{
		Rds: &hcm.Rds{
			ConfigSource:    rdsSource,
			RouteConfigName: getString(rawConfig, "route_config_name"),
		},
	}
}

func buildRoutes(rawObj interface{}) []envoy_api_v2_route.Route {
	if rawObj == nil {
		return make([]envoy_api_v2_route.Route, 0)
	}
	routes := rawObj.([]interface{})
	res := make([]envoy_api_v2_route.Route, len(routes))

	for i, route := range routes {
		routeMap := toMap(route)
		matchMap := toMap(routeMap["match"])
		routeRouteMap := toMap(routeMap["route"])
		res[i] = envoy_api_v2_route.Route{
			Match: envoy_api_v2_route.RouteMatch{
				PathSpecifier: &envoy_api_v2_route.RouteMatch_Prefix{
					Prefix: getString(matchMap, "prefix"),
				},
			},
			Action: &envoy_api_v2_route.Route_Route{
				Route: &envoy_api_v2_route.RouteAction{
					ClusterSpecifier: &envoy_api_v2_route.RouteAction_Cluster{
						Cluster: getString(routeRouteMap, "cluster"),
					},
				},
			},
		}
	}
	return res
}

func buildVHosts(rawVHosts interface{}) []envoy_api_v2_route.VirtualHost {
	if rawVHosts == nil {
		return make([]envoy_api_v2_route.VirtualHost, 0)
	}
	vhosts := rawVHosts.([]interface{})
	res := make([]envoy_api_v2_route.VirtualHost, len(vhosts))

	for i, vhost := range vhosts {
		vhostMap := toMap(vhost)
		res[i] = envoy_api_v2_route.VirtualHost{
			Name:    getString(vhostMap, "name"),
			Domains: getStringArray(vhostMap, "domains"),
			Routes:  buildRoutes(vhostMap["routes"]),
		}
	}

	return res
}

func buildRouteConfig(rawObj interface{}) hcm.HttpConnectionManager_RouteConfig {
	var routeConfigMap map[string]interface{}
	if rawObj != nil {
		routeConfigMap = rawObj.(map[string]interface{})
	} else {
		return hcm.HttpConnectionManager_RouteConfig{}
	}

	rConfig := v2.RouteConfiguration{
		Name:         getString(routeConfigMap, "name"),
		VirtualHosts: buildVHosts(routeConfigMap["virtual_hosts"]),
	}

	return hcm.HttpConnectionManager_RouteConfig{
		RouteConfig: &rConfig,
	}
}

func buildAccessLog(rawAlsArrayObj interface{}) []*alf.AccessLog {
	var rawAlsArray []interface{}
	if rawAlsArrayObj != nil {
		rawAlsArray = rawAlsArrayObj.([]interface{})
	} else {
		return make([]*alf.AccessLog, 0)
	}
	res := make([]*alf.AccessLog, len(rawAlsArray))
	for i, rawAls := range rawAlsArray {
		alsMap := rawAls.(map[string]interface{})
		configMap := alsMap["config"].(map[string]interface{})
		alsConfig := &als.FileAccessLog{
			Path:   getString(configMap, "path"),
			Format: getString(configMap, "format"),
		}

		alsConfigPbst, err := util.MessageToStruct(alsConfig)
		if err != nil {
			panic(err)
		}
		als := &alf.AccessLog{
			Name:   util.FileAccessLog,
			Config: alsConfigPbst,
		}
		res[i] = als
	}
	return res
}

type httpFilterConfig struct {
	PassThroughMode bool
	Endpoint        string
}

func (m *httpFilterConfig) Reset()         { *m = httpFilterConfig{} }
func (m *httpFilterConfig) String() string { return proto.CompactTextString(m) }
func (*httpFilterConfig) ProtoMessage()    {}

func (m *httpFilterConfig) GetPassThroughMode() bool {
	if m != nil {
		return m.PassThroughMode
	}
	return false
}

func (m *httpFilterConfig) GetEndpoint() string {
	if m != nil {
		return m.Endpoint
	}
	return ""
}

func buildHttpFilter(rawConfig interface{}) []*hcm.HttpFilter {
	if rawConfig == nil {
		return make([]*hcm.HttpFilter, 0)
	}
	filters := rawConfig.([]interface{})
	res := make([]*hcm.HttpFilter, len(filters))

	for i, filter := range filters {
		filterMap := toMap(filter)

		fc := &hcm.HttpFilter{
			Name: getString(filterMap, "name"),
		}
		if filterMap["config"] != nil {
			configMap := toMap(filterMap["config"])
			hfconfig2 := hc.HealthCheck{
				PassThroughMode: &types.BoolValue{Value: getBoolean(configMap, "pass_through_mode")},
				Endpoint:        getString(configMap, "endpoint"),
			}
			pbConfig2, err := util.MessageToStruct(&hfconfig2)
			if err != nil {
				log.Panic(err)
			}
			fc.Config = pbConfig2
		}

		res[i] = fc
	}

	return res
}

func buildHttpConnectionManager(rawConfig map[string]interface{}) hcm.HttpConnectionManager {
	als := buildAccessLog(rawConfig["access_log"])
	codec := hcm.HttpConnectionManager_CodecType_value[getString(rawConfig, "codec_type")]
	manager := hcm.HttpConnectionManager{
		CodecType:   hcm.HttpConnectionManager_CodecType(codec),
		StatPrefix:  getString(rawConfig, "stat_prefix"),
		HttpFilters: buildHttpFilter(rawConfig["http_filters"]),
		AccessLog:   als,
	}

	if rawConfig["rds"] != nil {
		rds := buildRds(rawConfig["rds"])
		manager.RouteSpecifier = &rds
	} else if rawConfig["route_config"] != nil {
		// routeSpec = buildRds(rawConfig["rds"])
		routeConfig := buildRouteConfig(rawConfig["route_config"])
		manager.RouteSpecifier = &routeConfig
	} else {
		panic("Rds or Routeconfig should be present")
	}

	return manager
}

func buildFilterChains(rawFilterChains []interface{}) ([]listener.FilterChain, error) {
	pbFilterChains := make([]listener.FilterChain, len(rawFilterChains))
	for i, rawFilterChain := range rawFilterChains {
		filterChainMap := rawFilterChain.(map[string]interface{})
		rawFilters := filterChainMap["filters"].([]interface{})
		pbFilters := make([]listener.Filter, len(rawFilters))

		for j, rawFilter := range rawFilters {
			filterMap := rawFilter.(map[string]interface{})
			pbFilter := listener.Filter{}
			pbFilter.Name = getString(filterMap, "name")

			manager := buildHttpConnectionManager(filterMap["config"].(map[string]interface{}))
			pbst, err := util.MessageToStruct(&manager)
			if err != nil {
				panic(err)
			}

			pbFilter.Config = pbst
			pbFilters[j] = pbFilter
		}
		pbFilterChains[i] = listener.FilterChain{
			Filters: pbFilters,
		}
	}
	return pbFilterChains, nil
}

func (c *ListenerMapper) GetListener(clusterJson string) (retListener *v2.Listener, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("*********************************")
			log.Printf("Recovered %s from %s: %s\n", "GetListeners", r, debug.Stack())
			log.Println("*********************************")
			retErr = errors.New(fmt.Sprintf("%s", r))
		}
	}()
	var rawObj = make(map[string]interface{})
	var listenerObj = &v2.Listener{}
	err := json.Unmarshal([]byte(clusterJson), &rawObj)

	if err != nil {
		panic(err)
	}

	log.Println("*************")
	for k := range rawObj {
		log.Println(k)
	}
	log.Println("*************")

	listenerObj.Name = rawObj["name"].(string)
	addr, err := buildHost(rawObj["address"].(map[string]interface{}))
	if err != nil {
		return nil, err
	}
	listenerObj.Address = addr
	filterChains, _ := buildFilterChains(rawObj["filter_chains"].([]interface{}))
	listenerObj.FilterChains = filterChains
	return listenerObj, nil
}

func (l *ListenerMapper) GetResources(configJson string) ([]types.Any, error) {
	typeUrl := cache.ListenerType
	resources := make([]types.Any, 1)

	protoVal, err := l.GetListener(configJson)
	if err != nil {
		log.Printf("Error parsing listener config")
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

	return resources, nil
}
