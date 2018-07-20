package server

import (
	"Envoy-xDS/cmd/server/manager"
	"Envoy-xDS/cmd/server/mapper"
	"Envoy-xDS/cmd/server/model"
	"Envoy-xDS/cmd/server/storage"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/google/uuid"
)

const envoySubscriberKey = "envoySubscriber"

func (s *Server) StreamClusters(stream v2.ClusterDiscoveryService_StreamClustersServer) error {
	fmt.Printf("-------------- Starting a stream ------------------\n")

	serverCtx, cancel := context.WithCancel(context.Background())
	serverCtx = context.WithValue(serverCtx, envoySubscriberKey, &model.EnvoySubscriber{})

	dispatchChannel := make(chan bool)
	i := 0

	for {
		req, err := stream.Recv()
		// util.Check(err)

		if err != nil {
			fmt.Println("Disconnecting client")
			fmt.Println(err)
			cancel()
			return err
		}

		if manager.IsACK(req) || !manager.IsOutDated(req.VersionInfo) {
			fmt.Println("No updates ignoring request....")
			continue
		}

		if i == 0 {
			go consulPoll(serverCtx, dispatchChannel)
			go dispatchCluster(stream, dispatchChannel, serverCtx)
			dao := storage.GetXdsConfigDao()
			dao.RegisterSubscriber(req.Node.Cluster, req.Node.Id)
			// dao.Register
			i++
		}

		fmt.Println("Out req channel..")
	}
}

func dispatchCluster(stream v2.ClusterDiscoveryService_StreamClustersServer,
	dispatchChannel chan bool,
	ctx context.Context) {
	for range dispatchChannel {
		select {
		case <-ctx.Done():
			return
		default:
		}
		responseUUID := uuid.New().String()
		responseVersion := "1"

		m := mapper.ClusterMapper{}

		response := &v2.DiscoveryResponse{
			VersionInfo: responseVersion,
			Resources:   m.GetResources(cache.ClusterType),
			TypeUrl:     cache.ClusterType,
			Nonce:       responseUUID,
		}
		fmt.Printf("%+v\n", response)

		err := stream.Send(response)
		if err != nil {
			fmt.Println("error sending to client")
			fmt.Println(err)
		}
		manager.UpdateMap(response)
	}
}

func (s *Server) FetchClusters(ctx context.Context, in *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	fmt.Printf("%+v\n", in)
	return &v2.DiscoveryResponse{VersionInfo: "2"}, nil
}

func (s *Server) IncrementalClusters(_ v2.ClusterDiscoveryService_IncrementalClustersServer) error {
	return errors.New("not implemented")
}

func consulPoll(ctx context.Context, dispatchChannel chan bool) {
	for {
		time.Sleep(10 * time.Second)
		select {
		case <-ctx.Done():
			return
		default:
		}
		fmt.Println("Checking consul..")
		dao := storage.GetXdsConfigDao()
		latestVersion := dao.GetLatestVersion()
		lastUpdatedVersion := ctx.Value(envoySubscriberKey).(string)
		if latestVersion != lastUpdatedVersion {
			dispatchChannel <- true
		}
	}
}
