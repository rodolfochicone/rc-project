package kinds

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSessionAwaitingInputPermissionMarshalsOptions(t *testing.T) {
	t.Parallel()

	payload := SessionAwaitingInputPayload{
		Index:        3,
		ACPSessionID: "acp-1",
		Kind:         AwaitingInputKindPermission,
		PromptID:     "p-1",
		Text:         "Run make verify?",
		Options: []SessionInputOption{
			{OptionID: "allow_once", Label: "Allow once"},
			{OptionID: "reject", Label: "Reject"},
		},
	}

	got := mustMarshalMap(t, payload)
	want := map[string]any{
		"index":          float64(3),
		"acp_session_id": "acp-1",
		"kind":           "permission",
		"prompt_id":      "p-1",
		"text":           "Run make verify?",
		"options": []any{
			map[string]any{"option_id": "allow_once", "label": "Allow once"},
			map[string]any{"option_id": "reject", "label": "Reject"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("awaiting-input permission JSON mismatch: got %#v want %#v", got, want)
	}
}

func TestSessionAwaitingInputQuestionOmitsEmptyFields(t *testing.T) {
	t.Parallel()

	payload := SessionAwaitingInputPayload{
		Index:    5,
		Kind:     AwaitingInputKindQuestion,
		PromptID: "p-2",
		Text:     "Which problem should this solve first?",
	}

	got := mustMarshalMap(t, payload)
	if _, ok := got["options"]; ok {
		t.Errorf("options should be omitted for a question, got %#v", got["options"])
	}
	if _, ok := got["acp_session_id"]; ok {
		t.Errorf("acp_session_id should be omitted when empty, got %#v", got["acp_session_id"])
	}
	if got["kind"] != "question" {
		t.Errorf("kind = %v, want question", got["kind"])
	}
	if got["prompt_id"] != "p-2" {
		t.Errorf("prompt_id = %v, want p-2", got["prompt_id"])
	}
}

func TestSessionAwaitingInputRoundTripsJSON(t *testing.T) {
	t.Parallel()

	original := SessionAwaitingInputPayload{
		Index:        7,
		ACPSessionID: "acp-9",
		Kind:         AwaitingInputKindPermission,
		PromptID:     "p-3",
		Text:         "Write file?",
		Options:      []SessionInputOption{{OptionID: "allow_always", Label: "Always"}},
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded SessionAwaitingInputPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("round-trip mismatch: got %#v want %#v", decoded, original)
	}
}
