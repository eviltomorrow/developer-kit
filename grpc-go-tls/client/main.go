package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/eviltomorrow/developer-kit/grpc-go-tls/gen"
	"github.com/eviltomorrow/developer-kit/grpc-go-tls/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	host = flag.String("host", "127.0.0.1", "host for server")
	port = flag.String("port", "45202", "port for server")
)

func generateServerCertAndKey(caCertPath, caKeyPath, clientCertPath, clientKeyPath string) error {
	_, err := os.Stat(clientCertPath)
	if err == nil {
		return nil
	}

	os.Remove(clientCertPath)
	os.Remove(clientKeyPath)

	caCert, err := gen.ReadCertificate(caCertPath)
	if err != nil {
		return fmt.Errorf("ParseCertificate CA cert failure, nest error: %v", err)
	}

	caKey, err := gen.ReadPKCS1PrivateKey(caKeyPath)
	if err != nil {
		return fmt.Errorf("ReadPKCS1PrivateKey CA key failure, nest error: %v", err)
	}

	serverPrivBytes, serverCertBytes, err := gen.GenerateCertificate(caKey, caCert, 2048, &gen.ApplicationInformation{
		CertificateConfig: &gen.CertificateConfig{
			ExpirationTime: 24 * time.Hour * 365,
		},
		CommonName:           "www.roigo.com",
		CountryName:          "CN",
		ProvinceName:         "BeiJing",
		LocalityName:         "BeiJing",
		OrganizationName:     "Roigo Inc",
		OrganizationUnitName: "Development",
	})
	if err != nil {
		return fmt.Errorf("generateCertificate client cert/key failure, nest error: %v", err)
	}

	if err := gen.WriteCertificate(clientCertPath, serverCertBytes); err != nil {
		return fmt.Errorf("write client cert failure, nest error: %v", err)
	}
	if err := gen.WritePKCS8PrivateKey(clientKeyPath, serverPrivBytes); err != nil {
		return fmt.Errorf("write client key failure, nest error: %v", err)
	}
	return nil
}

func main() {
	flag.Parse()

	var (
		caCertPath     = "certs/ca.crt"
		caKeyPath      = "certs/ca.key"
		clientCertPath = "certs/client.crt"
		clientKeyPath  = "certs/client.pem"
		serverName     = "www.roigo.com"
	)

	if err := generateServerCertAndKey(caCertPath, caKeyPath, clientCertPath, clientKeyPath); err != nil {
		log.Fatalf("Generate client cert and key failure, nest error: %v\r\n", err)
	}
	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		log.Fatalf("LoadX509KeyPair failure: %v", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatalf("Load caCertFile failure: %v", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("AppendCertsFromPEM failure: %v", err)
	}

	// Create the TLS credentials for transport
	creds := credentials.NewTLS(&tls.Config{
		ServerName:   serverName,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	var addr = net.JoinHostPort(*host, *port)
	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
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
	log.Printf("Response: %v\r\n", repley.Message)
}
