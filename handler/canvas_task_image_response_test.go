package handler

import "testing"

func TestImageURLsFromAIResponseResolvesRelativeMediaURL(t *testing.T) {
	payload := []byte(`{"created":1794288019,"data":[{"url":"/v1/media/result.webp"}]}`)
	urls, _, _, err := imageURLsFromAIResponse(payload, "application/json", false, false, "http://199.255.96.145:18610/v1")
	if err != nil {
		t.Fatalf("imageURLsFromAIResponse returned error: %v", err)
	}
	if len(urls) != 1 || urls[0] != "http://199.255.96.145:18610/v1/media/result.webp" {
		t.Fatalf("urls = %v", urls)
	}
}
