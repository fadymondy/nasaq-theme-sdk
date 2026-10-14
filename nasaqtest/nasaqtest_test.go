package nasaqtest_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/fadymondy/nasaq-theme-sdk/nasaqtest"
)

func TestCallerOwnsServerLifecycle(t *testing.T) {
	fake := nasaqtest.NewServer("example")
	defer fake.Close()
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get(fake.URL + "/api/themes")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status %d", response.StatusCode)
	}
	fake.Close()
	if response, err := client.Get(fake.URL + "/api/themes"); err == nil {
		response.Body.Close()
		t.Fatal("caller-closed server still accepts requests")
	}
}
