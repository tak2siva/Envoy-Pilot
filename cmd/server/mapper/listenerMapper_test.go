package mapper

import (
	"log"
	"testing"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_api_v2_core1 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	als "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	alf "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	util "github.com/envoyproxy/go-control-plane/pkg/util"
	"github.com/go-test/deep"
	"github.com/gogo/protobuf/jsonpb"
)

var listenerMapper ListenerMapper

func init() {
	listenerMapper = ListenerMapper{}
}
func TestListenerMapper_GetListener(t *testing.T) {
	jsonString := `
	{
		"name": "listener_0",
		"address": {
		   "socket_address": {
			  "address": "127.0.0.1",
			  "port_value": 10000
		   }
		},
		"filter_chains": [
		   {
			  "filters": [
				 {
					"name": "envoy.http_connection_manager",
					"config": {
					   "stat_prefix": "ingress_http",
					   "access_log": [
						  {
							 "name": "envoy.file_access_log",
							 "config": {
								"path": "/dev/stdout",
								"format": "some-format"
							 }
						  }
					   ],
					   "codec_type": "AUTO",
					   "rds": {
						  "route_config_name": "local_route",
						  "config_source": {
							 "api_config_source": {
								"api_type": "GRPC",
								"grpc_services": {
								   "envoy_grpc": {
									  "cluster_name": "xds_cluster"
								   }
								}
							 }
						  }
					   },
					   "http_filters": [
						  {
							 "name": "envoy.router"
						  }
					   ]
					}
				 }
			  ]
		   }
		]
	 }
	`

	rdsSource := core.ConfigSource{}
	rdsSource.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			ApiType: core.ApiConfigSource_GRPC,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}

	alsConfig := &als.FileAccessLog{
		Path:   "/var/access_log.log",
		Format: "some-format",
	}

	alsConfigPbst, err := util.MessageToStruct(alsConfig)
	if err != nil {
		panic(err)
	}

	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    rdsSource,
				RouteConfigName: "local_route",
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: util.Router,
		}},
		AccessLog: []*alf.AccessLog{{
			Name:   util.FileAccessLog,
			Config: alsConfigPbst,
		}},
	}
	pbst, err := util.MessageToStruct(manager)

	expectedListener := v2.Listener{
		Name: "listener_0",
		Address: envoy_api_v2_core1.Address{
			Address: &envoy_api_v2_core1.Address_SocketAddress{
				SocketAddress: &envoy_api_v2_core1.SocketAddress{
					Address:       "127.0.0.1",
					PortSpecifier: &envoy_api_v2_core1.SocketAddress_PortValue{PortValue: 1234},
				},
			},
		},
		FilterChains: []listener.FilterChain{
			{
				Filters: []listener.Filter{
					{
						Name:   util.HTTPConnectionManager,
						Config: pbst,
					},
				},
			},
		},
	}
	// _ = jsonString

	actualObj, _ := listenerMapper.GetListener(jsonString)

	marshaler := &jsonpb.Marshaler{}
	actJson, err := marshaler.MarshalToString(actualObj)
	exJson, err := marshaler.MarshalToString(&expectedListener)

	log.Println("*************")
	log.Println(string(actJson))
	log.Println("*************")
	log.Println(string(exJson))
	log.Println("*************")

	if diff := deep.Equal(actualObj, &expectedListener); diff != nil {
		t.Error(diff)
	}
}
