// service_test.go verifies model service capability derivation and filtering.
// internal/features/ai/model/app/model/service_test.go
package model

import (
	"testing"

	aiinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/ports"
)

// TestDeriveCapabilityIDsFromTagsMapsSemanticTags validates semantic capability derivation from tags.
func TestDeriveCapabilityIDsFromTagsMapsSemanticTags(t *testing.T) {

	capabilityIDs := deriveCapabilityIDsFromTags([]string{
		"image_edit",
		"vision_segmentation_image",
		"speech_asr",
	})

	if !containsAll(capabilityIDs, []string{
		"vision.edit.image",
		"vision.segmentation.promptable_image",
		"speech.asr",
	}) {
		t.Fatalf("expected derived capability ids to include mapped semantic capabilities, got %#v", capabilityIDs)
	}
}

// TestParseCapabilityIDsFromMetadataReadsKnownKeys validates capability id metadata parsing.
func TestParseCapabilityIDsFromMetadataReadsKnownKeys(t *testing.T) {

	metadata := `{"capabilityIds":["vision.edit.image"],"semantic_capabilities":["vision.segmentation.promptable_image"]}`
	capabilityIDs := parseCapabilityIDsFromMetadata(metadata)

	if !containsAll(capabilityIDs, []string{
		"vision.edit.image",
		"vision.segmentation.promptable_image",
	}) {
		t.Fatalf("expected capability ids parsed from metadata, got %#v", capabilityIDs)
	}
}

// TestParseSystemTagsFromMetadataReadsKnownKeys validates system tag metadata parsing.
func TestParseSystemTagsFromMetadataReadsKnownKeys(t *testing.T) {

	metadata := `{"systemTags":["image_edit"],"tags":["vision_segmentation_image"]}`
	systemTags := parseSystemTagsFromMetadata(metadata)

	if !containsAll(systemTags, []string{"image_edit", "vision_segmentation_image"}) {
		t.Fatalf("expected system tags parsed from metadata, got %#v", systemTags)
	}
}

// TestMatchesModelFilterByCapabilities validates capability-aware model filtering.
func TestMatchesModelFilterByCapabilities(t *testing.T) {

	profile := buildCapabilityProfile(ModelSummaryRecord{
		ModelCapabilitiesRecord: ModelCapabilitiesRecord{
			SupportsStreaming:        true,
			SupportsToolCalling:      true,
			SupportsStructuredOutput: false,
			SupportsVision:           true,
			InputModalities:          []string{"text", "image"},
			OutputModalities:         []string{"text"},
		},
	}, []string{"image_edit"})

	filter := aiinterfaces.ModelListFilter{
		RequiredInputModalities: []string{"image"},
		RequiredCapabilityIDs:   []string{"vision.edit.image"},
		RequiredSystemTags:      []string{"image_edit"},
	}

	if !matchesModelFilter(profile, filter) {
		t.Fatalf("expected model profile to satisfy capability-aware filter")
	}

	filter.RequiredCapabilityIDs = []string{"vision.segmentation.promptable_image"}
	if matchesModelFilter(profile, filter) {
		t.Fatalf("expected model profile to fail mismatched capability filter")
	}
}
