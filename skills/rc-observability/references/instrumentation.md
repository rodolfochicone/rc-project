# Instrumentação — logs, métricas, traces

Objetivo: dado um serviço em produção (Lambda, task ECS, consumer de SQS, handler de
EventBridge), garantir que dá pra responder "o que aconteceu?" sem precisar reproduzir
o bug localmente. Ordem de prioridade: **correlation ID → log estruturado → métrica → trace**.
Um trace sem correlation ID não conecta com o log do serviço vizinho; comece pela propagação.

## Os três pilares — quando usar cada um

| Pilar | Responde | Quando é o sinal certo |
| ----- | -------- | ---------------------- |
| **Logs** | "O que aconteceu, exatamente, nesta execução?" | Evento discreto, payload de erro, decisão de negócio, debug pontual |
| **Métricas** | "Como está a saúde do sistema ao longo do tempo?" | Taxa de erro, latência, throughput, saturação de recurso — o que vira alerta e dashboard |
| **Traces** | "Por onde essa requisição/mensagem passou e onde ela gastou tempo?" | Fluxo cross-service (API → Lambda → SQS → outro consumer), latência distribuída |

Não confunda os três: métrica não substitui log (perde o "porquê"), log não substitui métrica
(não agrega barato em escala), trace não substitui os dois (mostra o caminho, não o conteúdo).

## Logging estruturado

Sempre JSON, um objeto por linha, nunca `print`/`console.log` de string livre em produção.

```json
{
  "timestamp": "2026-07-08T14:32:01.104Z",
  "level": "error",
  "service": "order-processor",
  "correlation_id": "c4f2b8e1-...",
  "message": "failed to persist order after retries",
  "order_id": "ord_9182",
  "retry_count": 3,
  "error": "ConditionalCheckFailedException"
}
```

Níveis — use-os com disciplina, não todo log é `error`:

| Nível | Quando |
| ----- | ------ |
| `error` | Falhou e alguém (humano ou processo) precisa agir; entra em causa raiz de incidente |
| `warn` | Degradado mas seguiu (retry que funcionou, fallback ativado, config ausente com default) |
| `info` | Marco de negócio relevante (pedido criado, mensagem processada, deploy concluído) |
| `debug` | Detalhe só útil investigando; nunca ligado por padrão em produção de alto volume |

O que logar: identificadores (order_id, user_id opaco, request_id), estado da decisão,
contagem de retry, nome/tipo do erro, duração da operação.

O que **nunca** logar: senha, token, chave de API, dado de cartão, CPF/e-mail/telefone
em claro, corpo completo de payload que contenha dado de cliente. Se o campo pode identificar
uma pessoa ou autenticar algo, ele não entra no log — nem mascarado "só os 4 últimos dígitos"
sem uma política explícita para isso.

## Correlation ID / trace ID propagado

Toda execução carrega um ID que atravessa todos os serviços que ela toca. Em uma arquitetura
de filas/eventos isso não é automático como numa chamada HTTP síncrona — o ID precisa ser
propagado manualmente no envelope da mensagem.

```json
{
  "detail-type": "OrderCreated",
  "detail": { "orderId": "ord_9182" },
  "traceContext": {
    "correlationId": "c4f2b8e1-...",
    "parentSpanId": "a1b2c3..."
  }
}
```

- Gere o `correlation_id` na borda (API Gateway, primeiro handler) se não vier de upstream.
- Propague-o em todo `MessageAttributes` do SQS e em `Detail`/metadata do EventBridge — não
  enterre dentro do payload de negócio, use um campo de contexto dedicado.
- Todo log de todo serviço na cadeia usa o mesmo `correlation_id` — é isso que permite
  reconstruir o caminho de uma mensagem entre Lambda → SQS → outra Lambda sem trace distribuído.

## Métricas RED (serviços) e USE (recursos)

**RED** — para todo serviço/endpoint/handler:

| Métrica | O que mede |
| ------- | ---------- |
| **Rate** | Requisições/execuções por segundo |
| **Errors** | Taxa de falha (erros / total) |
| **Duration** | Latência (p50/p90/p99, não só média) |

**USE** — para todo recurso finito (fila, conexão de DB, memória da Lambda, thread pool):

| Métrica | O que mede |
| ------- | ---------- |
| **Utilization** | Fração do recurso em uso |
| **Saturation** | Trabalho em espera além da capacidade (profundidade de fila SQS, connection pool exhausted) |
| **Errors** | Erros do próprio recurso (throttle, timeout de conexão) |

Exemplo prático numa fila SQS: `ApproximateNumberOfMessagesVisible` é USE (saturação);
taxa de mensagens processadas com sucesso vs. enviadas para DLQ é RED do consumer.

## Tracing distribuído

Um **trace** é a árvore de **spans** de uma execução ponta a ponta; cada span tem início, fim,
e um `parent_span_id` que o liga ao span que o chamou. Em Lambda/ECS instrumentado com
OpenTelemetry (ou X-Ray), abra um span por unidade de trabalho relevante — não por linha de
código:

```
trace c4f2b8e1
├─ span: api-gateway → lambda:create-order (120ms)
│  └─ span: dynamodb:put-item (18ms)
└─ span: sqs:consume → lambda:notify-customer (340ms)
   └─ span: ses:send-email (310ms)
```

Propague o contexto do trace do mesmo jeito que o correlation ID: no `MessageAttributes` do
SQS ou nos atributos do evento do EventBridge, para o span filho saber quem é seu pai mesmo
atravessando uma fila assíncrona.

## Cuidado com cardinalidade

Toda label/tag de métrica que carrega um valor de alta cardinalidade (user_id, order_id,
request_id, IP) explode o número de séries de tempo e pode derrubar o backend de métricas ou
custar uma fortuna. Esses valores vão no **log** e no **trace** (que suportam cardinalidade
alta), nunca como label de métrica. Labels de métrica ficam restritas a valores de baixo
cardinal: nome do serviço, ambiente, status code, nome da fila, tipo de evento.

## Checklist de aceite

- [ ] Todo log é JSON estruturado com nível correto; nenhum PII/secret/payload sensível em claro
- [ ] Correlation ID gerado na borda e propagado em todo envelope de mensagem (SQS/EventBridge), presente em todo log da cadeia
- [ ] Serviço expõe RED (rate/errors/duration com percentis); recursos finitos expõem USE
- [ ] Se houver chamada cross-service, existe span de trace com contexto propagado entre elas
- [ ] Nenhuma métrica usa label de alta cardinalidade (ID de entidade, IP, request_id)
