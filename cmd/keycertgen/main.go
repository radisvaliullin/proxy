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
	var isCAFlag = flag.Bool("isca", false, "create ca cert")
	var parentKeyNameFlag = flag.String("parentkey", "cakey", "name of key file")
	var parentCertNameFlag = flag.String("parentcert", "cacert", "name of certificate file")
	var keyNameFlag = flag.String("key", "key", "name of key file")
	var certNameFlag = flag.String("cert", "cert", "name of certificate file")
	var clientIDFlag = flag.String("clientid", "default@default.org", "client id, uniq client identificator, for example client email")
	flag.Parse()
	isCA := *isCAFlag
	keyPath := fmt.Sprintf("./sec/%s.pem", *keyNameFlag)
	certPath := fmt.Sprintf("./sec/%s.pem", *certNameFlag)
	commonName := *clientIDFlag
	parentKeyPath := fmt.Sprintf("./sec/%s.pem", *parentKeyNameFlag)
	parentCertPath := fmt.Sprintf("./sec/%s.pem", *parentCertNameFlag)

	// create key and self-signed cert files
	genKeyAndSelfSignedCert(isCA, keyPath, certPath, commonName, parentKeyPath, parentCertPath)
}

func genKeyAndSelfSignedCert(isCA bool, keyPath, certPath, commonName, parentKeyPath, parentCertPath string) {

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
	keyUsage := x509.KeyUsageDigitalSignature
	if isCA {
		keyUsage |= x509.KeyUsageCertSign
	}
	certTemplate := x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			Organization:  []string{"Galley"},
			Country:       []string{"US"},
			Province:      []string{"CA"},
			Locality:      []string{"SF"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94000"},
			CommonName:    commonName,
		},
		IsCA:                  isCA,
		DNSNames:              []string{"localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * 30 * time.Hour),
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// gen cert
	// for CA self-signed, or provide parent CA cert, key.
	ct := &certTemplate
	parentCt := &certTemplate
	certPubKey := &key.PublicKey
	certSignPrivateKey := key
	if !isCA {
		certSignPrivateKey, parentCt = getParentKeyCertFromFile(parentKeyPath, parentCertPath)
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, ct, parentCt, certPubKey, certSignPrivateKey)
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

func getParentKeyCertFromFile(parentKeyPath, parentCertPath string) (*ecdsa.PrivateKey, *x509.Certificate) {
	var parentCt *x509.Certificate
	var parentKey *ecdsa.PrivateKey
	// parent cert
	parentCtRaw, err := os.ReadFile(parentCertPath)
	if err != nil {
		log.Fatalf("fail: read parent cert file: %v", err)
	}
	parentCtPem, _ := pem.Decode(parentCtRaw)
	if parentCtPem == nil {
		log.Fatalf("fail: decode parent cert pem")
	}
	parentCt, err = x509.ParseCertificate(parentCtPem.Bytes)
	if err != nil {
		log.Fatalf("fail: parse parent cert file: %v", err)
	}
	// parent key
	parentKeyRaw, err := os.ReadFile(parentKeyPath)
	if err != nil {
		log.Fatalf("fail: read parent key file: %v", err)
	}
	parentKeyPem, _ := pem.Decode(parentKeyRaw)
	if parentKeyPem == nil {
		log.Fatalf("fail: decode parent key pem")
	}
	parentKeyAny, err := x509.ParsePKCS8PrivateKey(parentKeyPem.Bytes)
	if err != nil {
		log.Fatalf("fail: parse parent key pem: %v", err)
	}
	parentKey, ok := parentKeyAny.(*ecdsa.PrivateKey)
	if !ok {
		log.Printf("fail: parsed parent key, wrong type")
	}
	return parentKey, parentCt
}
