/*
Copyright 2026 The Kubermatic Kubernetes Platform contributors.

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

package main

import (
	"errors"
	"net/http"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

type fakeManager struct {
	elected chan struct{}
	checks  map[string]healthz.Checker
}

func (f *fakeManager) AddReadyzCheck(name string, check healthz.Checker) error {
	f.checks[name] = check
	return nil
}

func (f *fakeManager) Elected() <-chan struct{} {
	return f.elected
}

func TestLeaderOnlyReadyzCheck(t *testing.T) {
	mgr := &fakeManager{elected: make(chan struct{}), checks: map[string]healthz.Checker{}}
	failing := func(_ *http.Request) error { return errors.New("not reconciled") }

	if err := leaderOnlyReadyzCheck(mgr)("test", failing); err != nil {
		t.Fatalf("failed to register check: %v", err)
	}

	if err := mgr.checks["test"](nil); err != nil {
		t.Errorf("standby replica should be ready, got: %v", err)
	}

	close(mgr.elected)

	if err := mgr.checks["test"](nil); err == nil {
		t.Error("leader should report the wrapped check's error")
	}
}
