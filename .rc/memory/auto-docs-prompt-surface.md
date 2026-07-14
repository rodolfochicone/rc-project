---
title: auto-docs.yml — a fronteira que segura o prompt injection
scope: context
key: auto-docs-prompt-surface
tags: [seguranca, github-actions, prompt-injection, codeql]
source: rc-memory (distilled 2026-07-14)
created: 2026-07-14
updated: 2026-07-14
---

O `auto-docs.yml` roda o `claude-code-action` com `contents: write` e `pull-requests: write`, e o
**título do PR flui para dentro do prompt**. Isso não é injeção de shell (essa foi corrigida em
2026-07-14: texto não-confiável agora só chega ao shell como env var) — é superfície de **prompt
injection**, e o CodeQL não a enxerga.

**O que segura hoje, e é a fronteira inteira:** o gatilho é `pull_request: types: [closed]` com
`if: merged == true`. Ou seja, **um mantenedor já olhou e mergeou** antes de o agente rodar. Some
a isso o `--allowedTools` restrito (sem `Bash(curl:*)`, sem `Bash(go:*)`).

**Quando isso deixa de valer:** se alguém mudar o gatilho para `pull_request` aberto (ou
`pull_request_target`), o título/corpo passam a ser texto de estranho chegando num agente com
permissão de escrita no repo — aí a fronteira sumiu e precisa ser redesenhada **antes** do merge
dessa mudança, não depois.
