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

// HelloService hello
type HelloService struct {
}

// SayHello hello
func (hs *HelloService) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Request: %v\r\n", request.Name)
	return &pb.HelloReply{Message: "Hello " + request.Name}, nil
}

var (
	host   = flag.String("host", "0.0.0.0", "host for server")
	port   = flag.String("port", "45202", "port for server")
	certIp = flag.String("cert-ip", "127.0.0.1", "ip for cert")
)

func generateServerCertAndKey(caCertPath, caKeyPath, serverCertPath, serverKeyPath string) error {
	_, err := os.Stat(serverCertPath)
	if err == nil {
		return nil
	}

	os.Remove(serverCertPath)
	os.Remove(serverKeyPath)

	localIP, err := gen.GetLocalIP()
	if err != nil {
		return fmt.Errorf("get local ip failure, nest error: %v", err)
	}

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
			IP:             []net.IP{net.ParseIP(localIP), net.ParseIP(*certIp)},
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
		return fmt.Errorf("generateCertificate server cert/key failure, nest error: %v", err)
	}

	if err := gen.WriteCertificate(serverCertPath, serverCertBytes); err != nil {
		return fmt.Errorf("write server cert failure, nest error: %v", err)
	}
	if err := gen.WritePKCS8PrivateKey(serverKeyPath, serverPrivBytes); err != nil {
		return fmt.Errorf("write server key failure, nest error: %v", err)
	}
	return nil
}

func main() {
	flag.Parse()

	var addr = net.JoinHostPort(*host, *port)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Listen addr[%s] failure, nest error: %v\r\n", addr, err)
	}

	var (
		caCertPath     = "certs/ca.crt"
		caKeyPath      = "certs/ca.key"
		serverCertPath = "certs/server.crt"
		serverKeyPath  = "certs/server.pem"
	)

	if err := generateServerCertAndKey(caCertPath, caKeyPath, serverCertPath, serverKeyPath); err != nil {
		log.Fatalf("Generate server cert and key failure, nest error: %v\r\n", err)
	}

	cert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		log.Fatalf("LoadX509KeyPair failure, nest error: %v", err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatalf("Read ca pem file failure, nest error: %v", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("AppendCertsFromPEM failure")
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		},
	})

	s := grpc.NewServer(grpc.Creds(creds))

	// Register EchoServer on the server.
	pb.RegisterGreeterServer(s, &HelloService{})

	log.Println("Server start")
	if err := s.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
