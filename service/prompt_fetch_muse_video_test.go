package service

import (
	"strings"
	"testing"
)

func TestBuildMuseVideoPromptsParsesAttributionAndMetadata(t *testing.T) {
	items, err := buildMuseVideoPromptsFromJSON([]byte(`[
		{
			"id":"MV001",
			"title":"阳台晨光",
			"category":"自拍 / Vlog",
			"language":"Chinese",
			"aspect":"9:16",
			"seconds":10,
			"needs_reference":true,
			"author":"creator",
			"date":"2026-11-04",
			"url":"https://x.com/creator/status/1",
			"text_status":"excerpt",
			"text":"人物按参考图保持一致。"
		}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %d", len(items))
	}
	item := items[0]
	if item.ID != "muse-video-prompts-MV001" || item.Title != "MV001 · 阳台晨光" || item.Prompt != "人物按参考图保持一致。" {
		t.Fatalf("item = %#v", item)
	}
	if !containsPromptTag(item.Tags, "muse-video") || !containsPromptTag(item.Tags, "9:16") || !containsPromptTag(item.Tags, "需要参考图") || !containsPromptTag(item.Tags, "摘录") {
		t.Fatalf("tags = %#v", item.Tags)
	}
	if item.Preview == "" || !containsAll(item.Preview, "@creator", "https://x.com/creator/status/1", "https://github.com/apimodels/muse-video-prompts") {
		t.Fatalf("preview = %q", item.Preview)
	}
}

func containsPromptTag(tags []string, expected string) bool {
	for _, tag := range tags {
		if tag == expected {
			return true
		}
	}
	return false
}

func containsAll(value string, values ...string) bool {
	for _, expected := range values {
		if !strings.Contains(value, expected) {
			return false
		}
	}
	return true
}
