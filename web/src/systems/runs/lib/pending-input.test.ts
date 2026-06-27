import { describe, expect, it } from "vitest";

import type { RunFeedEvent } from "./event-store";
import { parseAwaitingInput, resolvePendingInput } from "./pending-input";
import type { RunPendingInput } from "../types";

function event(kind: string, payload: unknown, seq: number): RunFeedEvent {
  return {
    id: `evt-${seq}`,
    seq,
    kind,
    runId: "run-1",
    timestamp: null,
    payload,
    receivedAt: seq,
  };
}

describe("parseAwaitingInput", () => {
  it("Should map a permission payload with options", () => {
    const parsed = parseAwaitingInput({
      prompt_id: "p1",
      kind: "permission",
      text: "Allow?",
      options: [{ option_id: "allow", label: "Allow once" }, { option_id: "reject" }],
    });
    expect(parsed).toEqual({
      prompt_id: "p1",
      kind: "permission",
      text: "Allow?",
      options: [{ option_id: "allow", label: "Allow once" }, { option_id: "reject" }],
    });
  });

  it("Should return null without a prompt id", () => {
    expect(parseAwaitingInput({ kind: "question", text: "hi" })).toBeNull();
    expect(parseAwaitingInput(null)).toBeNull();
  });
});

describe("resolvePendingInput", () => {
  const snapshotPending: RunPendingInput = { prompt_id: "snap", kind: "question", text: "snap?" };

  it("Should fall back to the snapshot field when no awaiting event is present", () => {
    expect(resolvePendingInput(snapshotPending, [])).toEqual(snapshotPending);
    expect(
      resolvePendingInput(snapshotPending, [event("session.update", { update: {} }, 1)])
    ).toEqual(snapshotPending);
  });

  it("Should prefer a live awaiting event over the snapshot field", () => {
    const live = [event("session.awaiting_input", { prompt_id: "live", kind: "question" }, 1)];
    expect(resolvePendingInput(snapshotPending, live)?.prompt_id).toBe("live");
  });

  it("Should clear the prompt when a later session update supersedes it", () => {
    const events = [
      event("session.awaiting_input", { prompt_id: "live", kind: "question" }, 1),
      event("session.update", { update: { kind: "agent_message_chunk" } }, 2),
    ];
    expect(resolvePendingInput(snapshotPending, events)).toBeNull();
  });

  it("Should clear the prompt on a terminal run event", () => {
    const events = [
      event("session.awaiting_input", { prompt_id: "live", kind: "question" }, 1),
      event("run.completed", {}, 2),
    ];
    expect(resolvePendingInput(snapshotPending, events)).toBeNull();
  });

  it("Should surface the latest of several awaiting prompts", () => {
    const events = [
      event("session.awaiting_input", { prompt_id: "first", kind: "question" }, 1),
      event("session.awaiting_input", { prompt_id: "second", kind: "permission" }, 2),
    ];
    expect(resolvePendingInput(snapshotPending, events)?.prompt_id).toBe("second");
  });
});
