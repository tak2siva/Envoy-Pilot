package server

import (
	"Envoy-xDS/cmd/server/constant"
	"context"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func (s *Server) FetchListeners(context.Context, *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	panic("Not implemented")
}

func (s *Server) StreamListeners(stream v2.ListenerDiscoveryService_StreamListenersServer) error {
	return s.BiDiStreamFor(constant.SUBSCRIBE_LDS, stream)
}
