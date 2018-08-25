package mapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"

	pilot_util "Envoy-Pilot/cmd/server/util"

	envoy_api_v2_core1 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_api_v2_endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
)

type EndpointMapper struct{}

func (e *EndpointMapper) GetSocketAddress(rawObj interface{}) *envoy_api_v2_core1.Address {
	if rawObj == nil {
		return &envoy_api_v2_core1.Address{}
	}
	addressMap := toMap(rawObj)

	if addressMap == nil {
		return &envoy_api_v2_core1.Address{}
	}

	socketAddressMap := toMap(addressMap["socket_address"])

	if addressMap == nil {
		return &envoy_api_v2_core1.Address{}
	}

	port, err := getUInt(socketAddressMap, "port_value")
	pilot_util.CheckAndPanic(err)

	return &envoy_api_v2_core1.Address{&envoy_api_v2_core1.Address_SocketAddress{
		SocketAddress: &envoy_api_v2_core1.SocketAddress{
			Address: getString(socketAddressMap, "address"),
			PortSpecifier: &envoy_api_v2_core1.SocketAddress_PortValue{
				PortValue: port,
			},
		},
	}}
}

func (e *EndpointMapper) GetEndpoint(rawObj interface{}) *envoy_api_v2_endpoint.Endpoint {
	if rawObj == nil {
		return &envoy_api_v2_endpoint.Endpoint{}
	}
	endpointMap := toMap(rawObj)
	rawAddress := toMap(endpointMap["address"])
	return &envoy_api_v2_endpoint.Endpoint{
		Address: e.GetSocketAddress(rawAddress),
	}
}

func (e *EndpointMapper) GetLbEndpoint(rawObj interface{}) envoy_api_v2_endpoint.LbEndpoint {
	if rawObj == nil {
		return envoy_api_v2_endpoint.LbEndpoint{}
	}
	lbEndpointMap := toMap(rawObj)

	return envoy_api_v2_endpoint.LbEndpoint{
		Endpoint: e.GetEndpoint(lbEndpointMap["endpoint"]),
	}
}

func (e *EndpointMapper) GetLbEndpoints(rawObj interface{}) []envoy_api_v2_endpoint.LbEndpoint {
	if rawObj == nil {
		return make([]envoy_api_v2_endpoint.LbEndpoint, 0)
	}
	rawLbEndpoints := toArray(rawObj)
	lbEndpoints := make([]envoy_api_v2_endpoint.LbEndpoint, len(rawLbEndpoints))

	for i, rawLbEndpoint := range rawLbEndpoints {
		lbEndpoints[i] = e.GetLbEndpoint(rawLbEndpoint)
	}

	return lbEndpoints
}

func (e *EndpointMapper) GetLocalityLbEndpoints(rawObj interface{}) []envoy_api_v2_endpoint.LocalityLbEndpoints {
	if rawObj == nil {
		return make([]envoy_api_v2_endpoint.LocalityLbEndpoints, 0)
	}
	rawLocalityLbEndpoints := toArray(rawObj)
	localityLbEndpoints := make([]envoy_api_v2_endpoint.LocalityLbEndpoints, len(rawLocalityLbEndpoints))

	for i, rawlocalityLbEndpoint := range rawLocalityLbEndpoints {
		localityLbEndpointMap := toMap(rawlocalityLbEndpoint)
		localityLbEndpoints[i] = envoy_api_v2_endpoint.LocalityLbEndpoints{
			LbEndpoints: e.GetLbEndpoints(getString(localityLbEndpointMap, "lb_endpoints")),
		}
	}
	return localityLbEndpoints
}

func (e *EndpointMapper) GetClusterLoadAssignment(rawObj interface{}) (retEndpoints *v2.ClusterLoadAssignment, retErr error) {
	if rawObj == nil {
		return &v2.ClusterLoadAssignment{}, nil
	}
	objMap := toMap(rawObj)
	cla := &v2.ClusterLoadAssignment{
		ClusterName: getString(objMap, "cluster_name"),
	}
	return cla, nil
}

func (e *EndpointMapper) GetClusterLoadAssignments(endpointsJson string) (retEndpoints []*v2.ClusterLoadAssignment, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("*********************************")
			log.Printf("Recovered %s from %s: %s\n", "GetEndpoints", r, debug.Stack())
			log.Println("*********************************")
			retErr = errors.New(fmt.Sprintf("%s", r))
		}
	}()
	var rawArr []interface{}
	err := json.Unmarshal([]byte(endpointsJson), &rawArr)
	if err != nil {
		panic(err)
	}

	var endpoints = make([]*v2.ClusterLoadAssignment, len(rawArr))
	for i, rawEndpoint := range rawArr {
		val, err := e.GetClusterLoadAssignment(rawEndpoint)
		if err != nil {
			panic(err)
		}
		endpoints[i] = val
	}
	return endpoints, nil
}

func (e *EndpointMapper) GetResources(configJson string) ([]types.Any, error) {
	typeUrl := cache.EndpointType

	endpoints, err := e.GetClusterLoadAssignments(configJson)
	if err != nil {
		log.Printf("Error parsing endpoint config")
		return nil, err
	}
	resources := make([]types.Any, len(endpoints))

	for i, endpoint := range endpoints {
		data, err := proto.Marshal(endpoint)
		if err != nil {
			log.Printf("Error marshalling endpoint config")
			log.Panic(err)
		}

		resources[i] = types.Any{
			Value:   data,
			TypeUrl: typeUrl,
		}
	}

	return resources, nil
}
