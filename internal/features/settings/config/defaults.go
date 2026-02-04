// define default application configuration values.
// internal/features/settings/config/defaults.go
package config

// DefaultConfig returns the baseline application configuration.
func DefaultConfig() AppConfig {

	return AppConfig{
		Providers: []ProviderConfig{
			{
				Type:            "openai",
				Name:            "openai",
				DisplayName:     "OpenAI",
				BaseURL:         "https://api.openai.com/v1",
				UpdateFrequency: UpdateFrequencyManual,
				Models:          nil,
			},
			{
				Type:            "gemini",
				Name:            "gemini",
				DisplayName:     "Google Gemini",
				BaseURL:         "https://generativelanguage.googleapis.com/v1beta",
				UpdateFrequency: UpdateFrequencyManual,
				Models:          nil,
			},
		},
	}
}
