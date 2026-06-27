package daemon

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/rodolfochicone/rc-project/internal/core/model"
)

// submitWhenReady polls Submit until the waiter for resp.PromptID is registered,
// yielding the scheduler between attempts. It avoids arbitrary sleeps while
// keeping the concurrent Await/Submit test deterministic.
func submitWhenReady(t *testing.T, c *inputCoordinator, resp model.UserResponse) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if err := c.Submit(resp); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("Submit never found a waiter for %q", resp.PromptID)
		}
		runtime.Gosched()
	}
}

func TestInputCoordinatorAwaitReturnsSubmittedResponse(t *testing.T) {
	t.Parallel()

	c := newInputCoordinator()
	prompt := model.PendingInput{ID: "p1", Kind: model.PendingInputKindPermission}
	want := model.UserResponse{PromptID: "p1", OptionID: "allow_once"}

	type result struct {
		resp model.UserResponse
		err  error
	}
	done := make(chan result, 1)
	go func() {
		resp, err := c.Await(context.Background(), prompt)
		done <- result{resp, err}
	}()

	submitWhenReady(t, c, want)

	got := <-done
	if got.err != nil {
		t.Fatalf("Await returned error: %v", got.err)
	}
	if got.resp != want {
		t.Fatalf("Await response = %+v, want %+v", got.resp, want)
	}
}

func TestInputCoordinatorAwaitHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	c := newInputCoordinator()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Await(ctx, model.PendingInput{ID: "p2"})
	if err == nil {
		t.Fatal("Await returned nil error after context cancellation, want context.Canceled")
	}
	// After cancellation the waiter must be cleaned up, so a late Submit fails.
	if subErr := c.Submit(model.UserResponse{PromptID: "p2"}); subErr == nil {
		t.Fatal("Submit succeeded after the awaiter was canceled, want no-waiter error")
	}
}

func TestInputCoordinatorSubmitWithoutWaiterErrors(t *testing.T) {
	t.Parallel()

	c := newInputCoordinator()
	if err := c.Submit(model.UserResponse{PromptID: "missing"}); err == nil {
		t.Fatal("Submit for an unregistered prompt returned nil, want descriptive error")
	}
}

func TestInputCoordinatorSecondSubmitErrorsAfterResolution(t *testing.T) {
	t.Parallel()

	c := newInputCoordinator()
	prompt := model.PendingInput{ID: "p3"}
	resp := model.UserResponse{PromptID: "p3", Text: "answer"}

	done := make(chan struct{})
	go func() {
		_, _ = c.Await(context.Background(), prompt)
		close(done)
	}()

	submitWhenReady(t, c, resp)
	<-done

	if err := c.Submit(resp); err == nil {
		t.Fatal("second Submit after resolution returned nil, want no-waiter error")
	}
}

func TestInputCoordinatorRejectsDuplicateWaiter(t *testing.T) {
	t.Parallel()

	c := newInputCoordinator()
	prompt := model.PendingInput{ID: "p4"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	registered := make(chan struct{})
	go func() {
		close(registered)
		_, _ = c.Await(ctx, prompt)
	}()
	<-registered

	// The first Await may not have registered yet; poll until a duplicate Await
	// is rejected, proving exactly one waiter per id is allowed.
	deadline := time.Now().Add(2 * time.Second)
	for {
		_, err := c.Await(ctx, prompt)
		if err != nil {
			return // duplicate correctly rejected
		}
		if time.Now().After(deadline) {
			t.Fatal("duplicate Await was never rejected")
		}
		runtime.Gosched()
	}
}
