package pixellab

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"strings"
)

const defaultBaseURL = "https://api.pixellab.ai"

// ImageSize specifies the dimensions of the generated image.
type ImageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Base64Image represents a base64-encoded image for API requests and responses.
type Base64Image struct {
	Type   string `json:"type"`
	Base64 string `json:"base64"` // data URL: "data:image/png;base64,..."
}

// Request holds all parameters for image generation.
type Request struct {
	Description         string    `json:"description"`
	ImageSize           ImageSize `json:"image_size"`
	TextGuidanceScale   *float64  `json:"text_guidance_scale,omitempty"`
	NegativeDescription *string   `json:"negative_description,omitempty"`
	NoBackground        *bool     `json:"no_background,omitempty"`
	Outline             *string   `json:"outline,omitempty"`
	Shading             *string   `json:"shading,omitempty"`
	Detail              *string   `json:"detail,omitempty"`
	Seed                *int      `json:"seed,omitempty"`
}

// AnimationRequest holds all parameters for animation generation.
type AnimationRequest struct {
	Description       string      `json:"description"`
	Action            string      `json:"action"`
	ImageSize         ImageSize   `json:"image_size"`
	ReferenceImage    Base64Image `json:"reference_image"`
	View              string      `json:"view,omitempty"`
	Direction         string      `json:"direction,omitempty"`
	NFrames           *int        `json:"n_frames,omitempty"`
	GuidanceScale     *float64    `json:"guidance_scale,omitempty"`
	InitImageStrength *int        `json:"init_image_strength,omitempty"`
	ForcedPalette     []string    `json:"forced_palette,omitempty"`
}

// response is the JSON envelope returned by the image generation API.
type response struct {
	Image Base64Image `json:"image"`
	Usage struct {
		Type    string `json:"type"`
		Credits int    `json:"credits"`
	} `json:"usage"`
}

// animationResponse is the JSON envelope returned by the animation API.
type animationResponse struct {
	Images []Base64Image `json:"images"`
	Usage  struct {
		Type    string `json:"type"`
		Credits int    `json:"credits"`
	} `json:"usage"`
}

// APIError represents a non-200 response from the Pixellab API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("pixellab API error (HTTP %d): %s", e.StatusCode, e.Body)
}

// Client communicates with the Pixellab API.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// New creates a Pixellab client with the given API key.
func New(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    http.DefaultClient,
	}
}

// SetEndpoint overrides the API base URL (useful for testing).
func (c *Client) SetEndpoint(url string) {
	c.baseURL = url
}

// post sends a JSON POST request and returns the raw response body.
// It handles marshalling, auth headers, and non-200 status codes.
func (c *Client) post(ctx context.Context, path string, reqBody any) ([]byte, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	return respBody, nil
}

// decodeBase64Image decodes a Base64Image (with optional data URL prefix) into an image.Image.
func decodeBase64Image(b64img Base64Image) (image.Image, error) {
	b64 := b64img.Base64
	if i := strings.IndexByte(b64, ','); i >= 0 {
		b64 = b64[i+1:]
	}

	pngData, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("decode base64 image: %w", err)
	}

	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("decode PNG: %w", err)
	}

	return img, nil
}

// EncodeBase64Image encodes an image.Image as a PNG Base64Image suitable for API requests.
func EncodeBase64Image(img image.Image) (Base64Image, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return Base64Image{}, fmt.Errorf("encode PNG: %w", err)
	}
	return Base64Image{
		Type:   "base64",
		Base64: base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

// Generate sends a generation request and returns the decoded PNG image.
func (c *Client) Generate(ctx context.Context, req Request) (image.Image, error) {
	respBody, err := c.post(ctx, "/v2/create-image-pixflux", req)
	if err != nil {
		return nil, err
	}

	var apiResp response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return decodeBase64Image(apiResp.Image)
}

// GenerateAnimation sends an animation request and returns the decoded frames.
func (c *Client) GenerateAnimation(ctx context.Context, req AnimationRequest) ([]image.Image, error) {
	respBody, err := c.post(ctx, "/v2/animate-with-text", req)
	if err != nil {
		return nil, err
	}

	var apiResp animationResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(apiResp.Images) == 0 {
		return nil, fmt.Errorf("API returned no frames")
	}

	frames := make([]image.Image, len(apiResp.Images))
	for i, b64img := range apiResp.Images {
		img, err := decodeBase64Image(b64img)
		if err != nil {
			return nil, fmt.Errorf("decode frame %d: %w", i, err)
		}
		frames[i] = img
	}

	return frames, nil
}
