export const meta = {
  name: 'review-squad',
  description: 'Run the parallel review squad (routed by language) on a diff and synthesize one verdict. Reviews a provided diff, or captures the current git diff if none is given. Falls back to a general-purpose agent with an embedded lens when a specialized reviewer agent is not registered.',
  whenToUse: 'Standalone: review your current changes or a PR diff. Composed: called per task by review-pipeline. Pass {diff, language, title} or {task, impl}, or nothing to review the current git diff.',
  phases: [
    { title: 'Review', detail: 'parallel review squad, routed by language' },
    { title: 'Synthesize', detail: 'dedup findings + one verdict' },
  ],
}

const CAPTURE_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['diff'],
  properties: {
    diff: { type: 'string' },
    files: { type: 'array', items: { type: 'string' } },
  },
}

const FINDINGS_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['findings'],
  properties: {
    findings: {
      type: 'array',
      items: {
        type: 'object',
        additionalProperties: false,
        required: ['severity', 'title'],
        properties: {
          severity: { type: 'string', enum: ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW'] },
          title: { type: 'string' },
          file: { type: 'string' },
          line: { type: 'integer' },
          issue: { type: 'string' },
          why: { type: 'string' },
          fix: { type: 'string' },
        },
      },
    },
  },
}

const VERDICT_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['taskId', 'verdict', 'summary'],
  properties: {
    taskId: { type: 'string' },
    verdict: { type: 'string', enum: ['APPROVE', 'WARN', 'BLOCK'] },
    critical: { type: 'integer' },
    high: { type: 'integer' },
    medium: { type: 'integer' },
    summary: { type: 'string' },
    topIssues: { type: 'array', items: { type: 'string' } },
  },
}

// Embedded lens used when the specialized reviewer agent is not registered in this
// environment (custom ~/.claude/agents are loaded at startup, not mid-session).
const LENS = {
  'code-reviewer': 'You are a senior code reviewer. Assess correctness, error handling, security, and maintainability.',
  'security-reviewer': 'You are a security reviewer. Look for injection, auth flaws, secret/credential exposure, unsafe input handling, SSRF, and resource leaks.',
  'silent-failure-hunter': 'You are hunting silent failures: swallowed errors, ignored return values, empty catch blocks, bad fallbacks, and missing error propagation.',
  'go-reviewer': 'You are an expert Go reviewer. Focus on idiomatic Go, concurrency safety, error wrapping (%w), integer conversion/overflow bounds (gosec G115), defer correctness, and nil handling.',
  'typescript-reviewer': 'You are an expert TypeScript reviewer. Focus on type safety, any/as misuse, strict-null violations, async correctness, and unhandled promise rejections.',
  'vue-reviewer': 'You are an expert Vue reviewer. Focus on reactivity correctness, v-html/template security, composable cleanup, and prop/emit contracts.',
  'svelte-reviewer': 'You are an expert Svelte reviewer. Focus on runes reactivity, {@html} security, store subscription leaks, and SvelteKit load/SSR safety.',
}

function reviewersFor(language) {
  const base = ['code-reviewer', 'security-reviewer', 'silent-failure-hunter']
  const byLang = {
    go: ['go-reviewer'],
    typescript: ['typescript-reviewer'],
    node: ['typescript-reviewer'],
    vue: ['vue-reviewer', 'typescript-reviewer'],
    svelte: ['svelte-reviewer', 'typescript-reviewer'],
  }
  const all = base.concat(byLang[language] || [])
  return all.filter((r, i) => all.indexOf(r) === i)
}

function reviewPrompt(rmeta, diff) {
  return [
    `Review this change for "${rmeta.title}" (${rmeta.language}).`,
    'Review ONLY the diff provided below — do not assume other files changed and do not fetch the repo state.',
    'Report findings in your structured output. Only report issues you are >80% confident are real. Consolidate duplicates. Skip pure style nits.',
    'If you find nothing wrong, return an empty findings array — do NOT invent issues.',
    '',
    rmeta.summary ? `SUMMARY: ${rmeta.summary}` : '',
    rmeta.filesChanged && rmeta.filesChanged.length ? `FILES: ${rmeta.filesChanged.join(', ')}` : '',
    rmeta.notes ? `AUTHOR NOTES: ${rmeta.notes}` : '',
    '',
    'DIFF:',
    diff || '(no diff)',
  ].filter(Boolean).join('\n')
}

function synthPrompt(title, taskId, reviews) {
  const merged = reviews.flatMap((r) => (r.findings || []).map((f) => ({ reviewer: r.reviewer, ...f })))
  return [
    `You are the review lead. Consolidate findings from multiple reviewers for "${title}".`,
    `Set taskId to "${taskId}".`,
    'Deduplicate findings that describe the same issue (same file+line, or same root cause). Keep the single clearest statement of each.',
    'Verdict rule: BLOCK if any CRITICAL or HIGH survives; WARN if only MEDIUM survives; APPROVE only if there are genuinely zero findings.',
    'Count critical/high/medium AFTER dedup. topIssues = up to 5 one-line summaries, most severe first.',
    '',
    'RAW FINDINGS (JSON):',
    JSON.stringify(merged),
  ].join('\n')
}

async function runReviewer(r, rmeta, diff) {
  const prompt = reviewPrompt(rmeta, diff)
  try {
    const f = await agent(prompt, { label: `review:${r}`, phase: 'Review', agentType: r, schema: FINDINGS_SCHEMA })
    return { reviewer: r, mode: 'agent', findings: (f && f.findings) || [] }
  } catch (e) {
    const preamble = LENS[r] || 'You are a senior code reviewer.'
    const f = await agent(`${preamble}\n\n${prompt}`, {
      label: `review:${r}(fallback)`,
      phase: 'Review',
      agentType: 'general-purpose',
      schema: FINDINGS_SCHEMA,
    })
    return { reviewer: r, mode: 'fallback', findings: (f && f.findings) || [] }
  }
}

let input = args || {}
if (typeof input === 'string') {
  try {
    input = JSON.parse(input)
  } catch (e) {
    input = {}
  }
}
const task = input.task || {}
const impl = input.impl || {}
const title = input.title || task.title || 'current changes'
const taskId = input.taskId || task.id || 'adhoc'
const language = input.language || task.language || 'other'
const summary = input.summary || impl.summary || ''
const filesChanged = input.filesChanged || impl.filesChanged || []
const notes = input.notes || impl.notes || ''
let diff = input.diff || impl.diff || ''

phase('Review')

if (!diff.trim()) {
  log('No diff provided — capturing current git diff.')
  const captured = await agent(
    'Capture the current repository diff. Run `git diff --staged`; if that is empty run `git diff`; if still empty run `git show --patch HEAD`. Return ONLY the raw unified diff text in the "diff" field and the changed file paths in "files".',
    { label: 'capture-diff', phase: 'Review', schema: CAPTURE_SCHEMA }
  )
  diff = (captured && captured.diff) || ''
  if (!filesChanged.length && captured && captured.files) filesChanged.push(...captured.files)
}

if (!diff.trim()) {
  log('Nothing to review.')
  return { taskId, title, language, filesChanged, reviewersAttempted: 0, reviewersSucceeded: 0, verdict: { taskId, verdict: 'APPROVE', summary: 'No changes to review.', critical: 0, high: 0, medium: 0 } }
}

const reviewers = reviewersFor(language)
log(`Reviewing "${title}" (${language}) with ${reviewers.length} reviewer(s): ${reviewers.join(', ')}`)

const reviews = (
  await parallel(reviewers.map((r) => () => runReviewer(r, { title, language, summary, filesChanged, notes }, diff)))
).filter(Boolean)

const attempted = reviewers.length
const got = reviews.length
const degraded = reviews.filter((r) => r.mode === 'fallback').length

if (got === 0) {
  log(`All ${attempted} reviewer(s) failed to run — cannot produce a trustworthy verdict.`)
  return {
    taskId,
    title,
    language,
    filesChanged,
    reviewersAttempted: attempted,
    reviewersSucceeded: 0,
    verdict: {
      taskId,
      verdict: 'ERROR',
      summary: `Review did not run: none of the ${attempted} reviewer agents could be invoked (agentType not found and general-purpose fallback also failed). This is NOT an approval.`,
      critical: 0,
      high: 0,
      medium: 0,
    },
  }
}

if (degraded) log(`${degraded}/${got} reviewer(s) ran in fallback mode (specialized agent not registered — restart to load ~/.claude/agents).`)
if (got < attempted) log(`Warning: only ${got}/${attempted} reviewers ran — verdict is partial.`)

const verdict = await agent(synthPrompt(title, taskId, reviews), { label: 'synth', phase: 'Synthesize', schema: VERDICT_SCHEMA })

return {
  taskId,
  title,
  language,
  filesChanged,
  reviewersAttempted: attempted,
  reviewersSucceeded: got,
  reviewersDegraded: degraded,
  verdict,
}
