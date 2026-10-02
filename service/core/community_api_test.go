package core

import (
	"github.com/safing/portmaster/base/api"
	"testing"
)

func TestCommunityUpdateEndpoints(t *testing.T) {
	if err := registerAPIEndpoints(); err != nil {
		t.Fatal(err)
	}
	if _, err := api.GetEndpointByPath("updates/from-url"); err == nil {
		t.Fatal("arbitrary URL binary upgrade must not be registered")
	}
	for _, path := range []string{"updates/check", "updates/apply"} {
		endpoint, err := api.GetEndpointByPath(path)
		if err != nil {
			t.Fatal(err)
		}
		if endpoint.Write != api.PermitUser || endpoint.WriteMethod != "POST" {
			t.Fatalf("%s must require user permission and POST", path)
		}
	}
}
