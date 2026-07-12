# Skill: /update-readme

Analisa o codebase completo do projeto e atualiza o `README.md` com informacoes precisas e atualizadas.

## O que fazer ao invocar este skill

Execute TODAS as fases abaixo, na ordem. Ao final, reescreva o `README.md` com as informacoes coletadas.

---

## Fase 1 — Coleta de contexto documental

Leia os seguintes arquivos para entender os padroes e arquitetura do projeto:

1. `CLAUDE.md` — regras obrigatorias e padroes do projeto
2. `agent-docs/ARCHITECTURE.md` — camadas, Clean Architecture, fluxo de dados
3. `agent-docs/API.md` — endpoints, contratos, eventos
4. `agent-docs/INFRA_AND_DEPLOY.md` — deploy, CI/CD, ambientes
5. `README.md` atual — para entender a estrutura existente e preservar o que ainda for valido

---

## Fase 2 — Analise do codebase

Analise o codigo-fonte real para extrair informacoes atualizadas. Execute cada sub-fase:

### 2.1 Entidades de dominio

Leia todos os arquivos em `src/domain/` para identificar:
- Todas as entidades existentes (interfaces/types)
- Campos de cada entidade com tipos TypeScript
- Interfaces de repositorio e seus metodos
- Interfaces de unit-of-work
- Interfaces de gateways externos (CRM, etc.)
- Exceptions de dominio
- Factories

Para cada entidade, documente: nome, campos (nome, tipo, descricao inferida do contexto).

### 2.2 Usecases (servicos de aplicacao)

Leia todos os arquivos em `src/usecases/` para identificar:
- Todos os services existentes
- Dependencias injetadas (repositorios, gateways, publishers)
- Metodo `execute()` — parametros e retorno
- Regras de negocio relevantes (validacoes, fluxos condicionais)
- Testes co-localizados existentes

### 2.3 API Endpoints (handlers Lambda)

Leia todos os handlers em `src/interfaces/lambda/` para identificar:
- Todos os endpoints HTTP expostos
- Metodo HTTP + path
- Schema de validacao Zod (request body, query params, path params)
- Responses possiveis (status codes, formato)
- Consumers SQS e scheduled functions

### 2.4 Serverless configuration

Leia `serverless.yaml` e `serverless.dev.yaml` para identificar:
- Todas as functions registradas (HTTP, SQS, scheduled)
- Paths e metodos HTTP
- Event sources (SQS queues, schedule expressions)
- Configuracoes de memoria, timeout
- Custom domains
- Plugins utilizados

### 2.5 Database migrations

Leia os arquivos em `migrations/` para identificar:
- Todas as tabelas criadas/alteradas
- Schema atual de cada tabela (colunas, tipos, constraints, indices)
- Ordem cronologica das migrations

### 2.6 Stack tecnologica

Leia `package.json` para identificar:
- Versoes reais das dependencias principais (TypeScript, Node, Vitest, Zod, etc.)
- Scripts disponiveis
- Registry configurado

### 2.7 CI/CD Pipelines

Leia os arquivos em `.github/workflows/` para identificar:
- Workflows existentes e seus triggers
- Steps de cada pipeline
- Ambientes de deploy

### 2.8 Variaveis de ambiente

Leia `environments/.env.example` (ou equivalente) e o `serverless.yaml` para identificar:
- Todas as variaveis de ambiente utilizadas
- Quais sao obrigatorias vs opcionais
- Valores default quando aplicavel

### 2.9 Estrutura do projeto

Execute `find` ou use Glob para mapear a estrutura real de diretorios do projeto:
- Diretorios principais e seu proposito
- Arquivos de configuracao na raiz

---

## Fase 3 — Comparacao e identificacao de gaps

Compare o `README.md` atual com os dados coletados:

- [ ] Entidades listadas no README batem com as entidades reais no `src/domain/`?
- [ ] Campos de cada entidade estao atualizados (novos campos, campos removidos)?
- [ ] Endpoints listados batem com os handlers reais em `src/interfaces/lambda/`?
- [ ] Request/response bodies estao corretos (baseados nos schemas Zod)?
- [ ] Consumers SQS e scheduled functions estao atualizados?
- [ ] Stack tecnologica reflete as versoes reais do `package.json`?
- [ ] Variaveis de ambiente estao completas?
- [ ] Estrutura de diretorios reflete a estrutura real?
- [ ] Comandos/scripts estao corretos?
- [ ] CI/CD pipelines estao documentados corretamente?
- [ ] Novas features adicionadas apos a ultima atualizacao do README estao faltando?

---

## Fase 4 — Reescrita do README.md

Reescreva o `README.md` completo seguindo a estrutura abaixo. Use APENAS dados reais coletados nas fases anteriores — nunca invente informacoes.

### Estrutura obrigatoria do README

O README deve conter TODAS as secoes abaixo, nesta ordem:

```
1. Header (titulo, descricao, badges)
2. Indice
3. Visao Geral
4. Arquitetura (diagrama Clean Architecture + fluxo de dados)
5. Entidades de Dominio (todas, com tabela de campos)
6. API Endpoints (todos, com request/response bodies)
7. Eventos SQS e Funcoes Scheduled
8. Stack Tecnologica (versoes reais do package.json)
9. Inicio Rapido (pre-requisitos, setup)
10. Comandos (todos os scripts do package.json)
11. Estrutura do Projeto (arvore real de diretorios)
12. Variaveis de Ambiente (tabela completa)
13. Database Migrations (tabelas, comandos)
14. Testes (runner, convencoes, comandos)
15. Deploy (ambientes, pipelines, serverless, custom domains)
16. Contribuindo (regras obrigatorias)
17. Footer
```

### Regras de escrita

- **Idioma:** Portugues (sem acentos nos textos tecnicos, consistente com o README atual)
- **Formato:** Markdown com tabelas, blocos de codigo, diagramas ASCII
- **Precisao:** Todos os dados devem vir da analise real do codebase (versoes, campos, endpoints, etc.)
- **Completude:** Cada entidade deve ter tabela completa de campos. Cada endpoint deve ter request/response body documentado.
- **Badges:** Manter os badges existentes, atualizar versoes se necessario
- **Diagramas:** Manter/atualizar os diagramas ASCII de arquitetura e fluxo de dados
- **Exemplos de codigo:** Usar exemplos reais baseados nos schemas Zod encontrados nos handlers
- **Sem invencao:** Se alguma informacao nao puder ser determinada pela analise, omita ou marque como `[TODO]`

### Detalhamento por secao

#### Entidades de Dominio
Para cada entidade encontrada em `src/domain/`:
- Descricao curta do proposito
- Tabela com TODOS os campos: `| Campo | Tipo | Descricao |`
- Tipos devem refletir o TypeScript real (ex: `string | null`, `Date`, `JSONB`)

#### API Endpoints
Para cada handler HTTP:
- Tabela resumo: `| Metodo | Path | Descricao |`
- Detalhamento de cada endpoint com:
  - Request body (JSON example baseado no schema Zod)
  - Query parameters (se aplicavel)
  - Response body (JSON example)
  - Status codes possiveis
  - Formato de erro padrao

#### Eventos SQS e Funcoes Scheduled
- Tabela de consumers com fila, descricao, batch size, timeout
- Tabela de scheduled functions com schedule expression e descricao
- Eventos suportados com descricao do processamento

#### Stack Tecnologica
- Versoes REAIS extraidas do package.json (nao hardcoded)

#### Variaveis de Ambiente
- Tabela completa com: variavel, descricao, obrigatoria (sim/nao), valor default

#### Estrutura do Projeto
- Arvore de diretorios real (nao inventada)
- Descricao de cada diretorio/arquivo relevante

---

## Fase 5 — Validacao final

Apos reescrever o README, valide:

- [ ] Todas as entidades de `src/domain/` estao documentadas?
- [ ] Todos os endpoints de `src/interfaces/lambda/` estao documentados?
- [ ] Todos os consumers SQS estao documentados?
- [ ] Todas as scheduled functions estao documentadas?
- [ ] Versoes da stack batem com `package.json`?
- [ ] Scripts listados batem com `package.json`?
- [ ] Variaveis de ambiente estao completas?
- [ ] Nenhuma informacao foi inventada?

Se alguma validacao falhar, corrija antes de finalizar.

---

## Output

Ao final, apresente um resumo das mudancas feitas:

```
README.md - Atualizacao Completa
================================

SECOES ATUALIZADAS:
- [lista de secoes que mudaram]

PRINCIPAIS MUDANCAS:
- [lista das mudancas mais relevantes: novas entidades, novos endpoints, campos adicionados, etc.]

SECOES SEM ALTERACAO:
- [lista de secoes que permaneceram iguais]
```
