import { useState, type FormEvent, type ReactElement } from "react";

import {
  Alert,
  Button,
  Markdown,
  StatusBadge,
  SurfaceCard,
  SurfaceCardBody,
  SurfaceCardDescription,
  SurfaceCardEyebrow,
  SurfaceCardHeader,
  SurfaceCardTitle,
} from "@rodolfochicone/ui";

import { parseQuestionOptions } from "../lib/option-parser";
import type { RunInputRequest, RunPendingInput } from "../types";

const fieldClass =
  "w-full rounded-[var(--radius-md)] border border-border bg-[color:var(--surface-inset)] px-3 py-2 text-sm text-foreground transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/60";

export interface RunInputPanelProps {
  pendingInput: RunPendingInput;
  onSubmit: (input: RunInputRequest) => void;
  isSubmitting: boolean;
  error?: string | null;
  testId?: string;
}

/**
 * RunInputPanel renders the prompt a paused run is awaiting plus a response area:
 * option buttons (ACP options for a permission, or A/B/C/D parsed from a skill
 * question) and a free-text box that is always available as a fallback.
 */
export function RunInputPanel({
  pendingInput,
  onSubmit,
  isSubmitting,
  error = null,
  testId = "run-detail-input",
}: RunInputPanelProps): ReactElement {
  const [text, setText] = useState("");

  const isPermission = pendingInput.kind === "permission";
  const permissionOptions = pendingInput.options ?? [];
  const parsed = isPermission
    ? { options: [], remainder: "" }
    : parseQuestionOptions(pendingInput.text ?? "");
  const promptText = isPermission
    ? (pendingInput.text ?? "")
    : parsed.remainder || pendingInput.text || "";

  const submit = (input: RunInputRequest): void => {
    if (isSubmitting) {
      return;
    }
    onSubmit(input);
  };

  const handleTextSubmit = (event: FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    const trimmed = text.trim();
    if (!trimmed) {
      return;
    }
    submit({ prompt_id: pendingInput.prompt_id, text: trimmed });
  };

  return (
    <SurfaceCard data-testid={testId}>
      <SurfaceCardHeader>
        <div>
          <SurfaceCardEyebrow>Awaiting your input</SurfaceCardEyebrow>
          <SurfaceCardTitle>{isPermission ? "Permission requested" : "Question"}</SurfaceCardTitle>
          <SurfaceCardDescription>The run is paused until you respond.</SurfaceCardDescription>
        </div>
        <StatusBadge pulse tone="warning">
          {isPermission ? "permission" : "question"}
        </StatusBadge>
      </SurfaceCardHeader>
      <SurfaceCardBody className="space-y-4">
        {promptText ? (
          // Render the prompt as Markdown (bold, lists, headings, inline code)
          // and cap the height so a long question stays scrollable without
          // pushing the answer box off-screen.
          <div
            className="max-h-80 max-w-none overflow-y-auto pr-1 text-sm leading-6 text-foreground"
            data-testid={`${testId}-prompt`}
          >
            <Markdown>{promptText}</Markdown>
          </div>
        ) : null}

        {isPermission && permissionOptions.length > 0 ? (
          <div className="flex flex-wrap gap-2" data-testid={`${testId}-options`}>
            {permissionOptions.map(option => (
              <Button
                data-testid={`${testId}-option-${option.option_id}`}
                disabled={isSubmitting}
                key={option.option_id}
                onClick={() =>
                  submit({ prompt_id: pendingInput.prompt_id, option_id: option.option_id })
                }
                size="sm"
                variant="secondary"
              >
                {option.label ?? option.option_id}
              </Button>
            ))}
          </div>
        ) : null}

        {!isPermission && parsed.options.length > 0 ? (
          // Question options are full-sentence choices, so stack them one per row
          // (left-aligned) instead of wrapping side by side.
          <div className="flex flex-col items-start gap-2" data-testid={`${testId}-options`}>
            {parsed.options.map(option => (
              <Button
                className="text-left"
                data-testid={`${testId}-option-${option.value}`}
                disabled={isSubmitting}
                key={option.value}
                onClick={() => submit({ prompt_id: pendingInput.prompt_id, text: option.value })}
                size="sm"
                variant="secondary"
              >
                {option.value}) {option.label}
              </Button>
            ))}
          </div>
        ) : null}

        <form className="space-y-2" data-testid={`${testId}-form`} onSubmit={handleTextSubmit}>
          <label className="block text-sm font-medium text-foreground" htmlFor={`${testId}-text`}>
            Your answer
          </label>
          <textarea
            className={`${fieldClass} min-h-[5rem] resize-y`}
            data-testid={`${testId}-text`}
            disabled={isSubmitting}
            id={`${testId}-text`}
            onChange={event => setText(event.target.value)}
            placeholder="Type a response…"
            value={text}
          />
          <div className="flex justify-end">
            <Button
              data-testid={`${testId}-submit`}
              disabled={isSubmitting || text.trim().length === 0}
              loading={isSubmitting}
              size="sm"
              type="submit"
              variant="primary"
            >
              Send response
            </Button>
          </div>
        </form>

        {error ? (
          <Alert data-testid={`${testId}-error`} variant="error">
            {error}
          </Alert>
        ) : null}
      </SurfaceCardBody>
    </SurfaceCard>
  );
}
