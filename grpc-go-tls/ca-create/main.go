package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/eviltomorrow/developer-kit/grpc-go-tls/gen"
)

func main() {
	var baseDir = "certs"
	// 生成 根 证书和密钥
	caPrivBytes, caCertBytes, err := gen.GenerateCertificate(nil, nil, 2048, &gen.ApplicationInformation{
		CertificateConfig: &gen.CertificateConfig{
			IsCA:           true,
			ExpirationTime: 24 * time.Hour * 365 * 3,
		},
		CommonName:           "www.roigo.com",
		CountryName:          "CN",
		ProvinceName:         "BeiJing",
		LocalityName:         "BeiJing",
		OrganizationName:     "Roigo Inc",
		OrganizationUnitName: "Development",
	})
	if err != nil {
		log.Fatalf("GenerateCertificate failure, nest error: %v", err)
	}
	if err := gen.WritePKCS1PrivateKey(filepath.Join(baseDir, "ca.key"), caPrivBytes); err != nil {
		log.Fatalf("Write ca key failure, nest error: %v\r\n", err)
	}
	if err := gen.WriteCertificate(filepath.Join(baseDir, "ca.crt"), caCertBytes); err != nil {
		log.Fatalf("Write ca cert failure, nest error: %v\r\n", err)
	}

}
