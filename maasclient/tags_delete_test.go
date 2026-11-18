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
	"net/http"
	"testing"
)

func TestTagsDelete(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/2.0/tags/maas-lxd-hcp-op-abc/" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Tags().Delete(context.Background(), "maas-lxd-hcp-op-abc"); err != nil {
		t.Fatalf("Tags.Delete: %v", err)
	}
}

func TestTagsDeleteNotFoundIsDone(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	if err := c.Tags().Delete(context.Background(), "already-gone"); err != nil {
		t.Fatalf("Tags.Delete 404: %v", err)
	}
}
