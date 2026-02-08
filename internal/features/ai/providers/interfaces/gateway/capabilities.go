// capabilities.go defines gateway semantic capability descriptors.
// internal/features/providers/interfaces/gateway/capabilities.go
package gateway

// CapabilityID identifies a semantic gateway capability.
type CapabilityID string

const (
	CapabilityChatText                     CapabilityID = "chat.text"
	CapabilityChatMultimodalToText         CapabilityID = "chat.multimodal_to_text"
	CapabilityGenerateImage                CapabilityID = "gen.image"
	CapabilityGenerateVideo                CapabilityID = "gen.video"
	CapabilitySpeechASR                    CapabilityID = "speech.asr"
	CapabilitySpeechTTS                    CapabilityID = "speech.tts"
	CapabilityRealtimeDuplexAudio          CapabilityID = "realtime.duplex_audio"
	CapabilityVisionSegmentationImage      CapabilityID = "vision.segmentation.promptable_image"
	CapabilityVisionSegmentationVideo      CapabilityID = "vision.segmentation.promptable_video"
	CapabilityRetrievalEmbedText           CapabilityID = "retrieval.embed.text"
	CapabilityRetrievalEmbedMultimodal     CapabilityID = "retrieval.embed.multimodal"
	CapabilityRankRerank                   CapabilityID = "rank.rerank"
	CapabilitySafetyModeration             CapabilityID = "safety.moderation"
	CapabilityAgentToolUse                 CapabilityID = "agent.tool_use"
)

// ControlDescriptor describes a provider-specific control exposed for a capability.
type ControlDescriptor struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

// CapabilityDescriptor describes one semantic capability and its I/O contract.
type CapabilityDescriptor struct {
	ID          CapabilityID        `json:"id"`
	Inputs      []InputType         `json:"inputs"`
	Outputs     []OutputType        `json:"outputs"`
	Interaction InteractionType     `json:"interaction"`
	Controls    []ControlDescriptor `json:"controls,omitempty"`
}

// CapabilityAdvertiser is implemented by providers that expose capability metadata.
type CapabilityAdvertiser interface {
	GatewayCapabilities() []CapabilityDescriptor
}
