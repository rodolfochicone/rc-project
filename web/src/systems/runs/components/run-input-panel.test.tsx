import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { RunInputPanel } from "./run-input-panel";
import type { RunPendingInput } from "../types";

const question: RunPendingInput = {
  prompt_id: "p1",
  kind: "question",
  text: "Which approach? A) Keep B) Plan",
};

describe("RunInputPanel", () => {
  it("Should render parsed question options and an always-available text box", () => {
    render(<RunInputPanel isSubmitting={false} onSubmit={vi.fn()} pendingInput={question} />);
    expect(screen.getByTestId("run-detail-input-prompt")).toHaveTextContent("Which approach?");
    expect(screen.getByTestId("run-detail-input-option-A")).toHaveTextContent("Keep");
    expect(screen.getByTestId("run-detail-input-text")).toBeInTheDocument();
  });

  it("Should submit the option letter as text when a parsed option is clicked", async () => {
    const onSubmit = vi.fn();
    render(<RunInputPanel isSubmitting={false} onSubmit={onSubmit} pendingInput={question} />);
    await userEvent.click(screen.getByTestId("run-detail-input-option-B"));
    expect(onSubmit).toHaveBeenCalledWith({ prompt_id: "p1", text: "B" });
  });

  it("Should submit free text and clear it through the mutation", async () => {
    const onSubmit = vi.fn();
    render(<RunInputPanel isSubmitting={false} onSubmit={onSubmit} pendingInput={question} />);
    await userEvent.type(screen.getByTestId("run-detail-input-text"), "do it manually");
    await userEvent.click(screen.getByTestId("run-detail-input-submit"));
    expect(onSubmit).toHaveBeenCalledWith({ prompt_id: "p1", text: "do it manually" });
  });

  it("Should disable the controls and not submit while a response is in flight", async () => {
    const onSubmit = vi.fn();
    render(<RunInputPanel isSubmitting onSubmit={onSubmit} pendingInput={question} />);
    expect(screen.getByTestId("run-detail-input-text")).toBeDisabled();
    expect(screen.getByTestId("run-detail-input-submit")).toBeDisabled();
    await userEvent.click(screen.getByTestId("run-detail-input-option-A"));
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("Should reset the text box when keyed by a new prompt id, as the detail view renders it", async () => {
    // RunDetailView renders <RunInputPanel key={pendingInput.prompt_id} />, so a
    // new prompt remounts the panel and clears any draft from the prior prompt.
    function Harness({ pending }: { pending: RunPendingInput }) {
      return (
        <RunInputPanel
          isSubmitting={false}
          key={pending.prompt_id}
          onSubmit={vi.fn()}
          pendingInput={pending}
        />
      );
    }
    const { rerender } = render(
      <Harness pending={{ prompt_id: "pA", kind: "question", text: "First?" }} />
    );
    await userEvent.type(screen.getByTestId("run-detail-input-text"), "draft answer");
    expect(screen.getByTestId("run-detail-input-text")).toHaveValue("draft answer");

    rerender(<Harness pending={{ prompt_id: "pB", kind: "question", text: "Second?" }} />);
    expect(screen.getByTestId("run-detail-input-text")).toHaveValue("");
  });

  it("Should surface a submission error", () => {
    render(
      <RunInputPanel
        error="run is not awaiting input"
        isSubmitting={false}
        onSubmit={vi.fn()}
        pendingInput={question}
      />
    );
    expect(screen.getByTestId("run-detail-input-error")).toHaveTextContent(
      "run is not awaiting input"
    );
  });
});
