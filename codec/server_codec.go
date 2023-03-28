package codec

import (
	"encoding/json"
	"learn/irpc/common"
	"reflect"
)

type ServerCodec struct {
}

func (c *ServerCodec) EncodeResponse(result interface{}) ([]byte, error) {
	// 将返回体解析为json
	resp := &common.Response{}
	mr, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	// 将response解析为json
	resp.Body = mr
	r, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *ServerCodec) DecodeRequest(req []byte, body interface{}) (*common.Request, error) {
	// 确保body是指针
	if reflect.TypeOf(body).Kind() != reflect.Ptr {
		return nil, ErrRequestBodyNotPtr
	}

	// 解析header
	cr := &common.Request{}
	err := json.Unmarshal(req, cr)
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
