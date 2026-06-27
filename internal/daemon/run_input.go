package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	apicore "github.com/rodolfochicone/rc-project/internal/api/core"
	"github.com/rodolfochicone/rc-project/internal/core/model"
	eventspkg "github.com/rodolfochicone/rc-project/pkg/rc/events"
	"github.com/rodolfochicone/rc-project/pkg/rc/events/kinds"
)

// ErrRunNotAwaitingInput indicates a SendInput call targeted a run that is not
// currently waiting for user input (terminal, not active, or no outstanding
// prompt). It is re-exported from the api/core package so the HTTP handler can
// map it to 409 without importing this package (which would be a cycle).
var ErrRunNotAwaitingInput = apicore.ErrRunNotAwaitingInput

func runInputToUserResponse(input apicore.RunInput) model.UserResponse {
	return model.UserResponse{
		PromptID: input.PromptID,
		OptionID: input.OptionID,
		Text:     input.Text,
		Canceled: input.Canceled,
	}
}

// inputCoordinator is the per-run channel mailbox that brokers user responses
// between the HTTP layer (Submit) and a paused agent turn or permission callback
// (Await). One waiter may be registered per PendingInput.ID at a time.
//
// When constructed with a run journal it also emits a session.awaiting_input
// event on Await and tracks the outstanding prompt so the run snapshot can
// surface pending_input. The journal is optional: a zero-value coordinator is a
// pure mailbox with no event emission.
type inputCoordinator struct {
	runID   string
	journal submitter
	logger  *slog.Logger

	mu      sync.Mutex
	waiters map[string]chan model.UserResponse
	pending *model.PendingInput
}

var _ model.InputCoordinator = (*inputCoordinator)(nil)

// newInputCoordinator returns an empty pure-mailbox coordinator (no event
// emission).
func newInputCoordinator() *inputCoordinator {
	return &inputCoordinator{waiters: make(map[string]chan model.UserResponse)}
}

// newRunInputCoordinator returns a coordinator that emits session.awaiting_input
// events to the run journal and tracks the outstanding prompt for the snapshot.
func newRunInputCoordinator(runID string, journal submitter, logger *slog.Logger) *inputCoordinator {
	c := newInputCoordinator()
	c.runID = runID
	c.journal = journal
	c.logger = logger
	return c
}

// Await registers a waiter for prompt.ID, emits the awaiting-input event, and
// blocks until a matching response is submitted or ctx is done (returning
// ctx.Err()). The waiter and pending prompt are always cleared before returning.
// It errors if a waiter for prompt.ID already exists.
func (c *inputCoordinator) Await(
	ctx context.Context,
	prompt model.PendingInput,
) (model.UserResponse, error) {
	ch := make(chan model.UserResponse, 1)

	c.mu.Lock()
	if _, exists := c.waiters[prompt.ID]; exists {
		c.mu.Unlock()
		return model.UserResponse{}, fmt.Errorf(
			"input coordinator: prompt %q is already awaiting input", prompt.ID,
		)
	}
	c.waiters[prompt.ID] = ch
	pendingCopy := prompt
	c.pending = &pendingCopy
	c.mu.Unlock()

	c.emitAwaiting(ctx, prompt)
	defer c.resolve(prompt.ID)

	select {
	case <-ctx.Done():
		return model.UserResponse{}, ctx.Err()
	case resp := <-ch:
		return resp, nil
	}
}

// Submit delivers resp to the waiter registered for resp.PromptID. It removes the
// waiter, never blocks, and returns an error when no waiter is registered.
func (c *inputCoordinator) Submit(resp model.UserResponse) error {
	c.mu.Lock()
	ch, ok := c.waiters[resp.PromptID]
	if ok {
		delete(c.waiters, resp.PromptID)
	}
	c.mu.Unlock()

	if !ok {
		return fmt.Errorf("input coordinator: no prompt %q is awaiting input", resp.PromptID)
	}
	// ch is buffered with capacity 1 and owned by exactly one waiter, so this
	// send never blocks. The pending prompt is cleared by the Await side.
	ch <- resp
	return nil
}

// PendingInput returns a copy of the prompt the run is currently awaiting, or nil
// when no prompt is outstanding. It is read by the run snapshot builder.
func (c *inputCoordinator) PendingInput() *model.PendingInput {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.pending == nil {
		return nil
	}
	pendingCopy := *c.pending
	return &pendingCopy
}

func (c *inputCoordinator) resolve(id string) {
	c.mu.Lock()
	delete(c.waiters, id)
	if c.pending != nil && c.pending.ID == id {
		c.pending = nil
	}
	c.mu.Unlock()
}

func (c *inputCoordinator) emitAwaiting(ctx context.Context, prompt model.PendingInput) {
	if c.journal == nil {
		return
	}
	payload := kinds.SessionAwaitingInputPayload{
		Kind:     prompt.Kind,
		PromptID: prompt.ID,
		Text:     prompt.Text,
		Options:  pendingOptionsToKinds(prompt.Options),
	}
	if err := submitSyntheticEvent(
		ctx, c.journal, c.runID, eventspkg.EventKindSessionAwaitingInput, payload,
	); err != nil && c.logger != nil {
		c.logger.Warn(
			"failed to emit awaiting-input event",
			"run_id", c.runID,
			"prompt_id", prompt.ID,
			"error", err,
		)
	}
}

func pendingOptionsToKinds(options []model.InputOption) []kinds.SessionInputOption {
	if len(options) == 0 {
		return nil
	}
	mapped := make([]kinds.SessionInputOption, 0, len(options))
	for _, option := range options {
		mapped = append(mapped, kinds.SessionInputOption{
			OptionID: option.OptionID,
			Label:    option.Label,
		})
	}
	return mapped
}
