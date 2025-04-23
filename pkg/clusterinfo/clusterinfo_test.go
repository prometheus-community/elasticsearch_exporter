// Copyright 2021 The Prometheus Authors
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

package clusterinfo

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/common/promslog"

	"github.com/blang/semver/v4"
)

const (
	nodeName      = "test-node-"
	clusterName   = "test-cluster-1"
	clusterUUID   = "r1bT9sBrR7S9-CamE41Qqg"
	versionNumber = "5.6.9"
	buildHash     = "877a590"
	buildDate     = "2018-04-12T16:25:14.838Z"
	buildSnapshot = false
	luceneVersion = "6.6.1"
	tagline       = "You Know, for Search"
)

type mockES struct{}

func (mockES) ServeHTTP(w http.ResponseWriter, _ *http.Request) {

	fmt.Fprintf(w, `{
  "name" : "%s",
  "cluster_name" : "%s",
  "cluster_uuid" : "%s",
  "version" : {
    "number" : "%s",
    "build_hash" : "%s",
    "build_date" : "%s",
    "build_snapshot" : %t,
    "lucene_version" : "%s"
  },
  "tagline" : "%s"
}`,
		nodeName,
		clusterName,
		clusterUUID,
		versionNumber,
		buildHash,
		buildDate,
		buildSnapshot,
		luceneVersion,
		tagline,
	)
}

type mockConsumer struct {
	name string

	mu   sync.RWMutex
	data *Response

	ch chan *Response
}

func newMockConsumer(ctx context.Context, name string, t *testing.T) *mockConsumer {
	mc := &mockConsumer{
		name: name,
		ch:   make(chan *Response),
	}
	go func() {
		for {
			select {
			case d := <-mc.ch:
				mc.mu.Lock()
				mc.data = d
				t.Logf("consumer %s received data from channel: %+v\n", mc, mc.data)
				mc.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
	return mc
}

func (mc *mockConsumer) getData() Response {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return *mc.data
}

func (mc *mockConsumer) String() string {
	return mc.name
}

func (mc *mockConsumer) ClusterLabelUpdates() *chan *Response {
	return &mc.ch
}

func TestNew(t *testing.T) {
	u, err := url.Parse("http://localhost:9200")
	if err != nil {
		t.Skipf("internal test error: %s", err)
	}
	r := New(promslog.NewNopLogger(), http.DefaultClient, u, 0)
	if r.url != u {
		t.Errorf("new Retriever mal-constructed")
	}
}

func TestRetriever_RegisterConsumer(t *testing.T) {
	mockES := httptest.NewServer(mockES{})
	u, err := url.Parse(mockES.URL)
	if err != nil {
		t.Fatalf("internal test error: %s", err)
	}
	retriever := New(promslog.NewNopLogger(), mockES.Client(), u, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumerNames := []string{"consumer-1", "consumer-2"}
	for _, n := range consumerNames {
		c := newMockConsumer(ctx, n, t)
		if err := retriever.RegisterConsumer(c); err != nil {
			t.Errorf("failed to register consumer: %s", err)
		}
	}
	if len(retriever.consumerChannels) != len(consumerNames) {
		t.Error("number of registered consumerChannels doesn't match the number of calls to the register func")
	}
}

func TestRetriever_fetchAndDecodeClusterInfo(t *testing.T) {
	// these override test package globals
	versionNumber, _ := semver.Make(versionNumber)
	luceneVersion, _ := semver.Make(luceneVersion)

	var expected = &Response{
		Name:        nodeName,
		ClusterName: clusterName,
		ClusterUUID: clusterUUID,
		Version: VersionInfo{
			Number:        versionNumber,
			BuildHash:     buildHash,
			BuildDate:     buildDate,
			BuildSnapshot: buildSnapshot,
			LuceneVersion: luceneVersion,
		},
		Tagline: tagline,
	}

	mockES := httptest.NewServer(mockES{})
	u, err := url.Parse(mockES.URL)
	if err != nil {
		t.Skipf("internal test error: %s", err)
	}
	retriever := New(promslog.NewNopLogger(), mockES.Client(), u, 0)
	ci, err := retriever.fetchAndDecodeClusterInfo()
	if err != nil {
		t.Fatalf("failed to retrieve cluster info: %s", err)
	}

	if !reflect.DeepEqual(ci, expected) {
		t.Errorf("unexpected response, want %v, got %v", expected, ci)
	}
}

func TestRetriever_Run(t *testing.T) {
	// setup mock ES
	mockES := httptest.NewServer(mockES{})
	u, err := url.Parse(mockES.URL)
	if err != nil {
		t.Fatalf("internal test error: %s", err)
	}

	// setup cluster info retriever
	retriever := New(promslog.New(&promslog.Config{Writer: os.Stdout}), mockES.Client(), u, 0)

	// setup mock consumer
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	mc := newMockConsumer(ctx, "test-consumer", t)
	if err := retriever.RegisterConsumer(mc); err != nil {
		t.Fatalf("failed to register consumer: %s", err)
	}

	// start retriever
	retriever.Run(ctx)

	// trigger update
	retriever.Update()
	time.Sleep(20 * time.Millisecond)
	// ToDo: check mockConsumers received data
	t.Logf("%+v\n", mc.getData())

	// check for deadlocks
	select {
	case <-ctx.Done():
		if err := ctx.Err(); err == context.DeadlineExceeded {
			t.Fatal("context timeout exceeded, caught deadlock")
		}
	default:
	}
}
