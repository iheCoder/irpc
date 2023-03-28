package common

import "reflect"

type MethodID uint8
type SrvID uint16
type KindID uint32

const (
	Bool KindID = iota
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Float32
	Float64
	Map
	Slice
	String
)

const InvalidKindID KindID = 1<<32 - 1

var KindMapKindID = map[reflect.Kind]KindID{
	reflect.Bool:    Bool,
	reflect.Int:     Int,
	reflect.Int8:    Int8,
	reflect.Int16:   Int16,
	reflect.Int32:   Int32,
	reflect.Int64:   Int64,
	reflect.Uint:    Uint,
	reflect.Uint8:   Uint8,
	reflect.Uint16:  Uint16,
	reflect.Uint32:  Uint32,
	reflect.Uint64:  Uint64,
	reflect.Float32: Float32,
	reflect.Float64: Float64,
	reflect.Map:     Map,
	reflect.Slice:   Slice,
	reflect.String:  String,
}

const ModelStartKindID = 16

type Request struct {
	Header ReqHeader `json:"header"`
	// json
	Body []byte `json:"body"`
}

type ReqHeader struct {
	SID SrvID
	MID MethodID
}

type Response struct {
	Body []byte `json:"body"`
}

type Models struct {
	ModelMap map[KindID]reflect.Type
}
