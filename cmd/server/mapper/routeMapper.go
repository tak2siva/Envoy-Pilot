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
)

type RouteMapper struct{}

func (r *RouteMapper) GetRoute(rawObj interface{}) (retRoutes *v2.RouteConfiguration, retErr error) {
	if rawObj == nil {
		return &v2.RouteConfiguration{}, nil
	}
	connManager := BuildRouteConfig(rawObj)
	return connManager.RouteConfig, nil
}

func (r *RouteMapper) GetRoutes(routesJson string) (retRoutes []*v2.RouteConfiguration, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("*********************************")
			log.Printf("Recovered %s from %s: %s\n", "GetRoutes", r, debug.Stack())
			log.Println("*********************************")
			retErr = errors.New(fmt.Sprintf("%s", r))
		}
	}()
	var rawArr []interface{}
	err := json.Unmarshal([]byte(routesJson), &rawArr)
	if err != nil {
		panic(err)
	}

	var routes = make([]*v2.RouteConfiguration, len(rawArr))
	for i, rawRoute := range rawArr {
		val, err := r.GetRoute(rawRoute)
		if err != nil {
			panic(err)
		}
		routes[i] = val
	}
	return routes, nil
}

func (r *RouteMapper) GetResources(configJson string) ([]types.Any, error) {
	typeUrl := cache.RouteType

	routes, err := r.GetRoutes(configJson)
	if err != nil {
		log.Printf("Error parsing route config")
		return nil, err
	}
	resources := make([]types.Any, len(routes))

	for i, route := range routes {
		data, err := proto.Marshal(route)
		if err != nil {
			log.Printf("Error marshalling route config")
			log.Panic(err)
		}

		resources[i] = types.Any{
			Value:   data,
			TypeUrl: typeUrl,
		}
	}

	return resources, nil
}
