package service

import (
	"testing"
)

func TestConfigInvokeService(t *testing.T) {
	//ssc := &config.ServicesConfig{Services: []*config.ServiceConfig{
	//	{
	//		ID:   1,
	//		Name: "DemoService",
	//		Methods: map[string]common.MethodID{
	//			"Add": 1,
	//		},
	//	},
	//}}
	mgr := NewServiceMgr("../config/services.yml")
	mgr.Register(&DemoService{})
	r := mgr.Invoke(1, 1, []interface{}{1, 2})
	if len(r) != 1 || r[0] != 3 {
		t.Fatal("unexpected result")
	}
}
