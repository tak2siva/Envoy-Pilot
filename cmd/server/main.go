package main

import (
	"Envoy-Pilot/cmd/server/constant"
	"Envoy-Pilot/cmd/server/dump"
	"Envoy-Pilot/cmd/server/metrics"
	"Envoy-Pilot/cmd/server/server"
	"Envoy-Pilot/cmd/server/service"
	"Envoy-Pilot/cmd/server/storage"
	myUtil "Envoy-Pilot/cmd/server/util"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	consul "github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var consulHandle *consul.KV

func init() {
	//host.docker.internal:8500
	initEnv()
	log.SetFlags(log.LstdFlags | log.Llongfile)
	if !constant.FILE_MODE {
		consulHealthCheck()
	}
	server.InitServerDeps()
}

func consulHealthCheck() {
	cwrapper := storage.GetConsulWrapper()
	cwrapper.GetUniqId()
}

func main() {
	go dump.SetUpHttpServer()
	go metrics.StartMetricsServer()
	go service.ConsulPollLoop()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 7777))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	v2.RegisterClusterDiscoveryServiceServer(s, &server.Server{})
	v2.RegisterListenerDiscoveryServiceServer(s, &server.Server{})
	v2.RegisterRouteDiscoveryServiceServer(s, &server.Server{})
	v2.RegisterEndpointDiscoveryServiceServer(s, &server.Server{})
	discovery.RegisterAggregatedDiscoveryServiceServer(s, &server.Server{})

	reflection.Register(s)

	log.Print("Started grpc server..")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}

func initEnv() {
	err := godotenv.Load(constant.ENV_PATH)
	if err != nil {
		log.Print(err)
		log.Fatal("Error loading .env file")
	}

	if len(os.Getenv("POLL_INTERVAL")) > 0 {
		res, err := time.ParseDuration(os.Getenv("POLL_INTERVAL"))
		myUtil.Check(err)
		constant.POLL_INTERVAL = res
	}

	if len(os.Getenv("CONSUL_PREFIX")) > 0 {
		constant.CONSUL_PREFIX = os.Getenv("CONSUL_PREFIX")
	}

	fileMode := os.Getenv("FILE_MODE")
	if fileMode == "true" {
		constant.FILE_MODE = true
	}

	if len(os.Getenv("FOLDER_PATH")) > 0 {
		constant.FOLDER_PATH = os.Getenv("FOLDER_PATH")
	}

	if constant.FILE_MODE && len(constant.FOLDER_PATH) == 0 {
		panic("Missing config folder path env variable FOLDER_PATH..\n")
	}

	log.Printf("------- ENV VALUES -----\n")
	log.Printf("FILE_MODE: %t", constant.FILE_MODE)
	log.Printf("FOLDER_PATH: %s", constant.FOLDER_PATH)
	log.Printf("CONSUL_PREFIX: %s", constant.CONSUL_PREFIX)
	log.Printf("------------------------\n")
}
