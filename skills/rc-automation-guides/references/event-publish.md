# Skill: /event-publish

Guia para publicar e consumir eventos via EventBridge/SQS, seguindo `agent-docs/API.md` secao 7.

## O que fazer ao invocar este skill

1. Leia os documentos de contexto:
   - `agent-docs/API.md` — secao 7 (Padrao de Eventos)
   - `agent-docs/ARCHITECTURE.md` — secao Consumer SQS

2. Pergunte ao usuario:
   - Vai **publicar** ou **consumir** um evento?
   - Qual o recurso (`object`)? (ex: `order`, `payment`, `invoice`)
   - Qual a acao (`action`)? (ex: `created`, `updated`, `approved`, `cancelled`)
   - Quais dados no `body` do evento?

---

## Publicar evento

### Passo 1 — Usar `@escaletech/event-publisher`

```typescript
import { logger } from '@escaletech/logger';
import { EscaleEventPublisher } from '@escaletech/event-publisher';

const publisher = new EscaleEventPublisher(process.env.APP_ENV, logger);
```

### Passo 2 — Envelope do evento

```typescript
await publisher.publishEvent([{
  object: 'order',           // recurso em singular, minusculas
  action: 'created',         // acao em minusculas
  state: 'pending',          // estado apos a acao
  body: {                    // dados relevantes — sem dados sensiveis
    id: order.id,
    total: order.total,
  },
  coi: order.id,             // identificador central da operacao
  partner: context.partner,  // propagado do header
  product: context.product,  // propagado do header
  operation: context.operation,
  source: 'nome-do-servico', // nome do servico que publica
}]);
```

### Passo 3 — Nomenclatura

```
object: 'order'    action: 'created'       -- CORRETO (singular, minusculas)
object: 'orders'   action: 'createOrder'   -- ERRADO (plural, verbo composto)
```

### Passo 4 — Regras obrigatorias

- Publicar **apos** a operacao ser persistida com sucesso — NUNCA antes
- Sempre propagar `partner`, `product`, `coi` do contexto original
- `body` contem apenas dados necessarios para consumidores — payloads enxutos
- Nunca incluir dados sensiveis no `body` (senhas, tokens, cartoes)

---

## Consumir evento

### Passo 1 — Registrar fila no builder-event-platform

A configuracao de roteamento EventBridge -> SQS e feita no repositorio `builder-event-platform`:

1. Abra PR no `builder-event-platform` definindo a regra de roteamento e fila SQS
2. Configure DLQ (Dead Letter Queue) para a fila
3. O servico consumer implementa o handler apontando para a fila provisionada

**Nunca configure filas diretamente no servico.**

### Passo 2 — Criar Consumer

Em `src/interfaces/sqs/consumers/`:

```typescript
/**
 * Consumer para processar eventos de order.created
 */
export class OrderCreatedConsumer {
  constructor(private readonly processOrder: ProcessOrderUseCase) {}

  async handle(message: SQSMessage): Promise<void> {
    const input = JSON.parse(message.Body) as ProcessOrderInput;
    await this.processOrder.execute(input);
  }
}
```

### Passo 3 — Regras do consumer

- Consumer **nunca contem logica de negocio** — delega ao usecase imediatamente
- Consumer deve ser **idempotente** — a mesma mensagem pode chegar mais de uma vez
- Use `coi` + `object` + `action` para deduplicacao quando necessario
- Em caso de erro, **nao delete a mensagem** — deixe o SQS reprocessar via retry
- Valide o body da mensagem com Zod antes de processar
- Defina correlation ID: `logger.setCorrectionId(message.MessageId)`

### Passo 4 — Dependencies e registro

- Adicione o consumer em `dependencies.ts` (server e lambda)
- Registre o consumer no `consumer.ts` da infraestrutura
- Configure o handler SQS no `serverless.yaml` se aplicavel

---

## Fluxo completo

```
Servico (publisher)
     |  publishEvent()
     v
AWS EventBridge (bus central)
     |  regras configuradas no builder-event-platform
     v
AWS SQS (fila do consumer)
     |
     v
Servico (consumer) -> UseCase -> Domain
```
