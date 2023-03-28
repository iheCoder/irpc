package config

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

const (
	AlpnQuicTransport = "wq-vvv-01"
)

func GenerateServerTLSConfig(tlsCertPath, tlsKeyPath string, nextProtos []string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   nextProtos,
	}, nil
}

func GenerateClientTLSConfig(tlsCertPath string, nextProtos []string) (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(tlsCertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
		NextProtos:         nextProtos,
	}, nil
}
