// image_types.go defines types related to image generation.
// internal/features/ai/providers/ports/gateway/image_types.go
package gateway

// ImageGenerationOptions defines parameters for image generation requests.
type ImageGenerationOptions struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`               // Number of images to generate
	Size           string `json:"size,omitempty"`            // 256x256, 512x512, 1024x1024
	Quality        string `json:"quality,omitempty"`         // standard, hd
	Style          string `json:"style,omitempty"`           // vivid, natural
	ResponseFormat string `json:"response_format,omitempty"` // url, b64_json
	User           string `json:"user,omitempty"`
}

// ImageEditOptions defines parameters for image editing requests.
type ImageEditOptions struct {
	Model  string `json:"model"`
	Image  string `json:"image"`          // Path or Base64
	Mask   string `json:"mask,omitempty"` // Path or Base64 (optional)
	Prompt string `json:"prompt"`
	N      int    `json:"n,omitempty"`
	Size   string `json:"size,omitempty"`
}

// ImageResult represents the result of an image generation request.
type ImageResult struct {
	Created int64       `json:"created"`
	Data    []ImageData `json:"data"`
}

// ImageData contains the generated image info.
type ImageData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}
