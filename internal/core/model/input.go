package model

import "context"

// Pending input kinds distinguish why a run paused for the user.
const (
	// PendingInputKindPermission marks a pause for an ACP permission request,
	// where the user picks one of the offered Options.
	PendingInputKindPermission = "permission"
	// PendingInputKindQuestion marks a pause for a free-text skill question,
	// where the user answers with Text (Options is empty).
	PendingInputKindQuestion = "question"
)

// InputCoordinator brokers user responses to a run paused waiting for input.
//
// A non-interactive run carries a nil coordinator, which callers MUST treat as
// "no interactivity" and skip without error. Implementations live outside this
// package (the per-run runtime wiring constructs one); a concrete type satisfies
// this interface with the usual assertion:
//
//	var _ model.InputCoordinator = (*coordinator)(nil)
type InputCoordinator interface {
	// Await blocks until a UserResponse whose PromptID matches prompt.ID is
	// submitted, or until ctx is done (returning ctx.Err()). It is called from
	// the agent permission callback and the executor turn loop.
	Await(ctx context.Context, prompt PendingInput) (UserResponse, error)
	// Submit delivers a user response from the HTTP layer to a waiting Await. It
	// returns a non-nil error when no prompt with resp.PromptID is currently
	// awaiting, and never blocks.
	Submit(resp UserResponse) error
}

// PendingInput describes a single point at which a run paused for the user.
type PendingInput struct {
	// ID uniquely correlates this pause with the response that resolves it.
	ID string
	// Kind is one of the PendingInputKind* constants.
	Kind string
	// Text is the permission or question prompt shown to the user.
	Text string
	// Options are the selectable choices for a permission request; it is empty
	// for a free-text question.
	Options []InputOption
}

// InputOption is one selectable choice offered to the user, mapped from an ACP
// permission option.
type InputOption struct {
	// OptionID is the stable identifier returned when the option is selected.
	OptionID string
	// Label is the human-readable text for the option.
	Label string
}

// UserResponse carries the user's answer to a PendingInput back to the run.
type UserResponse struct {
	// PromptID must match the PendingInput.ID being awaited.
	PromptID string
	// OptionID is the selected permission option, when answering with a choice.
	OptionID string
	// Text is the free-text answer, when answering a question.
	Text string
	// Canceled is true when the user declined to answer; the run treats it as a
	// canceled permission outcome or the end of the interactive loop.
	Canceled bool
}
