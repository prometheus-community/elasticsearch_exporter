package clusterinfo

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/blang/semver"
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

func (mockES) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
	data *Response
	ch   chan *Response
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
				mc.data = d
				t.Logf("consumer %s received data from channel: %+v\n", mc, mc.data)
			case <-ctx.Done():
				t.Logf("shutting down consumer %s", mc)
				return
			}
		}
	}()
	return mc
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
	r := New(log.NewNopLogger(), http.DefaultClient, u, 0)
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
	retriever := New(log.NewNopLogger(), mockES.Client(), u, 0)
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
	buildDate, _ := time.Parse(time.RFC3339, buildDate)

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
	retriever := New(log.NewNopLogger(), mockES.Client(), u, 0)
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
	retriever := New(log.NewLogfmtLogger(os.Stdout), mockES.Client(), u, 0)

	// setup mock consumer
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
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
	t.Logf("%+v\n", mc.data)

	// check for deadlocks
	select {
	case <-ctx.Done():
		if err := ctx.Err(); err == context.DeadlineExceeded {
			t.Fatal("context timeout exceeded, caught deadlock")
		}
	default:
	}
}
