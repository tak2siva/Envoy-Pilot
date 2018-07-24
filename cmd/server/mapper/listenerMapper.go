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
	als "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	alf "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

type ListenerMapper struct{}

func buildRds(rawConfig map[string]interface{}) hcm.HttpConnectionManager_Rds {
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

func buildAccessLog(rawAlsArray []interface{}) []*alf.AccessLog {
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

func buildHttpConnectionManager(rawConfig map[string]interface{}) hcm.HttpConnectionManager {
	rds := buildRds(rawConfig["rds"].(map[string]interface{}))
	als := buildAccessLog(rawConfig["access_log"].([]interface{}))
	manager := hcm.HttpConnectionManager{
		CodecType:      hcm.AUTO,
		StatPrefix:     getString(rawConfig, "stat_prefix"),
		RouteSpecifier: &rds,
		HttpFilters: []*hcm.HttpFilter{{
			Name: util.Router,
		}},
		AccessLog: als,
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
