package mapper

import (
	"log"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

type RouteMapper struct{}

func testRoute() (*v2.RouteConfiguration, error) {
	vhosts := make([]route.VirtualHost, 1)
	vhosts[0] = route.VirtualHost{
		Name:    "local_service",
		Domains: []string{"*"},
		Routes: []route.Route{
			{
				Match: route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: "bifrost",
						},
					},
				},
			},
		},
	}
	route := v2.RouteConfiguration{
		Name:         "local_route",
		VirtualHosts: vhosts,
	}
	return &route, nil
}

func (r *RouteMapper) GetResources(configJson string) ([]types.Any, error) {
	typeUrl := cache.RouteType
	resources := make([]types.Any, 1)

	protoVal, err := testRoute()
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
