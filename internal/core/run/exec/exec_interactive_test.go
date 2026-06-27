package exec

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/rodolfochicone/rc-project/internal/core/agent"
	"github.com/rodolfochicone/rc-project/internal/core/model"
	"github.com/rodolfochicone/rc-project/internal/core/run/internal/acpshared"
)

// scriptedCoordinator records each pending prompt and returns scripted responses
// in order, defaulting to a canceled response once the script is exhausted.
type scriptedCoordinator struct {
	mu        sync.Mutex
	prompts   []model.PendingInput
	responses []model.UserResponse
	idx       int
}

func (c *scriptedCoordinator) Await(
	_ context.Context,
	prompt model.PendingInput,
) (model.UserResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prompts = append(c.prompts, prompt)
	if c.idx >= len(c.responses) {
		return model.UserResponse{Canceled: true}, nil
	}
	resp := c.responses[c.idx]
	c.idx++
	return resp, nil
}

func (*scriptedCoordinator) Submit(model.UserResponse) error { return nil }

func (c *scriptedCoordinator) snapshotPrompts() []model.PendingInput {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]model.PendingInput(nil), c.prompts...)
}

// countingCoordinator records how many times Await is called so a non-interactive
// run can assert the coordinator is never consulted.
type countingCoordinator struct{ awaits int32 }

func (c *countingCoordinator) Await(
	context.Context,
	model.PendingInput,
) (model.UserResponse, error) {
	atomic.AddInt32(&c.awaits, 1)
	return model.UserResponse{Canceled: true}, nil
}

func (*countingCoordinator) Submit(model.UserResponse) error { return nil }

func newReplyingExecSession(t *testing.T, id, reply string) *capturingExecSession {
	t.Helper()
	session := newCapturingExecSession(id)
	session.updates <- model.SessionUpdate{
		Kind:   model.UpdateKindAgentMessageChunk,
		Status: model.StatusRunning,
		Blocks: []model.ContentBlock{preparedPromptTextContentBlock(t, reply)},
	}
	go session.finish(nil)
	return session
}

func TestExecutePreparedPromptRunsInteractiveTurns(t *testing.T) {
	workspaceRoot := workspaceRootForExecTest(t)

	coord := &scriptedCoordinator{responses: []model.UserResponse{{Text: "my answer"}}}
	var (
		gotContinue   agent.ContinueRequest
		continueCalls int
	)

	restore := acpshared.SwapNewAgentClientForTest(
		func(_ context.Context, _ agent.ClientConfig) (agent.Client, error) {
			return &capturingExecACPClient{
				createSessionFn: func(_ context.Context, _ agent.SessionRequest) (agent.Session, error) {
					return newReplyingExecSession(t, "sess-interactive", "turn one reply"), nil
				},
				continueFn: func(_ context.Context, req agent.ContinueRequest) (agent.Session, error) {
					continueCalls++
					gotContinue = req
					return newReplyingExecSession(t, "sess-interactive", "turn two reply"), nil
				},
			}, nil
		},
	)
	t.Cleanup(restore)

	result, err := ExecutePreparedPrompt(
		context.Background(),
		&model.RuntimeConfig{
			WorkspaceRoot:    workspaceRoot,
			IDE:              model.IDECodex,
			Model:            "gpt-5.5",
			AccessMode:       model.AccessModeDefault,
			OutputFormat:     model.OutputFormatText,
			Interactive:      true,
			InputCoordinator: coord,
		},
		"start interactive",
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("execute prepared prompt: %v", err)
	}
	if continueCalls != 1 {
		t.Fatalf("continue called %d times, want exactly 1", continueCalls)
	}
	if string(gotContinue.Prompt) != "my answer" {
		t.Fatalf("continue prompt = %q, want the user's answer", string(gotContinue.Prompt))
	}
	if result.Output != "turn two reply" {
		t.Fatalf("final output = %q, want the second turn's reply", result.Output)
	}

	prompts := coord.snapshotPrompts()
	if len(prompts) == 0 {
		t.Fatal("expected at least one pending prompt recorded")
	}
	if prompts[0].Kind != model.PendingInputKindQuestion {
		t.Fatalf("first pending kind = %q, want question", prompts[0].Kind)
	}
	if prompts[0].Text != "turn one reply" {
		t.Fatalf("question text = %q, want the first turn's reply", prompts[0].Text)
	}
}

func TestExecutePreparedPromptNonInteractiveNeverAwaits(t *testing.T) {
	workspaceRoot := workspaceRootForExecTest(t)
	coord := &countingCoordinator{}

	restore := acpshared.SwapNewAgentClientForTest(
		func(_ context.Context, _ agent.ClientConfig) (agent.Client, error) {
			return &capturingExecACPClient{
				createSessionFn: func(_ context.Context, _ agent.SessionRequest) (agent.Session, error) {
					return newReplyingExecSession(t, "sess-oneshot", "single reply"), nil
				},
			}, nil
		},
	)
	t.Cleanup(restore)

	result, err := ExecutePreparedPrompt(
		context.Background(),
		&model.RuntimeConfig{
			WorkspaceRoot:    workspaceRoot,
			IDE:              model.IDECodex,
			Model:            "gpt-5.5",
			AccessMode:       model.AccessModeDefault,
			OutputFormat:     model.OutputFormatText,
			Interactive:      false,
			InputCoordinator: coord,
		},
		"go once",
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("execute prepared prompt: %v", err)
	}
	if result.Output != "single reply" {
		t.Fatalf("output = %q, want single reply", result.Output)
	}
	if got := atomic.LoadInt32(&coord.awaits); got != 0 {
		t.Fatalf("coordinator awaited %d times for a non-interactive run, want 0", got)
	}
}
