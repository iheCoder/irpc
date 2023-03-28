package service

import (
	"testing"
)

var (
	demoServiceName = "DemoService"
	addMethodName   = "Add"
)

type DemoInterface interface {
	Add(x, y int) int
}

type DemoService struct {
}

type DemoService1 struct {
}

func (s DemoService1) Meaningless(x int64, name string, param map[int64]AddParam) ([]float32, bool) {
	return []float32{1.0}, true
}

//func (s *DemoService) Add(x, y int) int {
//	return x + y
//}

type AddParam struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type AddResult struct {
	r int
}

func (s *DemoService) AddWithStruct(p AddParam) AddResult {
	return AddResult{r: p.X + p.Y}
}

func (s *DemoService) Meaningless(x int64, name string, param map[int64]AddParam) ([]float32, bool) {
	return []float32{1.0}, true
}

func TestServiceCall(t *testing.T) {
	//ds := &DemoService{}
	//mv := reflect.ValueOf(ds.Add)
	//s := &service{
	//	methods: map[string]reflect.Value{
	//		addMethodName: mv,
	//	},
	//}
	//
	//// ... 可选参数会导致认为[]int也是interface{}
	//rs := s.call(addMethodName, []interface{}{1, 2})
	//if len(rs) != 1 || rs[0] != 3 {
	//	t.Fatal("wrong result")
	//}
}
