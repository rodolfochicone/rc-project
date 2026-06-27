import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";
import { describe, expect, it, vi } from "vitest";

import { RunDetailView } from "./run-detail-view";
import type { RunFeedEvent } from "../lib/event-store";
import type { RunInputRequest, RunPendingInput, RunSnapshot, RunTranscript } from "../types";
import type { RunStreamStatus } from "../hooks/use-run-stream";

function buildSnapshot(overrides: Partial<RunSnapshot> = {}): RunSnapshot {
  return {
    run: {
      run_id: "run-1",
      mode: "task",
      presentation_mode: "text",
      workspace_id: "ws-1",
      workflow_slug: "alpha",
      status: "running",
      started_at: "2026-01-01T00:00:00Z",
    },
    jobs: [
      {
        index: 0,
        job_id: "job-1",
        status: "running",
        updated_at: "2026-01-01T00:01:00Z",
      },
    ],
    transcript: [
      {
        content: "hello",
        role: "assistant",
        sequence: 1,
        stream: "stdout",
        timestamp: "2026-01-01T00:01:30Z",
      },
    ],
    usage: { input_tokens: 12, output_tokens: 7, total_tokens: 19 },
    ...overrides,
  };
}

function buildTranscript(overrides: Partial<RunTranscript> = {}): RunTranscript {
  return {
    run_id: "run-1",
    messages: [
      {
        id: "msg-1",
        role: "assistant",
        parts: [
          { type: "text", text: "hello" },
          {
            type: "dynamic-tool",
            toolCallId: "tool-1",
            toolName: "Bash",
            state: "output-available",
            input: { command: "echo ok" },
            output: { blocks: [{ type: "text", text: "ok" }] },
          },
        ],
      },
    ],
    ...overrides,
  } as RunTranscript;
}

interface RenderProps {
  snapshot?: RunSnapshot;
  streamStatus?: RunStreamStatus;
  streamEventCount?: number;
  lastHeartbeatAt?: number | null;
  overflowReason?: string | null;
  streamError?: string | null;
  cancelDisabled?: boolean;
  isCancelling?: boolean;
  cancelError?: string | null;
  cancelSuccess?: string | null;
  onReconnectStream?: () => void;
  onCancelRun?: () => void;
  onSendInput?: (input: RunInputRequest) => void;
  isSendingInput?: boolean;
  sendInputError?: string | null;
  liveEvents?: readonly RunFeedEvent[];
  isRefreshingSnapshot?: boolean;
  transcript?: RunTranscript;
  isLoadingTranscript?: boolean;
  isTranscriptError?: boolean;
  transcriptError?: string | null;
}

async function renderRunDetail(props: RenderProps = {}) {
  const rootRoute = createRootRoute();
  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/",
    component: function IndexRouteComponent(): ReactElement {
      return (
        <RunDetailView
          cancelDisabled={props.cancelDisabled ?? false}
          cancelError={props.cancelError ?? null}
          cancelSuccess={props.cancelSuccess ?? null}
          isCancelling={props.isCancelling ?? false}
          isRefreshingSnapshot={props.isRefreshingSnapshot ?? false}
          lastHeartbeatAt={props.lastHeartbeatAt ?? null}
          isSendingInput={props.isSendingInput ?? false}
          liveEvents={props.liveEvents ?? []}
          onCancelRun={props.onCancelRun ?? (() => {})}
          onReconnectStream={props.onReconnectStream ?? (() => {})}
          onSendInput={props.onSendInput ?? (() => {})}
          sendInputError={props.sendInputError ?? null}
          overflowReason={props.overflowReason ?? null}
          snapshot={props.snapshot ?? buildSnapshot()}
          streamError={props.streamError ?? null}
          streamEventCount={props.streamEventCount ?? 0}
          streamStatus={props.streamStatus ?? "connecting"}
          transcript={props.transcript ?? buildTranscript()}
          transcriptError={props.transcriptError ?? null}
          isLoadingTranscript={props.isLoadingTranscript ?? false}
          isTranscriptError={props.isTranscriptError ?? false}
        />
      );
    },
  });
  const workflowsRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/workflows",
    component: function WorkflowsRouteComponent(): ReactElement {
      return <div data-testid="workflows-stub" />;
    },
  });
  const router = createRouter({
    routeTree: rootRoute.addChildren([indexRoute, workflowsRoute]),
    history: createMemoryHistory({ initialEntries: ["/"] }),
    defaultPreload: false,
  });
  await router.load();
  render(<RouterProvider router={router} />);
  await act(async () => {
    await vi.dynamicImportSettled();
    await Promise.resolve();
  });
}

describe("RunDetailView", () => {
  it("Should render status badge, jobs, and transcript", async () => {
    await renderRunDetail();
    expect(screen.getByTestId("run-detail-view")).toBeInTheDocument();
    expect(screen.getByTestId("run-detail-status")).toHaveTextContent("running");
    expect(screen.getByTestId("run-detail-job-row-job-1")).toBeInTheDocument();
    expect(await screen.findByTestId("run-detail-transcript")).toHaveTextContent("hello");
    expect(await screen.findByTestId("run-transcript-tool-tool-1")).toHaveTextContent("Bash");
  });

  it("Should show the empty transcript state", async () => {
    await renderRunDetail({ transcript: buildTranscript({ messages: [] }) });
    expect(screen.getByTestId("run-detail-transcript-empty")).toBeInTheDocument();
  });

  it("Should show the stream status badge", async () => {
    await renderRunDetail({ streamStatus: "open", streamEventCount: 3 });
    expect(screen.getByTestId("run-detail-stream-status")).toHaveTextContent("stream open");
    expect(screen.getByTestId("run-detail-stream-events")).toHaveTextContent("3");
  });

  it("Should show the overflow notice when the stream overflowed", async () => {
    await renderRunDetail({
      streamStatus: "overflowed",
      overflowReason: "replay boundary exceeded",
    });
    expect(screen.getByTestId("run-detail-stream-overflow")).toHaveTextContent(
      "replay boundary exceeded"
    );
  });

  it("Should show a stream error when provided", async () => {
    await renderRunDetail({ streamError: "disconnected" });
    expect(screen.getByTestId("run-detail-stream-error")).toHaveTextContent("disconnected");
  });

  it("Should call the reconnect and cancel handlers", async () => {
    const onReconnectStream = vi.fn();
    const onCancelRun = vi.fn();
    await renderRunDetail({ onReconnectStream, onCancelRun });
    await userEvent.click(screen.getByTestId("run-detail-reconnect"));
    await userEvent.click(screen.getByTestId("run-detail-cancel"));
    expect(onReconnectStream).toHaveBeenCalledTimes(1);
    expect(onCancelRun).toHaveBeenCalledTimes(1);
  });

  it("Should disable the cancel action when requested", async () => {
    await renderRunDetail({ cancelDisabled: true });
    expect(screen.getByTestId("run-detail-cancel")).toBeDisabled();
  });

  it("Should render the cancel success banner", async () => {
    await renderRunDetail({ cancelSuccess: "Cancellation requested" });
    expect(screen.getByTestId("run-detail-cancel-success")).toHaveTextContent(
      "Cancellation requested"
    );
  });

  it("Should render the cancel error banner", async () => {
    await renderRunDetail({ cancelError: "could not cancel" });
    expect(screen.getByTestId("run-detail-cancel-error")).toHaveTextContent("could not cancel");
  });

  it("Should show a snapshot refresh indicator when provided", async () => {
    await renderRunDetail({ isRefreshingSnapshot: true });
    expect(screen.getByTestId("run-detail-jobs-refreshing")).toBeInTheDocument();
  });

  it("Should not render the input panel when the run is not awaiting input", async () => {
    await renderRunDetail();
    expect(screen.queryByTestId("run-detail-input")).not.toBeInTheDocument();
  });

  it("Should hide the input panel on a terminated run even with a stale pending input", async () => {
    const pendingInput: RunPendingInput = { prompt_id: "p1", kind: "question", text: "Continue?" };
    await renderRunDetail({
      snapshot: buildSnapshot({
        run: {
          run_id: "run-1",
          mode: "task",
          presentation_mode: "text",
          workspace_id: "ws-1",
          status: "completed",
          started_at: "2026-01-01T00:00:00Z",
        },
        pending_input: pendingInput,
      }),
    });
    expect(screen.queryByTestId("run-detail-input")).not.toBeInTheDocument();
  });

  it("Should render permission options from the snapshot pending input", async () => {
    const pendingInput: RunPendingInput = {
      prompt_id: "p1",
      kind: "permission",
      text: "Allow writing to disk?",
      options: [
        { option_id: "allow_once", label: "Allow once" },
        { option_id: "reject", label: "Reject" },
      ],
    };
    await renderRunDetail({ snapshot: buildSnapshot({ pending_input: pendingInput }) });
    expect(screen.getByTestId("run-detail-input")).toBeInTheDocument();
    expect(screen.getByTestId("run-detail-input-option-allow_once")).toHaveTextContent(
      "Allow once"
    );
    // The free-text box is always available even when option buttons render.
    expect(screen.getByTestId("run-detail-input-text")).toBeInTheDocument();
  });

  it("Should submit the matching option id when a permission option is clicked", async () => {
    const onSendInput = vi.fn();
    const pendingInput: RunPendingInput = {
      prompt_id: "p1",
      kind: "permission",
      options: [
        { option_id: "allow_once", label: "Allow once" },
        { option_id: "reject", label: "Reject" },
      ],
    };
    await renderRunDetail({
      snapshot: buildSnapshot({ pending_input: pendingInput }),
      onSendInput,
    });
    await userEvent.click(screen.getByTestId("run-detail-input-option-reject"));
    expect(onSendInput).toHaveBeenCalledWith({ prompt_id: "p1", option_id: "reject" });
  });

  it("Should render the input panel from a live awaiting-input event", async () => {
    const awaitingEvent: RunFeedEvent = {
      id: "evt-1",
      seq: 1,
      kind: "session.awaiting_input",
      runId: "run-1",
      timestamp: "2026-01-01T00:02:00Z",
      payload: { prompt_id: "p2", kind: "question", text: "Which approach? A) Keep B) Plan" },
      receivedAt: 1,
    };
    await renderRunDetail({ liveEvents: [awaitingEvent] });
    expect(screen.getByTestId("run-detail-input")).toBeInTheDocument();
    expect(screen.getByTestId("run-detail-input-option-A")).toHaveTextContent("Keep");
    expect(screen.getByTestId("run-detail-input-option-B")).toHaveTextContent("Plan");
  });

  it("Should clear the input panel once a later session update supersedes the prompt", async () => {
    const awaitingEvent: RunFeedEvent = {
      id: "evt-1",
      seq: 1,
      kind: "session.awaiting_input",
      runId: "run-1",
      timestamp: "2026-01-01T00:02:00Z",
      payload: { prompt_id: "p2", kind: "question", text: "Continue?" },
      receivedAt: 1,
    };
    const updateEvent: RunFeedEvent = {
      id: "evt-2",
      seq: 2,
      kind: "session.update",
      runId: "run-1",
      timestamp: "2026-01-01T00:02:30Z",
      payload: { update: { kind: "agent_message_chunk", blocks: [{ type: "text", text: "ok" }] } },
      receivedAt: 2,
    };
    await renderRunDetail({ liveEvents: [awaitingEvent, updateEvent] });
    expect(screen.queryByTestId("run-detail-input")).not.toBeInTheDocument();
  });

  it("Should render a user_message_chunk answer in the transcript", async () => {
    const userChunk: RunFeedEvent = {
      id: "evt-9",
      seq: 9,
      kind: "session.update",
      runId: "run-1",
      timestamp: "2026-01-01T00:03:00Z",
      payload: {
        update: { kind: "user_message_chunk", blocks: [{ type: "text", text: "Approved" }] },
      },
      receivedAt: 9,
    };
    await renderRunDetail({ liveEvents: [userChunk] });
    expect(await screen.findByTestId("run-detail-transcript")).toHaveTextContent("Approved");
  });
});
