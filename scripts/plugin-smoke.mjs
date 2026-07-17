#!/usr/bin/env node
// Plugin smoke test: validates that the RC Claude Code plugin is loadable purely from its
// declarative parts — no Go, no build. Checks skill/agent/command frontmatter, that every
// hook command references a script that exists and is executable, that no skill is orphaned
// (absent from the README catalog, and so undiscoverable), and that no CLI-era residue creeps
// back into the docs. Exits non-zero on any failure so CI can gate on it. Dependency-free
// (no YAML lib): frontmatter here is simple `key: value` lines between the first two `---` fences.
//
// Run: node scripts/plugin-smoke.mjs   (from the plugin root)

import { readFileSync, existsSync, statSync, readdirSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const VALID_MODELS = new Set(['sonnet', 'opus', 'haiku', 'fable', 'inherit']);

const problems = [];
const warnings = [];
let checked = 0;
const fail = (file, msg) => problems.push(`${file}: ${msg}`);
// Doctrine checks land as warnings, not failures: they report a real backlog (64 findings from the
// 2026-07-17 audit) that predates them, and a gate that is red on arrival is a gate people learn to
// skip. Promote to `fail` once the backlog is cleared — the count only goes down from here.
const warn = (file, msg) => warnings.push(`${file}: ${msg}`);

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

// --- orphan skills: a skill nobody can discover may as well not ship ---
// README.md is the skill catalog (COMMANDS.md is curated, so it is not required here).
// A skill absent from it loads but is invisible to humans — how rc-loop, rc-roadmap and
// rc-lessons shipped in 2.1.0 and stayed undiscoverable until 2.3.0.
const readme = existsSync(join(ROOT, 'README.md')) ? readFileSync(join(ROOT, 'README.md'), 'utf8') : '';
for (const name of listDirs(join(ROOT, 'skills'))) {
  if (!existsSync(join(ROOT, 'skills', name, 'SKILL.md'))) continue;
  checked++;
  if (!new RegExp(`\\b${name}\\b`).test(readme))
    fail('README.md', `skill \`${name}\` is not in the catalog — undiscoverable`);
}

// --- CLI-era residue: the `rc` binary/daemon is retired (see CLAUDE.md) ---
// Prescriptive mentions only: lines that *negate* the CLI ("there is no `rc exec` wrapper")
// are the fix, not the bug. CLAUDE.md is exempt — the rule has to name what it forbids.
const RESIDUE = [
  /`rc` binary|rc binary/i,
  /`rc (setup|exec|sync|init|tasks run|reviews)`/,
  /ACP runtime|home-scoped daemon/i,
  /--(auto-commit|dry-run|tui|concurrent|batch-size|persist|run-id|include-completed|include-resolved)\b/,
];
const NEGATED = /there is no|no longer|retired|never reference|reintroduce/i;
const DOC_ROOTS = ['README.md', 'COMMANDS.md', 'AGENTS.md'];
for (const dir of ['docs', 'commands', join('skills', 'rc', 'references')]) {
  for (const f of listFiles(join(ROOT, dir), '.md')) DOC_ROOTS.push(join(dir, f));
}
DOC_ROOTS.push(join('skills', 'rc', 'SKILL.md'));
for (const file of DOC_ROOTS) {
  const abs = join(ROOT, file);
  if (!existsSync(abs)) continue;
  checked++;
  readFileSync(abs, 'utf8').split('\n').forEach((line, i) => {
    if (NEGATED.test(line)) return;
    for (const re of RESIDUE) {
      if (re.test(line)) { fail(`${file}:${i + 1}`, `CLI-era residue: ${line.trim().slice(0, 70)}`); break; }
    }
  });
}

// --- dangling assets: a skill that points at a file it does not ship is broken on use ---
// Only markdown links `](path)` and backticked `` `path` `` into references/, assets/ or scripts/
// count — bare prose mentions are illustrative ("Read references/api-spec.md when needed").
// Scoped to the entry points (SKILL.md / AGENTS.md): files under references/ are prose and
// vendored examples whose paths belong to the example, not to this repo. `#anchors` are stripped.
// A path resolves if it exists under its own skill, the plugin root, or any sibling skill
// (skills cross-reference each other, e.g. `references/delegation-contract.md` in the rc skill).
const SKILL_DIRS = listDirs(join(ROOT, 'skills')).map((n) => join(ROOT, 'skills', n));
const PLACEHOLDER = /NNN|<|\{|_prd|_techspec|_tasks\b/;
const assetPaths = (text) => [
  ...text.matchAll(/\]\(((?:references|assets|scripts)\/[^)\s]+)\)/g),
  ...text.matchAll(/`((?:references|assets|scripts)\/[^`\s]+\.[a-z]+)(?=[`\s])/g), // may carry args: `scripts/x.sh --flag`
].map((m) => m[1].split('#')[0]);

for (const skillDir of SKILL_DIRS) {
  const files = [join(skillDir, 'SKILL.md'), join(skillDir, 'AGENTS.md')].filter(existsSync);
  for (const abs of files) {
    checked++;
    const rel = abs.slice(ROOT.length + 1);
    for (const p of new Set(assetPaths(readFileSync(abs, 'utf8')))) {
      if (PLACEHOLDER.test(p)) continue; // template placeholder, not a shipped file
      const found = existsSync(join(skillDir, p)) || existsSync(join(ROOT, p)) ||
        SKILL_DIRS.some((d) => existsSync(join(d, p)));
      if (!found) fail(rel, `dangling asset: \`${p}\` does not exist`);
    }
  }
}

// --- anti-trigger: a description without one over-fires and burns context on every session ---
// The skill spec wants both a positive trigger ("Use when…") and a negative one ("Do not use
// for…"); the 2026-07-17 doctrine audit found 15 skills carrying only the positive half — all of
// them vendored front-end skills copied in and never adapted (rc-react, rc-tailwindcss, rc-zod…).
// Phrasing is deliberately broad: `rc-zod` says "does NOT cover", not "do not use for", and a
// narrower regex flagged it as a false positive. Portuguese variants count — some skills ship pt-BR.
const ANTI_TRIGGER =
  /do not use|don't use|dont use|not for|never use|does not cover|doesn't cover|does not handle|not intended for|excludes|não use|nao use|não cobre|nao cobre/i;
for (const name of listDirs(join(ROOT, 'skills'))) {
  const abs = join(ROOT, 'skills', name, 'SKILL.md');
  if (!existsSync(abs)) continue;
  checked++;
  const fm = frontmatter(readFileSync(abs, 'utf8'));
  if (!fm?.description) continue; // already reported above
  if (!ANTI_TRIGGER.test(fm.description))
    warn(join('skills', name, 'SKILL.md'), 'description has no anti-trigger ("Do not use for…")');
}

// --- orphan bundled files: shipped weight the agent can never reach ---
// A file under a skill that SKILL.md never points at is dead weight every consumer downloads.
// `rc-shadcn-ui` ships 4,370 lines of `references/` with zero pointers; `rc-git` ships 587 and
// duplicates them inline instead. Reachability is generous on purpose — the naive version of this
// check (exact path only) reported 102 files at ~85% false positives, because skills legitimately
// point at a *directory* (`references/query/`) or name a file without its extension. Counting all
// four forms took it to 29 real orphans. If it ever gets noisy again, widen the forms, not the gate.
const walkFiles = (dir) =>
  readdirSync(dir, { withFileTypes: true }).flatMap((e) =>
    e.isDirectory() ? walkFiles(join(dir, e.name)) : [join(dir, e.name)]
  );
for (const name of listDirs(join(ROOT, 'skills'))) {
  const skillDir = join(ROOT, 'skills', name);
  const sp = join(skillDir, 'SKILL.md');
  if (!existsSync(sp)) continue;
  checked++;
  const text = readFileSync(sp, 'utf8');
  for (const abs of walkFiles(skillDir)) {
    if (abs === sp) continue;
    const rel = abs.slice(skillDir.length + 1);
    const base = rel.split('/').pop();
    const stem = base.replace(/\.[^.]+$/, '');
    const parent = rel.includes('/') ? rel.slice(0, rel.lastIndexOf('/')) : null;
    const reachable =
      text.includes(rel) || text.includes(base) || text.includes(stem) ||
      (parent && text.includes(parent + '/'));
    if (!reachable)
      warn(join('skills', name, 'SKILL.md'), `orphan file: \`${rel}\` is never pointed at`);
  }
}

// --- stray files at a skill root: the spec allows scripts/, references/, assets/ and nothing else ---
// Loose root files escape the dangling-asset check above (it only follows those three prefixes), so
// `rc-tdd`'s `[tests.md](tests.md)` links are unguarded — delete the file and nothing notices.
// They are also where vendoring debris lands: `rc-vitest/GENERATION.md` (a source SHA),
// `rc-systematic-debugging/test-pressure-*.md` (the upstream author's own test transcripts).
// Allowlist: AGENTS.md is a first-class entry point in this repo (see the dangling-asset check
// above, which reads it); LICENSE/NOTICE are legal requirements — Apache 2.0 §4(d) mandates NOTICE.
const ROOT_ALLOWED = /^(SKILL\.md|AGENTS\.md|LICENSE.*|NOTICE.*)$/i;
for (const name of listDirs(join(ROOT, 'skills'))) {
  const skillDir = join(ROOT, 'skills', name);
  if (!existsSync(join(skillDir, 'SKILL.md'))) continue;
  checked++;
  for (const e of readdirSync(skillDir, { withFileTypes: true })) {
    if (e.isDirectory() || ROOT_ALLOWED.test(e.name)) continue;
    warn(join('skills', name), `stray file at skill root: \`${e.name}\` — move under references/, assets/ or scripts/`);
  }
}

// --- CI toolchain fossils: a workflow must not build a stack this repo does not have ---
// ci.yml survived the de-fork setting up Go, Bun, Playwright and running `make verify` in a repo
// that ships plain markdown — so every push touching skills/ or scripts/ failed for months.
// The invariant is tied to reality (does the manifest exist?), not to a blacklist of names.
const TOOLCHAIN = [
  { re: /setup-go|go-version|\bgo (build|test|mod)\b/, needs: 'go.mod', what: 'Go' },
  { re: /setup-bun|bun install|bunx/, needs: 'bun.lock', what: 'Bun' },
  { re: /^\s*(run:.*|-\s+)make\s+\w+/m, needs: 'Makefile', what: 'make' },
];
for (const f of listFiles(join(ROOT, '.github', 'workflows'), '.yml')) {
  const file = join('.github', 'workflows', f);
  checked++;
  const text = readFileSync(join(ROOT, file), 'utf8');
  for (const { re, needs, what } of TOOLCHAIN) {
    if (re.test(text) && !existsSync(join(ROOT, needs)))
      fail(file, `sets up ${what}, but this repo has no \`${needs}\` — toolchain fossil`);
  }
}

// --- report ---
// Warnings print with `--warn` (or in CI via PLUGIN_SMOKE_WARN=1) so the default run stays a
// signal, not a wall of pre-existing backlog. They never affect the exit code.
const showWarnings = process.argv.includes('--warn') || process.env.PLUGIN_SMOKE_WARN === '1';
if (warnings.length) {
  if (showWarnings) {
    console.error(`plugin-smoke: ${warnings.length} doctrine warning(s):`);
    for (const w of warnings) console.error(`  ~ ${w}`);
  } else {
    console.error(`plugin-smoke: ${warnings.length} doctrine warning(s) — rerun with --warn to list.`);
  }
}
if (problems.length === 0) {
  console.log(`plugin-smoke: OK (${checked} components checked)`);
  process.exit(0);
}
console.error(`plugin-smoke: ${problems.length} problem(s) across ${checked} components:`);
for (const p of problems) console.error(`  - ${p}`);
process.exit(1);
