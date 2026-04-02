package pixellab

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
)

// encodePNG creates a minimal PNG of the given size and returns it as a Base64Image with data URL.
// encodePNG creates a minimal PNG of the given size and returns it as a base64 data URL string.
func encodePNG(t *testing.T, width, height int) Base64Image {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return Base64Image{
		Type:   "base64",
		Base64: "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}
}

func TestGenerate_Success(t *testing.T) {
	respImg := encodePNG(t, 32, 32)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/create-image-pixflux" {
			t.Errorf("expected path /v2/create-image-pixflux, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("expected application/json, got %s", got)
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Description != "red mushroom" {
			t.Errorf("expected 'red mushroom', got %q", req.Description)
		}
		if req.ImageSize.Width != 32 || req.ImageSize.Height != 32 {
			t.Errorf("expected 32x32, got %dx%d", req.ImageSize.Width, req.ImageSize.Height)
		}

		resp := response{Image: respImg}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := New("test-key")
	client.SetEndpoint(srv.URL)

	img, err := client.Generate(context.Background(), Request{
		Description: "red mushroom",
		ImageSize:   ImageSize{Width: 32, Height: 32},
	})
	if err != nil {
		t.Fatal(err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("expected 32x32 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerate_OptionalFields(t *testing.T) {
	respImg := encodePNG(t, 32, 32)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.TextGuidanceScale == nil || *req.TextGuidanceScale != 12.0 {
			t.Error("expected guidance scale 12.0")
		}
		if req.NoBackground == nil || !*req.NoBackground {
			t.Error("expected no_background true")
		}

		resp := response{Image: respImg}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := New("test-key")
	client.SetEndpoint(srv.URL)

	guidance := 12.0
	noBg := true
	_, err := client.Generate(context.Background(), Request{
		Description:       "test",
		ImageSize:         ImageSize{Width: 32, Height: 32},
		TextGuidanceScale: &guidance,
		NoBackground:      &noBg,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerate_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid api key"}`))
	}))
	defer srv.Close()

	client := New("bad-key")
	client.SetEndpoint(srv.URL)

	_, err := client.Generate(context.Background(), Request{
		Description: "test",
		ImageSize:   ImageSize{Width: 32, Height: 32},
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestGenerate_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	client := New("key")
	client.SetEndpoint(srv.URL)

	_, err := client.Generate(context.Background(), Request{
		Description: "test",
		ImageSize:   ImageSize{Width: 32, Height: 32},
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGenerate_InvalidBase64(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := response{Image: Base64Image{Type: "base64", Base64: "not-valid-base64!!!"}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := New("key")
	client.SetEndpoint(srv.URL)

	_, err := client.Generate(context.Background(), Request{
		Description: "test",
		ImageSize:   ImageSize{Width: 32, Height: 32},
	})
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestGenerate_InvalidPNG(t *testing.T) {
	// Valid base64 but not a PNG
	b64 := base64.StdEncoding.EncodeToString([]byte("not a png"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := response{Image: Base64Image{Type: "base64", Base64: "data:image/png;base64," + b64}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := New("key")
	client.SetEndpoint(srv.URL)

	_, err := client.Generate(context.Background(), Request{
		Description: "test",
		ImageSize:   ImageSize{Width: 32, Height: 32},
	})
	if err == nil {
		t.Fatal("expected error for invalid PNG data")
	}
}

func TestGenerateAnimation_Success(t *testing.T) {
	frame := encodePNG(t, 64, 64)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/animate-with-text" {
			t.Errorf("expected path /v2/animate-with-text, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req AnimationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Description != "a robot" {
			t.Errorf("expected 'a robot', got %q", req.Description)
		}
		if req.Action != "walk" {
			t.Errorf("expected 'walk', got %q", req.Action)
		}

		resp := animationResponse{
			Images: []Base64Image{frame, frame, frame, frame},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := New("test-key")
	client.SetEndpoint(srv.URL)

	nFrames := 4
	refImg := encodePNG(t, 64, 64)
	frames, err := client.GenerateAnimation(context.Background(), AnimationRequest{
		Description:    "a robot",
		Action:         "walk",
		ImageSize:      ImageSize{Width: 64, Height: 64},
		ReferenceImage: refImg,
		NFrames:        &nFrames,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(frames) != 4 {
		t.Fatalf("expected 4 frames, got %d", len(frames))
	}
	for i, f := range frames {
		bounds := f.Bounds()
		if bounds.Dx() != 64 || bounds.Dy() != 64 {
			t.Errorf("frame %d: expected 64x64, got %dx%d", i, bounds.Dx(), bounds.Dy())
		}
	}
}

func TestGenerateAnimation_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid api key"}`))
	}))
	defer srv.Close()

	client := New("bad-key")
	client.SetEndpoint(srv.URL)

	refImg := encodePNG(t, 64, 64)
	_, err := client.GenerateAnimation(context.Background(), AnimationRequest{
		Description:    "test",
		Action:         "walk",
		ImageSize:      ImageSize{Width: 64, Height: 64},
		ReferenceImage: refImg,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestGenerateAnimation_NoFrames(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := animationResponse{Images: []Base64Image{}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := New("key")
	client.SetEndpoint(srv.URL)

	refImg := encodePNG(t, 64, 64)
	_, err := client.GenerateAnimation(context.Background(), AnimationRequest{
		Description:    "test",
		Action:         "walk",
		ImageSize:      ImageSize{Width: 64, Height: 64},
		ReferenceImage: refImg,
	})
	if err == nil {
		t.Fatal("expected error for empty frames")
	}
}
