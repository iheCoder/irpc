package service

import (
	"learn/irpc/common"
	"reflect"
)

type method struct {
	f             reflect.Value
	inParamTypes  []common.KindID
	outParamTypes []common.KindID
}

func (m *method) call(argv []interface{}) []interface{} {
	args := make([]reflect.Value, len(argv))
	for i, arg := range argv {
		args[i] = reflect.ValueOf(arg)
	}
	rvs := m.f.Call(args)
	rs := make([]interface{}, len(rvs))
	for i, rv := range rvs {
		rs[i] = rv.Interface()
	}
	return rs
}
