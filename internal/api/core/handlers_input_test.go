package core_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/rodolfochicone/rc-project/internal/api/core"
	"github.com/rodolfochicone/rc-project/internal/store/globaldb"
)

func newInputTestEngine(t *testing.T, runs core.RunService) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	handlers := core.NewHandlers(&core.HandlerConfig{
		TransportName: "test",
		Runs:          runs,
	})
	engine := gin.New()
	engine.Use(core.RequestIDMiddleware())
	engine.Use(core.ErrorMiddleware())
	core.RegisterRoutes(engine, handlers)
	return engine
}

func postRunInput(t *testing.T, engine *gin.Engine, runID, body string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/runs/"+runID+"/input",
		bytes.NewBufferString(body),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response
}

func TestSendInputDeliversAnswerAndReturns202(t *testing.T) {
	t.Parallel()

	runs := &fakeRunService{}
	engine := newInputTestEngine(t, runs)

	response := postRunInput(t, engine, "run-1", `{"prompt_id":"p1","option_id":"opt-a"}`)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusAccepted)
	}
	if runs.sendInputCalls != 1 {
		t.Fatalf("SendInput called %d times, want 1", runs.sendInputCalls)
	}
	if runs.sendInputRunID != "run-1" {
		t.Fatalf("SendInput run id = %q, want run-1", runs.sendInputRunID)
	}
	if runs.sendInputArg.PromptID != "p1" || runs.sendInputArg.OptionID != "opt-a" {
		t.Fatalf("SendInput arg = %+v, want prompt_id=p1 option_id=opt-a", runs.sendInputArg)
	}
}

func TestSendInputRejectsMissingPromptIDBeforeService(t *testing.T) {
	t.Parallel()

	runs := &fakeRunService{}
	engine := newInputTestEngine(t, runs)

	// option_id is present, but prompt_id (the correlation id) is missing.
	response := postRunInput(t, engine, "run-1", `{"option_id":"opt-a"}`)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if runs.sendInputCalls != 0 {
		t.Fatalf("SendInput called %d times, want 0 (validation must precede the service)", runs.sendInputCalls)
	}
}

func TestSendInputRejectsBodyWithoutAnswerBeforeService(t *testing.T) {
	t.Parallel()

	runs := &fakeRunService{}
	engine := newInputTestEngine(t, runs)

	// prompt_id is present but none of option_id, text, or canceled is set.
	response := postRunInput(t, engine, "run-1", `{"prompt_id":"p1"}`)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if runs.sendInputCalls != 0 {
		t.Fatalf("SendInput called %d times, want 0", runs.sendInputCalls)
	}
}

func TestSendInputMapsUnknownRunTo404(t *testing.T) {
	t.Parallel()

	runs := &fakeRunService{sendInputErr: globaldb.ErrRunNotFound}
	engine := newInputTestEngine(t, runs)

	response := postRunInput(t, engine, "missing", `{"prompt_id":"p1","text":"yes"}`)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestSendInputMapsNotAwaitingTo409(t *testing.T) {
	t.Parallel()

	runs := &fakeRunService{sendInputErr: core.ErrRunNotAwaitingInput}
	engine := newInputTestEngine(t, runs)

	response := postRunInput(t, engine, "run-1", `{"prompt_id":"p1","canceled":true}`)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
}

type captureExecService struct {
	got core.ExecRequest
}

func (s *captureExecService) Start(_ context.Context, req core.ExecRequest) (core.Run, error) {
	s.got = req
	return core.Run{}, nil
}

func TestStartExecRunForwardsInteractiveFlag(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	exec := &captureExecService{}
	handlers := core.NewHandlers(&core.HandlerConfig{TransportName: "test", Exec: exec})
	engine := gin.New()
	engine.Use(core.RequestIDMiddleware())
	engine.Use(core.ErrorMiddleware())
	core.RegisterRoutes(engine, handlers)

	request := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/exec",
		bytes.NewBufferString(`{"workspace_path":"/tmp/ws","prompt":"hi","interactive":true}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if !exec.got.Interactive {
		t.Fatal("ExecRequest.Interactive = false, want true forwarded from the request body")
	}
}
