# Idempotência e deduplicação

Objetivo: garantir que processar a mesma mensagem duas vezes produz o mesmo resultado que
processar uma vez. Ordem de prioridade: **entenda por que duplicata acontece → classifique a
operação → escolha a chave → implemente o store**.

## Por que at-least-once torna idempotência obrigatória

EventBridge, SQS e a maioria dos brokers entregam **at-least-once**, não exactly-once. Causas
comuns de duplicata:

- O consumidor processa a mensagem, mas cai antes de fazer `ack`/`DeleteMessage` — a fila
  reentrega após o visibility timeout.
- O produtor retenta uma publicação que na verdade teve sucesso (timeout na resposta, não na
  operação).
- Redrive manual de uma DLQ reprocessa uma mensagem que já tinha sido parcialmente aplicada.

Não existe forma de eliminar duplicata na entrega sem custo de latência/disponibilidade
(two-phase commit distribuído). A saída prática é aceitar a duplicata e neutralizá-la no
consumidor.

## "Exactly-once" é praticamente mito

O que os brokers chamam de "exactly-once" (ex.: SQS FIFO) é **at-least-once na entrega +
deduplicação na entrada da fila**, dentro de uma janela de tempo (5 minutos no SQS FIFO). Fora
dessa janela, ou entre sistemas diferentes, a garantia não existe. Trate exactly-once como
**at-least-once + idempotência no consumidor** — essa combinação é o único "exactly-once"
que se sustenta ponta a ponta.

## Classifique a operação antes de decidir a estratégia

| Tipo | Exemplo | Idempotente por natureza? |
| ---- | ------- | -------------------------- |
| `SET` (estado final explícito) | `UPDATE pedido SET status = 'pago'` | Sim — reaplicar não muda o resultado |
| `PUT` (substituição completa) | `S3 PutObject`, upsert por chave | Sim |
| `INCREMENT`/`APPEND` | `saldo += valor`, `INSERT` sem chave única | Não — duplicata soma/duplica |
| Efeito colateral externo | Enviar e-mail, cobrar cartão, chamar webhook | Não — o mundo externo não desfaz |

Operações do tipo `SET`/`PUT` já são seguras. Para `INCREMENT` e efeitos colaterais externos,
é obrigatório uma chave de idempotência.

## Chave de idempotência: de onde tirar

Na ordem de preferência:

1. **Id de negócio determinístico** — `pedido_id + evento_tipo` (ex.: `pay-order-4521`).
   Preferível porque sobrevive a reprocessamento de qualquer origem, não só reentrega de fila.
2. **Dedup id do evento** — `detail.id` do EventBridge ou um id gerado pelo produtor e
   propagado no payload.
3. **`messageId` da fila** — último recurso. Cobre reentrega da mesma mensagem física, mas
   não cobre duplicata lógica (dois eventos diferentes para a mesma intenção de negócio).

```python
# ponytail: chave de negócio, não messageId — sobrevive a replay de DLQ e a reenvio manual
idempotency_key = f"charge:{order_id}:{event_type}"
```

## Store de idempotência

Uma tabela (DynamoDB, Redis, Postgres) com a chave de idempotência como chave primária e TTL.

```
PK: idempotency_key
attributes: status (PROCESSING | DONE), result (opcional), expires_at (TTL)
```

Padrão de uso — "check-then-act" com escrita condicional (evita corrida entre duas execuções
concorrentes da mesma chave):

```python
try:
    table.put_item(
        Item={"pk": idempotency_key, "status": "PROCESSING", "expires_at": now + ttl},
        ConditionExpression="attribute_not_exists(pk)",
    )
except ConditionalCheckFailedException:
    return  # já processado ou em processamento — não reprocessa

processar_evento(evento)
table.update_item(Key={"pk": idempotency_key}, UpdateExpression="SET #s = :done",
                   ExpressionAttributeNames={"#s": "status"}, ExpressionAttributeValues={":done": "DONE"})
```

TTL deve cobrir a janela realista de duplicata (visibility timeout + tempo de redrive manual),
não só os minutos do broker — redrive de DLQ pode acontecer dias depois.

## SQS FIFO: message group e dedup id

SQS FIFO oferece deduplicação nativa por 5 minutos via `MessageDeduplicationId` (explícito ou
derivado do `ContentBasedDeduplication`). Use quando:

- A ordem de processamento importa dentro de uma partição (`MessageGroupId` = ex. `order_id`).
- A duplicata a evitar é de curto prazo (reenvio imediato do produtor).

Não use FIFO como substituto da idempotência no consumidor — ele não cobre replay de DLQ,
reprocessamento manual, nem duplicata entre sistemas diferentes.

## Checklist de aceite

- [ ] Toda operação `INCREMENT`/`APPEND` ou com efeito colateral externo tem chave de idempotência
- [ ] A chave é um id de negócio ou dedup id do evento — não depende só do `messageId` da fila
- [ ] Existe store de idempotência com escrita condicional (sem corrida entre execuções concorrentes)
- [ ] TTL do store cobre replay de DLQ e reprocessamento manual, não só o visibility timeout
- [ ] Nenhuma dependência de "exactly-once" do broker sem idempotência no consumidor por trás
