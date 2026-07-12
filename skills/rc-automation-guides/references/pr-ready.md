# Skill: /pr-ready

Verifica se o codigo esta pronto para merge, executando os criterios de aceite definidos nos `agent-docs/`.

## O que fazer ao invocar este skill

Execute os passos abaixo em ordem. Se qualquer passo falhar, pare, reporte o erro ao usuario e ajude a corrigir antes de prosseguir.

---

## Passo 1 — Testes

```bash
pnpm test
```

**Criterio de aceite:** `0 failed` no output do Vitest.

Se falhar: identifique os testes que falharam, analise a causa raiz e corrija.

---

## Passo 2 — Build

```bash
pnpm build
```

Equivale a: `tsc --project tsconfig.build.json`

**Criterio de aceite:** sem erros de compilacao TypeScript.

Se falhar: reporte os erros de tipo e corrija-os.

---

## Passo 3 — Verificacao de cobertura de testes

Revise o diff atual e verifique:

- Todo novo service de usecase criado tem um arquivo `.test.ts` co-localizado?
- Os testes usam implementacoes `InMemory*` (preferido) ou `infra: 'mock-file'` via Factory?
- Os testes nunca instanciam repositorios concretos (Postgres, HTTP) diretamente?

Se algum service novo nao tiver teste: crie o teste seguindo o padrao em `agent-docs/TESTING.md`.

---

## Passo 4 — Verificacao de DI

Se novos services foram adicionados, verifique:

- `src/interfaces/server/dependencies.ts` foi atualizado?
- `src/interfaces/lambda/dependencies.ts` foi atualizado?

Ambos precisam estar sincronizados.

---

## Passo 5 — Verificacao de arquitetura

Revise o diff e verifique regras de importacao:

- `domain/` nao importa nada do projeto?
- `usecases/` importa apenas de `domain/`?
- `infrastructure/` importa apenas de `domain/`?
- Sem dependencias cruzadas entre `server/`, `lambda/` e `sqs/`?
- Logica de negocio esta em usecases (nunca em controllers/handlers/consumers)?

---

## Passo 6 — Verificacao de seguranca e padroes

Revise o diff e verifique:

- Inputs validados com Zod na fronteira (controllers/handlers)?
- Zero `console.log`, `console.error`, `console.warn`, `console.debug`?
- Sem uso explicito de `any`?
- Queries SQL parametrizadas (sem concatenacao de string com input)?
- Nenhum secret hardcoded em codigo-fonte?
- Dados sensiveis nao logados?
- Stack traces nao expostos em respostas?

---

## Passo 7 — Verificacao de API (se endpoint novo/alterado)

- OpenAPI (`/docs/openapi.yaml`) foi atualizado?
- Rotas no plural, kebab-case, sem trailing slash?
- Resposta segue contrato padrao (recurso unico sem envelope, colecao com `data` + `pagination`)?
- Codigos HTTP corretos (nunca `200` com body de erro)?
- Erros de validacao retornam `400` com `{ error: { code, message, details } }`?

---

## Passo 8 — Verificacao de banco de dados (se schema alterado)

- Campos obrigatorios presentes (`id UUID`, `created_at`, `updated_at`, `deleted_at`)?
- Soft delete como padrao? Queries filtram `WHERE deleted_at IS NULL`?
- Tabelas plural, snake_case? Colunas snake_case?
- Valores monetarios com `NUMERIC(15, 2)` (nunca `FLOAT`)?
- Migracao imutavel com rollback documentado?

---

## Passo 9 — Verificacao de processo

- Branch segue formato `<JIRA_TICKET>/<descricao-em-kebab-case>`?
- Commits seguem Conventional Commits?
- serverless.yaml e serverless.dev.yaml atualizados (se Lambda nova)?
- JSDoc nas funcoes e classes relevantes?
- Comentarios nos handlers do serverless.yaml?

---

## Resultado

Reporte ao usuario:

```
PR READY - Resultado
=====================
pnpm test:          [OK (X testes) / FALHOU]
pnpm build:         [OK / FALHOU]
Cobertura testes:   [OK / PENDENTE - descreva]
DI sincronizado:    [OK / PENDENTE / N/A]
Arquitetura:        [OK / ISSUE - descreva]
Seguranca/padroes:  [OK / ISSUE - descreva]
API:                [OK / N/A]
Banco de dados:     [OK / N/A]
Processo:           [OK / ISSUE - descreva]
```

Ou, se houver falhas bloqueantes:

```
PR BLOQUEADO
- [listar cada falha com descricao e sugestao de correcao]
```
