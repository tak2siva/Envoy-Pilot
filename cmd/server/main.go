package main

import (
	"Envoy-Pilot/cmd/server/dump"
	"Envoy-Pilot/cmd/server/server"
	"Envoy-Pilot/cmd/server/storage"
	"fmt"
	"log"
	"net"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	consul "github.com/hashicorp/consul/api"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var consulHandle *consul.KV

func init() {
	//host.docker.internal:8500
	log.SetFlags(log.LstdFlags | log.Llongfile)
	consulHealthCheck()
}

func consulHealthCheck() {
	cwrapper := storage.GetConsulWrapper()
	cwrapper.GetUniqId()
}

func main() {
	go dump.SetUpHttpServer()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 7777))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	v2.RegisterClusterDiscoveryServiceServer(s, &server.Server{})
	v2.RegisterListenerDiscoveryServiceServer(s, &server.Server{})
	v2.RegisterRouteDiscoveryServiceServer(s, &server.Server{})
	discovery.RegisterAggregatedDiscoveryServiceServer(s, &server.Server{})

	reflection.Register(s)

	log.Print("Started grpc server..")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}
