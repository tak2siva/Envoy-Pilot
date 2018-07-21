package server

import (
	"Envoy-xDS/cmd/server/manager"
	"Envoy-xDS/cmd/server/mapper"
	"Envoy-xDS/cmd/server/model"
	"Envoy-xDS/cmd/server/service"
	"Envoy-xDS/cmd/server/storage"
	"context"
	"errors"
	"log"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/google/uuid"
)

const envoySubscriberKey = "envoySubscriber"

var clusterService *service.ClusterService
var xdsConfigDao *storage.XdsConfigDao

func init() {
	clusterService = service.GetClusterService()
	xdsConfigDao = storage.GetXdsConfigDao()
}

func (s *Server) StreamClusters(stream v2.ClusterDiscoveryService_StreamClustersServer) error {
	log.Printf("-------------- Starting a stream ------------------\n")

	serverCtx, cancel := context.WithCancel(context.Background())
	dispatchChannel := make(chan bool)
	i := 0
	var subscriber *model.EnvoySubscriber

	for {
		req, err := stream.Recv()
		if i == 0 {
			subscriber = &model.EnvoySubscriber{Cluster: req.Node.Cluster, Node: req.Node.Id}
			serverCtx = context.WithValue(serverCtx, envoySubscriberKey, subscriber)
			dao := storage.GetXdsConfigDao()
			dao.RegisterSubscriber(subscriber)

			go consulPoll(serverCtx, dispatchChannel)
			go dispatchCluster(stream, dispatchChannel, serverCtx)
			i++
		}

		if err != nil {
			log.Printf("Disconnecting client %s\n", subscriber.BuildInstanceKey())
			log.Println(err)
			cancel()
			return err
		}

		if xdsConfigDao.IsACKPresent(subscriber, req.ResponseNonce) {
			log.Printf("Received ACK %s from %s", req.ResponseNonce, subscriber.BuildInstanceKey())
			continue
		}

		log.Println("Out req channel..")
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
		log.Printf("%+v\n", response)

		err := stream.Send(response)
		if err != nil {
			log.Println("error sending to client")
			log.Println(err)
		}
		manager.UpdateMap(response)
	}
}

func (s *Server) FetchClusters(ctx context.Context, in *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	log.Printf("%+v\n", in)
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
		subscriber := ctx.Value(envoySubscriberKey).(*model.EnvoySubscriber)
		log.Printf("Checking consul for %d..\n", subscriber.Id)
		if !xdsConfigDao.IsRepoPresent(subscriber) {
			continue
		}
		if clusterService.IsOutdated(subscriber) {
			log.Println("Found update dispatching for %s", subscriber.BuildInstanceKey())
			dispatchChannel <- true
		}
	}
}
