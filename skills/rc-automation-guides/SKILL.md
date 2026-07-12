---
name: rc-automation-guides
description: Guias de desenvolvimento e operação do repositório rc-automation (plataforma de automações da Escale). Use ao trabalhar no rc-automation — criar endpoint REST, migration, evento EventBridge/SQS, feature nova, revisão de código, checagem de segurança, deploy, docs (README/OpenAPI/Postman/arquitetura), onboarding ou refactor. Carrega o guia certo por tarefa a partir de references/. Não use em outros repositórios.
---

# rc-automation — guias por tarefa

Guias do padrão do repositório `rc-automation` (monorepo: `automation-registry` TS/Lambda,
`automation-audience` e `automation-handler` Go/ECS, módulo `actions/` stdlib-only).
Leia o guia da tarefa em `references/` antes de agir — cada um contém o passo a passo,
os arquivos envolvidos e os critérios de aceite. Complementam o `CLAUDE.md` e o
`agent-docs/` do próprio repo, que continuam sendo a fonte das regras sempre ativas.

## Roteamento

| Tarefa | Guia |
| ------ | ---- |
| Feature end-to-end (checklist de 10 passos) | `references/new-feature.md` |
| Novo endpoint REST | `references/api-endpoint.md` |
| Migration / alteração de schema | `references/db-migration.md` |
| Publicar/consumir evento (EventBridge + SQS) | `references/event-publish.md` |
| Revisão de código (checklist do repo) | `references/code-review.md` |
| Checagem de segurança por PR (OWASP) | `references/security-check.md` |
| Critérios de aceite pré-merge (test + build) | `references/pr-ready.md` |
| Refactor guiado | `references/refactor.md` |
| Branch + Conventional Commits | `references/branch-commit.md` |
| Deploy (staging vs production) | `references/deploy.md` |
| Atualizar README | `references/update-readme.md` |
| Gerar/atualizar OpenAPI (`.techdocs/docs/openapi.yaml`) | `references/update-openapi.md` |
| Sincronizar coleção Postman | `references/update-postman.md` |
| Atualizar doc de arquitetura geral | `references/update-arch-doc.md` |
| Auditoria de docs vs codebase | `references/audit-docs.md` |
| Carregar contexto certo por tipo de tarefa | `references/load-context.md` |
| Onboarding de engenheiro | `references/onboarding.md` |
| Melhorar/reestruturar um prompt | `references/enhance-prompt.md` |

Runbooks operacionais (diagnóstico de automação, health-check de frota, reprocesso,
migrations em produção) NÃO vivem nesta skill: referenciam `db.json` com conexões e
permanecem locais, fora de qualquer distribuição.
