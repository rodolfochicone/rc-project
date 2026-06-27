export interface ParsedOption {
  /** The letter the agent expects as the answer, e.g. "A". */
  value: string;
  /** The human-readable choice text, e.g. "Keep the plan". */
  label: string;
}

export interface ParsedQuestion {
  options: ParsedOption[];
  /** The question text with the labeled options stripped off. */
  remainder: string;
}

// Matches a single-letter option marker (`A)` / `B.`) at the start of the
// string or after whitespace, followed by at least one space before the label.
const OPTION_MARKER = /(?:^|\s)([A-Za-z])[).]\s+/g;

/**
 * parseQuestionOptions heuristically extracts labeled options (`A) … B) …`) from
 * a free-text skill question. It is a pure function isolated from rendering so it
 * can be exhaustively unit-tested. When fewer than two markers are present the
 * text is treated as prose and no options are returned — the caller always falls
 * back to a free-text box.
 */
export function parseQuestionOptions(text: string): ParsedQuestion {
  const trimmed = (text ?? "").trim();
  if (!trimmed) {
    return { options: [], remainder: "" };
  }

  const markers: { value: string; markerStart: number; labelStart: number }[] = [];
  const matcher = new RegExp(OPTION_MARKER);
  let match: RegExpExecArray | null;
  while ((match = matcher.exec(trimmed)) !== null) {
    const letter = match[1];
    if (!letter) {
      continue;
    }
    markers.push({
      value: letter.toUpperCase(),
      markerStart: match.index,
      labelStart: matcher.lastIndex,
    });
  }

  // A lone marker is more likely prose ("see item A) above") than an option
  // list; require at least two labeled choices before parsing them out.
  const firstMarker = markers[0];
  if (!firstMarker || markers.length < 2) {
    return { options: [], remainder: trimmed };
  }

  const options: ParsedOption[] = [];
  for (let index = 0; index < markers.length; index += 1) {
    const marker = markers[index];
    if (!marker) {
      continue;
    }
    const next = markers[index + 1];
    const labelEnd = next ? next.markerStart : trimmed.length;
    const label = trimmed.slice(marker.labelStart, labelEnd).trim();
    if (label) {
      options.push({ value: marker.value, label });
    }
  }

  return { options, remainder: trimmed.slice(0, firstMarker.markerStart).trim() };
}
