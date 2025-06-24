package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CoreKitMDK/corekit-service-tracing/v2/pkg/tracing"
)

func TestTracingConfiguration(t *testing.T) {
	config := tracing.NewConfiguration()
	config.UseConsole = true
	config.UseNATS = true
	config.NatsURL = "internal-tracing-broker-nats-client"
	config.NatsPassword = "internal-tracing-broker"
	config.NatsUsername = "internal-tracing-broker"

	tracerMaster := config.Init()

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := tracerMaster.TraceHttpRequest(r).Start()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
		tracer.TraceHttpResponseWriter(w).End()
	}))
	defer ts.Close()

	// Make test request
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	time.Sleep(2 * time.Second)
}
