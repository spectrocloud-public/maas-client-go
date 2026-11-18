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
	"testing"
	"time"
)

func TestDefaultRateLimiter(t *testing.T) {
	c := NewAuthenticatedClientSet("http://example.com/MAAS", "consumer:token:secret")
	cs, ok := c.(*authenticatedClientSet)
	if !ok {
		t.Fatalf("unexpected client type %T", c)
	}
	if cs.client.limiter == nil {
		t.Fatal("default rate limiter is nil")
	}
	if cs.client.limiter.rate != 5 || cs.client.limiter.burst != 10 {
		t.Fatalf("limiter = %v/%v, want 5/10", cs.client.limiter.rate, cs.client.limiter.burst)
	}
}

func TestRateLimiterOption(t *testing.T) {
	c := NewAuthenticatedClientSet("http://example.com/MAAS", "consumer:token:secret", WithRateLimit(1, 2))
	cs := c.(*authenticatedClientSet)
	if cs.client.limiter.rate != 1 || cs.client.limiter.burst != 2 {
		t.Fatalf("limiter = %v/%v, want 1/2", cs.client.limiter.rate, cs.client.limiter.burst)
	}

	c = NewAuthenticatedClientSet("http://example.com/MAAS", "consumer:token:secret", WithRateLimit(0, 0))
	cs = c.(*authenticatedClientSet)
	if cs.client.limiter != nil {
		t.Fatal("qps=0 should disable the limiter")
	}
}

func TestRateLimiterWaitHonorsContext(t *testing.T) {
	l := newRateLimiter(1, 1)
	ctx := context.Background()

	if err := l.wait(ctx); err != nil {
		t.Fatalf("first wait: %v", err)
	}

	deadline, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	if err := l.wait(deadline); err == nil {
		t.Fatal("expected the second wait to hit the context deadline")
	}
}
