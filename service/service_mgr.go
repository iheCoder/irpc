package service

import (
	"errors"
	common2 "learn/irpc/common"
	config2 "learn/irpc/config"
	"log"
	"reflect"
	"sync"
)

type serviceConfigInfo struct {
	id      common2.SrvID
	Methods map[string]common2.MethodID
}

var (
	ErrUnsupportedSliceMapElemType = errors.New("service_mgr: unsupported slice or map elem type")
	ErrUnsupportedType             = errors.New("service_mgr: unsupported type")
	ErrNotExistSrv                 = errors.New("service_mgr: not exist srv")
	ErrNotExistMethod              = errors.New("service_mgr: not exist method")
)

type Mgr struct {
	// 配置文件中srvId与服务名的对应关系。也就是可能导致idSrvName存在的服务，而services不存在
	idSrvName map[string]*serviceConfigInfo
	// 注册的服务名和服务信息
	services map[common2.SrvID]*service
	mu       *sync.Mutex
	// 注册的所有models
	models *common2.Models
	// 已经注册的model名称，假定不存在重名的model
	registeredModels map[string]common2.KindID
	// TODO: 生成储存Kid的文件
	// 全局记录model id
	kid common2.KindID
}

func NewServiceMgr(configPath string) *Mgr {
	mgr := &Mgr{
		idSrvName:        make(map[string]*serviceConfigInfo),
		services:         make(map[common2.SrvID]*service),
		mu:               &sync.Mutex{},
		models:           &common2.Models{ModelMap: make(map[common2.KindID]reflect.Type)},
		registeredModels: make(map[string]common2.KindID),
		kid:              common2.ModelStartKindID,
	}

	sc, err := config2.ParseToServicesConfig(configPath)
	if err != nil {
		log.Fatalf("Mgr NewServiceMgr: parse config file failed %s", err)
	}
	mgr.initFromConfig(sc)

	return mgr
}

func (m *Mgr) initFromConfig(sc *config2.ServicesConfig) {
	for _, s := range sc.Services {
		m.idSrvName[s.Name] = convertServiceConfigToConfigInfo(s)
	}
}

func convertServiceConfigToConfigInfo(sc *config2.ServiceConfig) *serviceConfigInfo {
	return &serviceConfigInfo{
		id:      sc.ID,
		Methods: sc.Methods,
	}
}

// RegisterServices 为mgr注册方法，应该在client以及server都调用该方法，并且保持server、client注册服务一致
func (m *Mgr) RegisterServices(srvs ...interface{}) error {
	for _, srv := range srvs {
		err := m.Register(srv)
		if err != nil {
			return err
		}
	}

	return nil
}

// Register 不支持并发注册。srv要求指针或接口类型。注册方法出入参数值确保必须大写
func (m *Mgr) Register(srv interface{}) error {
	// 获取srv名称
	// TypeOf也无法获取服务名，那该怎么获取呢？
	srvName := common2.GetServiceName(srv)

	// 检查是否配置过该服务，并获取srvId
	sci, configured := m.idSrvName[srvName]
	if !configured {
		log.Fatalf("Mgr Register: unconfiged srvName %s", srvName)
	}

	// 检查serviceName是否已经存在
	if _, exists := m.services[sci.id]; exists {
		log.Fatalf("Mgr Register: service name %s exists", srvName)
	}

	// 获取service方法，并注册service方法入参、出参
	ms, err := m.registerMethods(srv, sci.Methods)
	if err != nil {
		return err
	}
	m.services[sci.id] = &service{methods: ms}
	return nil
}

func (m *Mgr) registerMethods(srv interface{}, minfo map[string]common2.MethodID) (map[common2.MethodID]*method, error) {
	st := reflect.TypeOf(srv)
	sv := reflect.ValueOf(srv)
	num := st.NumMethod()
	ms := make(map[common2.MethodID]*method, num)
	for i := 0; i < num; i++ {
		sm := sv.Method(i)
		mt := st.Method(i)
		mn := mt.Name
		inTypes, outTypes, err := m.registerMethodModels(mt.Type)
		if err != nil {
			return nil, err
		}
		ms[minfo[mn]] = &method{
			f:             sm,
			inParamTypes:  inTypes,
			outParamTypes: outTypes,
		}
	}

	return ms, nil
}

func (m *Mgr) registerMethodModels(f reflect.Type) ([]common2.KindID, []common2.KindID, error) {
	// 注册入参
	numIn := f.NumIn()
	inKids := make([]common2.KindID, 0, numIn)
	// start from 1, for 0 is func receiver
	for i := 1; i < numIn; i++ {
		param := f.In(i)
		paramName := param.Name()
		kids, err := m.getKindID(param, paramName)
		if err != nil {
			return nil, nil, err
		}
		kid := kids[0]
		m.models.ModelMap[kid] = param
		inKids = append(inKids, kids...)
	}

	// 注册出参
	numOut := f.NumOut()
	outKids := make([]common2.KindID, 0, numOut)
	for i := 0; i < numOut; i++ {
		param := f.Out(i)
		paramName := param.Name()
		kids, err := m.getKindID(param, paramName)
		if err != nil {
			return nil, nil, err
		}
		kid := kids[0]
		m.models.ModelMap[kid] = param
		outKids = append(outKids, kids...)
	}

	return inKids, outKids, nil
}

func (m *Mgr) getKindID(rt reflect.Type, paramName string) ([]common2.KindID, error) {
	// 预先定义基本类型，直接返回kid
	if kid, ok := common2.KindMapKindID[rt.Kind()]; ok {
		// 若是slice、map类型，还要检查元素类型是否注册过
		if kid == common2.Slice {
			ekid, err := m.getElemKids(rt)
			if err != nil {
				return nil, err
			}

			return []common2.KindID{kid, ekid}, nil
		}

		if kid == common2.Map {
			keyType := rt.Key()
			kkid, ok := common2.KindMapKindID[keyType.Kind()]
			if !ok {
				return nil, ErrUnsupportedSliceMapElemType
			}

			ekid, err := m.getElemKids(rt)
			if err != nil {
				return nil, err
			}

			return []common2.KindID{kid, kkid, ekid}, nil
		}

		return []common2.KindID{kid}, nil
	}

	// 结构体类型，若已经存在，返回已经存在kid，否则采用新kid，并递增
	if rt.Kind() == reflect.Struct {
		if kid, exists := m.registeredModels[paramName]; exists {
			return []common2.KindID{kid}, nil
		}
		kid := m.kid
		m.registeredModels[paramName] = kid
		m.kid++
		return []common2.KindID{kid}, nil
	}

	// 其他类型返回无效类型
	return nil, ErrUnsupportedType
}

// 必须要是slice、map类型
func (m *Mgr) getElemKids(rt reflect.Type) (common2.KindID, error) {
	elemType := rt.Elem()
	if ekid, ok := common2.KindMapKindID[elemType.Kind()]; ok {
		return ekid, nil
	}

	if elemType.Kind() == reflect.Struct {
		ekids, err := m.getKindID(elemType, elemType.Name())
		if err != nil {
			return common2.InvalidKindID, err
		}
		ekid := ekids[0]
		return ekid, nil
	}

	return common2.InvalidKindID, ErrUnsupportedType
}

func (m *Mgr) Invoke(srvID common2.SrvID, mID common2.MethodID, args []interface{}) []interface{} {
	srv, exists := m.services[srvID]
	if !exists {
		log.Fatalf("Mgr Invoke: service %d not exists", srvID)
	}

	return srv.call(mID, args)
}

func (m *Mgr) GetSrvMethodID(srvName, methodName string) (common2.SrvID, common2.MethodID, error) {
	cfg, ok := m.idSrvName[srvName]
	if !ok {
		return 0, 0, ErrNotExistSrv
	}

	mid, ok := cfg.Methods[methodName]
	if !ok {
		return 0, 0, ErrNotExistMethod
	}

	srvID := cfg.id
	return srvID, mid, nil
}

func (m *Mgr) GetKindIDsByMethod(sid common2.SrvID, mid common2.MethodID) ([]common2.KindID, []common2.KindID) {
	f := m.services[sid].methods[mid]
	return f.inParamTypes, f.outParamTypes
}

func (m *Mgr) GetModels() *common2.Models {
	return m.models
}

// 	case reflect.Bool:
//		return common.Bool
//	case reflect.Int:
//		return common.Int
//	case reflect.Int8:
//		return common.Int8
//	case reflect.Int16:
//		return common.Int16
//	case reflect.Int32:
//		return common.Int32
//	case reflect.Int64:
//		return common.Int64
//	case reflect.Uint:
//		return common.Uint
//	case reflect.Uint8:
//		return common.Uint8
//	case reflect.Uint16:
//		return common.Uint16
//	case reflect.Uint32:
//		return common.Uint32
//	case reflect.Uint64:
//		return common.Uint64
//	case reflect.Float32:
//		return common.Float32
//	case reflect.Float64:
//		return common.Float64
//	case reflect.Map:
//		return common.Map
//	case reflect.Slice:
//		return common.Slice
//	case reflect.String:
//		return common.String
