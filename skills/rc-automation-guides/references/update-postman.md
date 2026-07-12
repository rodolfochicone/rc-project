# Skill: /update-postman

Analisa os lambdas HTTP expostos no `serverless.yaml`, extrai os schemas Zod dos handlers e atualiza o arquivo `docs/RC Session Manager.postman_collection.json` para manter a collection sincronizada com o codigo atual.

## Quando usar

- Apos adicionar, remover ou alterar endpoints HTTP no `serverless.yaml`
- Apos modificar schemas Zod nos handlers Lambda
- Antes de rodar testes manuais via Postman para garantir que a collection esta atualizada

---

## Instrucoes de execucao

### Passo 1 — Mapear endpoints HTTP do serverless.yaml

Leia `serverless.yaml` e identifique todas as functions com evento `http`.
Para cada uma, extraia:

- Nome da funcao
- Path da rota (ex: `/sessions/{sessionId}`)
- Metodo HTTP (GET, POST, PUT, DELETE, PATCH)
- Caminho do handler (ex: `dist/interfaces/lambda/session/getSession.handler`)

Ignore funcoes com eventos `schedule` ou `sqs` — elas nao sao endpoints HTTP.

### Passo 2 — Extrair schemas Zod dos handlers

Para cada handler identificado no passo 1, leia o arquivo fonte correspondente em `src/interfaces/lambda/` (troque `dist/` por `src/` e `.handler` por `.ts`).

Extraia os schemas Zod definidos no arquivo:

- **Path parameters**: schema que valida `event.pathParameters` (ex: `sessionId`, `dealId`)
- **Body**: schema que valida o body do request (campos obrigatorios, opcionais, tipos, restricoes)
- **Query parameters**: schema que valida `event.queryStringParameters` (se existir)

Para cada campo, registre:
- Nome
- Tipo (string, number, boolean, array, object)
- Obrigatorio ou opcional (`.optional()`)
- Valor default se existir (`.default()`)
- Restricoes (`.uuid()`, `.min()`, `.max()`, `.enum()`, etc.)

### Passo 3 — Ler a collection atual

Leia o arquivo `docs/RC Session Manager.postman_collection.json`.
Identifique a estrutura de pastas e items existentes.
Preserve:

- A estrutura de pastas (FASE 1, FASE 2, etc.)
- Endpoints SQS consumer (simulacao local)
- Endpoints Trigger (scheduled functions)
- Variaveis da collection (`{{baseUrl}}`, `{{sessionId}}`, etc.)
- Metadados do collection (`info`, `_postman_id`, `schema`, `_exporter_id`)

### Passo 4 — Atualizar os requests HTTP

Para cada endpoint HTTP mapeado no passo 1, atualize (ou crie) o request correspondente na collection:

**URL:**
- Use `{{baseUrl}}` como host
- Path deve refletir exatamente a rota do `serverless.yaml`
- Path variables devem usar a sintaxe Postman `:variavel` (ex: `:sessionId`)
- Configure a secao `variable` da URL com valores de exemplo (UUIDs para IDs)

**Headers:**
- Inclua `Content-Type: application/json` para metodos com body (POST, PUT, PATCH)

**Body (para POST, PUT, PATCH):**
- Gere um body JSON de exemplo com todos os campos do schema Zod
- Campos obrigatorios: use valores de exemplo realistas
- Campos opcionais: inclua com valores de exemplo (para facilitar testes)
- UUIDs: use formato valido (ex: `c7e8fc66-7a43-4d72-b6fb-891e4a91c77e`)
- Phones: use formato E.164 (ex: `+5511999990001`)
- Strings genericas: use valores descritivos do campo

**Descricao:**
- Atualize a descricao do request com base no JSDoc do handler

### Passo 5 — Gerar pasta "Docs" na collection

Crie (ou substitua completamente) uma pasta chamada **"Docs"** na collection Postman. Esta pasta serve como documentacao navegavel de todos os endpoints HTTP do projeto — nao contem requests executaveis.

A pasta "Docs" deve ser o **ultimo item** da collection (apos todas as pastas de fases, SQS e Trigger).

A cada execucao do skill, a pasta "Docs" e **regenerada do zero** com base nos endpoints mapeados no Passo 1 e schemas extraidos no Passo 2.

#### Estrutura de cada item na pasta "Docs"

Para cada endpoint HTTP, crie um item com:

**Nome:**
```
[METODO] /rota/completa — Descricao curta
```
Exemplo: `[POST] /sessions/v1/resolve — Criar ou Obter Sessao`

**Request:** `null` (item puramente documental, sem request executavel)

**Descricao (campo `description` do item):** Use markdown Postman com as seguintes secoes:

```markdown
# [METODO] /rota/completa

**Proposito:** Descricao do que o endpoint faz (extraida do JSDoc do handler).

---

## Headers

| Header | Valor | Obrigatorio |
|---|---|---|
| Content-Type | application/json | Sim (se tem body) |
| Authorization | {{authorization}} | Sim (staging) |
| operation | {{operation}} | Sim (staging) |

---

## Path Parameters

| Parametro | Tipo | Obrigatorio | Exemplo | Descricao |
|---|---|---|---|---|
| sessionId | string (UUID) | Sim | `a1b2c3d4-e5f6-7890-abcd-ef1234567890` | ID da sessao |

_Se nao houver path parameters, omitir esta secao._

---

## Query Parameters

| Parametro | Tipo | Obrigatorio | Default | Exemplo | Descricao |
|---|---|---|---|---|---|
| page | number | Nao | 1 | `1` | Pagina atual |
| limit | number | Nao | 20 | `20` | Itens por pagina |

_Se nao houver query parameters, omitir esta secao._

---

## Request Body

| Campo | Tipo | Obrigatorio | Restricoes | Exemplo | Descricao |
|---|---|---|---|---|---|
| welcomeAgentId | string | Sim | UUID | `c7e8fc66-7a43-4d72-b6fb-891e4a91c77e` | ID do agente |
| channel | string | Nao | - | `whatsapp` | Canal de comunicacao |

**Exemplo completo:**
\```json
{
  "welcomeAgentId": "c7e8fc66-7a43-4d72-b6fb-891e4a91c77e",
  "channel": "whatsapp"
}
\```

_Se o metodo nao aceitar body (GET, DELETE), omitir esta secao._

---

## Responses

| Status | Descricao |
|---|---|
| 200 | Sucesso — retorna a entidade |
| 201 | Criado — nova entidade criada |
| 400 | Bad Request — erro de validacao |
| 404 | Not Found — recurso nao encontrado |
| 409 | Conflict — conflito de estado |

_Liste apenas os status codes aplicaveis ao endpoint._
```

#### Regras da pasta "Docs"

- Os valores de exemplo devem seguir as mesmas **regras de body examples** definidas na secao "Regra dos body examples" deste skill
- A descricao de cada campo deve ser extraida do contexto do schema Zod e do JSDoc do handler
- A ordem dos items na pasta "Docs" deve seguir a mesma ordem dos endpoints no `serverless.yaml`
- Se o handler nao tiver JSDoc, use o nome da funcao para inferir o proposito (ex: `getSession` → "Obter uma sessao pelo ID")
- Os status codes de response devem ser inferidos a partir do handler (retornos explicitos como `statusCode: 200`, `201`, etc.) e dos erros possiveis (validacao Zod → 400, not found → 404, etc.)

#### Formato Postman do item documental

```json
{
  "name": "[POST] /sessions/v1/resolve — Criar ou Obter Sessao",
  "item": [],
  "description": "# [POST] /sessions/v1/resolve\n\n**Proposito:** Cria uma nova sessao ou retorna uma sessao ativa existente...\n\n---\n\n## Headers\n\n| Header | Valor | Obrigatorio |\n|---|---|---|\n..."
}
```

> **Nota:** No Postman, folders (pastas) aceitam o campo `description` em markdown. Cada endpoint documental e representado como uma **sub-folder vazia** (com `item: []`) dentro da pasta "Docs", permitindo que a descricao apareca como documentacao navegavel no Postman.

### Passo 6 — Criar/Atualizar Environments Postman


Crie (ou atualize) dois arquivos de ambiente Postman na pasta `docs/`:

#### 6a — Environment Local (`docs/RC Session Manager - Local.postman_environment.json`)

Variaveis:
- `baseUrl`: `http://localhost:3000` — URL do serverless-offline local
- `sessionId`: `00000000-0000-0000-0000-000000000000` — UUID placeholder para testes
- `dealId`: `d1e2f3a4-b5c6-7890-abcd-ef1234567890` — UUID placeholder para testes

Nao incluir headers de autenticacao — o ambiente local nao exige auth.

#### 6b — Environment Staging (`docs/RC Session Manager - Staging.postman_environment.json`)

Variaveis:
- `baseUrl`: `https://api.staging.escale.com.br` — URL do API Gateway em staging
- `sessionId`: `00000000-0000-0000-0000-000000000000` — UUID placeholder para testes
- `dealId`: `d1e2f3a4-b5c6-7890-abcd-ef1234567890` — UUID placeholder para testes
- `operation`: `6926a9fe-6332-46a5-9a08-b2353ecdf431` — Header obrigatorio de operacao
- `authorization`: `Basic <base64 de client_id:client_secret — obter no vault/1Password do time; NUNCA commitar o valor real>` — Header de autenticacao Basic

Para staging, adicione os seguintes headers em **todos os requests** da collection (ou configure auth herdavel no nivel da collection):
- `operation: {{operation}}`
- `Authorization: {{authorization}}`

Estes headers devem ser configurados via variavel de ambiente para facilitar a troca entre ambientes.

#### Formato do arquivo de environment

Use o schema padrao do Postman:
```json
{
  "id": "<uuid>",
  "name": "<nome do environment>",
  "values": [
    {
      "key": "<nome>",
      "value": "<valor>",
      "type": "default",
      "enabled": true
    }
  ],
  "_postman_variable_scope": "environment",
  "_postman_exported_at": "<ISO 8601>",
  "_postman_exported_using": "Postman/11.0.0"
}
```

#### Regra dos body examples

Ao gerar os bodies de exemplo nos requests, use valores que correspondam aos **tipos reais das entidades do dominio** (`src/domain/`):
- `id`, `sessionId`, `dealId`, `transferId` → UUID valido (ex: `a1b2c3d4-e5f6-7890-abcd-ef1234567890`)
- `senderPhoneNumber`, `receiverPhoneNumber` → formato E.164 (ex: `+5511999990001`)
- `channel` → string (ex: `whatsapp`)
- `modality` → string (ex: `messaging`)
- `status` → string do enum correspondente (ex: `active`, `ended`, `created`)
- `endedReason` → string (ex: `closed`, `timeout`, `connection_lost`, `system_terminated`)
- `currentAgentId`, `fromAgentId`, `toAgentId` → string (ex: `bot-ia-001`, `humano-001`)
- `currentAgentType`, `toAgentType` → string (ex: `bot`, `human`)
- `transferReason` → string (ex: `escalation`, `timeout`, `manual`)
- `startedAt`, `endedAt`, `timestamp` → ISO 8601 datetime (ex: `2026-03-05T10:00:00.000Z`)
- `content` → `Record<string, unknown>` (ex: `{ "text": "Ola, preciso de ajuda" }`)
- `metadata` → `Record<string, unknown> | null` (ex: `{ "closedBy": "user" }`)
- `isForwarded` → boolean (ex: `false`)
- `messageType` → string (ex: `text`, `image`, `audio`)
- `senderSide` → string (ex: `customer`, `agent`)
- `protocol` → string (ex: `20260305-DEAL123-001`)
- `dealProtocol` → string | null (ex: `DEAL-2026-001`)
- Campos nullable devem aparecer como `null` nos exemplos

### Passo 7 — Validar e salvar

- Certifique-se de que o JSON resultante e valido (collection + environments)
- Mantenha compatibilidade com Postman Collection v2.1.0
- Salve os arquivos:
  - `docs/RC Session Manager.postman_collection.json`
  - `docs/RC Session Manager - Local.postman_environment.json`
  - `docs/RC Session Manager - Staging.postman_environment.json`

### Passo 8 — Reportar diferencas

Apresente um resumo das alteracoes feitas:

```
## Resumo da atualizacao

### Endpoints atualizados
- [metodo] /path — descricao da mudanca

### Endpoints adicionados
- [metodo] /path — novo endpoint

### Endpoints removidos (da collection, nao existem mais no serverless.yaml)
- [metodo] /path

### Sem alteracao
- [metodo] /path — ja estava sincronizado

### Pasta Docs
- Total de endpoints documentados: N
- Endpoints adicionados a Docs: [lista]
- Endpoints removidos de Docs: [lista]
```

---

## Regras

- **Nao altere** a variavel `{{baseUrl}}` nem outras variaveis da collection
- **Nao remova** endpoints SQS consumer ou Trigger — eles sao uteis para testes locais
- **Mantenha** o schema `https://schema.getpostman.com/json/collection/v2.1.0/collection.json`
- Se um endpoint HTTP do serverless.yaml nao existir na collection, adicione-o na pasta mais adequada
- Se um endpoint HTTP da collection nao existir mais no serverless.yaml, remova-o e reporte
- **Mantenha** os arquivos de environment sincronizados com as variaveis usadas na collection
- **Staging** sempre deve incluir as variaveis `operation` e `authorization` nos headers
- **Body examples** devem usar valores coerentes com os tipos das entidades do dominio (`src/domain/`)
- **Pasta "Docs"** e regenerada do zero a cada execucao — nunca faca merge incremental, substitua completamente
- **Pasta "Docs"** deve ser o ultimo item da collection (apos todas as outras pastas)
