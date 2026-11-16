package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/textproto"
	"testing"

	"github.com/tigerowo/infinite-canvas/model"
)

func TestNormalizeVideoCreateBodyConvertsMuseMultipartToJSON(t *testing.T) {
	var raw bytes.Buffer
	writer := multipart.NewWriter(&raw)
	for key, value := range map[string]string{
		"model":             "muse-video",
		"prompt":            "小猫起床跳舞",
		"seconds":           "6",
		"size":              "1280x720",
		"input_reference[]": "data:image/webp;base64,ZmFrZQ==",
	} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	body, contentType, err := normalizeVideoCreateBody(raw.Bytes(), writer.FormDataContentType(), "muse-video", model.ModelChannel{Protocol: "openai"}, "/videos")
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "application/json" {
		t.Fatalf("contentType = %q", contentType)
	}
	var payload struct {
		Prompt   string `json:"prompt"`
		Duration int    `json:"duration"`
		Size     string `json:"size"`
		Images   []struct {
			URL string `json:"url"`
		} `json:"images"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Prompt != "小猫起床跳舞" || payload.Duration != 6 || payload.Size != "16:9" || len(payload.Images) != 1 || payload.Images[0].URL != "data:image/webp;base64,ZmFrZQ==" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestNormalizeVideoCreateBodyIncludesMuseAudioReference(t *testing.T) {
	var raw bytes.Buffer
	writer := multipart.NewWriter(&raw)
	for key, value := range map[string]string{
		"model":   "muse-video",
		"prompt":  "根据音频节奏生成视频",
		"seconds": "10",
		"size":    "720x1280",
	} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="audio_reference[]"; filename="voice.mp3"`)
	header.Set("Content-Type", "audio/mpeg")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("fake-mp3")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	body, contentType, err := normalizeVideoCreateBody(raw.Bytes(), writer.FormDataContentType(), "muse-video", model.ModelChannel{Protocol: "openai"}, "/videos")
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "application/json" {
		t.Fatalf("contentType = %q", contentType)
	}
	var payload struct {
		Audios []struct {
			URL string `json:"url"`
		} `json:"audios"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Audios) != 1 || payload.Audios[0].URL != "data:audio/mpeg;base64,ZmFrZS1tcDM=" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestParseVideoTaskPayloadReadsMuseResultURL(t *testing.T) {
	task := parseVideoTaskPayload([]byte(`{"id":"task_123","status":"succeeded","progress":100,"result":{"url":"/v1/media/result.mp4"}}`), "muse-video")
	if task.VideoURL != "/v1/media/result.mp4" {
		t.Fatalf("VideoURL = %q", task.VideoURL)
	}
}
