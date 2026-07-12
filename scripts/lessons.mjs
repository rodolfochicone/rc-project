#!/usr/bin/env node
// lessons.mjs — deterministic bookkeeping for the RC loop-lessons layer (the plugin-native
// port of the tlc-spec-driven lessons machine). The agent supplies judgment (which failure
// happened, how to phrase the lesson, what signal grounds it); this script owns everything
// mechanical: IDs, distinct-feature recurrence counting, candidate->confirmed promotion,
// pruning, demotion, and rendering the agent-readable playbook. Bookkeeping by hand is exactly
// what rots a lessons file, so it lives here (a script that must hold every time), not in prose.
//
// Canonical state:  .rc/lessons.json   (machine-owned — do NOT hand-edit)
// Rendered view:    .rc/LESSONS.md      (regenerated on every write)
//
// Dependency-free (Node stdlib only). Run from the project root (the dir that contains .rc), or
// pass --root. Run --selftest to verify the lifecycle offline.
//
// Commands:
//   add        Record a grounded lesson from a verification signal.
//   list       Print lessons (default: confirmed) for loading at plan/design time.
//   penalize   Mark a confirmed lesson as having failed when applied (-> quarantine).
//   prune      Drop stale uncorroborated candidates (also runs automatically on add/list).
//   status     Print counts.
//   init       Create empty store + rendered file.
//
// Exit codes: 0 ok, 2 usage/validation error (e.g. missing grounding).

import { readFileSync, writeFileSync, existsSync, mkdirSync, mkdtempSync, rmSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';

const STORE_REL = join('.rc', 'lessons.json');
const RENDER_REL = join('.rc', 'LESSONS.md');

const SIGNALS = {
  ac_gap: 'Acceptance criterion not covered / failed',
  surviving_mutant: 'Discrimination sensor mutant survived (weak test)',
  spec_precision_gap: 'Spec did not define a precise outcome',
  spec_deviation: 'Implementation diverged from spec/design (SPEC_DEVIATION)',
  gate_fail: 'Build-level gate check failed (build/lint/typecheck/test)',
};

const DEFAULTS = { promote_threshold: 2, window_days: 45, quarantine_threshold: 2 };

function now() {
  return new Date().toISOString().replace(/\.\d{3}Z$/, 'Z');
}

function parseDate(s) {
  const d = new Date(s);
  return Number.isNaN(d.getTime()) ? new Date() : d;
}

function storePath(root) {
  return join(root, STORE_REL);
}

function renderPath(root) {
  return join(root, RENDER_REL);
}

function load(root) {
  const path = storePath(root);
  if (!existsSync(path)) {
    return {
      schema: 1,
      promote_threshold: DEFAULTS.promote_threshold,
      window_days: DEFAULTS.window_days,
      quarantine_threshold: DEFAULTS.quarantine_threshold,
      next_id: 1,
      lessons: [],
    };
  }
  const data = JSON.parse(readFileSync(path, 'utf-8'));
  for (const [k, v] of Object.entries(DEFAULTS)) if (data[k] === undefined) data[k] = v;
  if (data.schema === undefined) data.schema = 1;
  if (data.next_id === undefined) data.next_id = 1;
  if (data.lessons === undefined) data.lessons = [];
  return data;
}

function save(root, data) {
  mkdirSync(join(root, '.rc'), { recursive: true });
  writeFileSync(storePath(root), JSON.stringify(data, null, 2) + '\n', 'utf-8');
  render(root, data);
}

// Normalized dedup key: lowercase, strip punctuation, collapse whitespace. Exact-after-
// normalization only (no semantic matching) — phrase lessons tersely and canonically so
// recurrences actually merge.
function norm(text) {
  return text
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

function keyOf(signal, text) {
  return signal + '::' + norm(text);
}

// Drop candidates that never recurred within the window. Mutates data; returns dropped ids.
function autoPrune(data) {
  const threshold = data.promote_threshold;
  const window = data.window_days;
  const nowMs = Date.now();
  const kept = [];
  const dropped = [];
  for (const l of data.lessons) {
    if (l.status === 'candidate' && l.recurrence < threshold) {
      const seen = parseDate(l.last_seen || l.created || now());
      const ageDays = Math.floor((nowMs - seen.getTime()) / 86400000);
      if (ageDays > window) {
        dropped.push(l.id);
        continue;
      }
    }
    kept.push(l);
  }
  data.lessons = kept;
  return dropped;
}

function find(data, signal, text) {
  const k = keyOf(signal, text);
  return data.lessons.find((l) => l.key === k) || null;
}

function render(root, data) {
  const lines = [];
  lines.push('# LESSONS — auto-maintained by scripts/lessons.mjs');
  lines.push('');
  lines.push('> Machine-owned. Do NOT hand-edit. Changes are overwritten on the next `lessons.mjs` write.');
  lines.push('> Canonical state lives in `.rc/lessons.json`. Edit lessons only via the script.');
  lines.push(
    `> promote_threshold=${data.promote_threshold} distinct features · window_days=${data.window_days} · quarantine_threshold=${data.quarantine_threshold}`
  );
  lines.push('');

  const byStatus = { confirmed: [], candidate: [], quarantined: [] };
  for (const l of data.lessons) (byStatus[l.status] || byStatus.candidate).push(l);

  const block = (title, items, note) => {
    const out = [`## ${title}`, ''];
    if (note) out.push(note, '');
    if (!items.length) {
      out.push('_none_', '');
      return out;
    }
    for (const l of [...items].sort((a, b) => a.id.localeCompare(b.id))) {
      const scope = l.scope ? ` · scope: \`${l.scope}\`` : '';
      out.push(`### ${l.id} — ${l.text}`);
      out.push(
        `- signal: \`${l.signal}\` · recurrence: ${l.recurrence} feature(s)${scope} · harmful: ${l.harmful || 0}`
      );
      out.push(`- features: ${(l.features || []).join(', ') || '—'}`);
      const ev = l.evidence || [];
      if (ev.length) out.push(`- evidence: ${ev[0]}` + (ev.length > 1 ? ` (+${ev.length - 1} more)` : ''));
      out.push(`- last seen: ${l.last_seen || '—'}`);
      out.push('');
    }
    return out;
  };

  lines.push(
    ...block(
      'Confirmed (load these at plan/design time)',
      byStatus.confirmed,
      'Corroborated across multiple features. Safe to apply as guidance.'
    )
  );
  lines.push(
    ...block(
      'Candidates (under observation — do NOT load as guidance yet)',
      byStatus.candidate,
      'Seen once or not yet corroborated. Tracked, not trusted.'
    )
  );
  lines.push(
    ...block(
      'Quarantined (failed when applied — ignore)',
      byStatus.quarantined,
      'A confirmed lesson that recurred alongside failure. Kept for the maintainer to review.'
    )
  );

  writeFileSync(renderPath(root), lines.join('\n').replace(/\s+$/, '') + '\n', 'utf-8');
}

// ----------------------------- commands -----------------------------

function cmdInit(root) {
  const data = load(root);
  save(root, data);
  console.log(`Initialized lessons store at ${storePath(root)} and ${renderPath(root)}`);
  return 0;
}

function cmdAdd(root, args) {
  const signal = args.signal;
  const source = (args.source || '').trim();
  const text = (args.text || '').trim();
  const feature = (args.feature || '').trim();

  // Grounding is enforced here, deterministically — not left to the prompt.
  if (!SIGNALS[signal]) {
    console.error(`ERROR: --signal must be one of ${Object.keys(SIGNALS).sort().join(', ')}`);
    return 2;
  }
  if (!feature) {
    console.error('ERROR: --feature is required (the feature/slug the signal came from).');
    return 2;
  }
  if (!source) {
    console.error('ERROR: --source is required (file:line / AC id / mutant id / SPEC_DEVIATION ref).');
    console.error('       A lesson with no grounding in the verification record is an opinion, not a lesson. Refused.');
    return 2;
  }
  if (text.length < 12) {
    console.error('ERROR: --text too short. State the actionable lesson in one terse sentence.');
    return 2;
  }

  const data = load(root);
  autoPrune(data);
  const existing = find(data, signal, text);
  const ts = now();

  if (existing) {
    if (!existing.features.includes(feature)) existing.features.push(feature);
    existing.recurrence = existing.features.length;
    existing.last_seen = ts;
    const ev = args.scope ? `${source} (${args.scope})` : source;
    if (!existing.evidence.includes(ev)) existing.evidence.push(ev);
    let promoted = false;
    if (existing.status === 'candidate' && existing.recurrence >= data.promote_threshold) {
      existing.status = 'confirmed';
      promoted = true;
    }
    save(root, data);
    let msg = `UPDATED ${existing.id} (recurrence=${existing.recurrence}, status=${existing.status})`;
    if (promoted) msg += ' — PROMOTED to confirmed';
    console.log(msg);
  } else {
    const lid = `L-${String(data.next_id).padStart(3, '0')}`;
    data.next_id += 1;
    data.lessons.push({
      id: lid,
      key: keyOf(signal, text),
      text,
      signal,
      scope: (args.scope || '').trim(),
      status: 'candidate',
      features: [feature],
      recurrence: 1,
      harmful: 0,
      evidence: [args.scope ? `${source} (${args.scope})` : source],
      created: ts,
      last_seen: ts,
    });
    save(root, data);
    console.log(`ADDED ${lid} (status=candidate, recurrence=1)`);
  }
  return 0;
}

function cmdPenalize(root, args) {
  const data = load(root);
  const target = data.lessons.find((l) => l.id.toLowerCase() === (args.id || '').toLowerCase());
  if (!target) {
    console.error(`ERROR: no lesson with id ${args.id}`);
    return 2;
  }
  target.harmful = (target.harmful || 0) + 1;
  target.last_seen = now();
  if (target.harmful >= data.quarantine_threshold) target.status = 'quarantined';
  save(root, data);
  console.log(`PENALIZED ${target.id} (harmful=${target.harmful}, status=${target.status})`);
  return 0;
}

function cmdList(root, args) {
  const data = load(root);
  if (autoPrune(data).length) save(root, data);
  const want = args.status || 'confirmed';
  const q = (args.query || '').toLowerCase().trim();
  const scope = (args.scope || '').toLowerCase().trim();
  const rows = data.lessons.filter((l) => {
    if (want !== 'all' && l.status !== want) return false;
    if (q && !l.text.toLowerCase().includes(q)) return false;
    if (scope && !(l.scope || '').toLowerCase().includes(scope)) return false;
    return true;
  });
  if (!rows.length) {
    console.log(`(no ${want} lessons${q || scope ? ` matching '${q || scope}'` : ''})`);
    return 0;
  }
  for (const l of rows.sort((a, b) => a.id.localeCompare(b.id))) {
    const sc = l.scope ? ` [scope:${l.scope}]` : '';
    console.log(`${l.id} (${l.status}, x${l.recurrence})${sc}: ${l.text}`);
  }
  return 0;
}

function cmdPrune(root) {
  const data = load(root);
  const dropped = autoPrune(data);
  save(root, data);
  console.log(`Pruned ${dropped.length} stale candidate(s): ${dropped.join(', ') || '—'}`);
  return 0;
}

function cmdStatus(root) {
  const data = load(root);
  const counts = { confirmed: 0, candidate: 0, quarantined: 0 };
  for (const l of data.lessons) counts[l.status] = (counts[l.status] || 0) + 1;
  console.log(
    `lessons: ${data.lessons.length} total | confirmed=${counts.confirmed} candidate=${counts.candidate} quarantined=${counts.quarantined}`
  );
  return 0;
}

// ----------------------------- arg parsing -----------------------------

function parseArgs(argv) {
  const out = { _: [] };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a.startsWith('--')) {
      const key = a.slice(2);
      const next = argv[i + 1];
      if (next === undefined || next.startsWith('--')) out[key] = true;
      else {
        out[key] = next;
        i++;
      }
    } else out._.push(a);
  }
  return out;
}

function main(argv) {
  const args = parseArgs(argv);
  if (args.selftest) return selftest();
  const root = args.root ? String(args.root) : '.';
  const cmd = args._[0];
  switch (cmd) {
    case 'init':
      return cmdInit(root);
    case 'add':
      return cmdAdd(root, args);
    case 'penalize':
      return cmdPenalize(root, args);
    case 'list':
      return cmdList(root, args);
    case 'prune':
      return cmdPrune(root);
    case 'status':
      return cmdStatus(root);
    default:
      console.error('Usage: lessons.mjs <init|add|list|penalize|prune|status> [--root DIR] [flags]');
      console.error('       lessons.mjs --selftest');
      return 2;
  }
}

// ----------------------------- selftest -----------------------------

function assert(cond, msg) {
  if (!cond) throw new Error('SELFTEST FAIL: ' + msg);
}

function selftest() {
  const dir = mkdtempSync(join(tmpdir(), 'rc-lessons-'));
  try {
    // Grounding is refused without source/feature/signal.
    assert(cmdAdd(dir, { signal: 'ac_gap', text: 'a real lesson here', feature: 'f1' }) === 2, 'missing source must fail');
    assert(cmdAdd(dir, { signal: 'bogus', source: 'x', text: 'a real lesson here', feature: 'f1' }) === 2, 'bad signal must fail');
    assert(cmdAdd(dir, { signal: 'ac_gap', source: 'x', text: 'short', feature: 'f1' }) === 2, 'short text must fail');

    // First sighting -> candidate.
    assert(cmdAdd(dir, { signal: 'ac_gap', source: 'spec.md:10', text: 'assert per placement not just count', feature: 'phase-a' }) === 0, 'add1');
    let data = load(dir);
    assert(data.lessons.length === 1 && data.lessons[0].status === 'candidate', 'first is candidate');

    // Same lesson, same feature -> no promotion (distinct-feature counting).
    cmdAdd(dir, { signal: 'ac_gap', source: 'spec.md:11', text: 'Assert per placement, not just count!', feature: 'phase-a' });
    data = load(dir);
    assert(data.lessons[0].recurrence === 1 && data.lessons[0].status === 'candidate', 'same feature does not promote');
    assert(data.lessons[0].evidence.length === 2, 'evidence accrues');

    // Second distinct feature -> promoted to confirmed.
    cmdAdd(dir, { signal: 'ac_gap', source: 'other.md:3', text: 'assert per placement not just count', feature: 'phase-b' });
    data = load(dir);
    assert(data.lessons[0].recurrence === 2 && data.lessons[0].status === 'confirmed', 'second feature promotes');

    // Penalize twice -> quarantined.
    cmdPenalize(dir, { id: 'L-001' });
    data = load(dir);
    assert(data.lessons[0].status === 'confirmed' && data.lessons[0].harmful === 1, 'one penalty keeps confirmed');
    cmdPenalize(dir, { id: 'L-001' });
    data = load(dir);
    assert(data.lessons[0].status === 'quarantined', 'two penalties quarantine');

    // Rendered file exists and reflects state.
    const md = readFileSync(renderPath(dir), 'utf-8');
    assert(md.includes('## Quarantined') && md.includes('L-001'), 'render reflects quarantine');

    // Auto-prune drops a stale candidate past the window.
    cmdAdd(dir, { signal: 'gate_fail', source: 'build.log:1', text: 'stale candidate to be pruned', feature: 'phase-c' });
    data = load(dir);
    const stale = data.lessons.find((l) => l.id === 'L-002');
    stale.last_seen = '2000-01-01T00:00:00Z';
    save(dir, data);
    cmdPrune(dir);
    data = load(dir);
    assert(!data.lessons.find((l) => l.id === 'L-002'), 'stale candidate pruned');

    console.log('SELFTEST OK');
    return 0;
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
}

process.exit(main(process.argv.slice(2)));
