# Skill: /branch-commit

Guia para nomenclatura de branches e Conventional Commits, baseado em `agent-docs/DEVELOPMENT-PROCESS.md`.

## O que fazer ao invocar este skill

Pergunte ao usuario o que precisa:
1. Criar uma nova branch?
2. Escrever uma mensagem de commit?
3. Preparar titulo de PR?

---

## Branches

### Formato obrigatorio

```
<JIRA_TICKET>/<descricao-em-kebab-case>
```

### Regras

- Ticket Jira em maiusculas (`ENG-123`, nao `eng-123`)
- Descricao em kebab-case, sem acentos, maximo 5 palavras
- Sem prefixos de tipo (`feature/`, `bugfix/`) — o tipo fica no commit

### Exemplos

| Correto | Errado |
|---|---|
| `ENG-123/add-payment-method` | `feature/add-payment` |
| `ENG-456/fix-order-timeout` | `fix-order-timeout` |
| `ENG-789/refactor-cep-service` | `ENG789-refactor` |

### Comando para criar

```bash
git checkout -b ENG-XXX/descricao-em-kebab-case
```

---

## Conventional Commits

### Formato obrigatorio

```
<type>(<scope>): <descricao no imperativo>
```

### Tipos

| Tipo | Quando usar |
|---|---|
| `feat` | Nova funcionalidade |
| `fix` | Correcao de bug |
| `refactor` | Refatoracao sem mudanca de comportamento |
| `test` | Adicao ou correcao de testes |
| `docs` | Documentacao |
| `chore` | Manutencao (deps, config, build) |
| `ci` | Mudancas em pipelines |

### Scope

Opcional. Indica o dominio afetado:

```
feat(orders): add cancel endpoint
fix(auth): handle expired token correctly
refactor(cep): extract address mapper
```

### Regras

- Descricao em minusculas, no imperativo (`add`, `fix` — nao `added`, `fixes`)
- Sem ponto final na primeira linha
- Maximo de 72 caracteres na primeira linha
- Breaking change: adicione `!` apos o tipo e descreva no footer

```
feat(orders)!: change pagination response envelope

BREAKING CHANGE: campo `items` renomeado para `data` na resposta paginada
```

### Exemplos

```
feat(payments): add pix payment method
fix(orders): prevent double processing on retry
refactor(cep): replace callback with async/await
test(coverage): add unit tests for StatesCoverageService
docs(api): update openapi for payment endpoints
chore(deps): upgrade @escaletech/logger to 2.1.0
ci(pipeline): add security scan to production workflow
```

---

## Titulo de PR

O titulo da PR deve seguir o formato Conventional Commit:

```
feat(orders): add cancel endpoint
```

### Template de PR

```markdown
## O que foi feito
-

## Como testar
-

## Checklist
- [ ] `pnpm test` passou
- [ ] OpenAPI atualizado (se endpoint foi adicionado ou alterado)
- [ ] `server/dependencies.ts` e `lambda/dependencies.ts` sincronizados (se aplicavel)
```

Para fechar ticket automaticamente no Jira, inclua `Closes ENG-123` no body da PR.

---

## Assistencia

Se o usuario descrever a mudanca, gere automaticamente:
1. Nome da branch sugerido
2. Mensagem de commit formatada
3. Titulo de PR
