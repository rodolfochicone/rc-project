package agent

import (
	"context"
	"testing"

	acp "github.com/coder/acp-go-sdk"

	"github.com/rodolfochicone/rc-project/internal/core/model"
)

// captureCoordinator records the PendingInput it received and returns a scripted
// response, letting tests assert option mapping and outcome selection.
type captureCoordinator struct {
	got  model.PendingInput
	resp model.UserResponse
	err  error
}

func (f *captureCoordinator) Await(
	_ context.Context,
	prompt model.PendingInput,
) (model.UserResponse, error) {
	f.got = prompt
	return f.resp, f.err
}

func (f *captureCoordinator) Submit(model.UserResponse) error { return nil }

// blockingCoordinator blocks until the context is done, mimicking a user who
// never answers while the run is being canceled.
type blockingCoordinator struct{}

func (blockingCoordinator) Await(
	ctx context.Context,
	_ model.PendingInput,
) (model.UserResponse, error) {
	<-ctx.Done()
	return model.UserResponse{}, ctx.Err()
}

func (blockingCoordinator) Submit(model.UserResponse) error { return nil }

// selectedOption returns the chosen permission option id, or "" when the outcome
// is a declined/canceled (non-selection) result. It avoids referencing the ACP
// SDK's British-spelled outcome field directly in assertions.
func selectedOption(resp acp.RequestPermissionResponse) string {
	if resp.Outcome.Selected == nil {
		return ""
	}
	return string(resp.Outcome.Selected.OptionId)
}

func permissionParams() acp.RequestPermissionRequest {
	title := "Run make verify"
	return acp.RequestPermissionRequest{
		SessionId: "sess-1",
		ToolCall: acp.RequestPermissionToolCall{
			ToolCallId: "tool-1",
			Title:      &title,
		},
		Options: []acp.PermissionOption{
			{OptionId: "allow_once", Name: "Allow once"},
			{OptionId: "reject", Name: "Reject"},
		},
	}
}

func TestRequestPermissionInteractiveSelectsUserOption(t *testing.T) {
	t.Parallel()

	coord := &captureCoordinator{
		resp: model.UserResponse{PromptID: "perm-tool-1", OptionID: "reject"},
	}
	client := &clientImpl{interactive: true, inputCoordinator: coord}

	resp, err := client.RequestPermission(context.Background(), permissionParams())
	if err != nil {
		t.Fatalf("RequestPermission returned error: %v", err)
	}
	if got := selectedOption(resp); got != "reject" {
		t.Fatalf("selected option = %q, want reject", got)
	}

	// Verify the pending prompt handed to the coordinator: kind, text, and the
	// ACP options mapped to InputOptions preserving id and label.
	if coord.got.Kind != model.PendingInputKindPermission {
		t.Errorf("pending kind = %q, want permission", coord.got.Kind)
	}
	if coord.got.ID != "perm-tool-1" {
		t.Errorf("pending id = %q, want perm-tool-1", coord.got.ID)
	}
	if coord.got.Text != "Run make verify" {
		t.Errorf("pending text = %q, want the tool-call title", coord.got.Text)
	}
	if len(coord.got.Options) != 2 {
		t.Fatalf("mapped options = %d, want 2", len(coord.got.Options))
	}
	if coord.got.Options[0] != (model.InputOption{OptionID: "allow_once", Label: "Allow once"}) {
		t.Errorf("first mapped option = %+v", coord.got.Options[0])
	}
}

func TestRequestPermissionInteractiveCancelsOnDeclinedResponse(t *testing.T) {
	t.Parallel()

	coord := &captureCoordinator{resp: model.UserResponse{Canceled: true}}
	client := &clientImpl{interactive: true, inputCoordinator: coord}

	resp, err := client.RequestPermission(context.Background(), permissionParams())
	if err != nil {
		t.Fatalf("RequestPermission returned error: %v", err)
	}
	if got := selectedOption(resp); got != "" {
		t.Fatalf("expected no option selected on decline, got %q", got)
	}
}

func TestRequestPermissionInteractiveCancelsOnContextCancellation(t *testing.T) {
	t.Parallel()

	client := &clientImpl{interactive: true, inputCoordinator: blockingCoordinator{}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resp, err := client.RequestPermission(ctx, permissionParams())
	if err != nil {
		t.Fatalf("RequestPermission returned error: %v", err)
	}
	if got := selectedOption(resp); got != "" {
		t.Fatalf("expected no option selected on context cancellation, got %q", got)
	}
}

func TestRequestPermissionNonInteractiveSelectsFirstOption(t *testing.T) {
	t.Parallel()

	client := &clientImpl{interactive: false}
	resp, err := client.RequestPermission(context.Background(), permissionParams())
	if err != nil {
		t.Fatalf("RequestPermission returned error: %v", err)
	}
	if got := selectedOption(resp); got != "allow_once" {
		t.Fatalf("expected first option auto-approved, got %q", got)
	}
}

func TestRequestPermissionWithoutOptionsCancels(t *testing.T) {
	t.Parallel()

	client := &clientImpl{interactive: true, inputCoordinator: &captureCoordinator{}}
	resp, err := client.RequestPermission(context.Background(), acp.RequestPermissionRequest{})
	if err != nil {
		t.Fatalf("RequestPermission returned error: %v", err)
	}
	if got := selectedOption(resp); got != "" {
		t.Fatalf("expected no option selected with empty options, got %q", got)
	}
}

func TestClientContinueValidatesRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		client *clientImpl
		req    ContinueRequest
	}{
		{
			name:   "missing session id",
			client: &clientImpl{sessions: map[string]*sessionImpl{}},
			req:    ContinueRequest{},
		},
		{
			name:   "client not started",
			client: &clientImpl{sessions: map[string]*sessionImpl{}},
			req:    ContinueRequest{SessionID: "sess-1"},
		},
		{
			name:   "client closed",
			client: &clientImpl{sessions: map[string]*sessionImpl{}, started: true, closed: true},
			req:    ContinueRequest{SessionID: "sess-1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := tc.client.Continue(context.Background(), tc.req); err == nil {
				t.Fatal("expected Continue to return an error")
			}
		})
	}
}

func TestClientContinueSendsAnotherPromptTurnOnLiveSession(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	scenario := helperScenario{
		SessionID:   "sess-multi",
		ExpectedCWD: cwd,
		StopReason:  string(acp.StopReasonEndTurn),
	}
	client := newTestClient(t, scenario)
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	})

	first, err := client.CreateSession(context.Background(), SessionRequest{
		WorkingDir: cwd,
		Prompt:     []byte("turn one"),
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	_ = collectSessionUpdates(t, first)

	continuer, ok := client.(SessionContinuer)
	if !ok {
		t.Fatal("test client does not implement SessionContinuer")
	}
	second, err := continuer.Continue(context.Background(), ContinueRequest{
		SessionID:  "sess-multi",
		WorkingDir: cwd,
		Prompt:     []byte("turn two"),
	})
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}

	updates := collectSessionUpdates(t, second)
	if len(updates) == 0 || updates[len(updates)-1].Status != model.StatusCompleted {
		t.Fatalf("expected second turn to complete, got %+v", updates)
	}
	if second.Err() != nil {
		t.Fatalf("unexpected second-turn error: %v", second.Err())
	}
}
