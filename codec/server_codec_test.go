package codec

import (
	"encoding/json"
	"learn/irpc/common"
	"testing"
)

type CodecHello struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestDecodeReq(t *testing.T) {
	b := CodecHello{
		ID:   1,
		Name: "hello",
	}
	bytes, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	req := &common.Request{
		Header: common.ReqHeader{
			SID: 1,
			MID: 1,
		},
		Body: bytes,
	}

	r1, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	sc := &ServerCodec{}
	bp := &CodecHello{}
	request, err := sc.DecodeRequest(r1, bp)
	if err != nil {
		t.Fatal(err)
	}
	if request.Header.SID != 1 || request.Header.MID != 1 || bp.ID != 1 || bp.Name != "hello" {
		t.Fatal("wrong parse")
	}
}

func TestC2SReq(t *testing.T) {
	cc := &ClientCodec{}
	b := CodecHello{
		ID:   1,
		Name: "hello",
	}
	h := common.ReqHeader{
		SID: 1,
		MID: 1,
	}
	req, err := cc.EncodeRequest(b, h)
	if err != nil {
		t.Fatal(err)
	}

	sc := &ServerCodec{}
	bp := &CodecHello{}
	request, err := sc.DecodeRequest(req, bp)
	if err != nil {
		t.Fatal(err)
	}
	if request.Header.SID != 1 || request.Header.MID != 1 || bp.ID != 1 || bp.Name != "hello" {
		t.Fatal("wrong parse")
	}
}

func TestS2CRes(t *testing.T) {
	cc := &ClientCodec{}
	sc := &ServerCodec{}

	b := CodecHello{
		ID:   1,
		Name: "hello",
	}
	rbytes, err := sc.EncodeResponse(b)
	if err != nil {
		t.Fatal(err)
	}

	bp := &CodecHello{}
	_, err = cc.DecodeResponse(rbytes, bp)
	if err != nil {
		t.Fatal(err)
	}
	if bp.ID != 1 || bp.Name != "hello" {
		t.Fatal("wrong parse")
	}
}
