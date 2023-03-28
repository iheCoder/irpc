package client

import (
	"encoding/binary"
	"io"
	common2 "learn/irpc/common"
)

type StreamCodec struct {
	parser *common2.Parser
}

func NewStreamCodec(parser *common2.Parser) *StreamCodec {
	return &StreamCodec{parser: parser}
}

func (c *StreamCodec) EncodeToRequest(req *common2.Request) ([]byte, error) {
	// json marshal content
	contentLen := len(req.Body)
	r := make([]byte, 2+1+4+contentLen)

	// 编码srvID
	binary.BigEndian.PutUint16(r[:2], uint16(req.Header.SID))

	// 编码mID
	r[2] = byte(req.Header.MID)

	// 编码content len
	binary.BigEndian.PutUint32(r[3:3+4], uint32(contentLen))

	// 编码content
	copy(r[7:], req.Body)

	return r, nil
}

func (c *StreamCodec) EncodeBody(kids []common2.KindID, params ...interface{}) ([]byte, error) {
	return c.parser.EncodeBody(kids, params...)
}

func (c *StreamCodec) ReadResponse(reader io.Reader) (*common2.Response, error) {
	var contentLen uint32
	err := binary.Read(reader, binary.BigEndian, &contentLen)
	if err != nil {
		return nil, err
	}

	body := make([]byte, contentLen)
	_, err = reader.Read(body)
	if err != nil {
		return nil, err
	}

	res := &common2.Response{Body: body}
	return res, nil
}

func (c *StreamCodec) ParseResponseBody(body []byte, kids []common2.KindID) ([]interface{}, error) {
	return c.parser.ParseBody(body, kids)
}
