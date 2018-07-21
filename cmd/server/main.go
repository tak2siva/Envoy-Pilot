package main

import (
	"Envoy-xDS/cmd/server/server"
	"fmt"
	"log"
	"net"

	consul "github.com/hashicorp/consul/api"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var consulHandle *consul.KV

func init() {
	consulClient, err := consul.NewClient(&consul.Config{Address: "host.docker.internal:8500"})
	if err != nil {
		panic(err)
	}
	consulHandle = consulClient.KV()
	log.SetFlags(log.Llongfile)
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 7777))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	v2.RegisterClusterDiscoveryServiceServer(s, &server.Server{})
	reflection.Register(s)

	log.Print("Started grpc server..")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}
