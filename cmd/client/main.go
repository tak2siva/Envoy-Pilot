package main

import (
	"google.golang.org/grpc"
	"log"
	"grpc_server/lib/proto"
	"time"
	"golang.org/x/net/context"
)

func main()  {
	conn, err := grpc.Dial("localhost:7777", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	c := api.NewPingClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &api.PingMessage{Greeting:"Hello from client"})

	if err != nil {
		log.Fatalf("could not ping: %v", err)
	}

	log.Printf("Response from server %s", r.Greeting)
}