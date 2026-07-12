# Skill: /audit-docs

Auditoria completa dos guias em `agent-docs/`, do `CLAUDE.md` e dos commands em `.claude/commands/`. Verifica se o codebase esta dentro dos padroes documentados.

## O que fazer ao invocar este skill

Execute TODAS as fases abaixo, na ordem. Ao final, gere um relatorio consolidado.

---

## Fase 1 — Inventario de documentos

Leia todos os arquivos:

1. `CLAUDE.md`
2. `agent-docs/ARCHITECTURE.md`
3. `agent-docs/CODING-STANDARDS.md`
4. `agent-docs/TESTING.md`
5. `agent-docs/API.md`
6. `agent-docs/DATA-MODELING.md`
7. `agent-docs/DATABASE.md`
8. `agent-docs/SECURITY.md`
9. `agent-docs/DEVELOPMENT-PROCESS.md`
10. `agent-docs/INFRA_AND_DEPLOY.md`
11. `agent-docs/HOW_TO_USE.md`
12. Todos os arquivos em `.claude/commands/`

Verifique:
- [ ] Todos os agent-docs listados existem e sao legiveis?
- [ ] Todos os commands referenciados no CLAUDE.md existem em `.claude/commands/`?
- [ ] Todos os commands existentes estao listados no CLAUDE.md?

---

## Fase 2 — Consistencia entre CLAUDE.md e agent-docs

Para cada regra no CLAUDE.md, verifique se ha respaldo em pelo menos um agent-doc. Para cada regra importante nos agent-docs, verifique se esta refletida no CLAUDE.md.

Verificar especificamente:

### Arquitetura (ARCHITECTURE.md)
- [ ] Regras de importacao entre camadas no CLAUDE.md batem com ARCHITECTURE.md secao 8?
- [ ] Camada `sqs/` esta mencionada em ambos?
- [ ] Checklist de feature no CLAUDE.md e consistente com secao 10 do ARCHITECTURE.md?
- [ ] Principios DRY e YAGNI estao refletidos?

### Coding Standards (CODING-STANDARDS.md)
- [ ] Nomenclatura no CLAUDE.md bate com secao 3 do CODING-STANDARDS.md?
- [ ] Regras de logging (niveis, securityLog, correlation ID) consistentes?
- [ ] Regras de TypeScript (strict, sem any, named exports, sem path aliases) consistentes?
- [ ] Organizacao de arquivos (uma classe por arquivo, testes co-localizados) mencionada?

### Testing (TESTING.md)
- [ ] Regras de teste no CLAUDE.md batem com TESTING.md?
- [ ] mock-file, Factory, beforeAll mencionados consistentemente?

### API (API.md)
- [ ] Regras de rotas (plural, kebab-case) consistentes?
- [ ] Contrato de resposta (recurso unico, colecao paginada, erro) consistente?
- [ ] Paginacao (LIMIT/OFFSET, cursor) documentada?
- [ ] Eventos (EventBridge/SQS) consistentes entre CLAUDE.md e API.md secao 7?
- [ ] Versionamento de API documentado?

### Data Modeling (DATA-MODELING.md) + Database (DATABASE.md)
- [ ] Campos obrigatorios consistentes entre os dois docs e o CLAUDE.md?
- [ ] **CONFLITO CONHECIDO:** DATABASE.md secao "Colunas Base" usa `SERIAL` e `TIMESTAMP`, mas DATA-MODELING.md exige `UUID` e `TIMESTAMPTZ`. Verificar se o conflito persiste e reportar.
- [ ] Tabelas de juncao: DATA-MODELING.md nao menciona, DATABASE.md define convencao. Consistente?
- [ ] Soft delete documentado consistentemente em ambos?
- [ ] ENUMs: UPPER_SNAKE_CASE em ambos?

### Security (SECURITY.md)
- [ ] Validacao com Zod consistente entre CLAUDE.md e SECURITY.md?
- [ ] OWASP Top 10 — regras relevantes refletidas no CLAUDE.md?
- [ ] Gestao de secrets consistente?
- [ ] Politica de dependencias mencionada?

### Development Process (DEVELOPMENT-PROCESS.md)
- [ ] Branches, Conventional Commits, PRs consistentes com CLAUDE.md?
- [ ] SemVer e release documentados?
- [ ] Definition of Done refletida no /pr-ready?

### Infra and Deploy (INFRA_AND_DEPLOY.md)
- [ ] Ambientes (staging/production) consistentes com /deploy?
- [ ] Plugins do Serverless Framework documentados?
- [ ] Variaveis de ambiente criticas listadas?

---

## Fase 3 — Consistencia entre commands e agent-docs

Para cada command, verifique se as instrucoes estao alinhadas com os agent-docs que referencia:

| Command | Docs que deve refletir |
|---|---|
| `/new-feature` | ARCHITECTURE.md, CODING-STANDARDS.md, TESTING.md, API.md, SECURITY.md |
| `/api-endpoint` | API.md, SECURITY.md, ARCHITECTURE.md |
| `/db-migration` | DATA-MODELING.md, DATABASE.md |
| `/event-publish` | API.md (secao 7), ARCHITECTURE.md |
| `/pr-ready` | TESTING.md, CODING-STANDARDS.md, SECURITY.md, DEVELOPMENT-PROCESS.md |
| `/code-review` | Todos os agent-docs |
| `/security-check` | SECURITY.md |
| `/deploy` | INFRA_AND_DEPLOY.md, DEVELOPMENT-PROCESS.md |
| `/branch-commit` | DEVELOPMENT-PROCESS.md |
| `/load-context` | HOW_TO_USE.md |
| `/onboarding` | HOW_TO_USE.md, todos os agent-docs |

Verifique:
- [ ] Commands nao contem regras contraditorias com os agent-docs?
- [ ] Commands nao estao desatualizados em relacao aos agent-docs?
- [ ] Exemplos de codigo nos commands seguem os padroes dos agent-docs?

---

## Fase 4 — Auditoria do codebase contra os padroes

Analise o codigo-fonte real do projeto e verifique se segue os padroes documentados.

### 4.1 Arquitetura e importacoes

Verifique os imports reais nos arquivos do projeto:

```
src/domain/       -> nao deve importar de usecases, infrastructure, interfaces
src/usecases/     -> deve importar apenas de domain
src/infrastructure/ -> deve importar apenas de domain
src/interfaces/server/ -> nunca importa de lambda ou sqs
src/interfaces/lambda/ -> nunca importa de server ou sqs
src/interfaces/sqs/    -> nunca importa de server ou lambda
```

- Grep por imports cruzados que violem as regras

### 4.2 Logging

- Grep por `console.log`, `console.error`, `console.warn`, `console.debug` em arquivos `.ts` (excluindo node_modules e arquivos de config)
- Verificar se services recebem `ILogger` via construtor

### 4.3 TypeScript

- Grep por `any` explicito em arquivos `.ts` (excluindo declaracoes de tipo de libs externas)
- Grep por `export default` (excluindo routers Express)
- Verificar se contratos de repositorio usam `interface` (nunca `class` com `throw`)

### 4.4 Validacao

- Verificar se controllers/handlers usam Zod para validar inputs
- Verificar se erros de validacao retornam `400` com contrato padrao

### 4.5 Testes

- Verificar se todo usecase service tem `.test.ts` co-localizado
- Verificar se testes usam `infra: 'mock-file'` via Factory
- Executar `pnpm test` e reportar resultado

### 4.6 DI

- Verificar se `src/interfaces/server/dependencies.ts` e `src/interfaces/lambda/dependencies.ts` estao sincronizados (mesmos services registrados)

### 4.7 Nomenclatura

- Verificar se classes seguem PascalCase com sufixos corretos
- Verificar se arquivos seguem PascalCase
- Verificar se variaveis de ambiente usam SCREAMING_SNAKE_CASE

### 4.8 Banco de dados (se houver migrations ou queries)

- Verificar se tabelas sao plural, snake_case
- Verificar se queries filtram `deleted_at IS NULL`
- Verificar se valores monetarios usam `NUMERIC` (nunca `FLOAT`)
- Verificar se IDs sao UUID (nunca SERIAL)

### 4.9 Seguranca

- Grep por concatenacao de SQL com variaveis (injection risk)
- Verificar se `.env` esta no `.gitignore`
- Grep por secrets hardcoded (patterns como `password =`, `token =`, `secret =` com valores literais)

### 4.10 serverless.yaml

- Verificar se handlers registrados tem comentarios descritivos
- Verificar se serverless.dev.yaml tem as mesmas funcoes

---

## Fase 5 — Secoes de melhorias e conflitos

Colete todas as secoes "MELHORIAS E CONFLITOS" dos agent-docs e liste-as com prioridade:

| Prioridade | Criterio |
|---|---|
| CRITICA | Conflito entre documentos que pode gerar codigo incorreto |
| ALTA | Melhoria que impacta qualidade ou seguranca |
| MEDIA | Inconsistencia menor ou melhoria de DX |
| BAIXA | Sugestao de melhoria sem impacto imediato |

---

## Relatorio final

Gere o relatorio consolidado no seguinte formato:

```
AUDIT-DOCS - Relatorio Completo
=================================

DATA: <data atual>

1. INVENTARIO
   - Agent-docs: X/Y existentes
   - Commands: X/Y existentes e listados no CLAUDE.md
   - Status: [OK / ISSUES]

2. CONSISTENCIA DOCS
   - CLAUDE.md vs agent-docs: [X conflitos encontrados]
   - Detalhes: [listar cada conflito]

3. CONSISTENCIA COMMANDS
   - Commands vs agent-docs: [X desatualizados]
   - Detalhes: [listar cada issue]

4. CODEBASE vs PADROES
   4.1 Arquitetura:      [OK / X violacoes]
   4.2 Logging:          [OK / X violacoes]
   4.3 TypeScript:       [OK / X violacoes]
   4.4 Validacao:        [OK / X violacoes]
   4.5 Testes:           [OK / X violacoes]
   4.6 DI:               [OK / DESSINCRONIZADO]
   4.7 Nomenclatura:     [OK / X violacoes]
   4.8 Banco de dados:   [OK / X violacoes / N/A]
   4.9 Seguranca:        [OK / X violacoes]
   4.10 serverless.yaml: [OK / X violacoes]

5. MELHORIAS E CONFLITOS (dos agent-docs)
   CRITICA: [listar]
   ALTA:    [listar]
   MEDIA:   [listar]
   BAIXA:   [listar]

6. ACOES RECOMENDADAS
   [lista priorizada de acoes para corrigir os problemas encontrados]
```

Para cada violacao no codebase, inclua:
- Arquivo e linha
- Regra violada
- Documento de referencia
- Sugestao de correcao
