// image.go defines canonical image DTOs for the application facade.
// internal/app/contracts/image.go
package contracts

// GenerateImageRequest contains image generation inputs.
type GenerateImageRequest struct {
	ProviderName   string `json:"providerName"`
	ModelName      string `json:"modelName,omitempty"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Style          string `json:"style,omitempty"`
	ResponseFormat string `json:"responseFormat,omitempty"`
	User           string `json:"user,omitempty"`
}

// EditImageRequest contains image edit inputs.
type EditImageRequest struct {
	ProviderName string `json:"providerName"`
	ModelName    string `json:"modelName,omitempty"`
	Prompt       string `json:"prompt"`
	ImagePath    string `json:"imagePath"`
	MaskPath     string `json:"maskPath,omitempty"`
	N            int    `json:"n,omitempty"`
	Size         string `json:"size,omitempty"`
}

// ImageBinaryResult contains binary image output metadata.
type ImageBinaryResult struct {
	Bytes         []byte `json:"bytes"`
	RevisedPrompt string `json:"revisedPrompt,omitempty"`
}
