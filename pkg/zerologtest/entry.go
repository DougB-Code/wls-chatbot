// entry.go models a single structured log entry for test assertions.
// pkg/zerologtest/entry.go
package zerologtest

// Contract field names for structured log assertions.
const (
	FieldTime      = "time"
	FieldCaller    = "caller"
	FieldLevel     = "level"
	FieldMessage   = "message"
	FieldEventCode = "event_code"
	FieldComponent = "component"
	FieldOp        = "op"
	FieldStatus    = "status"
	FieldErrKind   = "err_kind"
)

// EventCodeDecodeFailed flags invalid JSON writes to the recorder.
const EventCodeDecodeFailed = "LOG_DECODE_FAILED"

// PresenceOnlyValue marks fields that were intentionally normalized to presence-only.
const PresenceOnlyValue = "<present>"

// Entry stores one captured structured log event.
type Entry struct {
	Fields map[string]any
}

// Field returns the value for a field key.
func (entry Entry) Field(key string) (any, bool) {

	value, ok := entry.Fields[key]
	return value, ok
}

// StringField returns a string field value when present.
func (entry Entry) StringField(key string) (string, bool) {

	value, ok := entry.Field(key)
	if !ok {
		return "", false
	}
	str, ok := value.(string)
	if !ok {
		return "", false
	}
	return str, true
}

// EventCode returns the contracted event code.
func (entry Entry) EventCode() string {

	value, _ := entry.StringField(FieldEventCode)
	return value
}

// Level returns the zerolog level field.
func (entry Entry) Level() string {

	value, _ := entry.StringField(FieldLevel)
	return value
}

// Message returns the log message field.
func (entry Entry) Message() string {

	value, _ := entry.StringField(FieldMessage)
	return value
}

// copyEntry clones an entry to keep snapshots immutable.
func copyEntry(entry Entry) Entry {

	return Entry{Fields: copyFields(entry.Fields)}
}

// copyFields clones field maps to avoid shared mutable references.
func copyFields(fields map[string]any) map[string]any {

	if len(fields) == 0 {
		return map[string]any{}
	}

	cloned := make(map[string]any, len(fields))
	for key, value := range fields {
		cloned[key] = value
	}
	return cloned
}
