package config

import (
	"testing"
)

func TestParseToServicesConfig(t *testing.T) {
	path := "./yaml_test.yml"
	config, err := ParseToServicesConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(config.Services) == 1 {
		srv := config.Services[0]
		if srv.ID != 1 || srv.Name != "DemoService" {
			t.Fatal("wrong srv")
		}
		ms := srv.Methods
		if len(ms) == 1 {
			if ms["Add"] != 1 {
				t.Fatal("wrong method")
			}
		} else {
			t.Fatal("wrong methods len")
		}
	} else {
		t.Fatal("wrong service len")
	}
}
