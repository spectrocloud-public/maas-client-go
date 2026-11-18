//go:build integration

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
	"os"
	"strings"
	"testing"
)

func integrationClient(t *testing.T) ClientSetInterface {
	t.Helper()
	endpoint := os.Getenv("MAAS_ENDPOINT")
	apiKey := os.Getenv("MAAS_API_KEY")
	if endpoint == "" || apiKey == "" {
		t.Skip("MAAS_ENDPOINT and MAAS_API_KEY are required")
	}
	return NewAuthenticatedClientSet(endpoint, apiKey)
}

// TestIntegrationVMHostReadOnly lists vm-hosts and their populated machine
// list. It makes no mutations.
func TestIntegrationVMHostReadOnly(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	hosts, err := c.VMHosts().List(ctx, nil)
	if err != nil {
		t.Fatalf("VMHosts.List: %v", err)
	}
	t.Logf("found %d vm-hosts", len(hosts))
	if len(hosts) == 0 {
		t.Skip("MAAS has no vm-hosts")
	}

	okHosts := 0
	for _, host := range hosts {
		machines, err := host.Machines().ListPopulated(ctx)
		if err != nil {
			t.Logf("host id=%s name=%s type=%s: ListPopulated error: %v",
				host.SystemID(), host.Name(), host.Type(), err)
			continue
		}
		okHosts++
		t.Logf("host id=%s name=%s type=%s hostSystemID=%s machines=%d",
			host.SystemID(), host.Name(), host.Type(), host.HostSystemID(), len(machines))
		for _, m := range machines {
			t.Logf("  machine id=%s hostname=%s state=%s parent=%s tags=%v",
				m.SystemID(), m.Hostname(), m.State(), m.Parent(), m.Tags())
		}
	}
	if okHosts == 0 {
		t.Fatal("ListPopulated failed for every vm-host")
	}
}

// TestIntegrationVMHostRefresh refreshes a specific vm-host. Set
// MAAS_TEST_VM_HOST_ID to a disposable lab host; refresh mutates host state.
func TestIntegrationVMHostRefresh(t *testing.T) {
	id := os.Getenv("MAAS_TEST_VM_HOST_ID")
	if id == "" {
		t.Skip("MAAS_TEST_VM_HOST_ID is required")
	}
	c := integrationClient(t)
	ctx := context.Background()

	host, err := c.VMHosts().VMHost(id).Refresh(ctx)
	if err != nil {
		t.Fatalf("VMHost.Refresh: %v", err)
	}
	if host.SystemID() == "" {
		t.Fatal("refreshed host has an empty system id")
	}
	t.Logf("refreshed host id=%s name=%s type=%s", host.SystemID(), host.Name(), host.Type())
}

// TestIntegrationMachineMarkBroken marks a specific machine Broken. Set
// MAAS_TEST_BROKEN_MACHINE_ID to a disposable machine; this is destructive.
func TestIntegrationMachineMarkBroken(t *testing.T) {
	id := os.Getenv("MAAS_TEST_BROKEN_MACHINE_ID")
	if id == "" {
		t.Skip("MAAS_TEST_BROKEN_MACHINE_ID is required")
	}
	c := integrationClient(t)
	ctx := context.Background()

	m, err := c.Machines().Machine(id).MarkBroken(ctx, "PCP-8053 integration test")
	if err != nil {
		t.Fatalf("Machine.MarkBroken: %v", err)
	}
	t.Logf("machine id=%s state=%s", m.SystemID(), m.State())
}

func TestIntegrationMachineMarkBrokenNotFound(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	_, err := c.Machines().Machine("pcp8053-does-not-exist").MarkBroken(ctx, "PCP-8053 integration test")
	if err == nil {
		t.Fatal("expected MarkBroken to fail for a nonexistent machine")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "404") && !strings.Contains(msg, "not found") && !strings.Contains(msg, "no machine") {
		t.Fatalf("unexpected MarkBroken error: %v", err)
	}
	t.Logf("MarkBroken nonexistent machine returned: %v", err)
}
