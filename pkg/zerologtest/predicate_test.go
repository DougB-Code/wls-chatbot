// predicate_test.go verifies predicate matching and assertion helpers.
// pkg/zerologtest/predicate_test.go
package zerologtest

import "testing"

// TestPredicatesMatchStructuredEntries verifies primitive field predicates.
func TestPredicatesMatchStructuredEntries(t *testing.T) {

	recorder := NewRecorder()
	line := `{"level":"info","event_code":"TASK_DONE","component":"worker","op":"sync","status":"ok","attempt":2,"message":"task completed"}`
	if _, err := recorder.Write([]byte(line + "\n")); err != nil {
		t.Fatalf("write line: %v", err)
	}

	entry := recorder.Entries()[0]
	if !HasEventCode("TASK_DONE").Match(entry) {
		t.Fatalf("expected HasEventCode to match")
	}
	if !HasLevel("info").Match(entry) {
		t.Fatalf("expected HasLevel to match")
	}
	if !HasField("attempt").Match(entry) {
		t.Fatalf("expected HasField to match")
	}
	if !FieldEq("attempt", 2).Match(entry) {
		t.Fatalf("expected FieldEq numeric match for decoded JSON number")
	}
	if !FieldContains(FieldMessage, "completed").Match(entry) {
		t.Fatalf("expected FieldContains to match")
	}
}

// TestPredicateCombinators verifies And, Or, and Not behavior.
func TestPredicateCombinators(t *testing.T) {

	entry := Entry{
		Fields: map[string]any{
			FieldLevel:     "error",
			FieldEventCode: "SYNC_FAILED",
			FieldErrKind:   "timeout",
		},
	}

	if !And(HasLevel("error"), HasEventCode("SYNC_FAILED")).Match(entry) {
		t.Fatalf("expected And predicate to match")
	}
	if !Or(HasEventCode("OTHER"), HasEventCode("SYNC_FAILED")).Match(entry) {
		t.Fatalf("expected Or predicate to match")
	}
	if !Not(HasEventCode("OTHER")).Match(entry) {
		t.Fatalf("expected Not predicate to match")
	}
}

// TestContainsAndInOrderHelpers verifies contains and ordering helper behavior.
func TestContainsAndInOrderHelpers(t *testing.T) {

	entries := []Entry{
		{Fields: map[string]any{FieldEventCode: "SERVER_STARTING", FieldLevel: "info"}},
		{Fields: map[string]any{FieldEventCode: "MIGRATIONS_DONE", FieldLevel: "info"}},
		{Fields: map[string]any{FieldEventCode: "SERVER_READY", FieldLevel: "info"}},
	}

	if !Contains(entries, HasEventCode("MIGRATIONS_DONE")) {
		t.Fatalf("expected Contains to find MIGRATIONS_DONE")
	}
	if Contains(entries, HasEventCode("MISSING")) {
		t.Fatalf("expected Contains to not match MISSING")
	}
	if !InOrder(entries, HasEventCode("SERVER_STARTING"), HasEventCode("SERVER_READY")) {
		t.Fatalf("expected InOrder to match event sequence")
	}
	if InOrder(entries, HasEventCode("SERVER_READY"), HasEventCode("SERVER_STARTING")) {
		t.Fatalf("expected InOrder to fail out-of-order sequence")
	}
}

// TestAssertionHelpersPassOnValidData verifies assertion wrappers on passing cases.
func TestAssertionHelpersPassOnValidData(t *testing.T) {

	recorder := NewRecorder()
	if _, err := recorder.Write([]byte(`{"level":"warn","event_code":"CACHE_STALE","component":"catalog","op":"refresh","status":"warn"}` + "\n")); err != nil {
		t.Fatalf("write line: %v", err)
	}
	if _, err := recorder.Write([]byte(`{"level":"info","event_code":"CACHE_REFRESHED","component":"catalog","op":"refresh","status":"ok"}` + "\n")); err != nil {
		t.Fatalf("write line: %v", err)
	}

	entries := recorder.Entries()
	AssertContains(t, entries, HasEventCode("CACHE_STALE"))
	AssertNotContains(t, entries, HasEventCode("MISSING_EVENT"))
	AssertInOrder(t, entries, []Predicate{
		HasEventCode("CACHE_STALE"),
		HasEventCode("CACHE_REFRESHED"),
	})
	AssertRecorderContains(t, recorder, HasEventCode("CACHE_REFRESHED"))
	AssertRecorderNotContains(t, recorder, HasEventCode("SOMETHING_ELSE"))
	AssertRecorderInOrder(t, recorder, []Predicate{
		HasEventCode("CACHE_STALE"),
		HasEventCode("CACHE_REFRESHED"),
	})
}
