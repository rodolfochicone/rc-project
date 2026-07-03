export const meta = {
  name: 'review-pipeline',
  description: 'Plan a goal into independently-implementable tasks, implement each in an isolated git worktree, then delegate review to the review-squad workflow and collect one verdict per task',
  whenToUse: 'When you have a feature/change goal and want it planned, implemented per-task in isolation, and reviewed by the review-squad. Pass the goal as args.',
  phases: [
    { title: 'Plan', detail: 'decompose the goal into independently-implementable tasks', model: 'opus' },
    { title: 'Execute', detail: 'implement each task in an isolated git worktree' },
    { title: 'Review', detail: 'delegate to review-squad per task (parallel squad + synthesis)' },
  ],
}

const TASKS_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['tasks'],
  properties: {
    tasks: {
      type: 'array',
      items: {
        type: 'object',
        additionalProperties: false,
        required: ['id', 'title', 'description', 'language'],
        properties: {
          id: { type: 'string' },
          title: { type: 'string' },
          description: { type: 'string' },
          files: { type: 'array', items: { type: 'string' } },
          language: { type: 'string', enum: ['go', 'typescript', 'node', 'vue', 'svelte', 'other'] },
        },
      },
    },
  },
}

const IMPL_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['taskId', 'summary', 'diff'],
  properties: {
    taskId: { type: 'string' },
    summary: { type: 'string' },
    filesChanged: { type: 'array', items: { type: 'string' } },
    diff: { type: 'string' },
    notes: { type: 'string' },
  },
}

function planPrompt(goal) {
  return [
    'You are a senior planning specialist. Decompose the goal below into a small set of INDEPENDENTLY-implementable tasks.',
    'Each task must be self-contained: implementable and reviewable on its own, without waiting on another task first.',
    'Prefer FEWER, well-scoped tasks over many tiny ones. Do not invent scope beyond the goal. Do not gold-plate.',
    'Set "language" to the dominant language of the files the task touches (go, typescript, node, vue, svelte, or other).',
    'If the goal is already a single change, return exactly one task.',
    '',
    'GOAL:',
    goal,
  ].join('\n')
}

function implPrompt(task) {
  return [
    'Implement this task. Make the MINIMUM change that satisfies it — surgical, matching existing code style. Do not refactor adjacent code, do not add speculative abstractions.',
    `Task id: ${task.id}`,
    `Title: ${task.title}`,
    `Description: ${task.description}`,
    task.files && task.files.length ? `Likely files: ${task.files.join(', ')}` : '',
    '',
    "You are in an isolated git worktree — your edits will not touch the user's main working tree.",
    'When done, return:',
    '- taskId: this task id',
    '- summary: what you changed and why',
    '- filesChanged: files you edited',
    '- diff: the unified diff of your changes (run `git add -A` then `git diff --staged`). If it exceeds ~400 lines, include the most important hunks and clearly note the truncation.',
    '- notes: assumptions, follow-ups, or anything the reviewer must know',
    'Do NOT commit and do NOT merge. Leave the changes staged in this worktree.',
  ].filter(Boolean).join('\n')
}

phase('Plan')
const goal = typeof args === 'string' ? args : (args && args.goal) || ''
if (!goal.trim()) {
  log('No goal provided. Pass the feature/change description as args, e.g. Workflow({name:"review-pipeline", args:"add retry with backoff to the HTTP client"}).')
  return { error: 'missing goal' }
}

log(`Planning: ${goal.slice(0, 120)}`)
const plan = await agent(planPrompt(goal), { agentType: 'planner', phase: 'Plan', schema: TASKS_SCHEMA })
if (!plan || !plan.tasks || plan.tasks.length === 0) {
  log('Planner returned no tasks — nothing to do.')
  return { goal, taskCount: 0, results: [] }
}
log(`Planned ${plan.tasks.length} task(s): ${plan.tasks.map((t) => t.language).join(', ')}`)

const results = await pipeline(
  plan.tasks,
  (task) =>
    agent(implPrompt(task), {
      label: `impl:${task.id}`,
      phase: 'Execute',
      isolation: 'worktree',
      schema: IMPL_SCHEMA,
    }),
  async (impl, task) => {
    if (!impl || !impl.diff) {
      return {
        taskId: task.id,
        task,
        impl: impl || null,
        verdict: { taskId: task.id, verdict: 'BLOCK', summary: 'Implementation failed or was skipped.', critical: 0, high: 0, medium: 0 },
      }
    }
    const reviewed = await workflow({ scriptPath: 'C:\\Users\\rodol\\code\\rc-project\\.claude\\workflows\\review-squad.js' }, { task, impl })
    return { taskId: task.id, task, impl, verdict: (reviewed && reviewed.verdict) || { taskId: task.id, verdict: 'WARN', summary: 'Review produced no verdict.', critical: 0, high: 0, medium: 0 } }
  }
)

const clean = results.filter(Boolean)
const blocked = clean.filter((r) => r.verdict && r.verdict.verdict === 'BLOCK')
const warned = clean.filter((r) => r.verdict && r.verdict.verdict === 'WARN')
log(`Done: ${clean.length}/${plan.tasks.length} task(s) — ${blocked.length} BLOCK, ${warned.length} WARN`)

return {
  goal,
  taskCount: plan.tasks.length,
  blocked: blocked.length,
  warned: warned.length,
  note: 'Each task was implemented in its own git worktree and left staged (not committed, not merged). Inspect the worktrees and integrate the ones you accept.',
  results: clean.map((r) => ({
    taskId: r.taskId,
    title: r.task.title,
    language: r.task.language,
    filesChanged: r.impl ? r.impl.filesChanged || [] : [],
    verdict: r.verdict,
  })),
}
