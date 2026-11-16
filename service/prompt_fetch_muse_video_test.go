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

func TestMuseVideoPromptMediaLinksOnlineVideoToOriginalPost(t *testing.T) {
	media := museVideoPromptMediaFromHTML(`
		<script type="application/ld+json">{
			"@context":"https://schema.org",
			"@type":"ItemList",
			"itemListElement":[{"item":{"url":"https://apimodels.app/zh/muse-video-prompts#prompt-1","isBasedOn":"https://x.com/creator/status/1"}}]
		}</script>
		<script type="application/ld+json">{
			"@context":"https://schema.org",
			"@type":"VideoObject",
			"url":"https://apimodels.app/zh/muse-video-prompts#prompt-1",
			"contentUrl":"https://video.example.com/clip.mp4",
			"thumbnailUrl":"https://image.example.com/cover.jpg"
		}</script>`)

	item := media["https://x.com/creator/status/1"]
	if item.VideoURL != "https://video.example.com/clip.mp4" || item.CoverURL != "https://image.example.com/cover.jpg" {
		t.Fatalf("media = %#v", item)
	}
}

func TestBuildMuseVideoPromptsIncludesOnlineMedia(t *testing.T) {
	items, err := buildMuseVideoPromptsFromJSONWithMedia([]byte(`[
		{"id":"MV001","title":"阳台晨光","url":"https://x.com/creator/status/1","text":"人物按参考图保持一致。"}
	]`), map[string]museVideoPromptMedia{
		"https://x.com/creator/status/1": {CoverURL: "https://image.example.com/cover.jpg", VideoURL: "https://video.example.com/clip.mp4"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].CoverURL != "https://image.example.com/cover.jpg" || items[0].VideoURL != "https://video.example.com/clip.mp4" {
		t.Fatalf("items = %#v", items)
	}
}

func TestBuildMuseVideoPromptsMatchesOnlineMediaByPromptText(t *testing.T) {
	items, err := buildMuseVideoPromptsFromJSONWithMediaIndex([]byte(`[
		{"id":"MV001","title":"阳台晨光","url":"https://x.com/creator/status/1","text":"人物按参考图保持一致，镜头保持稳定。"}
	]`), museVideoPromptMediaIndex{
		BySource: map[string]museVideoPromptMedia{},
		ByText: map[string]museVideoPromptMedia{
			museVideoPromptTextKey("人物按参考图保持一致，镜头保持稳定。完整原文还在后面。 "): {CoverURL: "https://image.example.com/cover.jpg", VideoURL: "https://video.example.com/clip.mp4"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].CoverURL != "https://image.example.com/cover.jpg" || items[0].VideoURL != "https://video.example.com/clip.mp4" {
		t.Fatalf("items = %#v", items)
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
