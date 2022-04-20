package gen

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

type ApplicationInformation struct {
	CertificateConfig    *CertificateConfig
	CommonName           string
	CountryName          string
	ProvinceName         string
	LocalityName         string
	OrganizationName     string
	OrganizationUnitName string
}

type CertificateConfig struct {
	IsCA           bool
	IP             []net.IP
	DNS            []string
	ExpirationTime time.Duration
}

func GenerateCertificate(caKey *rsa.PrivateKey, caCert *x509.Certificate, bits int, info *ApplicationInformation) ([]byte, []byte, error) {
	if !info.CertificateConfig.IsCA {
		if caKey == nil || caCert == nil {
			return nil, nil, fmt.Errorf("miss ca key/cert")
		}
	}

	priv, err := rsa.GenerateKey(cryptorand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	var template = x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:         info.CommonName,
			Country:            []string{info.CountryName},
			Province:           []string{info.ProvinceName},
			Locality:           []string{info.LocalityName},
			Organization:       []string{info.OrganizationName},
			OrganizationalUnit: []string{info.OrganizationUnitName},
		},
		NotBefore: time.Now().Add(-24 * time.Hour),
		NotAfter:  time.Now().Add(info.CertificateConfig.ExpirationTime),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		BasicConstraintsValid: true,
	}

	if info.CertificateConfig.IsCA {
		template.IsCA = true
	} else {
		if i := net.ParseIP(info.CommonName); i != nil {
			template.IPAddresses = append(template.IPAddresses, i)
		} else {
			template.DNSNames = append(template.DNSNames, info.CommonName)
		}
		template.IPAddresses = append(template.IPAddresses, info.CertificateConfig.IP...)
		template.DNSNames = append(template.DNSNames, info.CertificateConfig.DNS...)
	}

	var key *rsa.PrivateKey

	if info.CertificateConfig.IsCA {
		caCert = &template
		key = priv
	} else {
		key = caKey
	}

	certBytes, err := x509.CreateCertificate(cryptorand.Reader, &template, caCert, &priv.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}

	return x509.MarshalPKCS1PrivateKey(priv), certBytes, nil
}

func ReadCertificate(path string) (*x509.Certificate, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(buffer)
	if block == nil {
		return nil, fmt.Errorf("decode certificate failure, block is nil")
	}

	return x509.ParseCertificate(block.Bytes)
}

func WriteCertificate(path string, cert []byte) error {
	_, err := x509.ParseCertificate(cert)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := pem.Encode(&buffer, &pem.Block{Type: "CERTIFICATE", Bytes: cert}); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buffer.Bytes())
	return err
}

func ReadPKCS1PrivateKey(path string) (*rsa.PrivateKey, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(buffer)
	if block == nil {
		return nil, fmt.Errorf("decode private key failure, block is nil")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func ReadPKCS8PrivateKey(path string) (*rsa.PrivateKey, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(buffer)
	if block == nil {
		return nil, fmt.Errorf("decode private key failure, block is nil")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	if val, ok := key.(*rsa.PrivateKey); ok {
		return val, nil
	}
	return nil, fmt.Errorf("ParsePKCS8PrivateKey failure")
}

func WritePKCS1PrivateKey(path string, privKey []byte) error {
	_, err := x509.ParsePKCS1PrivateKey(privKey)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := pem.Encode(&buffer, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privKey}); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buffer.Bytes())
	return err
}

func WritePKCS8PrivateKey(path string, privKey []byte) error {
	priv, err := x509.ParsePKCS1PrivateKey(privKey)
	if err != nil {
		return err
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := pem.Encode(&buffer, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buffer.Bytes())
	return err
}

var _ = demo

func demo() {
	var baseDir = "certs"
	// 生成 根 证书和密钥
	caPrivBytes, caCertBytes, err := GenerateCertificate(nil, nil, 2048, &ApplicationInformation{
		CertificateConfig: &CertificateConfig{
			IsCA:           true,
			ExpirationTime: 24 * time.Hour * 365 * 3,
		},
		CommonName:           "www.eviltomorrow.com",
		CountryName:          "CN",
		ProvinceName:         "BeiJing",
		LocalityName:         "BeiJing",
		OrganizationName:     "Eviltomorrow Inc",
		OrganizationUnitName: "Development",
	})
	if err != nil {
		log.Fatalf("GenerateCertificate failure, nest error: %v", err)
	}
	WritePKCS1PrivateKey(filepath.Join(baseDir, "ca.key"), caPrivBytes)
	WriteCertificate(filepath.Join(baseDir, "ca.crt"), caCertBytes)

	// 生成 Server 证书
	caKey, err := x509.ParsePKCS1PrivateKey(caPrivBytes)
	if err != nil {
		log.Fatalf("ParsePKCS1PrivateKey CA key failure, nest error: %v", err)
	}
	caCert, err := x509.ParseCertificate(caCertBytes)
	if err != nil {
		log.Fatalf("ParseCertificate CA cert failure, nest error: %v", err)
	}

	// // 生成 Server 证书
	// caKey, err := ReadPKCS8PrivateKey(filepath.Join(baseDir, "ca.key"))
	// if err != nil {
	// 	log.Fatalf("ReadPKCS1PrivateKey CA key failure, nest error: %v", err)
	// }

	// caCert, err := ReadCertificate(filepath.Join(baseDir, "ca.crt"))
	// if err != nil {
	// 	log.Fatalf("ParseCertificate CA cert failure, nest error: %v", err)
	// }

	serverPrivBytes, serverCertBytes, err := GenerateCertificate(caKey, caCert, 2048, &ApplicationInformation{
		CertificateConfig: &CertificateConfig{
			IP:             []net.IP{net.ParseIP("127.0.0.1")},
			ExpirationTime: 24 * time.Hour * 365,
		},
		CommonName:           "localhost",
		CountryName:          "CN",
		ProvinceName:         "BeiJing",
		LocalityName:         "BeiJing",
		OrganizationName:     "Eviltomorrow Inc",
		OrganizationUnitName: "Development",
	})
	if err != nil {
		log.Fatalf("GenerateCertificate CA cert failure, nest error: %v", err)
	}
	WritePKCS1PrivateKey(filepath.Join(baseDir, "server.key"), serverPrivBytes)
	WritePKCS8PrivateKey(filepath.Join(baseDir, "server.pem"), serverPrivBytes)
	WriteCertificate(filepath.Join(baseDir, "server.crt"), serverCertBytes)

	clientPrivBytes, clientCertBytes, err := GenerateCertificate(caKey, caCert, 2048, &ApplicationInformation{
		CertificateConfig: &CertificateConfig{
			ExpirationTime: 24 * time.Hour * 365,
		},
		CommonName:           "localhost",
		CountryName:          "CN",
		ProvinceName:         "BeiJing",
		LocalityName:         "BeiJing",
		OrganizationName:     "Apple Inc",
		OrganizationUnitName: "Dev",
	})
	if err != nil {
		log.Fatalf("GenerateCertificate CA cert failure, nest error: %v", err)
	}
	WritePKCS1PrivateKey(filepath.Join(baseDir, "client.key"), clientPrivBytes)
	WritePKCS8PrivateKey(filepath.Join(baseDir, "client.pem"), clientPrivBytes)
	WriteCertificate(filepath.Join(baseDir, "client.crt"), clientCertBytes)
}