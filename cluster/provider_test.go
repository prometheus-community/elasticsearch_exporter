// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/prometheus/common/promslog"
)

func TestInfoProvider_GetInfo(t *testing.T) {
	timesURLCalled := 0
	expectedInfo := Info{
		ClusterName: "test-cluster-1",
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		timesURLCalled++
		_, _ = w.Write([]byte(`{
			"name": "test-node-abcd",
			"cluster_name": "test-cluster-1",
			"cluster_uuid": "r1bT9sBrR7S9-CamE41Qqg",
			"version": {
				"number": "5.6.9",
				"build_hash": "877a590",
				"build_date": "2018-04-12T16:25:14.838Z",
				"build_snapshot": false,
				"lucene_version": "6.6.1"
			}
		}`))
	}))
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse test server URL: %v", err)
	}
	defer ts.Close()

	i := NewInfoProvider(promslog.New(&promslog.Config{Writer: os.Stdout}), http.DefaultClient, tsURL, time.Second)

	if timesURLCalled != 0 {
		t.Errorf("expected no initial URL calls, got %d", timesURLCalled)
	}

	got, err := i.GetInfo(context.Background())
	if err != nil {
		t.Errorf("InfoProvider.GetInfo() error = %v, wantErr %v", err, false)
		return
	}

	if !reflect.DeepEqual(got, expectedInfo) {
		t.Errorf("InfoProvider.GetInfo() = %v, want %v", got, expectedInfo)
	}

	if timesURLCalled != 1 {
		t.Errorf("expected URL to be called once, got %d", timesURLCalled)
	}

	// Call again to ensure cached value is returned
	got, err = i.GetInfo(context.Background())
	if err != nil {
		t.Errorf("InfoProvider.GetInfo() error on second call = %v, wantErr %v", err, false)
		return
	}
	if !reflect.DeepEqual(got, expectedInfo) {
		t.Errorf("InfoProvider.GetInfo() on second call = %v, want %v", got, expectedInfo)
	}
	if timesURLCalled != 1 {
		t.Errorf("expected URL to be called only once, got %d", timesURLCalled)
	}

	// Call again after delay to ensure we refresh the cache
	time.Sleep(2 * time.Second)
	got, err = i.GetInfo(context.Background())
	if err != nil {
		t.Errorf("InfoProvider.GetInfo() error on second call = %v, wantErr %v", err, false)
		return
	}
	if !reflect.DeepEqual(got, expectedInfo) {
		t.Errorf("InfoProvider.GetInfo() on second call = %v, want %v", got, expectedInfo)
	}
	if timesURLCalled != 2 {
		t.Errorf("expected URL to be called only once, got %d", timesURLCalled)
	}
}
