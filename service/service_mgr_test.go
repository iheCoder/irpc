package service

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	common2 "learn/irpc/common"
	"reflect"
	"sync"
	"testing"
)

func TestRegisterService(t *testing.T) {
	//ds := &DemoService{}
	//sm := NewServiceMgr()
	//sm.Register((DemoInterface)(ds))
	//srvName := GetServiceName(ds)
	//r := sm.Invoke(srvName, "Add", []interface{}{1, 2})
	//fmt.Println(r)
	//rs := s.call(addMethodName, []interface{}{1, 2})
	//if len(rs) != 1 || rs[0] != 3 {
	//	t.Fatal("wrong result")
	//}
}

func TestWhatIfConcRead(t *testing.T) {
	m := make(map[string]string)
	m["hello"] = "world"
	count := 1000
	wg := &sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			_, ok := m["hello"]
			if ok {
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestGetServiceNameAuto(t *testing.T) {
	ds := DemoService{}
	st := reflect.TypeOf(ds)
	fmt.Println(st.Name())
}

// https://blog.csdn.net/weixin_43705457/article/details/109107288
func TestGetMethodNameAuto(t *testing.T) {
	ds := &DemoService{}
	v := reflect.TypeOf(ds)
	num := v.NumMethod()
	ms := make(map[string]reflect.Method, num)
	for i := 0; i < num; i++ {
		m := v.Method(i)
		//mn := runtime.FuncForPC(m.Pointer()).Name()
		mn := m.Name

		ms[mn] = m
	}

	fmt.Println(ms)
}

func TestGetMethodArgsReturnType(t *testing.T) {
	ds := &DemoService{}
	v := reflect.TypeOf(ds)
	m := v.Method(0).Type

	params := make([]reflect.Type, 0, m.NumIn())
	for i := 0; i < m.NumIn(); i++ {
		params = append(params, m.In(i))
	}

	//outs := make([]reflect.Type, 0, m.Type().NumOut())
	//for i := 0; i < m.Type().NumOut(); i++ {
	//	outs = append(outs, m.Type().Out(i))
	//}

	ap := AddParam{
		X: 1,
		Y: 1,
	}
	apj, err := json.Marshal(ap)
	if err != nil {
		t.Fatal(err)
	}

	//p := reflect.New(params[0])
	// 先测试能否根据反射数据类型解析正确结构
	// 并不能
	// // New returns a Value representing a pointer to a new zero value
	//// for the specified type. That is, the returned Value's Type is PtrTo(typ).
	// 所以得到的是指针，要先通过Elem找到结构体，而后找到interface
	p := reflect.New(params[1])
	pi := p.Elem()
	px := pi.Interface().(AddParam)
	err = json.Unmarshal(apj, &px)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWrapReflectValue(t *testing.T) {
	ap := AddParam{
		X: 1,
		Y: 1,
	}
	apv := reflect.ValueOf(ap)
	apvj, err := json.Marshal(apv)
	if err != nil {
		t.Fatal(err)
	}
	pv := &reflect.Value{}
	err = json.Unmarshal(apvj, pv)
	if err != nil {
		t.Fatal(err)
	}
	// 并不行，Value里面的属性可全是私有的啊
	// 其实还有一个方式——就是自定义reflect.Value
	v := pv.Interface().(AddParam)
	fmt.Println(v.X)
}

func TestGetMethodsName(t *testing.T) {
	//r := registerMethods(&DemoService{})
	//fmt.Println(r)
}

type Kind uint

const (
	Bool Kind = iota
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

func TestIProto(t *testing.T) {
	type Test struct {
		ID   int64
		Name string
	}
	// 编号结构体
	var TypesMap = map[Kind]reflect.Type{
		16: reflect.TypeOf(Test{}),
	}

	// 编码
	// 假设方法参数分别为int64, Test，值为 1，Test{ID:2, Name: "hello"}
	x := int64(1)
	t1 := Test{
		ID:   2,
		Name: "Hello",
	}
	bytes, err := json.Marshal(t1)
	if err != nil {
		t.Fatal(err)
	}
	r := make([]byte, 8+len(bytes))
	binary.BigEndian.PutUint64(r[:8], uint64(x))
	copy(r[8:], bytes)

	// 解析为结构体
	paramIDs := []Kind{
		5,
		16,
	}

	r1 := make([]interface{}, len(paramIDs))
	var cost int
	for i, id := range paramIDs {
		switch id {
		case Int64:
			y := int64(binary.BigEndian.Uint64(r[cost:8]))
			r1[i] = y
			cost += 8
		default:
			s := TypesMap[id]
			// 长度怎么办呢？string这种可不是定长的啊
			// 神奇的是放入[]interface{}中根本无法json解析出来
			// reflect.Type也不行，需要转换成具体类型
			end, _ := common2.FindEndForPeerSepInByteArray(r, cost, 0, 0)
			err := json.Unmarshal(r[cost:end+1], &s)
			if err != nil {
				t.Fatal(err)
			}
			r1[i] = s
		}
	}

	if r1[0] != int64(1) {
		t.Fatal("first int64 arg wrong")
	}
	tr := r1[1].(Test)
	if tr.ID != 2 || tr.Name != "hello" {
		t.Fatal("struct wrong")
	}
}

func TestFindJsonEndInByteArray(t *testing.T) {
	type Test struct {
		ID   int64
		Name string
	}

	// 编码
	// 假设方法参数分别为int64, Test，值为 1，Test{ID:2, Name: "hello"}
	x := int64(1)
	t1 := Test{
		ID:   2,
		Name: "Hello",
	}
	bytes, err := json.Marshal(t1)
	if err != nil {
		t.Fatal(err)
	}
	r := make([]byte, 8+len(bytes))
	binary.BigEndian.PutUint64(r[:8], uint64(x))
	copy(r[8:], bytes)

	y, _ := common2.FindEndForPeerSepInByteArray(r, 8, common2.JsonLeftSep, common2.JsonRightSep)
	if y != len(r)-1 {
		t.Fatal("wrong end pos")
	}

	ax := make([]byte, 8)
	binary.BigEndian.PutUint64(ax, uint64(x))
	r = append(r, ax...)
	z, _ := common2.FindEndForPeerSepInByteArray(r, 8, common2.JsonLeftSep, common2.JsonRightSep)
	if z != 8+len(bytes)-1 {
		t.Fatal("wrong end pos")
	}
}

func TestUnmarshallByInterface(t *testing.T) {
	type Test struct {
		ID   int64
		Name string
		SS   []int
		M    map[int]string
	}

	t1 := Test{
		ID:   2,
		Name: "Hello",
		SS:   []int{1, 2},
		M: map[int]string{
			1: "world",
		},
	}
	bytes, err := json.Marshal(t1)
	if err != nil {
		t.Fatal(err)
	}

	t2 := interface{}(Test{})
	err = json.Unmarshal(bytes, &t2)
	if err != nil {
		t.Fatal(err)
	}
	t2 = t2.(Test)
}

func TestPopulateStruct(t *testing.T) {
	//type TestChild struct {
	//	Name string
	//}
	//
	//type Test struct {
	//	ID int64
	//	TC TestChild
	//}
	//
	//tt := reflect.TypeOf(Test{})
	//
	//tv := reflect.New(tt)
	//
	//fieldNum := tt.NumField()
	//for i := 0; i < fieldNum; i++ {
	//
	//}
}

//func populateValue(rt reflect.Type) interface{} {
//	v := reflect.New(rt)
//	fieldNum := rt.NumField()
//	for i := 0; i < fieldNum; i++ {
//		switch rt.Field(i).Type.Kind() {
//		case reflect.Int64:
//			v.Field(i).SetInt(1)
//		case reflect.Struct:
//			rv := populateStruct(rt.Field(i).Type)
//			v.Field(i).Set(rv)
//		case reflect.String:
//			v.Field(i).SetString("hello")
//		}
//	}
//	return v.Interface()
//}
//
//func populateStruct(rt reflect.Type) reflect.Value {
//
//}

func TestRegisterMethodModels(t *testing.T) {
	// 反射获取的方法类型居然不行，这真是奇怪，我明明在其他地方是可行的啊
	// 细节问题，为什么方法的type还要typeOf呢？
	ds := &DemoService{}
	dst := reflect.TypeOf(ds)
	mt := dst.Method(1).Type
	mgr := &Mgr{
		idSrvName:        make(map[string]*serviceConfigInfo),
		services:         make(map[common2.SrvID]*service),
		mu:               &sync.Mutex{},
		models:           &common2.Models{ModelMap: make(map[common2.KindID]reflect.Type)},
		registeredModels: make(map[string]common2.KindID),
		kid:              common2.ModelStartKindID,
	}
	inKids, outKids, err := mgr.registerMethodModels(mt)
	if err != nil {
		t.Fatal(err)
	}

	if len(inKids) != 5 {
		t.Fatal("wrong in len")
	}

	if len(outKids) != 3 {
		t.Fatal("wrong out len")
	}

	if inKids[0] != common2.Int64 {
		t.Fatal("int64 wrong")
	}

	if inKids[1] != common2.String {
		t.Fatal("str wrong")
	}

	if inKids[2] != common2.Map || inKids[3] != common2.Int64 || inKids[4] != 16 {
		t.Fatal("map wrong")
	}

	if outKids[0] != common2.Slice || outKids[1] != common2.Float32 {
		t.Fatal("slice wrong")
	}

	if outKids[2] != common2.Bool {
		t.Fatal("bool wrong")
	}
}

type CallTest struct {
}

func (c *CallTest) Sum(arr []int) int {
	var sum int
	for _, a := range arr {
		sum += a
	}
	return sum
}

func TestCallWithInterfaceParam(t *testing.T) {
	ds := &CallTest{}
	dsv := reflect.ValueOf(ds)
	m := dsv.Method(0)

	inParams := []interface{}{
		1,
		2,
		3,
	}
	inParam := interface{}(inParams)
	inP := reflect.ValueOf(inParam)

	rvs := m.Call([]reflect.Value{inP})
	if rvs[0].Interface() != 6 {
		t.Fatal("wrong")
	}
}
