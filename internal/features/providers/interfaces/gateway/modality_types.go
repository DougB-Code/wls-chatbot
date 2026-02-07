// modality_types.go defines gateway modality and interaction classification types.
// internal/features/providers/interfaces/gateway/modality_types.go
package gateway

// InputType identifies normalized gateway input wire types.
type InputType string

const (
	InputText             InputType = "text"
	InputImage            InputType = "image"
	InputAudio            InputType = "audio"
	InputVideo            InputType = "video"
	InputDocument         InputType = "document"
	InputStructuredPrompt InputType = "structured_prompt"
)

// OutputType identifies normalized gateway output wire types.
type OutputType string

const (
	OutputText              OutputType = "text"
	OutputEmbedding         OutputType = "embedding"
	OutputImage             OutputType = "image"
	OutputAudio             OutputType = "audio"
	OutputVideo             OutputType = "video"
	OutputMasks             OutputType = "masks"
	OutputBoxes             OutputType = "boxes"
	OutputKeypoints         OutputType = "keypoints"
	OutputTracks            OutputType = "tracks"
	OutputSafetyLabels      OutputType = "safety_labels"
	OutputToolCalls         OutputType = "tool_calls"
	OutputRankingScores     OutputType = "ranking_scores"
	OutputActionInvocations OutputType = "action_invocations"
)

// InteractionType identifies the runtime interaction pattern for a gateway capability.
type InteractionType string

const (
	InteractionSingle    InteractionType = "single"
	InteractionStreaming InteractionType = "streaming"
	InteractionDuplex    InteractionType = "duplex"
	InteractionBatch     InteractionType = "batch"
)
