---
description: Gera uma mensagem de commit convencional a partir do diff staged (não faz commit).
allowed-tools: Bash(git status:*), Bash(git diff:*), Bash(git log:*)
---

Analise apenas as mudanças já em stage e proponha uma mensagem de commit.

Contexto:
- Status: !`git status --short`
- Diff staged: !`git diff --staged`
- Estilo recente do repositório: !`git log --oneline -10`

Gere uma mensagem no formato **Conventional Commits**, seguindo o estilo dos commits recentes acima:
- Linha de título: `tipo(escopo): resumo imperativo curto` (≤72 caracteres).
- Se necessário, corpo explicando o "porquê" (não o "o quê").
- NÃO adicione atribuição de co-autoria nem rodapé "Generated with Claude".

Apenas imprima a mensagem sugerida. **Não execute `git commit`** — eu faço isso manualmente.
