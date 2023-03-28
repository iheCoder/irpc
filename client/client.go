package client

import (
	"context"
	"crypto/tls"
	"github.com/lucas-clemente/quic-go"
	"io"
	"learn/irpc/common"
	"learn/irpc/service"
	"time"
)

type IrpcClient struct {
	ctx       context.Context
	cc        *StreamCodec
	mgr       *service.Mgr
	tlsConfig *tls.Config
	dialAddr  string
	conn      quic.Connection
	requester *QuicAdapter
}

const (
	defaultExpireDuration = time.Second
	defaultScanDuration   = 100 * time.Millisecond
)

func NewIrpcClient(ctx context.Context, cc *StreamCodec, mgr *service.Mgr, tlsConfig *tls.Config, dialAddr string) *IrpcClient {
	// 初始化client
	c := &IrpcClient{
		ctx:       ctx,
		cc:        cc,
		mgr:       mgr,
		tlsConfig: tlsConfig,
		dialAddr:  dialAddr,
	}

	qa := NewQuicAdapter(tlsConfig, dialAddr, defaultExpireDuration, defaultScanDuration)

	c.requester = qa

	return c
}

// Call 根据服务名、方法名以及参数去请求
func (c *IrpcClient) Call(srvName, methodName string, params ...interface{}) ([]interface{}, error) {
	// 根据srvName、methodName获取相应编号
	// 为什么不使得Mgr Invoke参数为srvName, methodName呢？反正srvName、methodName获取id也要通过Mgr啊
	// 错了，这是要传递id到服务端啊
	srvID, mid, err := c.mgr.GetSrvMethodID(srvName, methodName)
	if err != nil {
		return nil, err
	}

	// 构造请求
	req, err := c.constructReq(srvID, mid, params...)
	if err != nil {
		return nil, err
	}

	// 编码请求
	encodeReq, err := c.cc.EncodeToRequest(req)
	if err != nil {
		return nil, err
	}

	// 发送请求
	respReader, err := c.requester.Request(encodeReq)
	if err != nil {
		return nil, err
	}

	// 获取响应内容并解析
	resp, err := c.parseResp(respReader, srvID, mid)
	if err != nil {
		return nil, err
	}

	// 通知使用完毕
	err = respReader.Close()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *IrpcClient) parseResp(reader io.Reader, srvID common.SrvID, mid common.MethodID) ([]interface{}, error) {
	// 解析响应
	// server 写入了正确的response，可是在最后主动断开了该连接。这导致response根本没有返回
	// 问题在于请求过程中即使超过了时间，那么也不应该断开连接
	// 会有一直请求却没有得到回应的情况嘛？
	response, err := c.cc.ReadResponse(reader)
	if err != nil {
		return nil, err
	}

	// 根据methodID获取输出参数kindIDs
	_, outKids := c.mgr.GetKindIDsByMethod(srvID, mid)

	return c.cc.ParseResponseBody(response.Body, outKids)
}

func (c *IrpcClient) constructReq(srvID common.SrvID, mid common.MethodID, params ...interface{}) (*common.Request, error) {
	// 根据methodID获取输入参数kindIDs
	inKids, _ := c.mgr.GetKindIDsByMethod(srvID, mid)

	// 构造请求body
	body, err := c.cc.EncodeBody(inKids, params...)
	if err != nil {
		return nil, err
	}

	req := &common.Request{
		Header: common.ReqHeader{
			SID: srvID,
			MID: mid,
		},
		Body: body,
	}

	return req, nil
}
