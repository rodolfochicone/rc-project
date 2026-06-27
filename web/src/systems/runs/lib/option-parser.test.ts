import { describe, expect, it } from "vitest";

import { parseQuestionOptions } from "./option-parser";

describe("parseQuestionOptions", () => {
  it("Should parse inline A/B/C options in order", () => {
    const result = parseQuestionOptions("A) Keep B) Plan C) Both");
    expect(result.options).toEqual([
      { value: "A", label: "Keep" },
      { value: "B", label: "Plan" },
      { value: "C", label: "Both" },
    ]);
  });

  it("Should keep the question text as the remainder", () => {
    const result = parseQuestionOptions("Which approach? A) Keep B) Plan");
    expect(result.remainder).toBe("Which approach?");
    expect(result.options.map(option => option.label)).toEqual(["Keep", "Plan"]);
  });

  it("Should parse newline-separated options with period markers", () => {
    const result = parseQuestionOptions("Pick one:\nA. First\nB. Second");
    expect(result.options).toEqual([
      { value: "A", label: "First" },
      { value: "B", label: "Second" },
    ]);
  });

  it("Should return no options for free-form text", () => {
    const result = parseQuestionOptions("Please describe what you want to do.");
    expect(result.options).toHaveLength(0);
    expect(result.remainder).toBe("Please describe what you want to do.");
  });

  it("Should treat a single lone marker as prose, not an option list", () => {
    const result = parseQuestionOptions("See item A) referenced above.");
    expect(result.options).toHaveLength(0);
  });

  it("Should return empty results for empty input", () => {
    expect(parseQuestionOptions("")).toEqual({ options: [], remainder: "" });
  });
});
