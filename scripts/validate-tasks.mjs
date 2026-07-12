#!/usr/bin/env node
// Structural validator for a RC feature slug's task files — the plugin-native replacement for
// `rc tasks validate`. Checks the invariants documented by the rc-create-tasks skill:
//   - each task_NN.md has the required frontmatter (status, title, type, complexity, dependencies)
//   - numbering is sequential (task_01..task_NN, no gaps) and consistent with the _tasks.md table
//   - each task file contains the required section headers
// Hard-fails (exit 1) on frontmatter / numbering / consistency problems. Missing sections are
// warnings (exit 0) since section wording can vary. Dependency-free. Run --selftest to verify
// the logic offline.
//
// Usage:
//   node scripts/validate-tasks.mjs --slug <feature> [--dir <tasksRoot>]   (default dir: .rc/tasks)
//   node scripts/validate-tasks.mjs --selftest

import { readFileSync, existsSync, readdirSync } from 'node:fs';
import { join } from 'node:path';

const REQUIRED_FM = ['status', 'title', 'type', 'complexity', 'dependencies'];
const REQUIRED_SECTIONS = [
  '## Overview',
  '## Subtasks',
  '## Implementation Details',
  '## Deliverables',
  '## Tests',
  '## Success Criteria',
];

function frontmatter(text) {
  if (!text.startsWith('---')) return null;
  const end = text.indexOf('\n---', 3);
  if (end === -1) return null;
  const out = {};
  for (const line of text.slice(3, end).split('\n')) {
    const m = line.match(/^([A-Za-z0-9_-]+):\s*(.*)$/);
    if (m) out[m[1]] = m[2].trim();
  }
  return out;
}

// Pure core: given the _tasks.md text and the task files, return {errors, warnings}.
// files: array of { name, content } for every task_NN.md.
export function validateTaskSet({ tasksMd, files }) {
  const errors = [];
  const warnings = [];

  // Task file numbers, in file-name order.
  const numbered = files
    .map((f) => ({ ...f, n: (f.name.match(/^task_(\d+)\.md$/) || [])[1] }))
    .filter((f) => f.n !== undefined)
    .map((f) => ({ ...f, num: Number(f.n) }))
    .sort((a, b) => a.num - b.num);

  if (numbered.length === 0) errors.push('no task_NN.md files found');

  // Sequential numbering starting at 1.
  numbered.forEach((f, i) => {
    if (f.num !== i + 1)
      errors.push(`numbering gap: expected task_${String(i + 1).padStart(2, '0')}.md, found ${f.name}`);
  });

  // Per-file frontmatter + sections.
  for (const f of numbered) {
    const fm = frontmatter(f.content);
    if (!fm) { errors.push(`${f.name}: missing or malformed frontmatter`); continue; }
    for (const key of REQUIRED_FM)
      if (!(key in fm)) errors.push(`${f.name}: frontmatter missing \`${key}\``);
    for (const sec of REQUIRED_SECTIONS)
      if (!f.content.includes(sec)) warnings.push(`${f.name}: missing section \`${sec}\``);
  }

  // Consistency with the _tasks.md master table (numbers referenced as | NN | ...).
  if (tasksMd == null) {
    errors.push('_tasks.md not found');
  } else {
    const tableNums = [...tasksMd.matchAll(/^\|\s*(\d+)\s*\|/gm)].map((m) => Number(m[1]));
    const fileNums = new Set(numbered.map((f) => f.num));
    for (const n of tableNums)
      if (!fileNums.has(n)) errors.push(`_tasks.md lists task ${n} but task_${String(n).padStart(2, '0')}.md is missing`);
    for (const f of numbered)
      if (!tableNums.includes(f.num)) errors.push(`${f.name} exists but is not listed in _tasks.md`);
  }

  return { errors, warnings };
}

function runCli() {
  const argv = process.argv.slice(2);
  const arg = (flag) => { const i = argv.indexOf(flag); return i >= 0 ? argv[i + 1] : undefined; };
  const slug = arg('--slug');
  const root = arg('--dir') ?? '.rc/tasks';
  if (!slug) { console.error('usage: validate-tasks.mjs --slug <feature> [--dir <tasksRoot>]'); process.exit(2); }

  const dir = join(root, slug);
  if (!existsSync(dir)) { console.error(`validate-tasks: task directory not found: ${dir}`); process.exit(1); }

  const tasksMdPath = join(dir, '_tasks.md');
  const tasksMd = existsSync(tasksMdPath) ? readFileSync(tasksMdPath, 'utf8') : null;
  const files = readdirSync(dir)
    .filter((n) => /^task_\d+\.md$/.test(n))
    .map((n) => ({ name: n, content: readFileSync(join(dir, n), 'utf8') }));

  const { errors, warnings } = validateTaskSet({ tasksMd, files });
  for (const w of warnings) console.warn(`warn: ${w}`);
  if (errors.length === 0) {
    console.log(`validate-tasks: OK (${files.length} tasks, ${warnings.length} warning(s))`);
    process.exit(0);
  }
  console.error(`validate-tasks: ${errors.length} error(s):`);
  for (const e of errors) console.error(`  - ${e}`);
  process.exit(1);
}

function selftest() {
  let fail = 0;
  const eq = (label, got, want) => {
    if (got === want) console.log(`ok   ${label}`);
    else { console.error(`FAIL ${label}: got ${got}, want ${want}`); fail = 1; }
  };
  const fm = (extra = '') =>
    `---\nstatus: pending\ntitle: T\ntype: feature\ncomplexity: low\ndependencies: []\n---\n` +
    `## Overview\nx\n## Subtasks\n- a\n## Implementation Details\nx\n## Deliverables\nx\n## Tests\nx\n## Success Criteria\nx\n${extra}`;
  const tasksMd = '| # | Title |\n|---|---|\n| 01 | A |\n| 02 | B |\n';

  let r = validateTaskSet({ tasksMd, files: [
    { name: 'task_01.md', content: fm() }, { name: 'task_02.md', content: fm() }] });
  eq('valid-set errors', r.errors.length, 0);
  eq('valid-set warnings', r.warnings.length, 0);

  r = validateTaskSet({ tasksMd, files: [
    { name: 'task_01.md', content: fm() }, { name: 'task_03.md', content: fm() }] });
  eq('numbering-gap flagged', r.errors.some((e) => e.includes('numbering gap')), true);

  r = validateTaskSet({ tasksMd: '| # |\n| 01 |\n| 02 |\n', files: [
    { name: 'task_01.md', content: '---\nstatus: pending\n---\nbody' },
    { name: 'task_02.md', content: fm() }] });
  eq('missing-fm-key flagged', r.errors.some((e) => e.includes('frontmatter missing')), true);

  r = validateTaskSet({ tasksMd, files: [{ name: 'task_01.md', content: fm() }] });
  eq('tasksMd-mismatch flagged', r.errors.some((e) => e.includes('task 2')), true);

  r = validateTaskSet({ tasksMd: '| 01 |\n', files: [
    { name: 'task_01.md', content: '---\nstatus: pending\ntitle: T\ntype: t\ncomplexity: low\ndependencies: []\n---\nno sections' }] });
  eq('missing-section warns not errors', r.errors.length === 0 && r.warnings.length > 0, true);

  process.exit(fail);
}

if (process.argv.includes('--selftest')) selftest();
else runCli();
