package events

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/rodolfochicone/rc-project/pkg/rc/events/kinds"
)

func TestSessionAwaitingInputEnvelopeRoundTrip(t *testing.T) {
	t.Parallel()

	want := kinds.SessionAwaitingInputPayload{
		Index:        2,
		ACPSessionID: "acp-7",
		Kind:         kinds.AwaitingInputKindPermission,
		PromptID:     "prompt-42",
		Text:         "Run `make verify`?",
		Options:      []kinds.SessionInputOption{{OptionID: "allow_once", Label: "Allow once"}},
	}
	rawPayload, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	event := Event{
		SchemaVersion: SchemaVersion,
		RunID:         "run-1",
		Seq:           9,
		Timestamp:     time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC),
		Kind:          EventKindSessionAwaitingInput,
		Payload:       rawPayload,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	if decoded.Kind != EventKindSessionAwaitingInput {
		t.Fatalf("kind = %q, want %q", decoded.Kind, EventKindSessionAwaitingInput)
	}

	var gotPayload kinds.SessionAwaitingInputPayload
	if err := json.Unmarshal(decoded.Payload, &gotPayload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if !reflect.DeepEqual(gotPayload, want) {
		t.Fatalf("payload round-trip mismatch: got %#v want %#v", gotPayload, want)
	}
}
