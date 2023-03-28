package service

import (
	"learn/irpc/common"
	"log"
)

type service struct {
	methods map[common.MethodID]*method
}

func (s *service) call(mn common.MethodID, argv []interface{}) []interface{} {
	m, ok := s.methods[mn]
	if !ok {
		log.Fatalln("wrong method")
	}

	return m.call(argv)
}
