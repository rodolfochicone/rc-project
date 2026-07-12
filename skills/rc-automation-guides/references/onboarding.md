# Skill: /onboarding

Roteiro de onboarding para novos engenheiros no projeto, baseado em `agent-docs/HOW_TO_USE.md`.

## O que fazer ao invocar este skill

Apresente o roteiro completo de onboarding e guie o usuario por cada etapa.

---

## Roteiro de Onboarding

### 1. Entenda o big picture

Leia `agent-docs/ARCHITECTURE.md`:
- Microsservicos com Clean Architecture
- 4 camadas: `domain/`, `usecases/`, `infrastructure/`, `interfaces/`
- Dois modos de deploy: Server (ECS/Docker) e Lambda (Serverless)
- Comunicacao sincrona (HTTP/REST) e assincrona (EventBridge + SQS)

### 2. Conheca as regras de codigo

Leia `agent-docs/CODING-STANDARDS.md`:
- Nomenclatura (PascalCase para classes, sufixos corretos)
- TypeScript strict, sem `any`, sem path aliases
- Logging com `ILogger` (console.* proibido)
- Organizacao de arquivos (uma classe por arquivo, testes co-localizados)

### 3. Entenda como testar

Leia `agent-docs/TESTING.md`:
- Vitest
- Testes co-localizados (`NomeService.test.ts`)
- Sempre `infra: 'mock-file'` via Factory
- `pnpm test` deve passar com 0 falhas

### 4. Configure o ambiente

```bash
# Clone o repositorio
git clone <repo-url>
cd <projeto>

# Instale dependencias (sempre pnpm via Corepack, nunca npm/yarn)
pnpm install

# Configure variaveis de ambiente
cp environments/.env.example .env
# Edite o .env com suas configuracoes

# Inicie um servico localmente (Lambda-only)
pnpm automation-registry:dev  # serverless-offline (Lambda local, hot-reload)
```

### 5. Conheca os comandos disponiveis

```bash
pnpm test              # Roda todos os testes (dentro do diretorio do servico)
pnpm test --coverage   # Com relatorio de cobertura
pnpm build             # Compila TypeScript
pnpm release           # Cria tag v* para producao (a partir da raiz)
```

### 6. Entenda o processo de desenvolvimento

Leia `agent-docs/DEVELOPMENT-PROCESS.md`:
- Branches: `<JIRA_TICKET>/<descricao-em-kebab-case>`
- Commits: Conventional Commits
- PRs: minimo 1 aprovacao, titulo no formato Conventional Commit
- Deploy: staging (push em main) e production (tag v*)

### 7. Leia sobre seguranca

Leia `agent-docs/SECURITY.md`:
- Validacao de inputs com Zod na fronteira
- Queries parametrizadas
- Secrets nunca em codigo

### 8. Conheca os padroes de banco de dados

Leia `agent-docs/DATA-MODELING.md` e `agent-docs/DATABASE.md`:
- Tabelas plural, snake_case
- Campos obrigatorios (id UUID, created_at, updated_at, deleted_at)
- Soft delete como padrao

---

## Mapa de documentos

```
agent-docs/
  ARCHITECTURE.md          <- como o sistema e organizado
  INFRA_AND_DEPLOY.md      <- onde roda e como fazer deploy
  CODING-STANDARDS.md      <- regras de codigo e nomenclatura
  TESTING.md               <- estrategia e padroes de testes
  DEVELOPMENT-PROCESS.md   <- branches, commits, PRs e releases
  SECURITY.md              <- validacao, OWASP, secrets
  DATA-MODELING.md         <- schema, soft delete, migracoes
  DATABASE.md              <- nomenclatura de banco de dados
  API.md                   <- REST, paginacao, eventos
  HOW_TO_USE.md            <- este guia
```

---

## Skills uteis para o dia a dia

| Comando | O que faz |
|---|---|
| `/new-feature` | Workflow guiado para criar feature end-to-end |
| `/api-endpoint` | Guia para criar endpoint REST |
| `/db-migration` | Guia para migracoes de banco |
| `/pr-ready` | Verifica se o codigo esta pronto para merge |
| `/code-review` | Checklist de revisao de codigo |
| `/security-check` | Revisao de seguranca |
| `/branch-commit` | Ajuda com branches e commits |
| `/deploy` | Guia de deploy |
| `/load-context` | Carrega documentos certos por tipo de tarefa |
| `/event-publish` | Guia para eventos EventBridge/SQS |
