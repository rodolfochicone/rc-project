package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	apicore "github.com/rodolfochicone/rc-project/internal/api/core"
	"github.com/rodolfochicone/rc-project/internal/core/model"
	eventspkg "github.com/rodolfochicone/rc-project/pkg/rc/events"
	"github.com/rodolfochicone/rc-project/pkg/rc/events/kinds"
)

type fakeJournal struct {
	mu     sync.Mutex
	events []eventspkg.Event
}

func (j *fakeJournal) SubmitWithSeq(_ context.Context, event eventspkg.Event) (uint64, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.events = append(j.events, event)
	return uint64(len(j.events)), nil
}

func (j *fakeJournal) snapshot() []eventspkg.Event {
	j.mu.Lock()
	defer j.mu.Unlock()
	return append([]eventspkg.Event(nil), j.events...)
}

func pollUntil(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if cond() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("condition not met within deadline")
		}
		runtime.Gosched()
	}
}

func TestRunInputCoordinatorEmitsAwaitingInputAndTracksPending(t *testing.T) {
	t.Parallel()

	journal := &fakeJournal{}
	coordinator := newRunInputCoordinator("run-1", journal, nil)
	prompt := model.PendingInput{
		ID:      "p1",
		Kind:    model.PendingInputKindPermission,
		Text:    "Run make verify?",
		Options: []model.InputOption{{OptionID: "allow_once", Label: "Allow once"}},
	}

	done := make(chan struct{})
	go func() {
		_, _ = coordinator.Await(context.Background(), prompt)
		close(done)
	}()

	pollUntil(t, func() bool { return coordinator.PendingInput() != nil })
	if got := coordinator.PendingInput(); got == nil || got.ID != "p1" {
		t.Fatalf("PendingInput() = %+v, want prompt p1", got)
	}

	if err := coordinator.Submit(model.UserResponse{PromptID: "p1", OptionID: "allow_once"}); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	<-done
	if coordinator.PendingInput() != nil {
		t.Fatal("PendingInput() should be nil after resolution")
	}

	events := journal.snapshot()
	if len(events) != 1 {
		t.Fatalf("emitted %d events, want 1", len(events))
	}
	if events[0].Kind != eventspkg.EventKindSessionAwaitingInput {
		t.Fatalf("event kind = %q, want session.awaiting_input", events[0].Kind)
	}
	var payload kinds.SessionAwaitingInputPayload
	if err := json.Unmarshal(events[0].Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.PromptID != "p1" || payload.Kind != "permission" || len(payload.Options) != 1 {
		t.Fatalf("awaiting-input payload = %+v", payload)
	}
}

func TestPendingInputFromLastEvent(t *testing.T) {
	t.Parallel()

	awaiting, err := json.Marshal(kinds.SessionAwaitingInputPayload{
		Kind:     model.PendingInputKindQuestion,
		PromptID: "q1",
		Text:     "Which problem first?",
	})
	if err != nil {
		t.Fatalf("marshal awaiting payload: %v", err)
	}

	// The run is awaiting input when its last event is an awaiting-input event.
	pending := pendingInputFromLastEvent(&eventspkg.Event{
		Kind:    eventspkg.EventKindSessionAwaitingInput,
		Payload: awaiting,
	})
	if pending == nil || pending.PromptID != "q1" || pending.Kind != model.PendingInputKindQuestion {
		t.Fatalf("pendingInputFromLastEvent = %+v, want q1 question", pending)
	}

	// A later session update supersedes the prompt: the last event is no longer
	// an awaiting-input, so nothing is pending.
	update, err := json.Marshal(kinds.SessionUpdatePayload{
		Index: 0,
		Update: kinds.SessionUpdate{
			Kind:   kinds.UpdateKindAgentMessageChunk,
			Status: kinds.StatusRunning,
		},
	})
	if err != nil {
		t.Fatalf("marshal session update: %v", err)
	}
	if got := pendingInputFromLastEvent(&eventspkg.Event{
		Kind:    eventspkg.EventKindSessionUpdate,
		Payload: update,
	}); got != nil {
		t.Fatalf("pendingInputFromLastEvent after update = %+v, want nil", got)
	}

	// No last event (fresh run) means nothing is pending.
	if got := pendingInputFromLastEvent(nil); got != nil {
		t.Fatalf("pendingInputFromLastEvent(nil) = %+v, want nil", got)
	}
}

func TestRunManagerSendInputDeliversToAwaitingRun(t *testing.T) {
	started := make(chan string, 1)
	gotResp := make(chan model.UserResponse, 1)
	env := newRunManagerTestEnv(t, runManagerTestDeps{
		prepare: func(context.Context, *model.RuntimeConfig, model.RunScope) (*model.SolvePreparation, error) {
			return &model.SolvePreparation{}, nil
		},
		execute: func(ctx context.Context, _ *model.SolvePreparation, cfg *model.RuntimeConfig) error {
			go func() {
				resp, _ := cfg.InputCoordinator.Await(
					ctx, model.PendingInput{ID: "p1", Kind: model.PendingInputKindQuestion},
				)
				gotResp <- resp
			}()
			started <- cfg.RunID
			<-ctx.Done()
			return ctx.Err()
		},
	})

	run := env.startTaskRun(t, "task-run-input", nil)
	waitForString(t, started, run.RunID)

	pollUntil(t, func() bool {
		return env.manager.SendInput(
			context.Background(), run.RunID, apicore.RunInput{PromptID: "p1", Text: "the answer"},
		) == nil
	})

	select {
	case resp := <-gotResp:
		if resp.Text != "the answer" {
			t.Fatalf("awaiting goroutine got %+v, want text 'the answer'", resp)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("awaiting goroutine never received the submitted response")
	}
}

func TestRunManagerSendInputRejectsUnknownAndNotAwaitingRuns(t *testing.T) {
	started := make(chan string, 1)
	env := newRunManagerTestEnv(t, runManagerTestDeps{
		prepare: func(context.Context, *model.RuntimeConfig, model.RunScope) (*model.SolvePreparation, error) {
			return &model.SolvePreparation{}, nil
		},
		execute: func(ctx context.Context, _ *model.SolvePreparation, cfg *model.RuntimeConfig) error {
			started <- cfg.RunID
			<-ctx.Done()
			return ctx.Err()
		},
	})

	if err := env.manager.SendInput(
		context.Background(), "unknown-run", apicore.RunInput{PromptID: "x"},
	); err == nil {
		t.Fatal("expected an error sending input to an unknown run")
	}

	run := env.startTaskRun(t, "task-run-not-awaiting", nil)
	waitForString(t, started, run.RunID)

	err := env.manager.SendInput(
		context.Background(), run.RunID, apicore.RunInput{PromptID: "x"},
	)
	if !errors.Is(err, ErrRunNotAwaitingInput) {
		t.Fatalf("SendInput on a non-awaiting run = %v, want ErrRunNotAwaitingInput", err)
	}
}
