package server

import (
	"Envoy-xDS/cmd/server/manager"
	"Envoy-xDS/cmd/server/xdscluster"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"github.com/google/uuid"
)

func (s *Server) StreamClusters(stream v2.ClusterDiscoveryService_StreamClustersServer) error {
	fmt.Printf("-------------- Starting a stream ------------------\n")
	serverCtx, cancel := context.WithCancel(context.Background())
	i := false
	for {
		req, err := stream.Recv()
		// util.Check(err)

		if err != nil {
			fmt.Println("Disconnecting client")
			fmt.Println(err)
			cancel()
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
			cancel()
			return err
		}
		manager.UpdateMap(response)

		if i == false {
			go consulPoll(serverCtx)
			i = true
		}
		fmt.Println("Out req channel..")
	}
}

func (s *Server) FetchClusters(ctx context.Context, in *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	fmt.Printf("%+v\n", in)
	return &v2.DiscoveryResponse{VersionInfo: "2"}, nil
}

func (s *Server) IncrementalClusters(_ v2.ClusterDiscoveryService_IncrementalClustersServer) error {
	return errors.New("not implemented")
}

func consulPoll(ctx context.Context) {
	for {
		time.Sleep(10 * time.Second)
		select {
		case <-ctx.Done():
			return
		default:
		}
		fmt.Println("Checking consul..")
	}
}
