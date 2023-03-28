package common

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"math"
	"reflect"
	"strconv"
)

var (
	ErrNotRegisteredStruct   = errors.New("not register struct")
	ErrNotMatchedStruct      = errors.New("not matched struct")
	ErrNotMatchedBody        = errors.New("not matched body")
	ErrUnsupportedMapKeyType = errors.New("unsupported map key type")
)

const (
	SliceKidsLen = 2
	MapKidsLen   = 3
)

type Parser struct {
	models *Models
}

func NewParser(models *Models) *Parser {
	return &Parser{models: models}
}

func (p *Parser) ParseToStruct(js []byte, kid KindID) (interface{}, error) {
	// 检查是否是已经注册的类型
	rt, ok := p.models.ModelMap[kid]
	if !ok {
		return nil, ErrNotRegisteredStruct
	}

	// 解析json为map[string]interface{}
	m := make(map[string]interface{})
	err := json.Unmarshal(js, &m)
	if err != nil {
		return nil, err
	}

	// 遍历map，一个个赋值到创建的结构成员中
	op := reflect.New(rt)
	o := op.Elem()
	for k, v := range m {
		// field一直是nil的原因在于字符串中的name是小写，可是结构中的却是大写的
		field := o.FieldByName(k)
		err = p.assign(field, v)
		if err != nil {
			return nil, err
		}
	}

	// 为什么返回含slice结构的elem的Interface是空呢？
	// 即使我更新了slice value还是如此呢?
	// set value 之后slice确实有值啦，可是结构指针interface却还是有问题。真不明白为什么同样代码含有slice的rv指针就不行，但是不含的就行
	// 我的猜测是因为slice还是指针
	return o.Interface(), nil
}

func (p *Parser) assign(value reflect.Value, input interface{}) error {
	iv := reflect.ValueOf(input)
	if value.CanSet() {
		switch value.Kind() {
		case reflect.Bool:
			if iv.Kind() != reflect.Bool {
				return ErrNotMatchedStruct
			}
			value.SetBool(iv.Bool())

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if iv.Kind() != reflect.Float64 {
				return ErrNotMatchedStruct
			}
			value.SetInt(int64(iv.Float()))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if iv.Kind() != reflect.Float64 {
				return ErrNotMatchedStruct
			}
			value.SetUint(uint64(iv.Float()))

		case reflect.Float64, reflect.Float32:
			if iv.Kind() != reflect.Float64 {
				return ErrNotMatchedStruct
			}
			value.SetFloat(iv.Float())

		case reflect.Complex64, reflect.Complex128:
			if iv.Kind() != reflect.Complex64 && iv.Kind() != reflect.Complex128 {
				return ErrNotMatchedStruct
			}
			value.SetComplex(iv.Complex())

		case reflect.String:
			value.SetString(iv.String())

		case reflect.Slice:
			slice, ok := input.([]interface{})
			if !ok {
				return ErrNotMatchedStruct
			}
			// slice类型.Elem()就是里面值的类型
			//objType := value.Type().Elem()
			// makeSlice需要的是slice的类型
			s := reflect.MakeSlice(value.Type(), len(slice), len(slice))
			for k, v := range slice {
				o := s.Index(k)
				//op := reflect.New(objType)
				//o := op.Elem()
				// 不需要reflect.Value的指针也能set值
				if err := p.assign(o, v); err != nil {
					return err
				}
			}
			// value只是临时变量，改这个是没用的，只能set改变里面实际值的指针
			value.Set(s)

		case reflect.Map:
			m, ok := input.(map[string]interface{})
			if !ok {
				return ErrNotMatchedStruct
			}

			mt := value.Type()
			mkt := mt.Key()
			mvt := mt.Elem()
			rvm := reflect.MakeMapWithSize(mt, len(m))
			for k, v := range m {
				krvp := reflect.New(mkt)
				krv := krvp.Elem()
				if err := p.assignMapKey(krv, k); err != nil {
					return err
				}

				vrvp := reflect.New(mvt)
				vrv := vrvp.Elem()
				if err := p.assign(vrv, v); err != nil {
					return err
				}
				rvm.SetMapIndex(krv, vrv)
			}
			value.Set(rvm)

		case reflect.Struct:
			m, ok := input.(map[string]interface{})
			if !ok {
				return ErrNotMatchedStruct
			}
			for k, v := range m {
				field := value.FieldByName(k)
				if field.IsValid() == false || field.CanSet() == false {
					return ErrNotMatchedStruct
				}
				if err := p.assign(field, v); err != nil {
					return err
				}
			}

		default:
			return ErrNotMatchedStruct
		}
	} else {
		return ErrNotMatchedStruct
	}

	return nil
}

func (p *Parser) assignMapKey(value reflect.Value, input string) error {
	if value.CanSet() {
		switch value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			parseInt, err := strconv.ParseInt(input, 10, 64)
			if err != nil {
				return err
			}
			value.SetInt(parseInt)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			parseUint, err := strconv.ParseUint(input, 10, 64)
			if err != nil {
				return err
			}
			value.SetUint(parseUint)

		case reflect.Float64, reflect.Float32:
			parseFloat, err := strconv.ParseFloat(input, 64)
			if err != nil {
				return err
			}
			value.SetFloat(parseFloat)

		case reflect.String:
			value.SetString(input)

		default:
			return ErrNotMatchedStruct
		}
	} else {
		return ErrNotMatchedStruct
	}

	return nil
}

// kids还剩的情况就忽略吧
func (p *Parser) ParseBody(body []byte, kids []KindID) ([]interface{}, error) {
	if len(kids) == 0 {
		return nil, nil
	}

	var index int
	var items []byte
	var item byte
	r := make([]interface{}, 0, len(kids))
	var kid KindID
	for i := 0; i < len(kids); {
		kid = kids[i]
		switch kid {
		case Bool:
			item = body[index]
			r = append(r, p.parseBool(item))
			index++

		case Int8:
			item = body[index]
			r = append(r, p.parseInt8(item))
			index++

		case Uint8:
			item = body[index]
			r = append(r, p.parseUint8(item))
			index++

		case Int16:
			items = body[index : index+2]
			r = append(r, p.parseInt16(items))
			index += 2

		case Uint16:
			items = body[index : index+2]
			r = append(r, p.parseUint16(items))
			index += 2

		case Int32:
			items = body[index : index+4]
			r = append(r, p.parseInt32(items))
			index += 4

		case Uint32:
			items = body[index : index+4]
			r = append(r, p.parseUint32(items))
			index += 4

		case Float32:
			items = body[index : index+4]
			r = append(r, p.parseFloat32(items))
			index += 4

		case Int64:
			items = body[index : index+8]
			r = append(r, p.parseInt64(items))
			index += 8

		case Uint64:
			items = body[index : index+8]
			r = append(r, p.parseUint64(items))
			index += 8

		case Float64:
			items = body[index : index+8]
			r = append(r, p.parseFloat64(items))
			index += 8

		case Int:
			// 解析int，全部定为int64
			items = body[index : index+8]
			r = append(r, p.parseInt(items))
			index += 8

		case String:
			// 确保当前byte为分隔符
			item = body[index]
			if item != StringSep {
				log.Printf("common parser: find string start with %b", item)
				return nil, ErrNotMatchedBody
			}

			// 找到后一个分隔符位置，并返回index
			// 假设[]byte要是特别大，怎么办呢？
			s, steps := p.parseString(body[index:])
			r = append(r, s)

			index = index + steps

		case Slice:
			// slice前8个byte指示slice的长度
			lb := body[index : index+8]
			ul := binary.BigEndian.Uint64(lb)
			l := int(ul)
			index += 8

			if l == 0 {
				r = append(r, nil)
				i += SliceKidsLen
				index++
				continue
			}

			i++
			kid = kids[i]
			span, err := FindEndForPeerSepInByteArray(body[index:], 0, SliceLeftSep, SliceRightSep)
			if err != nil {
				log.Printf("common parser: find slice %s end fialed for %s", body[index:], err)
				return nil, ErrNotMatchedBody
			}
			s := p.parseSlice(kid, body[index+1:index+span], l)
			r = append(r, s)

			index = index + span + 1

		case Map:
			// slice前8个byte指示slice的长度
			lb := body[index : index+8]
			ul := binary.BigEndian.Uint64(lb)
			l := int(ul)
			index += 8

			if l == 0 {
				r = append(r, nil)
				i += MapKidsLen
				index++
				continue
			}

			// 获取map k、v kindID
			i++
			// TODO: 如果超出就panic吧！
			keyKid := kids[i]

			i++
			elemKid := kids[i]

			span, err := FindEndForPeerSepInByteArray(body[index:], 0, JsonLeftSep, JsonRightSep)
			if err != nil {
				log.Printf("common parser: find map %s end fialed for %s", body[index:], err)
				return nil, ErrNotMatchedStruct
			}
			m, err := p.parseMap(keyKid, elemKid, body[index+1:index+span], l)
			if err != nil {
				return nil, err
			}

			r = append(r, m)
			index = index + span + 1

		// default 认为是结构体
		default:
			rs, steps, err := p.parseStruct(body[index:], kid)
			if err != nil {
				return nil, err
			}
			r = append(r, rs)
			index = index + steps
		}
		i++
	}

	return r, nil
}

func (p *Parser) parseStruct(body []byte, kid KindID) (interface{}, int, error) {
	item := body[0]
	if item != JsonLeftSep {
		log.Printf("common parser: find struct start with %b", item)
		return nil, 0, ErrNotMatchedBody
	}

	span, _ := FindEndForPeerSepInByteArray(body, 0, JsonLeftSep, JsonRightSep)
	items := body[:span+1]
	rs, err := p.ParseToStruct(items, kid)
	if err != nil {
		return nil, 0, err
	}

	return rs, span + 1, nil
}

func (p *Parser) parseBool(item byte) bool {
	if item != byte(0) {
		return true
	} else {
		return false
	}
}

func (p *Parser) parseInt8(item byte) int8 {
	return int8(item)
}

func (p *Parser) parseUint8(item byte) uint8 {
	return item
}

func (p *Parser) parseInt16(items []byte) int16 {
	return int16(binary.BigEndian.Uint16(items))
}

func (p *Parser) parseUint16(items []byte) uint16 {
	return binary.BigEndian.Uint16(items)
}

func (p *Parser) parseInt32(items []byte) int32 {
	return int32(binary.BigEndian.Uint32(items))
}

func (p *Parser) parseUint32(items []byte) uint32 {
	return binary.BigEndian.Uint32(items)
}

func (p *Parser) parseFloat32(items []byte) float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(items))
}

func (p *Parser) parseInt64(items []byte) int64 {
	return int64(binary.BigEndian.Uint64(items))
}

func (p *Parser) parseUint64(items []byte) uint64 {
	return binary.BigEndian.Uint64(items)
}

func (p *Parser) parseFloat64(items []byte) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(items))
}

func (p *Parser) parseInt(items []byte) int {
	return int(binary.BigEndian.Uint64(items))
}

func (p *Parser) parseString(body []byte) (string, int) {
	span := FindEndForSepInByteArray(body, 0, StringSep)
	return string(body[1:span]), span + 1
}

// body 是去除map括号的map值
func (p *Parser) parseMap(kkid, vkid KindID, body []byte, l int) (interface{}, error) {
	if l == 0 {
		return nil, nil
	}

	// 获取key reflect.Type
	keyType, err := p.getBasicType(kkid)
	if err != nil {
		return nil, err
	}

	// 构建key、value kids序列
	mkids := make([]KindID, l*2)
	for j := 0; j < len(mkids); {
		mkids[j] = kkid
		j++
		mkids[j] = vkid
		j++
	}

	// 获取key、value序列
	mapItems, err := p.ParseBody(body, mkids)
	if err != nil {
		return nil, err
	}

	// 组装map
	m := reflect.MakeMapWithSize(reflect.MapOf(keyType, p.models.ModelMap[vkid]), l)
	for i := 0; i < 2*l; i += 2 {
		m.SetMapIndex(reflect.ValueOf(mapItems[i]), reflect.ValueOf(mapItems[i+1]))
	}

	// 返回map
	return m.Interface(), nil
}

func (p *Parser) getBasicType(kid KindID) (reflect.Type, error) {
	switch kid {
	case Bool:
		return reflect.TypeOf(false), nil
	case Int8:
		return reflect.TypeOf(int8(0)), nil
	case Uint8:
		return reflect.TypeOf(uint8(0)), nil
	case Int16:
		return reflect.TypeOf(int16(0)), nil
	case Uint16:
		return reflect.TypeOf(uint16(0)), nil
	case Int32:
		return reflect.TypeOf(int32(0)), nil
	case Uint32:
		return reflect.TypeOf(uint32(0)), nil
	case Float32:
		return reflect.TypeOf(float32(0)), nil
	case Int64:
		return reflect.TypeOf(int64(0)), nil
	case Uint64:
		return reflect.TypeOf(uint64(0)), nil
	case Float64:
		return reflect.TypeOf(float64(0)), nil
	case Int:
		return reflect.TypeOf(0), nil
	case String:
		return reflect.TypeOf(""), nil
	default:
		return nil, ErrUnsupportedMapKeyType
	}
}

func (p *Parser) parseSlice(ekid KindID, body []byte, l int) interface{} {
	if l == 0 {
		return nil
	}

	switch ekid {
	case Bool:
		r := make([]bool, l)
		for i := 0; i < l; i++ {
			r[i] = p.parseBool(body[i])
		}
		return r

	case Int8:
		r := make([]int8, l)
		for i := 0; i < l; i++ {
			r[i] = p.parseInt8(body[i])
		}
		return r

	case Uint8:
		r := make([]uint8, l)
		for i := 0; i < l; i++ {
			r[i] = p.parseUint8(body[i])
		}
		return r

	case Int16:
		r := make([]int16, l)
		var j int
		for i := 0; i < l; i++ {
			r[i] = p.parseInt16(body[j : j+2])
			j += 2
		}
		return r

	case Uint16:
		r := make([]uint16, l)
		var j int
		for i := 0; i < l; i++ {
			r[i] = p.parseUint16(body[j : j+2])
			j += 2
		}
		return r

	case Int32:
		r := make([]int32, l)
		var j int
		for i := 0; i < l; i++ {
			r[i] = p.parseInt32(body[j : j+4])
			j += 4
		}
		return r

	case Uint32:
		r := make([]uint32, l)
		var j int
		for i := 0; i < l; i++ {
			r[i] = p.parseUint32(body[j : j+4])
			j += 4
		}
		return r

	case Float32:
		r := make([]float32, l)
		var j int
		for i := 0; i < l; i++ {
			r[i] = p.parseFloat32(body[j : j+4])
			j += 4
		}
		return r

	case Int64:
		r := make([]int64, l)
		var j int
		for i := 0; i < l; i++ {
			r[i] = p.parseInt64(body[j : j+8])
			j += 8
		}
		return r

	case Uint64:
		r := make([]uint64, l)
		var j int
		for i := 0; i < l; i++ {
			r[i] = p.parseUint64(body[j : j+8])
			j += 8
		}
		return r

	case Float64:
		r := make([]float64, l)
		var j int
		for i := 0; i < l; i++ {
			r[i] = p.parseFloat64(body[j : j+8])
			j += 8
		}
		return r

	case Int:
		// 解析int，全部定为int64
		r := make([]int, l)
		var j int
		for i := 0; i < l; i++ {
			r[i] = p.parseInt(body[j : j+8])
			j += 8
		}
		return r

	case String:
		r := make([]string, l)
		var j int
		var steps int
		for i := 0; i < l; i++ {
			r[i], steps = p.parseString(body[j:])
			j += steps
		}
		return r

	default:
		rv := p.models.ModelMap[ekid]
		r := reflect.MakeSlice(reflect.SliceOf(rv), l, l)
		var j int
		for i := 0; i < l; i++ {
			s, steps, err := p.parseStruct(body[j:], ekid)
			if err != nil {
				return nil
			}
			j += steps
			r.Index(i).Set(reflect.ValueOf(s))
		}
		return r.Interface()
	}
}

func (p *Parser) EncodeBody(kids []KindID, params ...interface{}) ([]byte, error) {
	var kid KindID
	var index int
	r := make([]byte, 0)
	var item byte
	var items []byte
	var err error
	for i := 0; i < len(kids); index++ {
		kid = kids[i]
		switch kid {
		case Bool:
			if params[index].(bool) == false {
				item = byte(0)
			} else {
				item = byte(1)
			}
			r = append(r, item)

		case Int8:
			item = byte(params[index].(int8))
			r = append(r, item)

		case Uint8:
			item = params[index].(uint8)
			r = append(r, item)

		case Int16:
			items = make([]byte, 2)
			binary.BigEndian.PutUint16(items, uint16(params[index].(int16)))
			r = append(r, items...)

		case Uint16:
			items = make([]byte, 2)
			binary.BigEndian.PutUint16(items, params[index].(uint16))
			r = append(r, items...)

		case Int32:
			items = make([]byte, 4)
			binary.BigEndian.PutUint32(items, uint32(params[index].(int32)))
			r = append(r, items...)

		case Uint32:
			items = make([]byte, 4)
			binary.BigEndian.PutUint32(items, params[index].(uint32))
			r = append(r, items...)

		case Float32:
			items = make([]byte, 4)
			binary.BigEndian.PutUint32(items, math.Float32bits(params[index].(float32)))
			r = append(r, items...)

		case Int64:
			items = make([]byte, 8)
			binary.BigEndian.PutUint64(items, uint64(params[index].(int64)))
			r = append(r, items...)

		case Uint64:
			items = make([]byte, 8)
			binary.BigEndian.PutUint64(items, params[index].(uint64))
			r = append(r, items...)

		case Float64:
			items = make([]byte, 8)
			binary.BigEndian.PutUint64(items, math.Float64bits(params[index].(float64)))
			r = append(r, items...)

		case Int:
			// 解析int，全部定为int64
			items = make([]byte, 8)
			binary.BigEndian.PutUint64(items, uint64(params[index].(int)))
			r = append(r, items...)

		case String:
			// 确保当前byte为分隔符
			items = p.encodeString(params[index].(string))
			r = append(r, items...)

		case Slice:
			items = make([]byte, 8)

			// 判断是否是nil
			if params[index] == nil {
				binary.BigEndian.PutUint64(items, 0)
				i += SliceKidsLen
				r = append(r, items...)
				r = append(r, NilFlag)
				continue
			}

			// 获取长度
			sv := reflect.ValueOf(params[index])
			l := sv.Len()
			if l == 0 {
				binary.BigEndian.PutUint64(items, 0)
				i += SliceKidsLen
				r = append(r, items...)
				r = append(r, EmptyFlag)
				continue
			}

			// 放入长度
			ul := uint64(l)
			binary.BigEndian.PutUint64(items, ul)
			r = append(r, items...)

			// 递归编码
			r = append(r, SliceLeftSep)

			i++
			elemKid := kids[i]
			for j := 0; j < l; j++ {
				items, err = p.EncodeBody([]KindID{elemKid}, sv.Index(j).Interface())
				if err != nil {
					return nil, err
				}
				r = append(r, items...)
			}

			r = append(r, SliceRightSep)

		case Map:
			items = make([]byte, 8)

			// nil
			if params[index] == nil {
				binary.BigEndian.PutUint64(items, 0)
				i += MapKidsLen
				r = append(r, items...)
				r = append(r, NilFlag)
				continue
			}

			// 获取长度
			sv := reflect.ValueOf(params[index])
			l := sv.Len()
			if l == 0 {
				binary.BigEndian.PutUint64(items, 0)
				i += MapKidsLen
				r = append(r, items...)
				r = append(r, EmptyFlag)
				continue
			}

			// 放入长度
			ul := uint64(l)
			binary.BigEndian.PutUint64(items, ul)
			r = append(r, items...)

			// 递归编码
			r = append(r, JsonLeftSep)
			i++
			keyKid := kids[i]

			i++
			elemKid := kids[i]

			iter := sv.MapRange()
			for iter.Next() {
				k := iter.Key().Interface()
				items, err = p.EncodeBody([]KindID{keyKid}, k)
				if err != nil {
					return nil, err
				}
				r = append(r, items...)

				v := iter.Value().Interface()
				items, err = p.EncodeBody([]KindID{elemKid}, v)
				if err != nil {
					return nil, err
				}
				r = append(r, items...)
			}

			r = append(r, JsonRightSep)

		// default 认为是结构体
		default:
			items, err = json.Marshal(params[index])
			if err != nil {
				return nil, err
			}
			r = append(r, items...)

		}
		i++
	}

	return r, nil
}

func (p *Parser) encodeString(s string) []byte {
	r := make([]byte, len(s)+2)
	r[0] = StringSep
	copy(r[1:len(r)-1], s)
	r[len(r)-1] = StringSep

	return r
}

//func (p *Parser) recurAssign(m map[string]interface{}, rvp reflect.Value) error {
//	rv := rvp.Elem()
//	for s, i := range m {
//		obj := rv.FieldByName(s)
//		if err := p.assign(obj, i); err != nil {
//			return err
//		}
//	}
//	return nil
//}
