package main

import (
	"Envoy-xDS/lib"
	"Envoy-xDS/lib/envoy/api/v2"
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct{}

func (s *server) SayHello(ctx context.Context, in *api.PingMessage) (*api.PingMessage, error) {
	log.Printf("Serving request: %s", in.Greeting)
	fmt.Printf("%+v\n", in)
	return &api.PingMessage{Greeting: "Hello from server"}, nil
}

func (s *server) FetchClusters(ctx context.Context, in *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	fmt.Printf("%+v\n", in)
	return &v2.DiscoveryResponse{VersionInfo: "2"}, nil
}

func (s *server) StreamClusters(stream v2.ClusterDiscoveryService_StreamClustersServer) error {
	return nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 7777))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	api.RegisterPingServer(s, &server{})
	v2.RegisterClusterDiscoveryServiceServer(s, &server{})
	reflection.Register(s)

	log.Print("Started grpc server..")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}
