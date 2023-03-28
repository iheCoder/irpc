package server

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

func (p *StreamCodec) ReadRequest(reader io.Reader) (*common2.Request, error) {
	// 读取srvID
	var srvID uint16
	err := binary.Read(reader, binary.BigEndian, &srvID)
	if err != nil {
		return nil, err
	}

	// 读取methodID
	var methodID uint8
	err = binary.Read(reader, binary.BigEndian, &methodID)
	if err != nil {
		return nil, err
	}

	// 读取请求内容长度。默认不超过2^32
	var contentLen uint32
	err = binary.Read(reader, binary.BigEndian, &contentLen)
	if err != nil {
		return nil, err
	}

	// 读取内容
	content := make([]byte, contentLen)
	_, err = reader.Read(content)
	if err != nil {
		return nil, err
	}

	// 读取请求内容。log请求内容，并返回请求内容
	return &common2.Request{
		Header: common2.ReqHeader{
			SID: common2.SrvID(srvID),
			MID: common2.MethodID(methodID),
		},
		Body: content,
	}, nil
}

// WriteResponse 将result写入writer
// 返回结果是数组，但是result并不能编码为数组
func (p *StreamCodec) WriteResponse(writer io.Writer, resp *common2.Response) error {
	// 放入response长度
	resLen := len(resp.Body)
	res := make([]byte, 4+resLen)
	binary.BigEndian.PutUint32(res[:4], uint32(resLen))

	// 复制response
	copy(res[4:], resp.Body)

	_, err := writer.Write(res)
	if err != nil {
		return err
	}

	return nil
}

func (p *StreamCodec) ParseRequestBody(body []byte, kids []common2.KindID) ([]interface{}, error) {
	return p.parser.ParseBody(body, kids)
}

func (p *StreamCodec) EncodeBody(kids []common2.KindID, results ...interface{}) ([]byte, error) {
	return p.parser.EncodeBody(kids, results...)
}

// 总不能大于1<<8-1的时候，输入又是大端吧？虽然也不是不可以
// 低位的byte在前是小端
// 只能固定位数啦
//func (p *StreamCodec) parseSrvIDLen(reader io.Reader) (uint16, error) {
//	var srvIDLen uint16
//	sidLen := make([]byte, 1)
//	_, err := reader.Read(sidLen)
//	if err != nil {
//		return 0, err
//	}
//	V:=uint8(sidLen[0])
//	if V >= math.MaxUint8 {
//		sidLen1 := make([]byte, 1)
//		_, err = reader.Read(sidLen1)
//		if err != nil {
//			return 0, err
//		}
//		srvIDLen = binary.BigEndian.Uint16([]byte{sidLen[0], sidLen1[0]})
//	} else {
//		srvIDLen = uint16(sidLen[0])
//	}
//
//	return srvIDLen, nil
//}
