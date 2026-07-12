# Skill: /load-context

Carrega os documentos certos para o tipo de tarefa, seguindo o protocolo definido em `agent-docs/HOW_TO_USE.md`.

## Uso

```
/load-context <tipo>
```

Tipos aceitos: `new-feature`, `bugfix`, `infra`, `deploy`, `api`, `database`, `security`, `events`

---

## Protocolo por tipo de tarefa

### `new-feature` — Nova feature end-to-end

Leia os seguintes documentos antes de comecar:

1. `agent-docs/ARCHITECTURE.md` — onde cada artefato vai, regras de DI, checklist de 10 passos
2. `agent-docs/CODING-STANDARDS.md` — nomenclatura, TypeScript patterns, organizacao de arquivos
3. `agent-docs/TESTING.md` — como escrever o teste co-localizado com mock-file

Apos ler, use `/new-feature` para o workflow guiado.

---

### `bugfix` — Correcao de bug ou refatoracao

Leia os seguintes documentos:

1. `agent-docs/CODING-STANDARDS.md` — verifique se o codigo existente segue os padroes
2. Arquivos fontes relevantes ao bug reportado

Verifique se a correcao:
- Nao introduz `console.*` (use ILogger injetado)
- Nao coloca logica de negocio em controller/handler
- Mantem as convencoes de nomenclatura
- Nao cria importacoes cruzadas entre camadas

---

### `infra` — Nova infraestrutura ou repositorio

Leia os seguintes documentos:

1. `agent-docs/ARCHITECTURE.md` — secao "Repository Interface + Factory"
2. `agent-docs/INFRA_AND_DEPLOY.md` — como a infra esta organizada na AWS

Lembre-se:
- Interface no domain, implementacao na infrastructure
- Factory seleciona implementacao via variavel de ambiente
- Adicionar variavel de ambiente em `environments/`

---

### `deploy` — CI/CD ou processo de deploy

Leia o seguinte documento:

1. `agent-docs/INFRA_AND_DEPLOY.md` — pipelines, ambientes, Serverless Framework, Terraform

Use `/deploy` para o guia step-by-step.

---

### `api` — Criar ou alterar endpoint REST

Leia os seguintes documentos:

1. `agent-docs/API.md` — nomenclatura de rotas, paginacao, filtros, versionamento, resposta padrao
2. `agent-docs/SECURITY.md` — validacao de inputs com Zod, headers de contexto
3. `agent-docs/ARCHITECTURE.md` — onde colocar controller/handler, regras de importacao

Use `/api-endpoint` para o workflow guiado.

---

### `database` — Schema, migracoes ou queries

Leia os seguintes documentos:

1. `agent-docs/DATA-MODELING.md` — schema PostgreSQL, soft delete, migracoes, OpenSearch
2. `agent-docs/DATABASE.md` — nomenclatura de tabelas, colunas, chaves, ENUMs

Use `/db-migration` para o workflow guiado.

---

### `security` — Revisao de seguranca

Leia o seguinte documento:

1. `agent-docs/SECURITY.md` — validacao, OWASP, secrets, dependencias

Use `/security-check` para o checklist completo.

---

### `events` — Publicar ou consumir eventos

Leia os seguintes documentos:

1. `agent-docs/API.md` — secao 7 (Padrao de Eventos EventBridge + SQS)
2. `agent-docs/ARCHITECTURE.md` — secao Consumer SQS

Use `/event-publish` para o workflow guiado.

---

## Regra de ouro (sempre aplicavel)

Antes de propor qualquer solucao arquitetural, valide contra as regras de importacao em `agent-docs/ARCHITECTURE.md`. Alerte o usuario se a solucao:

- Viola a separacao entre camadas (`domain`, `usecases`, `infrastructure`, `interfaces`)
- Cria dependencias cruzadas entre `server/`, `lambda/` e `sqs/`
- Importa infra diretamente em usecases ou domain
