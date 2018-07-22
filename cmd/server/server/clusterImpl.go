package server

import (
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

// StreamClusters bi directional stream to update cluster config
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
			go dispatchCluster(serverCtx, stream, dispatchChannel)
			i++
		}

		log.Printf("Received Request from %s\n %+v\n", subscriber.BuildInstanceKey(), req)

		if err != nil {
			log.Printf("Disconnecting client %s\n", subscriber.BuildInstanceKey())
			log.Println(err)
			cancel()
			return err
		}

		if xdsConfigDao.IsACK(subscriber, req.ResponseNonce) {
			log.Printf("Received ACK %s from %s", req.ResponseNonce, subscriber.BuildInstanceKey())
			xdsConfigDao.RemoveNonce(subscriber, req.ResponseNonce)
			subscriber.LastUpdatedVersion = req.VersionInfo
			xdsConfigDao.UpdateEnvoySubscriber(subscriber)
			continue
		} else {
			log.Printf("Response nonce not recognized %s", req.ResponseNonce)
		}
	}
}

func dispatchCluster(ctx context.Context, stream v2.ClusterDiscoveryService_StreamClustersServer,
	dispatchChannel chan bool) {
	for range dispatchChannel {
		select {
		case <-ctx.Done():
			return
		default:
		}

		subscriber := ctx.Value(envoySubscriberKey).(*model.EnvoySubscriber)
		mapper := mapper.ClusterMapper{}
		configJson, version := xdsConfigDao.GetClusterConfigJson(subscriber)
		clusterObj, err := mapper.GetResources(configJson)

		if err != nil {
			log.Println(err)
			log.Printf("Unable to dispatch config for %s\n", subscriber.BuildInstanceKey())
			continue
		}

		responseUUID := uuid.New().String()
		response := &v2.DiscoveryResponse{
			VersionInfo: version,
			Resources:   clusterObj,
			TypeUrl:     cache.ClusterType,
			Nonce:       responseUUID,
		}

		log.Printf("%+v\n", response)
		log.Printf("Sending config to %s \n %+v \n", subscriber.BuildInstanceKey(), response)

		xdsConfigDao.SaveNonceForStreamClusters(subscriber, responseUUID)
		err = stream.Send(response)
		if err != nil {
			log.Println("error sending to client")
			log.Println(err)
			xdsConfigDao.RemoveNonce(subscriber, responseUUID)
		} else {
			log.Printf("Successfully Sent config to %s \n", subscriber.BuildInstanceKey())
		}
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
		log.Printf("Checking consul for %s..\n", subscriber.BuildInstanceKey())
		if !xdsConfigDao.IsRepoPresent(subscriber) {
			continue
		}
		if clusterService.IsOutdated(subscriber) {
			log.Printf("Found update dispatching for %s\n", subscriber.BuildInstanceKey())
			dispatchChannel <- true
		}
	}
}
