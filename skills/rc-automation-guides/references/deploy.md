# Skill: /deploy

Guia o processo de deploy baseado em `agent-docs/INFRA_AND_DEPLOY.md` e `agent-docs/DEVELOPMENT-PROCESS.md`.

## O que fazer ao invocar este skill

Leia `agent-docs/INFRA_AND_DEPLOY.md` e `agent-docs/DEVELOPMENT-PROCESS.md`, depois pergunte ao usuario:
- Qual o alvo do deploy? **Staging** ou **Production**?
- E um deploy de **Server (ECS)** ou **Lambda (Serverless)**?

---

## Deploy para Staging

**Trigger:** push na branch principal (automatico via GitHub Actions)

O pipeline executa automaticamente:
1. Build da aplicacao
2. Build/push da imagem Docker para ECR
3. Terraform plan/apply

**Nao e necessario executar manualmente.** Apenas garanta que o codigo esta na branch principal.

Antes de fazer o push, verifique com `/pr-ready` se o codigo esta pronto.

---

## Deploy para Production

**Trigger:** tag `v*` criada via `pnpm release`

### Criterios para release

Antes de criar a tag, confirme:
- [ ] `main` passou em todos os checks do pipeline (test, build)
- [ ] Funcionalidade foi validada em staging
- [ ] Nenhum issue critico em aberto relacionado ao que sera liberado

### Comando

```bash
pnpm release
```

Este comando cria a tag `v*` e faz push. O pipeline de producao inicia automaticamente.

### Pipeline de producao

1. Build da aplicacao
2. Build/push da imagem para ECR
3. **Security Scan:** Conviso AST + `security-gate.yml` — bloqueia se vulnerabilidade encontrada
4. Terraform plan/apply

**ATENCAO — Security Gate:** se vulnerabilidades forem encontradas, o deploy e bloqueado automaticamente. Resolva as vulnerabilidades antes de tentar novamente.

### Versionamento (SemVer)

| Versao | Quando incrementar |
|---|---|
| `MAJOR` (v**2**.0.0) | Breaking change — nova versao de API, quebra de contrato |
| `MINOR` (v1.**1**.0) | Nova funcionalidade retrocompativel |
| `PATCH` (v1.0.**1**) | Bugfix sem novo comportamento |

### O que e quebra de contrato (exige MAJOR)

| Quebra de contrato | Retrocompativel |
|---|---|
| Remover ou renomear campo da resposta | Adicionar campo opcional na resposta |
| Alterar tipo de um campo existente | Adicionar novo endpoint |
| Mudar semantica de um endpoint | Adicionar campo opcional no body |
| Remover endpoint | Adicionar novo codigo de erro |

Se houver quebra de contrato:
1. Crie branch `v{N}-maintenance` a partir de `main`
2. Desenvolva a nova versao na `main`
3. Configure API Gateway para mapear `/v{N}/*` e `/v{N+1}/*`
4. Deprece a versao antiga com prazo comunicado

---

## Adicionando nova funcao Lambda

Se o deploy inclui uma nova funcao Lambda, verifique se ela foi registrada em **ambos** os arquivos:

### serverless.yaml (staging e production)

```yaml
# Descricao do que a funcao faz
functions:
  nomeFuncao:
    handler: src/interfaces/lambda/<contexto>/<nome>.handler
    events:
      - http:
          path: /<rota>
          method: GET
          integration: lambda
          response: ${file(config/response.yaml)}
          request: ${file(config/request.yaml)}
```

### serverless.dev.yaml (desenvolvimento local)

Registre com `httpApi` para funcionar com `serverless-offline`.

Se esqueceu de registrar: a funcao nao sera deployada e nao respondera requisicoes.

---

## Testando localmente antes do deploy

```bash
pnpm automation-registry:dev  # serverless-offline (Lambda local, hot-reload via nodemon)
```

Configure o arquivo `.env` a partir de `environments/.env.example`.

---

## Variaveis de ambiente

| Contexto | Fonte |
|---|---|
| Local | arquivo `.env` (nunca commitado) |
| CI/CD | GitHub Secrets e GitHub Vars |
| Runtime producao | Variaveis injetadas pelo pipeline |

Variaveis criticas: `REPOSITORY_COVERED_PLACE`, `REPOSITORY_DEALERS`, `PORT`, `NODE_ENV`, `API_BASE_PATH`

Regras:
- `.env` no `.gitignore` — verificar antes de qualquer commit
- Secrets nunca em codigo-fonte, comentarios ou logs
- Nomeadas em `SCREAMING_SNAKE_CASE`
- IAM roles com least privilege — nenhuma Lambda com `*`

---

## Checklist pre-deploy

- [ ] `pnpm test` passou com 0 falhas
- [ ] `pnpm build` compilou sem erros
- [ ] Nova funcao Lambda registrada em `serverless.yaml` E `serverless.dev.yaml` (se aplicavel)
- [ ] Variaveis de ambiente configuradas no ambiente alvo
- [ ] `server/dependencies.ts` e `lambda/dependencies.ts` estao sincronizados
- [ ] OpenAPI atualizado (se endpoint novo ou alterado)
- [ ] Funcionalidade validada em staging (para producao)
- [ ] Nenhum issue critico em aberto (para producao)
