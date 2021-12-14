package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eviltomorrow/developer-kit/grpc-go/pb"
	"google.golang.org/grpc"
)

func main() {

	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	log.Println("connetion ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := pb.NewGreeterClient(conn)
	repley, err := client.SayHello(ctx, &pb.HelloRequest{Name: "shepard"})
	if err != nil {
		log.Fatalf("SayHello error: %v", err)
	}
	fmt.Println(repley.Message)
}
