// KeyCertGen generates key and self-signed certificates
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

func main() {
	log.Print("generate")

	// config
	// var keyPath, certPath, commonName string
	var keyNameFlag = flag.String("key", "key", "name of key file")
	var certNameFlag = flag.String("cert", "cert", "name of certificate file")
	var clientIDFlag = flag.String("clientid", "default@default.org", "client id, uniq client identificator, for example client email")
	flag.Parse()
	keyPath := fmt.Sprintf("./sec/%s.pem", *keyNameFlag)
	certPath := fmt.Sprintf("./sec/%s.pem", *certNameFlag)
	commonName := *clientIDFlag

	// create key and self-signed cert files
	genKeyAndSelfSignedCert(keyPath, certPath, commonName)
}

func genKeyAndSelfSignedCert(keyPath, certPath, commonName string) {

	// generate new key
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("fail: generate key: %v", err)
	}

	// certificate template
	serNumMaxLim := new(big.Int).Lsh(big.NewInt(1), 128)
	sn, err := rand.Int(rand.Reader, serNumMaxLim)
	if err != nil {
		log.Fatalf("fail: generate serial number: %v", err)
	}
	certTemplate := x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		DNSNames:              []string{"localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * 30 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// gen cert
	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &key.PublicKey, key)
	if err != nil {
		log.Fatalf("fail: create certificate: %v", err)
	}

	// create cert file
	pemCertBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if pemCertBytes == nil {
		log.Fatal("fail: encode certificate to PEM")
	}
	if err := os.WriteFile(certPath, pemCertBytes, 0644); err != nil {
		log.Fatalf("fail: write cert to file: %v", err)
	}
	log.Printf("created %v", certPath)

	// create key file
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		log.Fatalf("fail: marshal private key: %v", err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})
	if pemKey == nil {
		log.Fatal("fail: encode key to PEM")
	}
	if err := os.WriteFile(keyPath, pemKey, 0600); err != nil {
		log.Fatalf("fail: write key to file: %v", err)
	}
	log.Printf("created %v", keyPath)
}
