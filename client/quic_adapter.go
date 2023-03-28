package client

import (
	"crypto/tls"
	"time"
)

type QuicAdapter struct {
	ac *AdapterConn
}

func NewQuicAdapter(tlsConfig *tls.Config, dialAddr string, expireDuration, scanDuration time.Duration) *QuicAdapter {
	ac := NewAdapterConn(tlsConfig, dialAddr, expireDuration, scanDuration)

	// 初始化quicAdapter
	qa := &QuicAdapter{
		ac: ac,
	}

	return qa
}

func (a *QuicAdapter) Request(b []byte) (StreamConn, error) {
	streamConn, err := a.ac.AcquireStream()
	if err != nil {
		return nil, err
	}

	// write bytes in open stream
	_, err = streamConn.Write(b)
	if err != nil {
		return nil, err
	}

	return streamConn, nil
}
