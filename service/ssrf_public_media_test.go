package service

import (
	"net/http"
	"testing"
)

func TestPublicMediaHTTPClientUsesEnvironmentProxyAndRejectsRedirects(t *testing.T) {
	client := PublicMediaHTTPClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport = %T", client.Transport)
	}
	if transport.Proxy == nil {
		t.Fatal("public media client must preserve the environment proxy")
	}
	if client.CheckRedirect == nil {
		t.Fatal("public media client must reject redirects")
	}
	if err := client.CheckRedirect(&http.Request{}, nil); err != http.ErrUseLastResponse {
		t.Fatalf("redirect error = %v", err)
	}
}
