# Skill: /update-openapi

Analisa os lambdas HTTP expostos no `serverless.yaml`, extrai os schemas Zod dos handlers e as entidades do domain, e cria ou atualiza o arquivo `.techdocs/docs/openapi.yaml` para manter a especificacao OpenAPI sincronizada com o codigo atual.

## Quando usar

- Apos adicionar, remover ou alterar endpoints HTTP no `serverless.yaml`
- Apos modificar schemas Zod nos handlers Lambda
- Apos alterar entidades do domain que afetam responses
- Antes de publicar documentacao da API ou integrar com ferramentas como Swagger UI / Redoc
- Quando o `/api-endpoint` pedir para atualizar o OpenAPI (passo 10)

---

## Instrucoes de execucao

### Passo 1 — Verificar existencia do arquivo OpenAPI

Verifique se o arquivo `.techdocs/docs/openapi.yaml` ja existe:

- **Se existir**: leia o conteudo atual para comparacao no passo 6
- **Se nao existir**: sera criado do zero no passo 5

### Passo 2 — Mapear endpoints HTTP do serverless.yaml

Leia `serverless.yaml` e identifique todas as functions com evento `http`.
Para cada uma, extraia:

- Nome da funcao
- Path da rota (ex: `/v1/{sessionId}`)
- Metodo HTTP (GET, POST, PUT, DELETE, PATCH)
- Caminho do handler (ex: `dist/interfaces/lambda/session/getSession.handler`)
- Comentarios/descricao do bloco YAML (usados para `summary` e `description` do endpoint)

Ignore funcoes com eventos `schedule` ou `sqs` — elas nao sao endpoints HTTP.

### Passo 3 — Extrair schemas Zod dos handlers

Para cada handler identificado no passo 2, leia o arquivo fonte correspondente em `src/interfaces/lambda/` (troque `dist/` por `src/` e `.handler` por `.ts`).

Extraia os schemas Zod definidos no arquivo:

- **Path parameters**: schema que valida `event.pathParameters` (ex: `sessionId`, `dealId`)
- **Body**: schema que valida o body do request (campos obrigatorios, opcionais, tipos, restricoes)
- **Query parameters**: schema que valida `event.queryStringParameters` (se existir)

Para cada campo, registre:
- Nome
- Tipo (string, number, boolean, array, object)
- Obrigatorio ou opcional (`.optional()`)
- Valor default se existir (`.default()`)
- Nullable se existir (`.nullable()`)
- Restricoes (`.uuid()`, `.min()`, `.max()`, `.datetime()`, `.enum()`, etc.)
- Mensagens de erro customizadas (usadas para `description`)

### Passo 4 — Extrair response schemas do domain e handlers

Para cada endpoint, analise:

1. **Entidades do domain** em `src/domain/` — os campos da entidade definem o schema de response
2. **Transformacoes no handler** — ex: `{ id, ...rest }` vira `{ sessionId: id, ...rest }` (renomeacao de campos)
3. **Codigos HTTP** retornados no handler (`200`, `201`, `400`, `404`, `409`)

Mapeie os tipos TypeScript para tipos OpenAPI:
- `string` → `type: string`
- `string` com `.uuid()` → `type: string, format: uuid`
- `string` com `.datetime()` → `type: string, format: date-time`
- `number` → `type: number`
- `boolean` → `type: boolean`
- `Date` → `type: string, format: date-time`
- `null` / `.nullable()` → `nullable: true`
- `unknown` / `Record<string, unknown>` → `type: object, additionalProperties: true`

### Passo 5 — Gerar ou montar o OpenAPI

Monte o arquivo `.techdocs/docs/openapi.yaml` com a seguinte estrutura completa:

```yaml
openapi: 3.0.3

info:
  title: RC Session Manager API
  description: >-
    API para gerenciamento de sessoes de conversacao vinculadas a deals do CRM.
    Controla ciclo de vida de sessoes, transferencias de agentes e mapeamento
    de comunicacoes externas.
  version: <versao do package.json>
  contact:
    name: Escale Engineering
```

**servers** — ambientes staging e production:
```yaml
servers:
  - url: https://api.staging.escale.com.br/sessions
    description: Staging
  - url: https://api.escale.com.br/sessions
    description: Production
```

**security** — esquema de autenticacao Token:
```yaml
security:
  - tokenAuth: []

components:
  securitySchemes:
    tokenAuth:
      type: apiKey
      in: header
      name: Authorization
      description: Token de autenticacao gerenciado pelo API Gateway Authorizer
```

**tags** — agrupamento por dominio:
```yaml
tags:
  - name: Sessions
    description: Operacoes de ciclo de vida de sessoes
  - name: Transfers
    description: Transferencias de agente dentro de sessoes
  - name: Communication Mappings
    description: Mapeamento entre comunicacoes externas e internas
```

**paths** — para cada endpoint HTTP:
- `summary`: descricao curta (do comentario YAML ou JSDoc do handler)
- `description`: descricao detalhada da operacao
- `operationId`: nome da funcao do serverless.yaml (ex: `getSession`)
- `tags`: tag do dominio correspondente
- `parameters`: path params e query params extraidos do Zod
- `requestBody`: body schema extraido do Zod (quando aplicavel)
- `responses`: todos os codigos de retorno possiveis com schemas

**components/schemas** — schemas reutilizaveis:
- `Session` — entidade principal com todos os campos
- `Transfer` — entidade de transferencia
- `SessionCommunicationMapping` — entidade de mapeamento
- `ErrorResponse` — contrato padrao de erro:
  ```yaml
  ErrorResponse:
    type: object
    required: [error]
    properties:
      error:
        type: object
        required: [code, message, details]
        properties:
          code:
            type: string
            example: VALIDATION_ERROR
          message:
            type: string
            example: Falha na validacao da requisicao.
          details:
            type: array
            items:
              type: object
              properties:
                field:
                  type: string
                message:
                  type: string
  ```
- `ValidationErrorDetail` — detalhe de erro de validacao
- Request body schemas individuais (ex: `CreateOrGetSessionRequest`, `EndSessionRequest`, `TransferAgentRequest`, `GetSessionCommunicationMappingRequest`)

**Regras para os schemas:**
- Campos obrigatorios devem estar na lista `required`
- Campos opcionais com default devem ter `default` especificado
- Campos nullable devem ter `nullable: true`
- Usar `$ref` para referenciar schemas em `components/schemas` — nunca duplicar definicoes inline
- Incluir `example` para cada campo com valores realistas (UUIDs validos, phones E.164, etc.)

### Passo 6 — Comparar e aplicar diferencas (se arquivo ja existia)

Se o arquivo ja existia (passo 1), compare o OpenAPI atual com o gerado:

1. **Endpoints novos**: adicione ao `paths`
2. **Endpoints removidos**: remova do `paths` (nao existe mais no serverless.yaml)
3. **Endpoints alterados**: atualize parametros, body, responses conforme o Zod atual
4. **Schemas alterados**: atualize em `components/schemas`
5. **Preservar**: customizacoes manuais que nao conflitem (descricoes extras, exemplos adicionais)

### Passo 7 — Validar o YAML

- Certifique-se de que o YAML resultante e sintaticamente valido
- Certifique-se de que todas as `$ref` apontam para schemas existentes
- Certifique-se de que todos os campos `required` listam apenas propriedades existentes
- Certifique-se de que a spec segue OpenAPI 3.0.3

### Passo 8 — Salvar o arquivo

Salve o arquivo em `.techdocs/docs/openapi.yaml`.

### Passo 9 — Reportar resultado

Apresente um resumo das alteracoes:

```
## Resultado — OpenAPI

### Modo: [CRIACAO | ATUALIZACAO]

### Endpoints documentados
- [metodo] /path — descricao

### Endpoints adicionados (se atualizacao)
- [metodo] /path — novo endpoint

### Endpoints removidos (se atualizacao)
- [metodo] /path — removido do codebase

### Endpoints atualizados (se atualizacao)
- [metodo] /path — descricao da mudanca

### Schemas gerados
- Schema1, Schema2, ...

### Proximos passos
- Validar importando em https://editor.swagger.io
- Revisar descricoes e exemplos
```

---

## Regras

- **Formato**: YAML (nao JSON) — mais legivel e consistente com o resto do projeto
- **Versao**: OpenAPI 3.0.3
- **Idioma das descricoes**: portugues (consistente com as mensagens de log do projeto). Mensagens de erro nos exemplos de response devem ser em ingles
- **Nao invente endpoints**: documente apenas o que existe no `serverless.yaml` com evento `http`
- **Nao invente campos**: extraia apenas do Zod (request) e domain entities (response)
- **Schemas reutilizaveis**: use `$ref` para evitar duplicacao — principio DRY
- **Exemplos realistas**: UUIDs validos, phones no formato E.164, datas ISO 8601
- **Destino fixo**: sempre `.techdocs/docs/openapi.yaml` — nao crie em outro local
- **Nao altere codigo**: este command apenas gera/atualiza documentacao
