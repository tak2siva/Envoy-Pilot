package main

import (
	"Envoy-xDS/cmd/server/manager"
	xdscluster "Envoy-xDS/cmd/server/xdscluster"
	"Envoy-xDS/lib"
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	consul "github.com/hashicorp/consul/api"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"github.com/google/uuid"
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
	fmt.Printf("-------------- Starting a stream ------------------\n")

	for {
		req, err := stream.Recv()
		// util.Check(err)

		if err != nil {
			fmt.Println("Disconnecting client")
			fmt.Println(err)
			return err
		}

		if manager.IsACK(req) || !manager.IsOutDated(req) {
			fmt.Println("No updates ignoring request....")
			continue
		}

		responseUUID := uuid.New().String()
		responseVersion := "1"

		response := &v2.DiscoveryResponse{
			VersionInfo: responseVersion,
			Resources:   xdscluster.GetResources(req.TypeUrl),
			TypeUrl:     req.TypeUrl,
			Nonce:       responseUUID,
		}
		fmt.Printf("%+v\n", req)
		fmt.Printf("%+v\n", response)

		err = stream.Send(response)
		if err != nil {
			fmt.Println("error sending to client")
			fmt.Println(err)
			return err
		}
		manager.UpdateMap(response)
		fmt.Println("Out req channel..")
	}
}

func (s *server) IncrementalClusters(_ v2.ClusterDiscoveryService_IncrementalClustersServer) error {
	return errors.New("not implemented")
}

var consulHandle *consul.KV

func init() {
	consulClient, err := consul.NewClient(&consul.Config{Address: "host.docker.internal:8500"})
	if err != nil {
		panic(err)
	}
	consulHandle = consulClient.KV()
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
