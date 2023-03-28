package config

import (
	"learn/irpc/common"
)

type ServicesConfig struct {
	Services []*ServiceConfig `yaml:"services"`
}

// 根本不需要根据路径解析出service
// ID, Name 必须确保唯一，同一个服务中方法id、name也必须唯一
type ServiceConfig struct {
	ID      common.SrvID               `yaml:"id"`
	Name    string                     `yaml:"name"`
	Methods map[string]common.MethodID `yaml:"methods"`
}

type ServerConfig struct {
	ListenAddr  string `yaml:"listen_addr"`
	TLSCertPath string `yaml:"tls_cert_path"`
	TLSKeyPath  string `yaml:"tls_key_path"`
}

type ClientConfig struct {
	DialAddr    string `yaml:"dial_addr"`
	TLSCertPath string `yaml:"tls_cert_path"`
	TLSKeyPath  string `yaml:"tls_key_path"`
}
