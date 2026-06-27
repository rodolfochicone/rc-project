import { isTerminalKind, type RunFeedEvent } from "./event-store";
import type { RunInputOption, RunPendingInput } from "../types";

const AWAITING_INPUT_KIND = "session.awaiting_input";
const SESSION_UPDATE_KIND = "session.update";

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function stringValue(value: unknown): string {
  return typeof value === "string" ? value : "";
}

/**
 * parseAwaitingInput maps a `session.awaiting_input` event payload onto the same
 * shape as the snapshot's `pending_input` so both sources drive one panel. It
 * returns null when the payload is missing the prompt correlation id.
 */
export function parseAwaitingInput(payload: unknown): RunPendingInput | null {
  if (!isRecord(payload)) {
    return null;
  }
  const promptId = stringValue(payload.prompt_id);
  if (!promptId) {
    return null;
  }

  const options: RunInputOption[] = Array.isArray(payload.options)
    ? payload.options.flatMap(option => {
        if (!isRecord(option)) {
          return [];
        }
        const optionId = stringValue(option.option_id);
        if (!optionId) {
          return [];
        }
        const label = stringValue(option.label);
        return [label ? { option_id: optionId, label } : { option_id: optionId }];
      })
    : [];

  const text = stringValue(payload.text);
  return {
    prompt_id: promptId,
    kind: stringValue(payload.kind),
    ...(text ? { text } : {}),
    ...(options.length > 0 ? { options } : {}),
  };
}

/**
 * resolvePendingInput derives the prompt a run is currently awaiting from the
 * live event stream, falling back to the snapshot field. Live `awaiting_input`
 * events take precedence and are superseded by any later `session.update` or a
 * terminal run event (ADR-003: clear on the next update / termination). The
 * snapshot value is only used when the stream carries no awaiting signal yet,
 * which covers a paused run opened fresh before its stream replays.
 */
export function resolvePendingInput(
  snapshotPending: RunPendingInput | null | undefined,
  liveEvents: readonly RunFeedEvent[]
): RunPendingInput | null {
  let live: RunPendingInput | null = null;
  let sawAwaiting = false;
  let terminated = false;

  for (const event of liveEvents) {
    if (event.kind === AWAITING_INPUT_KIND) {
      const parsed = parseAwaitingInput(event.payload);
      if (parsed) {
        live = parsed;
        sawAwaiting = true;
        terminated = false;
      }
    } else if (isTerminalKind(event.kind)) {
      live = null;
      terminated = true;
    } else if (event.kind === SESSION_UPDATE_KIND && live) {
      live = null;
    }
  }

  if (terminated) {
    return null;
  }
  if (sawAwaiting) {
    return live;
  }
  return snapshotPending ?? null;
}
