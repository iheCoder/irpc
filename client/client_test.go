package client

import (
	"context"
	"fmt"
	"learn/irpc/common"
	"learn/irpc/config"
	"learn/irpc/service"
	"sync"
	"testing"
)

const (
	DefaultDialAddr = "127.0.0.1:4433"
	TLSCertPath     = "/Users/ihewe/GolandProjects/learn/net/http2/try4/root.pem"
	TLSKeyPath      = "/Users/ihewe/GolandProjects/learn/net/http2/try4/root.key"
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

func TestRunClientBasic(t *testing.T) {
	mgr := service.NewServiceMgr("../config/services.yml")
	err := mgr.Register(&ServerTest{})
	if err != nil {
		t.Fatal(err)
	}
	parser := common.NewParser(mgr.GetModels())

	cc := NewStreamCodec(parser)
	tlsConfig, err := config.GenerateClientTLSConfig(TLSCertPath, protos)
	if err != nil {
		t.Fatal(err)
	}

	client := NewIrpcClient(context.Background(), cc, mgr, tlsConfig, DefaultDialAddr)

	r, err := client.Call("ServerTest", "Add", 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(r)
}

func TestALotClientCall(t *testing.T) {
	mgr := service.NewServiceMgr("../config/services.yml")
	err := mgr.Register(&ServerTest{})
	if err != nil {
		t.Fatal(err)
	}
	parser := common.NewParser(mgr.GetModels())

	cc := NewStreamCodec(parser)
	tlsConfig, err := config.GenerateClientTLSConfig(TLSCertPath, protos)
	if err != nil {
		t.Fatal(err)
	}

	client := NewIrpcClient(context.Background(), cc, mgr, tlsConfig, DefaultDialAddr)

	times := 1000

	wg := &sync.WaitGroup{}
	wg.Add(times)
	for i := 0; i < times; i++ {
		go func() {
			r, err := client.Call("ServerTest", "Add", 1, 2)
			if err != nil {
				panic(err)
			}
			if r[0] != 3 {
				panic("wrong result")
			}

			wg.Done()
		}()
	}
	wg.Wait()
}

func TestCallParamStruct(t *testing.T) {
	mgr := service.NewServiceMgr("../config/services.yml")
	err := mgr.Register(&ServerTest{})
	if err != nil {
		t.Fatal(err)
	}
	parser := common.NewParser(mgr.GetModels())

	cc := NewStreamCodec(parser)
	tlsConfig, err := config.GenerateClientTLSConfig(TLSCertPath, protos)
	if err != nil {
		t.Fatal(err)
	}

	client := NewIrpcClient(context.Background(), cc, mgr, tlsConfig, DefaultDialAddr)

	times := 1000

	wg := &sync.WaitGroup{}
	wg.Add(times)
	x := X{1}
	y := Y{2}
	for i := 0; i < times; i++ {
		go func() {
			r, err := client.Call("ServerTest", "AddWithStruct", x, y)
			if err != nil {
				panic(err)
			}
			if r[0].(Z).V != 3 {
				emsg := fmt.Sprintf("wrong result with %v", r)
				panic(emsg)
			}

			wg.Done()
		}()
	}
	wg.Wait()
}
