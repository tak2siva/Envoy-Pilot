package server

import (
	"Envoy-Pilot/cmd/server/constant"
	"context"

	"log"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func (s *Server) FetchEndpoints(ctx context.Context, in *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	log.Printf("%+v\n", in)
	panic("Not implemented")
	return &v2.DiscoveryResponse{VersionInfo: "2"}, nil
}

// func (s *Server) IncrementalEndpoints(_ v2.ClusterDiscoveryService_IncrementalClustersServer) error {
// 	return errors.New("not implemented")
// }

// StreamClusters bi directional stream to update endpoints config
func (s *Server) StreamEndpoints(stream v2.EndpointDiscoveryService_StreamEndpointsServer) error {
	return s.BiDiStreamFor(constant.SUBSCRIBE_EDS, stream)
}
