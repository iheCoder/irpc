package common

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
)

type Simple struct {
	Name string
	ID   int64
}

//
//func TestParseSimple(t *testing.T) {
//	p := &Parser{map[KindID]reflect.Type{
//		16: reflect.TypeOf(Simple{}),
//	}}
//	s := Simple{
//		Name: "hello",
//		ID:   1,
//	}
//	bytes, err := json.Marshal(s)
//	if err != nil {
//		t.Fatal(err)
//	}
//	toStruct, err := p.ParseToStruct(bytes, 16)
//	if err != nil {
//		t.Fatal(err)
//	}
//	_, ok := toStruct.(Simple)
//	if !ok {
//		t.Fatal("parse wrong")
//	}
//}
//
//type WithSlice struct {
//	//Name string
//	//ID   int64
//	S []int
//}
//
//func TestParseWithSlice(t *testing.T) {
//	p := &Parser{map[KindID]reflect.Type{
//		16: reflect.TypeOf(WithSlice{}),
//	}}
//	s := WithSlice{
//		//Name: "hello",
//		//ID:   1,
//		S: []int{1, 2, 3},
//	}
//	bytes, err := json.Marshal(s)
//	if err != nil {
//		t.Fatal(err)
//	}
//	toStruct, err := p.ParseToStruct(bytes, 16)
//	if err != nil {
//		t.Fatal(err)
//	}
//	_, ok := toStruct.(WithSlice)
//	if !ok {
//		t.Fatal("parse wrong")
//	}
//}
//
//type WithMap struct {
//	M map[int][]int
//}
//
//func TestParseWithMap(t *testing.T) {
//	p := &Parser{map[KindID]reflect.Type{
//		16: reflect.TypeOf(WithMap{}),
//	}}
//	// json marshall 根本不支持map key为struct, key, pointer的
//	s := WithMap{
//		M: map[int][]int{
//			1: {1, 2, 3},
//		},
//	}
//	bytes, err := json.Marshal(s)
//	if err != nil {
//		t.Fatal(err)
//	}
//	toStruct, err := p.ParseToStruct(bytes, 16)
//	if err != nil {
//		t.Fatal(err)
//	}
//	_, ok := toStruct.(WithMap)
//	if !ok {
//		t.Fatal("parse wrong")
//	}
//}
//
//type WithStruct struct {
//	ID    int
//	Value float64
//	S     Simple
//}
//
//type Complex struct {
//	M           WithMap
//	S           WithSlice
//	WS          WithStruct
//	PlaceHolder string
//}
//
//func TestComplexParse(t *testing.T) {
//	p := &Parser{map[KindID]reflect.Type{
//		16: reflect.TypeOf(Complex{}),
//	}}
//	// json marshall 根本不支持map key为struct, key, pointer的
//	simple := Simple{
//		Name: "hello",
//		ID:   1,
//	}
//	m := WithMap{
//		M: map[int][]int{
//			1: {1, 2, 3},
//		},
//	}
//	s := WithSlice{
//		S: []int{1, 2, 3},
//	}
//	ws := WithStruct{
//		ID:    7,
//		Value: 9,
//		S:     simple,
//	}
//
//	c := Complex{
//		M:           m,
//		S:           s,
//		WS:          ws,
//		PlaceHolder: "place",
//	}
//	bytes, err := json.Marshal(c)
//	if err != nil {
//		t.Fatal(err)
//	}
//	toStruct, err := p.ParseToStruct(bytes, 16)
//	if err != nil {
//		t.Fatal(err)
//	}
//	cr, ok := toStruct.(Complex)
//	if !ok {
//		t.Fatal("parse wrong")
//	}
//	if cr.WS.S.ID != 1 {
//		t.Fatal("wrong value")
//	}
//}
//
//func TestRegisterFine(t *testing.T) {
//	m := registerServiceMethods(&DemoService{})
//	p := &Parser{&Models{ModelMap: m}}
//	s := AddParam{
//		X: 1,
//		Y: 1,
//	}
//	bytes, err := json.Marshal(s)
//	if err != nil {
//		t.Fatal(err)
//	}
//	toStruct, err := p.ParseToStruct(bytes, 16)
//	if err != nil {
//		t.Fatal(err)
//	}
//	_, ok := toStruct.(AddParam)
//	if !ok {
//		t.Fatal("parse wrong")
//	}
//}
//
//type DemoService struct {
//}
//
//type AddParam struct {
//	X int
//	Y int
//}
//
//type AddResult struct {
//	R int
//}
//
//func (s *DemoService) AddWithStruct(p AddParam) AddResult {
//	return AddResult{R: p.X + p.Y}
//}
//
//func registerServiceMethods(service interface{}) map[KindID]reflect.Type {
//	v := reflect.TypeOf(service)
//	mt := v.Method(0).Type
//	m := make(map[KindID]reflect.Type)
//
//	params := make([]reflect.Type, 0, mt.NumIn())
//	srvID := KindID(16)
//	for i := 1; i < mt.NumIn(); i++ {
//		param := mt.In(i)
//		params = append(params, param)
//		m[srvID] = mt.In(i)
//		srvID++
//	}
//
//	return m
//}

func TestParseBodyWithBasicTypes(t *testing.T) {
	p := &Parser{}
	kids := []KindID{
		Bool,
		Int8,
		Float64,
		String,
	}

	boolB := byte(0)

	int8B := byte(10)

	float64B := 3.14
	f64B := make([]byte, 8)
	binary.BigEndian.PutUint64(f64B, math.Float64bits(float64B))

	sB := "hello,world 你好啊"
	sBB := p.encodeString(sB)

	or := make([]byte, 1+1+8+len(sBB))
	or[0] = boolB
	or[1] = int8B
	copy(or[2:10], f64B)
	copy(or[10:], sBB)

	r, err := p.ParseBody(or, kids)
	if err != nil {
		t.Fatal(err)
	}

	if len(r) != 4 {
		t.Fatal("result len wrong")
	}

	if r[0] != false {
		t.Fatal("bool wrong")
	}

	if r[1] != int8(10) {
		t.Fatal("int8 wromg")
	}

	if r[2] != 3.14 {
		t.Fatal("float wrong")
	}

	if r[3] != sB {
		t.Fatal("string wrong")
	}
}

func TestParseBody(t *testing.T) {
	p := &Parser{
		models: &Models{ModelMap: map[KindID]reflect.Type{
			16: reflect.TypeOf(Simple{}),
		}},
	}
	kids := []KindID{
		Bool,
		Int,
		Float32,
		String,
		Slice,
		Uint64,
		Map,
		Int64,
		16,
		Slice,
		16,
		Map,
		16,
	}
	body, err := p.EncodeBody(kids, true, 10, float32(3.14), "你好啊", []uint64{1, 2, 3}, map[int64]Simple{
		1: {
			Name: "hello",
			ID:   100,
		},
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	result, err := p.ParseBody(body, kids)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 8 {
		t.Fatal("len wrong")
	}

	if result[0].(bool) != true {
		t.Fatal("bool wrong")
	}

	if result[1].(int) != 10 {
		t.Fatal("int wrong")
	}

	if result[2].(float32) != float32(3.14) {
		t.Fatal("float32 wrong")
	}

	if result[3].(string) != "你好啊" {
		t.Fatal("string wrong")
	}

	// 更大的问题是恐怕Call无法处理这种slice elem为[]interface 的interface，而实际值又是uint64
	s := result[4].([]uint64)
	if len(s) != 3 || s[0] != uint64(1) {
		t.Fatal("slice wrong")
	}

	m := result[5].(map[int64]Simple)
	if m[(int64(1))].ID != 100 {
		t.Fatal("map wrong")
	}

	if result[6] != nil {
		t.Fatal("nil slice wrong")
	}

	if result[7] != nil {
		t.Fatal("nil map wrong")
	}
}

func TestByteMeans(t *testing.T) {
	b := byte(' ')
	fmt.Println(b)
	s := "hello,world 你好啊"
	sb := []byte(s)
	if len(s) != len(sb) {
		t.Logf("mismatch")
	}
}

func TestSliceRelated(t *testing.T) {
	bs := make([]int, 3, 16)
	for i := 0; i < len(bs); i++ {
		bs[i] = i
	}
	bsi := (interface{})(bs)

	bsb, err := json.Marshal(bsi)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(bsb)

	bsiv := reflect.ValueOf(bsi)
	t.Log(bsiv.Len())
}
