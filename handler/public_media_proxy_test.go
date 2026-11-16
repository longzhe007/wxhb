package handler

import "testing"

func TestPublicVideoMediaURLAllowsTwitterVideoHost(t *testing.T) {
	parsed, err := publicVideoMediaURL("https://video.twimg.com/amplify_video/123/vid/avc1/720x1280/example.mp4?tag=29")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Host != "video.twimg.com" {
		t.Fatalf("host = %q", parsed.Host)
	}
}

func TestPublicVideoMediaURLRejectsOtherHosts(t *testing.T) {
	if _, err := publicVideoMediaURL("https://example.com/video.mp4"); err == nil {
		t.Fatal("expected non-Twitter video host to be rejected")
	}
}
