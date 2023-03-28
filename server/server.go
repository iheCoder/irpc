package server

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/lucas-clemente/quic-go"
	"io"
	common2 "learn/irpc/common"
	"learn/irpc/service"
	"log"
)

type IrpcServer struct {
	TLSConfig  *tls.Config
	ListenAddr string
	ctx        context.Context
	cc         *StreamCodec
	mgr        *service.Mgr
}

var (
	connFinishedErr = errors.New("conn handle finished")
)

func NewIrpcServer(tlsConfig *tls.Config, listenAddr string, ctx context.Context, cc *StreamCodec, mgr *service.Mgr) *IrpcServer {
	return &IrpcServer{
		TLSConfig:  tlsConfig,
		ListenAddr: listenAddr,
		ctx:        ctx,
		cc:         cc,
		mgr:        mgr,
	}
}

func (s *IrpcServer) Run() error {
	// 监听端口
	listener, err := quic.ListenAddr(s.ListenAddr, s.TLSConfig, nil)
	if err != nil {
		return err
	}

	for {
		// 接受连接
		conn, err := listener.Accept(context.Background())
		if err != nil {
			return err
		}
		log.Printf("irpcServer NewIrpcClient: conn accepted: %s", conn.RemoteAddr().String())

		// 处理连接。并在处理完毕后关闭
		go s.handleConn(conn)
	}
}

func (s *IrpcServer) handleConn(conn quic.Connection) {
	for {
		// 接受流
		stream, err := conn.AcceptStream(s.ctx)
		if err != nil {
			err = handleConnErr(err)
			if err == connFinishedErr {
				return
			}
			log.Printf("irpcServer handleConn: accept stream failed %s", err)
			return
		}

		// 处理流
		go s.handleStream(stream)
	}

}

func (s *IrpcServer) handleStream(stream quic.Stream) {
	for {
		// 解析请求
		request, err := s.cc.ReadRequest(stream)
		if err != nil {
			err = handleConnErr(err)
			if err == connFinishedErr || err == io.EOF {
				return
			}
			log.Printf("irpcServer handleStream: decode to req failed %s", err)
			return
		}

		// 解析请求参数
		inParams, _ := s.mgr.GetKindIDsByMethod(request.Header.SID, request.Header.MID)
		params, err := s.cc.ParseRequestBody(request.Body, inParams)
		if err != nil {
			log.Printf("irpcServer handleStream: parse req body %s failed %s", string(request.Body), err)
			return
		}

		// 调用方法
		result := s.mgr.Invoke(request.Header.SID, request.Header.MID, params)

		// 构造响应
		resp, err := s.constructResp(request.Header.SID, request.Header.MID, result...)
		if err != nil {
			log.Printf("irpcServer handleStream: construct response failed %s", err)
			return
		}

		// 编码结果为response
		err = s.cc.WriteResponse(stream, resp)
		if err != nil {
			err = handleConnErr(err)
			if err == connFinishedErr || err == io.EOF {
				return
			}
			log.Printf("irpcServer handleStream: write response failed %s", err)
			return
		}
	}
}

func handleConnErr(err error) error {
	switch e := err.(type) {
	case *quic.ApplicationError:
		if e.ErrorCode == common2.TooLongToUsedErrCode {
			log.Printf("info handleConnErr: remote client close")
			return connFinishedErr
		}
	case *quic.IdleTimeoutError:
		log.Printf("info handleConnErr: remote client idle timeout error")
		return connFinishedErr
	}

	return err
}

func (s *IrpcServer) constructResp(srvID common2.SrvID, mid common2.MethodID, result ...interface{}) (*common2.Response, error) {
	// 根据methodID获取输入参数kindIDs
	_, outKinds := s.mgr.GetKindIDsByMethod(srvID, mid)

	// 构造响应body
	body, err := s.cc.EncodeBody(outKinds, result...)
	if err != nil {
		return nil, err
	}

	resp := &common2.Response{Body: body}
	return resp, nil
}
