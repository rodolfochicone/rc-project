#!/usr/bin/env node
// Plugin smoke test: validates that the RC Claude Code plugin is loadable purely from its
// declarative parts — no Go, no build. Checks skill/agent/command frontmatter and that every
// hook command references a script that exists and is executable. Exits non-zero on any
// failure so CI can gate on it. Dependency-free (no YAML lib): frontmatter here is simple
// `key: value` lines between the first two `---` fences.
//
// Run: node scripts/plugin-smoke.mjs   (from the plugin root)

import { readFileSync, existsSync, statSync, readdirSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const VALID_MODELS = new Set(['sonnet', 'opus', 'haiku', 'fable', 'inherit']);

const problems = [];
let checked = 0;
const fail = (file, msg) => problems.push(`${file}: ${msg}`);

/** Parse the leading `---` frontmatter block into a flat {key: value} map (or null). */
function frontmatter(text) {
  if (!text.startsWith('---')) return null;
  const end = text.indexOf('\n---', 3);
  if (end === -1) return null;
  const lines = text.slice(3, end).split('\n');
  const out = {};
  for (let i = 0; i < lines.length; i++) {
    const m = lines[i].match(/^([A-Za-z0-9_-]+):\s*(.*)$/);
    if (!m) continue;
    let value = m[2].trim();
    // YAML block/folded scalar or plain continuation: gather following indented lines.
    if (value === '' || value === '|' || value === '>') {
      const cont = [];
      while (i + 1 < lines.length && /^\s+\S/.test(lines[i + 1])) cont.push(lines[++i].trim());
      if (cont.length) value = cont.join(' ');
    }
    out[m[1]] = value;
  }
  return out;
}

function listDirs(dir) {
  if (!existsSync(dir)) return [];
  return readdirSync(dir, { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => e.name);
}
function listFiles(dir, ext) {
  if (!existsSync(dir)) return [];
  return readdirSync(dir, { withFileTypes: true })
    .filter((e) => e.isFile() && e.name.endsWith(ext))
    .filter((e) => e.name.toLowerCase() !== 'readme.md') // docs, not components
    .map((e) => e.name);
}

// --- skills: each skills/<name>/SKILL.md needs name + description ---
for (const name of listDirs(join(ROOT, 'skills'))) {
  const file = join('skills', name, 'SKILL.md');
  const abs = join(ROOT, file);
  if (!existsSync(abs)) continue; // dirs without SKILL.md (e.g. references) are fine
  checked++;
  const fm = frontmatter(readFileSync(abs, 'utf8'));
  if (!fm) { fail(file, 'missing or malformed frontmatter'); continue; }
  if (!fm.name) fail(file, 'frontmatter missing `name`');
  if (!fm.description) fail(file, 'frontmatter missing `description`');
  if (fm.model && !VALID_MODELS.has(fm.model) && !fm.model.startsWith('claude-'))
    fail(file, `invalid model \`${fm.model}\``);
}

// --- agents: each agents/*.md needs name + description + valid model ---
for (const f of listFiles(join(ROOT, 'agents'), '.md')) {
  const file = join('agents', f);
  checked++;
  const fm = frontmatter(readFileSync(join(ROOT, file), 'utf8'));
  if (!fm) { fail(file, 'missing or malformed frontmatter'); continue; }
  if (!fm.name) fail(file, 'frontmatter missing `name`');
  if (!fm.description) fail(file, 'frontmatter missing `description`');
  if (fm.model && !VALID_MODELS.has(fm.model) && !fm.model.startsWith('claude-'))
    fail(file, `invalid model \`${fm.model}\``);
}

// --- commands: each commands/*.md must be non-empty ---
for (const f of listFiles(join(ROOT, 'commands'), '.md')) {
  const file = join('commands', f);
  checked++;
  if (readFileSync(join(ROOT, file), 'utf8').trim().length === 0) fail(file, 'empty command file');
}

// --- hooks: hooks.json parses and every referenced script exists + is executable ---
const hooksPath = join(ROOT, 'hooks', 'hooks.json');
if (existsSync(hooksPath)) {
  checked++;
  let hooks;
  try {
    hooks = JSON.parse(readFileSync(hooksPath, 'utf8'));
  } catch (e) {
    fail('hooks/hooks.json', `invalid JSON: ${e.message}`);
    hooks = null;
  }
  const commands = [];
  for (const groups of Object.values(hooks?.hooks ?? {})) {
    for (const group of groups ?? []) {
      for (const h of group.hooks ?? []) if (h.command) commands.push(h.command);
    }
  }
  for (const cmd of commands) {
    const m = cmd.match(/\$\{CLAUDE_PLUGIN_ROOT\}\/(\S+)/);
    if (!m) continue; // non-plugin-root commands (inline shell) aren't path-checked
    const rel = m[1];
    const abs = join(ROOT, rel);
    if (!existsSync(abs)) { fail('hooks/hooks.json', `references missing script: ${rel}`); continue; }
    if (!(statSync(abs).mode & 0o111)) fail('hooks/hooks.json', `script not executable: ${rel}`);
  }
}

// --- report ---
if (problems.length === 0) {
  console.log(`plugin-smoke: OK (${checked} components checked)`);
  process.exit(0);
}
console.error(`plugin-smoke: ${problems.length} problem(s) across ${checked} components:`);
for (const p of problems) console.error(`  - ${p}`);
process.exit(1);
