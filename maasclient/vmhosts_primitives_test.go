/*
Copyright 2021 Spectro Cloud

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package maasclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) ClientSetInterface {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewAuthenticatedClientSet(server.URL, "consumer:token:secret")
}

func TestVMHostRefresh(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/2.0/vm-hosts/7/" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "op=refresh") {
			t.Errorf("body missing op=refresh: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7,"name":"host-7","type":"lxd","zone":{"id":1,"name":"az1"},"pool":{"id":2,"name":"pool"}}`))
	})

	host, err := c.VMHosts().VMHost("7").Refresh(context.Background())
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if host.SystemID() != "7" || host.Name() != "host-7" {
		t.Fatalf("host = %s/%s", host.SystemID(), host.Name())
	}
	if host.Zone() == nil || host.Zone().Name() != "az1" {
		t.Fatalf("zone = %#v", host.Zone())
	}
}

func TestVMHostMachinesListPopulated(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/2.0/vm-hosts/7/machines/" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"system_id":"vm1","hostname":"vm-1","power_type":"lxd","status_name":"Deployed","power_state":"on","tags":["maas-lxd-wlc-cp"],"parent":{"system_id":"7"}}]`))
	})

	machines, err := c.VMHosts().VMHost("7").Machines().ListPopulated(context.Background())
	if err != nil {
		t.Fatalf("ListPopulated: %v", err)
	}
	if len(machines) != 1 {
		t.Fatalf("machines = %d, want 1", len(machines))
	}
	m := machines[0]
	if m.SystemID() != "vm1" || m.Hostname() != "vm-1" || m.State() != "Deployed" {
		t.Fatalf("machine = %+v", m)
	}
	if m.Parent() != "7" {
		t.Fatalf("parent = %q, want 7", m.Parent())
	}
	if len(m.Tags()) != 1 || m.Tags()[0] != "maas-lxd-wlc-cp" {
		t.Fatalf("tags = %v", m.Tags())
	}
}

func TestMachineMarkBroken(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/2.0/machines/abc/" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "op=mark_broken") || !strings.Contains(string(body), "comment=dead+host") {
			t.Errorf("body = %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"system_id":"abc","hostname":"host","status_name":"Broken","power_type":"ipmi","comment":"dead host"}`))
	})

	m, err := c.Machines().Machine("abc").MarkBroken(context.Background(), "dead host")
	if err != nil {
		t.Fatalf("MarkBroken: %v", err)
	}
	if m.SystemID() != "abc" || m.State() != "Broken" {
		t.Fatalf("machine = %+v", m)
	}
}

func TestVMHostMachinesListPopulatedFallback(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/2.0/vm-hosts/7/machines/":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Unknown API endpoint"}`))
		case r.URL.Path == "/api/2.0/vm-hosts/7/":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":7,"name":"host-7","type":"lxd","host":{"system_id":"host-7"}}`))
		case r.URL.Path == "/api/2.0/machines/":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{"system_id":"vm1","hostname":"vm-1","power_type":"lxd","status_name":"Deployed","tags":["maas-lxd-wlc-cp"],"parent":{"system_id":"host-7"}},
				{"system_id":"vm2","hostname":"vm-2","power_type":"lxd","status_name":"Deployed","parent":{"system_id":"other-host"}}
			]`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	machines, err := c.VMHosts().VMHost("7").Machines().ListPopulated(context.Background())
	if err != nil {
		t.Fatalf("ListPopulated fallback: %v", err)
	}
	if len(machines) != 1 {
		t.Fatalf("machines = %d, want 1", len(machines))
	}
	m := machines[0]
	if m.SystemID() != "vm1" || m.Parent() != "host-7" || m.Hostname() != "vm-1" {
		t.Fatalf("machine = %+v", m)
	}
}
