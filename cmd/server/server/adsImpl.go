package server

import (
	"Envoy-xDS/cmd/server/constant"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
)

//IncrementalAggregatedResources - Not implemented
func (s *Server) IncrementalAggregatedResources(_ discovery.AggregatedDiscoveryService_IncrementalAggregatedResourcesServer) error {
	panic("Not implemented")
}

// StreamAggregatedResources - ADS server impl
func (s *Server) StreamAggregatedResources(stream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error {
	return s.BiDiStreamFor(constant.SUBSCRIBE_ADS, stream)
}
