package server

import (
	"Envoy-xDS/cmd/server/constant"
	"context"

	"errors"
	"log"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func (s *Server) FetchClusters(ctx context.Context, in *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	log.Printf("%+v\n", in)
	panic("Not implemented")
	return &v2.DiscoveryResponse{VersionInfo: "2"}, nil
}

func (s *Server) IncrementalClusters(_ v2.ClusterDiscoveryService_IncrementalClustersServer) error {
	return errors.New("not implemented")
}

// StreamClusters bi directional stream to update cluster config
func (s *Server) StreamClusters(stream v2.ClusterDiscoveryService_StreamClustersServer) error {
	return s.BiDiStreamFor(constant.SUBSCRIBE_CDS, stream)
}
