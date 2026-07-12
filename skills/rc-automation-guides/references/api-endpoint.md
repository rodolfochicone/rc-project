# Skill: /api-endpoint

Guia para criar um novo endpoint REST seguindo `agent-docs/API.md` e `agent-docs/ARCHITECTURE.md`.

## O que fazer ao invocar este skill

1. Leia os documentos de contexto:
   - `agent-docs/API.md` — nomenclatura de rotas, paginacao, filtros, resposta padrao
   - `agent-docs/ARCHITECTURE.md` — camadas, DI, controller/handler
   - `agent-docs/SECURITY.md` — validacao de inputs com Zod

2. Pergunte ao usuario:
   - Qual o recurso? (ex: `orders`, `payments`, `invoices`)
   - Qual a operacao? (listar, criar, buscar por ID, atualizar, deletar, acao customizada)
   - Quais campos do request body/query params?
   - O endpoint ja existe no usecase ou precisa criar do zero?

3. Execute o checklist abaixo.

---

## Checklist

### Passo 1 — Definir a rota

Formato: `<METHOD> /v1/<recursos>`

```
GET    /v1/{recursos}              -> lista paginada
POST   /v1/{recursos}              -> criacao
GET    /v1/{recursos}/{id}         -> busca por ID
PUT    /v1/{recursos}/{id}         -> substituicao completa
PATCH  /v1/{recursos}/{id}         -> atualizacao parcial
DELETE /v1/{recursos}/{id}         -> remocao (soft delete)
```

Regras:
- Recursos sempre no **plural**, **kebab-case**
- Sem trailing slash, sem extensoes (`.json`)
- Acoes nao-CRUD: `POST /v1/orders/{id}/cancel` (sub-recurso com POST)
- Maximo 2 niveis de aninhamento

### Passo 2 — Schema de validacao (Zod)

Defina o schema junto ao controller/handler:

```typescript
import { z } from 'zod';

const CreateOrderSchema = z.object({
  customerId: z.string().uuid(),
  items: z.array(
    z.object({
      productId: z.string().uuid(),
      quantity: z.number().int().positive(),
    })
  ).min(1),
  total: z.number().positive(),
});

type CreateOrderInput = z.infer<typeof CreateOrderSchema>;
```

Validacoes minimas:
- IDs: `uuid()` | Strings: `min(1)`, `max(N)` | Enums: `z.enum([...])` | Datas: `z.string().datetime()`

### Passo 3 — Tratamento de erro de validacao

```typescript
const result = CreateOrderSchema.safeParse(req.body);
if (!result.success) {
  const details = result.error.issues.map(issue => ({
    field: issue.path.join('.'),
    message: issue.message,
  }));
  return res.status(400).json({
    error: { code: 'VALIDATION_ERROR', message: 'Request validation failed.', details },
  });
}
```

### Passo 4 — Contrato de resposta

**Recurso unico** — sem envelope:
```json
{ "id": "123", "status": "pending", "total": 99.90, "createdAt": "2025-01-15T10:30:00Z" }
```

**Colecao paginada:**
```json
{
  "data": [...],
  "pagination": { "page": 1, "limit": 20, "total": 150, "totalPages": 8, "nextPage": "...", "prevPage": null }
}
```

**Erro:**
```json
{ "error": { "code": "ORDER_NOT_FOUND", "message": "Order with id '123' was not found.", "details": [] } }
```

### Passo 5 — Codigos HTTP

| Status | Quando usar |
|---|---|
| `200 OK` | GET, PUT, PATCH com sucesso |
| `201 Created` | POST com criacao de recurso |
| `204 No Content` | DELETE ou acao sem retorno |
| `400 Bad Request` | Payload invalido ou erro de validacao |
| `404 Not Found` | Recurso nao encontrado |
| `409 Conflict` | Conflito de estado |
| `422 Unprocessable Entity` | Payload valido, mas invalido para o negocio |

Nunca retorne `200` com body de erro.

### Passo 6 — Paginacao e filtros (se listagem)

Query parameters padrao:
- `page` (default: 1), `limit` (default: 20, max: 100), `sort` (formato `campo:direcao`)
- Filtros: `filter[campo][operador]=valor`
- Operadores: igualdade (sem operador), `contains`, `gte`, `lte`, `is_null`
- Use `API_BASE_PATH` para construir `nextPage`/`prevPage`

### Passo 7 — Headers de contexto

Leia os headers injetados pelo API Gateway Authorizer:

```typescript
const context = {
  partner: req.headers['partner'] as string,
  product: req.headers['product'] as string,
  user: req.headers['user'] as string | undefined,
  email: req.headers['email'] as string | undefined,
};
```

### Passo 8 — Controller (server) ou Handler (lambda)

- Controller: arrow methods, delega para usecase, sem logica de negocio
- Handler: importar `Context` de `aws-lambda`, definir `logger.setCorrectionId(context.awsRequestId)`
- Atualizar `dependencies.ts` em AMBOS (`server/` e `lambda/`)

### Passo 9 — serverless.yaml

Registre a funcao em `serverless.yaml` E `serverless.dev.yaml`:

```yaml
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

### Passo 10 — OpenAPI

Atualize `/docs/openapi.yaml` com:
- Endpoint, parametros e responses documentados
- Schemas de request body e response
- Codigos de erro possiveis
- Headers de contexto documentados

---

## Verificacao final

```bash
pnpm test    # 0 falhas
pnpm build   # sem erros de compilacao
```
