package server

import (
	"bytes"
	"learn/irpc/client"
	"learn/irpc/common"
	"testing"
)

func TestParseSrvIDLen(t *testing.T) {
	//sp := &StreamCodec{}

	//l := uint16(10)
	//p := make([]byte, 2)
	//binary.LittleEndian.PutUint16(p, l)
	//reader := bytes.NewReader(p)
	//idLen, err := sp.parseSrvIDLen(reader)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//if idLen != 10 {
	//	t.Fatalf("unexpexcted %d", idLen)
	//}

	//l1 := uint16(1024)
	//p1 := make([]byte, 2)
	//binary.BigEndian.PutUint16(p1, l1)
	//t.Logf("p %d", binary.BigEndian.Uint16(p1))
	//reader1 := bytes.NewReader(p1)
	//
	//p2 := make([]byte, 2)
	//_, err := reader1.Read(p2)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Logf("first %d second %d", p1[0], p1[1])
	//t.Logf("reader %d", binary.BigEndian.Uint16(p2))
	//idLen1, err := sp.parseSrvIDLen(reader1)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//if idLen1 != 1024 {
	//	t.Fatalf("unexpexcted %d", idLen1)
	//}
}

func TestParseToRequest(t *testing.T) {
	req := &common.Request{
		Header: common.ReqHeader{
			SID: 1,
			MID: 1,
		},
		Body: []byte("hello"),
	}
	csc := &client.StreamCodec{}
	encodeReq, err := csc.EncodeToRequest(req)
	if err != nil {
		t.Fatal(err)
	}

	ssc := &StreamCodec{}
	reader := bytes.NewReader(encodeReq)
	request, err := ssc.ReadRequest(reader)
	if err != nil {
		t.Fatal(err)
	}

	if request.Header.SID != req.Header.SID || request.Header.MID != req.Header.MID || !bytes.Equal(request.Body, req.Body) {
		t.Fatal("req mismatch")
	}
}
