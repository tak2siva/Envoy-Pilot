package server

import (
	"Envoy-xDS/cmd/server/constant"
	"context"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func (s *Server) FetchRoutes(context.Context, *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	panic("Not implemented")
}

func (s *Server) IncrementalRoutes(v2.RouteDiscoveryService_IncrementalRoutesServer) error {
	panic("Not implemented")
}

func (s *Server) StreamRoutes(stream v2.RouteDiscoveryService_StreamRoutesServer) error {
	return s.BiDiStreamFor(constant.SUBSCRIBE_RDS, stream)
}
