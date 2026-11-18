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
	"fmt"
	"testing"
	"time"
)

// TestIntegrationTagsCreateDelete creates a throwaway tag and deletes its
// definition through Tags.Delete.
func TestIntegrationTagsCreateDelete(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()
	name := fmt.Sprintf("pcp-8053-test-%d", time.Now().UnixNano())

	if err := c.Tags().Create(ctx, name); err != nil {
		t.Fatalf("Tags.Create: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Tags().Delete(context.Background(), name)
	})

	if err := c.Tags().Delete(ctx, name); err != nil {
		t.Fatalf("Tags.Delete: %v", err)
	}
	t.Logf("created and deleted tag %s", name)
}

func TestIntegrationTagsDeleteNotFound(t *testing.T) {
	c := integrationClient(t)
	if err := c.Tags().Delete(context.Background(), "pcp-8053-does-not-exist"); err != nil {
		t.Fatalf("Tags.Delete 404 should be nil: %v", err)
	}
}
