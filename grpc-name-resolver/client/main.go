package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eviltomorrow/developer-kit/grpc-name-resolver/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	mySchema    = "shepard"
	myService   = "roigo.me"
	backendAddr = "localhost:8080"
)

func main() {
	var (
		addr = fmt.Sprintf("%s:///%s", mySchema, myService)
	)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
