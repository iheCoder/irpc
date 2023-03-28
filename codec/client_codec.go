package codec

import (
	"encoding/json"
	"learn/irpc/common"
	"reflect"
)

type ClientCodec struct {
}

func (c *ClientCodec) DecodeResponse(res []byte, body interface{}) (*common.Response, error) {
	// 确保body是指针
	if reflect.TypeOf(body).Kind() != reflect.Ptr {
		return nil, ErrRequestBodyNotPtr
	}

	// 解析header
	cr := &common.Response{}
	err := json.Unmarshal(res, cr)
	if err != nil {
		return nil, err
	}

	// 解析body放入Request
	err = json.Unmarshal(cr.Body, body)
	if err != nil {
		return nil, err
	}

	return cr, nil
}

func (c *ClientCodec) EncodeRequest(body interface{}, header common.ReqHeader) ([]byte, error) {
	// 将返回体解析为json
	req := common.Request{}
	mr, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// 将response解析为json
	req.Body = mr
	req.Header = header
	r, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return r, nil
}
