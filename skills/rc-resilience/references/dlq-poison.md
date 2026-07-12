# Dead-letter queue, poison message e redrive

Objetivo: garantir que uma mensagem que nunca vai passar não trava a fila e não se perde.
Ordem de prioridade: **detectar poison message → configurar DLQ → alarmar → redrive seguro**.

## Poison message

Mensagem que falha sempre, independente de quantas vezes é reentregue — payload malformado,
bug de código, dependência que rejeita permanentemente aquele dado. Sem uma DLQ, ela fica
sendo reentregue em loop (o visibility timeout expira, o SQS reentrega, falha de novo),
consumindo capacidade do consumidor e atrasando as mensagens boas atrás dela.

Detecção no SQS: `maxReceiveCount` na redrive policy. Depois de N tentativas
(`ApproximateReceiveCount` >= `maxReceiveCount`), o SQS move a mensagem automaticamente para
a DLQ associada.

```json
{
  "deadLetterTargetArn": "arn:aws:sqs:...:pedidos-dlq",
  "maxReceiveCount": 5
}
```

## Quando configurar DLQ

Toda fila (SQS) e toda regra de EventBridge com destino assíncrono (Lambda, outra fila)
**precisa** de DLQ — não é opcional. Sem DLQ, mensagem que falha permanentemente é
descartada silenciosamente (Lambda + EventBridge) ou fica em loop de redrive (SQS). As duas
opções são perda de dado; a diferença é só quando você descobre.

O que vai para a DLQ:

- A mensagem original completa (não um resumo) — precisa ser reprocessável depois do fix.
- Preserve/anexe o contexto do erro (`ReceiveCount`, timestamp da última falha, mensagem de
  exceção) se o destino suportar metadata, para não precisar cavar logs para saber por que
  caiu.

## Alarme sobre profundidade da DLQ

DLQ sem alarme é lixeira que ninguém olha — a mensagem "foi tratada" no sentido de não
travar a fila principal, mas o problema de negócio continua sem ser notado. Configure alarme
em `ApproximateNumberOfMessagesVisible` da DLQ (CloudWatch), com limiar baixo (>= 1 já é
sinal, dependendo do volume). Ver `rc-observability` para o desenho do alerta em si (SLI,
threshold, para onde ele notifica).

## Redrive/reprocesso seguro

Redrive (mover mensagens da DLQ de volta para a fila principal) só depois de:

1. **Corrigir a causa raiz.** Redrive sem fix reentrega a mesma poison message, que volta
   para a DLQ depois de `maxReceiveCount` tentativas — ciclo sem ganho.
2. **Confirmar idempotência no consumidor.** A mensagem redrivada pode já ter sido
   parcialmente processada antes de falhar. Idempotência (ver `idempotency.md`) é o que torna
   o replay seguro em vez de duplicar efeito.
3. Preferir a `StartMessageMoveTask` da AWS (redrive nativo do SQS) a script manual — evita
   erro de reimplementar paginação/retry do próprio redrive.

## Ordenação e o trade-off do FIFO

DLQ e retry quebram ordem por natureza: uma mensagem que falha e é redrivada minutos/dias
depois chega fora de ordem em relação às que vieram atrás dela. Se a ordem é requisito de
negócio (ex.: eventos de um mesmo pedido precisam ser aplicados em sequência):

- Use SQS FIFO com `MessageGroupId` por entidade (garante ordem dentro do grupo, não entre
  grupos — isso é o que permite paralelismo).
- Aceite que, quando algo cai na DLQ, mensagens **depois** dela no mesmo grupo ficam
  bloqueadas até o redrive (FIFO não pula a que falhou). Trate isso como trade-off explícito,
  não como bug.

## Visibility timeout e backpressure

`VisibilityTimeout` precisa ser maior que o tempo máximo de processamento do consumidor —
senão o SQS reentrega uma mensagem que ainda está sendo processada (falsa duplicata). Regra
prática AWS: visibility timeout >= timeout da função Lambda x 6 (para cobrir o batch inteiro
mais retries do SDK).

Backpressure: se o consumidor está mais lento que a chegada de mensagens, a fila cresce — isso
é esperado e correto (a fila é o buffer). O problema é quando a fila cresce **e** mensagens
começam a expirar/cair em DLQ por lentidão, não por erro de lógica; nesse caso o ataque é
escalar concorrência do consumidor, não aumentar `maxReceiveCount`.

## Partial batch failure (Lambda + SQS)

Lambda processando um batch de SQS: por padrão, se a função lança exceção, o **lote inteiro**
é considerado falho e reentregue — inclusive as mensagens do lote que já tinham sido
processadas com sucesso, gerando reprocessamento (e duplicata) do que já deu certo.

Ative `ReportBatchItemFailures` e reporte só os `itemIdentifier` que falharam:

```python
def handler(event, context):
    failures = []
    for record in event["Records"]:
        try:
            processar(record)
        except Exception:
            failures.append({"itemIdentifier": record["messageId"]})
    return {"batchItemFailures": failures}
```

Sem isso, todo consumidor de batch reprocessa itens bons a cada falha de um item ruim —
mais um motivo pelo qual todo processamento aqui precisa ser idempotente.

## Checklist de aceite

- [ ] Toda fila SQS e toda regra EventBridge assíncrona tem DLQ configurada com `maxReceiveCount` definido (não o default implícito)
- [ ] Alarme de profundidade da DLQ existe e notifica alguém (ver `rc-observability`)
- [ ] Redrive documentado como "corrigir causa → confirmar idempotência → redrive", nunca redrive automático sem fix
- [ ] Se ordem é requisito, FIFO + `MessageGroupId` estão configurados e o trade-off de bloqueio por grupo está aceito explicitamente
- [ ] `VisibilityTimeout` >= tempo máximo de processamento (Lambda: timeout x 6 como regra prática)
- [ ] Consumidor de batch Lambda+SQS usa `ReportBatchItemFailures` — não reprocessa o lote inteiro por um item ruim
