package common

import (
	"log"
	"reflect"
)

// GetServiceName srv要求指针或接口类型
func GetServiceName(srv interface{}) string {
	st := reflect.TypeOf(srv)
	if st.Kind() != reflect.Interface && st.Kind() != reflect.Ptr {
		log.Fatalln("common GetServiceName: service must pointer or interface")
	}

	return st.Elem().Name()
}
