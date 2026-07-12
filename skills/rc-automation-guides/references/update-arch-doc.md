# Skill: /update-arch-doc

Analisa o codebase do projeto (domain, usecases, infrastructure, migrations, serverless.yaml, handlers) e gera ou atualiza o arquivo `.techdocs/docs/arquitetura-geral.md` com a documentacao completa de arquitetura geral do componente.

## Quando usar

- Apos adicionar, remover ou alterar entidades do domain
- Apos criar novos endpoints, consumers ou scheduled jobs
- Apos alterar integracoes com sistemas externos (CRM, EventBridge, SQS)
- Apos criar ou alterar migrations que impactam o modelo de dados
- Quando precisar documentar o estado atual da arquitetura do componente
- Periodicamente para manter a documentacao sincronizada com o codigo

---

## Instrucoes de execucao

### Passo 1 — Verificar existencia do arquivo

Verifique se `.techdocs/docs/arquitetura-geral.md` ja existe:

- **Se existir**: leia o conteudo atual para preservar informacoes manuais que nao conflitem com o codigo
- **Se nao existir**: sera criado do zero no passo final

### Passo 2 — Coletar informacoes do codebase

Leia e analise os seguintes artefatos do projeto:

#### 2.1 Domain (entidades e contratos)

Leia todos os arquivos em `src/domain/` para identificar:

- Entidades principais (interfaces/types que representam objetos de negocio)
- Contratos de repositorio (interfaces de persistencia)
- Contratos de gateways (interfaces de integracao)
- Contratos de event publisher
- Enums e value objects

#### 2.2 Use cases

Leia todos os arquivos em `src/usecases/` para identificar:

- Servicos de caso de uso (classes com metodo `execute`)
- Inputs e outputs de cada use case
- Dependencias injetadas (repositorios, gateways, publishers)

#### 2.3 Endpoints HTTP

Leia `serverless.yaml` e os handlers em `src/interfaces/lambda/` para mapear:

- Metodo HTTP, path, nome do handler
- Schema Zod de entrada (path params, query params, body)
- Descricao/comentarios do handler

#### 2.4 Consumers SQS

Leia os consumers em `src/interfaces/sqs/` e/ou handlers de eventos em `src/interfaces/lambda/events/` para identificar:

- Filas consumidas
- Tipos de evento processados
- Logica de roteamento

#### 2.5 Scheduled Jobs

Leia `serverless.yaml` para identificar functions com evento `schedule`:

- Nome do job
- Frequencia/cron
- Proposito

#### 2.6 Migrations

Leia os arquivos em `migrations/` para mapear:

- Tabelas criadas e seus campos
- Tipos de dados, constraints, indices
- Evolucao do schema ao longo do tempo

#### 2.7 Integracoes

Leia `src/infrastructure/` para identificar:

- Gateways HTTP (ex: CRM)
- Publishers de eventos (ex: EventBridge)
- Outros servicos externos

### Passo 3 — Montar o documento

Monte o arquivo `.techdocs/docs/arquitetura-geral.md` seguindo **exatamente** a estrutura de topicos abaixo. Cada secao deve ser preenchida com informacoes reais extraidas do codebase.

---

## Estrutura obrigatoria do documento gerado

O documento DEVE seguir esta estrutura. Todas as secoes sao obrigatorias. Se uma secao nao se aplicar ao projeto atual, inclua-a com a nota "(Nao aplicavel ao estado atual do componente)".

```markdown
DOCUMENTO DE PRODUTO  ·  COMPONENTE

# **[Nome do Componente]**

[Descricao curta do componente]

[Lista de entidades/dominios principais separados por " · "]

Versao X.Y  ·  [Mes Ano]

---

## **1. O QUE E**

### 1. O que e

[Descricao do que o componente faz, qual seu papel na plataforma, e qual problema resolve]

#### 1.1 Responsabilidades do [Nome do Componente]

* [Responsabilidade 1]
* [Responsabilidade 2]
* [...]

#### 1.2 O que o [Nome do Componente] nao faz

**Fora do escopo deste componente**

* [O que nao faz 1 — e de responsabilidade de qual componente]
* [O que nao faz 2]
* [...]

#### 1.3 Modelo Multi-Tenant

[Descreva o modelo de isolamento por tenant. Se nao aplicavel, indique]

| Caracteristica | Detalhe |
| :---- | :---- |
| Isolamento de dados | [Descricao] |
| [Outros aspectos] | [Descricao] |

#### 1.4 Entidades

[Descricao geral das entidades do dominio e seus relacionamentos]

##### **[Nome da Entidade 1]**

[Descricao da entidade, seu papel e restricoes]

##### **[Nome da Entidade 2]**

[Descricao da entidade, seu papel e restricoes]

[... repita para cada entidade]

---

## **2. FEATURES**

### 2. Features

#### 2.1 [Nome da Feature/Dominio 1]

* [Capacidade 1]
* [Capacidade 2]
* [...]

#### 2.2 [Nome da Feature/Dominio 2]

* [Capacidade 1]
* [Capacidade 2]
* [...]

[... repita para cada feature/dominio]

---

## **3. REQUISITOS**

### 3. Requisitos

#### 3.1 Requisitos Funcionais

| ID | Descricao |
| :---- | :---- |
| **RF-01** | [Descricao extraida do codigo/comportamento] |
| **RF-02** | [Descricao] |
| [...]  | [...] |

#### 3.2 Requisitos Nao Funcionais

| ID | Descricao |
| :---- | :---- |
| **RNF-01** | [Descricao] |
| **RNF-02** | [Descricao] |
| [...]  | [...] |

#### 3.3 Regras de Negocio

| ID | Descricao |
| :---- | :---- |
| **RN-01** | [Regra extraida do domain/usecases] |
| **RN-02** | [Regra] |
| [...]  | [...] |

---

## **4. ARQUITETURA**

### 4. Arquitetura

[Descricao geral da arquitetura, camadas e organizacao]

### 4.1 Modelo de Infraestrutura

[Descreva como o componente e deployado: Lambda, containers, etc. Modelo multi-tenant se aplicavel]

### 4.2 Entidades e Campos

[Para cada tabela/entidade persistida, inclua uma tabela com os campos]

#### [NOME_DA_TABELA]

[Descricao da tabela]

| Campo | Tipo | Restricao | Descricao |
| :---- | :---- | :---- | :---- |
| id | UUID | PK, NOT NULL | Chave primaria interna |
| [campo] | [tipo] | [restricao] | [descricao] |
| created_at | TIMESTAMPTZ | NOT NULL | Data/hora de criacao |
| updated_at | TIMESTAMPTZ | NOT NULL | Data/hora da ultima atualizacao |
| deleted_at | TIMESTAMPTZ | NULLABLE | Data/hora do soft delete |

**Unicidade**

* [Constraints de unicidade]

**Soft delete**

* [Regras de soft delete, se aplicavel]

[... repita para cada tabela]

### 4.3 Campos Customizados por Operacao

[Se o componente suporta campos customizados via JSONB ou schema dinamico, descreva aqui. Caso contrario, indique "Nao aplicavel"]

### 4.4 Audit Log (Trilha de Auditoria)

[Se o componente mantém trilha de auditoria, descreva as tabelas e mecanismo. Caso contrario, indique "Nao aplicavel"]

### 4.5 Integracoes com outros sistemas

[Descreva cada integracao: qual sistema, como se comunica (HTTP, eventos, fila), e o que troca]

**Principios**

* [Principio 1 da integracao]
* [Principio 2]

#### 4.5.1 [Nome da Integracao 1]

[Descricao detalhada: meio de integracao, payload, idempotencia]

#### 4.5.2 [Nome da Integracao 2]

[Descricao detalhada]

### 4.6 Fluxos principais

[Descreva os fluxos de negocio principais do componente, passo a passo]

#### 4.6.1 [Nome do Fluxo 1]

1. [Passo 1]
2. [Passo 2]
3. [...]

#### 4.6.2 [Nome do Fluxo 2]

1. [Passo 1]
2. [...]

### 4.7 Regras de integridade, constraints e indices sugeridos

#### 4.7.1 Integridade referencial

* [tabela.campo] -> [tabela_referenciada.campo]
* [...]

#### 4.7.2 Uniqueness (ativos)

* [tabela]: UNIQUE([campos]) WHERE deleted_at IS NULL
* [...]

#### 4.7.3 Regras de integridade de dominio

* [Regra 1]
* [Regra 2]

#### 4.7.4 Indices recomendados (consultas criticas)

##### **[nome_tabela]**

* ([campos]) [filtro se aplicavel]
* [...]

[... repita para cada tabela]

### 4.8 Fluxo de Eventos e Auditoria

[Descreva como eventos sao publicados, o padrao outbox se usado, e o fluxo de auditoria]

#### 4.8.1 Eventos publicados

| Evento | Quando | Payload (resumo) |
| :---- | :---- | :---- |
| [NomeDoEvento] | [Condicao de disparo] | [Campos principais] |
| [...] | [...] | [...] |

#### 4.8.2 Eventos consumidos

| Evento | Origem | Acao no componente |
| :---- | :---- | :---- |
| [NomeDoEvento] | [Sistema de origem] | [O que o componente faz ao receber] |
| [...] | [...] | [...] |

#### 4.8.3 Padrao Outbox (se aplicavel)

[Descreva o mecanismo de outbox: tabela, poller, cleanup]

### 4.9 API — Contratos

[Descricao geral dos padroes de API: autenticacao, paginacao, formato de erro]

**Padroes gerais**

* [Padrao 1]
* [Padrao 2]

#### 4.9.1 [Dominio/Recurso 1] — Rotas

##### **Padroes de resposta ([Recurso])**

[Exemplo de shape JSON do objeto]

##### **Tabela de Rotas**

| Metodo | Rota | Descricao | Query params / Body (resumo) |
| :---- | :---- | :---- | :---- |
| GET | /v1/[recurso] | [Descricao] | [Params] |
| POST | /v1/[recurso] | [Descricao] | [Body] |
| [...]  | [...]  | [...]  | [...]  |

##### **Detalhes de comportamento**

* **[METODO] [rota]**: [descricao detalhada do comportamento, validacoes, regras]
* [...]

[... repita para cada dominio/recurso]

#### 4.9.N Scheduled Jobs

| Handler | Schedule | Descricao |
| :---- | :---- | :---- |
| [nome] | [cron/rate] | [O que faz] |
| [...] | [...] | [...] |

#### 4.9.N+1 SQS Consumers

| Handler | Fila | Eventos processados | Descricao |
| :---- | :---- | :---- | :---- |
| [nome] | [nome da fila] | [tipos de evento] | [O que faz] |
| [...] | [...] | [...] | [...] |
```

### Passo 4 — Preencher com dados reais

Substitua todos os placeholders `[...]` com informacoes reais extraidas do codebase nos passos 2.1 a 2.7.

Regras de preenchimento:

- **Entidades e campos**: extraia dos arquivos de migration mais recentes (estado final do schema)
- **Rotas**: extraia do `serverless.yaml` e dos schemas Zod dos handlers
- **Regras de negocio**: extraia das validacoes nos usecases e domain
- **Integracoes**: extraia dos gateways em `src/infrastructure/`
- **Eventos**: extraia dos tipos de evento e do event publisher
- **Indices**: extraia das migrations que criam indices
- **Versao**: use a versao do `package.json`
- **Data**: use o mes e ano atuais

### Passo 5 — Escrever o arquivo

- **Se o arquivo nao existia**: crie `.techdocs/docs/arquitetura-geral.md` com o conteudo completo
- **Se o arquivo ja existia**: atualize preservando a estrutura de topicos. Se houver secoes com conteudo manual que nao conflita com o codigo, preserve-as

### Passo 6 — Validacao final

Verifique que o documento gerado:

- [ ] Contem todas as secoes obrigatorias (1 a 4, incluindo subsecoes)
- [ ] Todas as tabelas de entidades tem Campo, Tipo, Restricao e Descricao
- [ ] Todas as tabelas de rotas tem Metodo, Rota, Descricao e Params/Body
- [ ] Tabelas de requisitos tem ID e Descricao
- [ ] Nenhum placeholder `[...]` permaneceu no documento
- [ ] Conteudo esta em portugues
- [ ] Dados correspondem ao estado atual do codebase

Reporte ao usuario o resultado da validacao e quaisquer secoes que precisaram de "(Nao aplicavel)".
