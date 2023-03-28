package server

import (
	"context"
	"learn/irpc/common"
	"learn/irpc/config"
	"learn/irpc/service"
	"testing"
)

const (
	DefaultListenAddr = "0.0.0.0:4433"
	TLSCertPath       = "/Users/ihewe/GolandProjects/learn/net/http2/try4/root.pem"
	TLSKeyPath        = "/Users/ihewe/GolandProjects/learn/net/http2/try4/root.key"
)

var protos = []string{config.AlpnQuicTransport}

type ServerTest struct {
}

func (s *ServerTest) Add(x, y int) int {
	return x + y
}

type X struct {
	V int
}

type Y struct {
	V int
}

type Z struct {
	V int
}

func (s *ServerTest) AddWithStruct(x X, y Y) Z {
	return Z{x.V + y.V}
}

func TestRunServerBasic(t *testing.T) {
	mgr := service.NewServiceMgr("../config/services.yml")
	err := mgr.Register(&ServerTest{})
	if err != nil {
		t.Fatal(err)
	}

	parser := common.NewParser(mgr.GetModels())

	cc := NewStreamCodec(parser)

	tlsConfig, err := config.GenerateServerTLSConfig(TLSCertPath, TLSKeyPath, protos)
	if err != nil {
		t.Fatal(err)
	}

	server := NewIrpcServer(tlsConfig, DefaultListenAddr, context.Background(), cc, mgr)
	err = server.Run()
	if err != nil {
		t.Fatal(err)
	}
}
