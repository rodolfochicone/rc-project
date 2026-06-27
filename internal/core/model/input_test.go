package model_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/rodolfochicone/rc-project/internal/core/model"
)

func TestUserResponseZeroValue(t *testing.T) {
	t.Parallel()

	var resp model.UserResponse
	if resp.Canceled {
		t.Errorf("zero UserResponse.Canceled = true, want false")
	}
	if resp.OptionID != "" {
		t.Errorf("zero UserResponse.OptionID = %q, want empty", resp.OptionID)
	}
	if resp.Text != "" {
		t.Errorf("zero UserResponse.Text = %q, want empty", resp.Text)
	}
	if resp.PromptID != "" {
		t.Errorf("zero UserResponse.PromptID = %q, want empty", resp.PromptID)
	}
}

func TestPendingInputRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input model.PendingInput
	}{
		{
			name: "permission with options",
			input: model.PendingInput{
				ID:   "prompt-1",
				Kind: model.PendingInputKindPermission,
				Text: "Run `make verify`?",
				Options: []model.InputOption{
					{OptionID: "allow_once", Label: "Allow once"},
					{OptionID: "reject", Label: "Reject"},
				},
			},
		},
		{
			name: "question without options",
			input: model.PendingInput{
				ID:   "prompt-2",
				Kind: model.PendingInputKindQuestion,
				Text: "Which problem should this solve first?",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			copied := tc.input
			if !reflect.DeepEqual(copied, tc.input) {
				t.Errorf("PendingInput did not round-trip: got %+v, want %+v", copied, tc.input)
			}
		})
	}
}

func TestRuntimeConfigNonInteractiveByDefault(t *testing.T) {
	t.Parallel()

	cfg := model.RuntimeConfig{}
	cfg.ApplyDefaults()

	if cfg.Interactive {
		t.Errorf("default RuntimeConfig.Interactive = true, want false")
	}
	if cfg.InputCoordinator != nil {
		t.Errorf("default RuntimeConfig.InputCoordinator = %v, want nil", cfg.InputCoordinator)
	}
}

func TestRuntimeConfigCarriesCoordinator(t *testing.T) {
	t.Parallel()

	coord := &stubCoordinator{}
	cfg := model.RuntimeConfig{Interactive: true, InputCoordinator: coord}
	cfg.ApplyDefaults()

	if !cfg.Interactive {
		t.Errorf("RuntimeConfig.Interactive = false, want true after explicit set")
	}
	if cfg.InputCoordinator != coord {
		t.Errorf("RuntimeConfig.InputCoordinator was not preserved through ApplyDefaults")
	}
}

// stubCoordinator proves the InputCoordinator interface is satisfiable and that
// RuntimeConfig can carry an implementation without import cycles. The concrete,
// behavior-bearing coordinator is implemented and tested separately (task_03).
type stubCoordinator struct{}

var _ model.InputCoordinator = (*stubCoordinator)(nil)

func (*stubCoordinator) Await(context.Context, model.PendingInput) (model.UserResponse, error) {
	return model.UserResponse{}, nil
}

func (*stubCoordinator) Submit(model.UserResponse) error { return nil }
