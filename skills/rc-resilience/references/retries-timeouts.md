# Timeout, retry e circuit breaker

Objetivo: falhar rápido, retentar só o que vale a pena retentar, e parar de bater numa
dependência que já está caída. Ordem de prioridade: **timeout em cada hop → classificar o
erro → backoff+jitter → circuit breaker/bulkhead**. Pular a classificação e retentar tudo é o
erro mais comum.

## Timeout em cada hop

Nenhuma chamada remota (HTTP, SDK da AWS, chamada a outra Lambda, query externa) sobe sem
timeout explícito. O timeout implícito da lib (ou nenhum) some com o comportamento sob
degradação: uma dependência lenta consome todas as conexões/threads do chamador.

```python
# ponytail: timeout explícito, não o default do cliente (que pode ser None)
response = http_client.post(url, json=payload, timeout=(2, 5))  # (connect, read)
```

Regra prática: o timeout do hop N deve ser menor que o budget de tempo restante do hop N-1
(ex.: Lambda com 10s de timeout não pode chamar algo com timeout de 15s).

## Classifique o erro: retryable vs não-retryable

| Retryable | Não-retryable |
| --------- | -------------- |
| 5xx (erro do servidor) | 4xx de validação (400, 422) |
| 429 / throttling | 401/403 (auth/permissão) |
| Timeout de conexão/leitura | 404 (recurso não existe) |
| Erro de rede transiente (conexão recusada, DNS) | Erro de parsing/schema do payload |

Retentar um erro não-retryable não conserta nada — só adia a falha e queima o orçamento de
retry. Decida essa tabela explicitamente por integração; não trate toda excepção como
retryable por padrão.

## Backoff exponencial + jitter

Retry sem espaçamento crescente gera **retry storm**: N clientes retentando no mesmo instante
batem juntos na dependência já degradada, prolongando a queda (efeito thundering herd).

```python
import random

def backoff_seconds(attempt, base=0.5, cap=30):
    exp = min(cap, base * (2 ** attempt))
    return random.uniform(0, exp)  # full jitter — não soma jitter ao exponencial, sorteia dentro dele
```

- **Backoff exponencial**: cada tentativa espera mais que a anterior (`base * 2^attempt`).
- **Jitter**: sorteia dentro do intervalo em vez de usar o valor exato — descorrelaciona os
  clientes que falharam juntos no mesmo evento.
- **Orçamento de retry**: número máximo de tentativas (3–5 é comum) ou um deadline total
  (ex.: nunca passar de 30s somando todas as tentativas). Sem limite, uma dependência caída
  vira uma fila que nunca esvazia.

## Idempotência é pré-requisito de retry seguro

Retentar uma chamada que já teve efeito colateral (cobrar, enviar e-mail, criar registro sem
chave única) duplica o efeito, não só a chamada. Antes de configurar retry numa operação,
confirme que ela é idempotente (ver `idempotency.md`) — senão o retry resolve o timeout e
cria um bug de duplicata.

## Circuit breaker

Evita continuar retentando (e aumentando latência do chamador) contra uma dependência que já
está claramente fora do ar.

| Estado | Comportamento |
| ------ | -------------- |
| `closed` | Chamadas passam normalmente; conta falhas |
| `open` | Falhas acima do limiar → rejeita chamadas imediatamente (sem nem tentar) por um período |
| `half-open` | Após o período, deixa passar uma amostra; sucesso volta para `closed`, falha volta para `open` |

Use quando a dependência tem custo de falha alto (latência de timeout se acumulando) e é
compartilhada por muitos chamadores — um breaker por dependência, não um global.

## Bulkhead

Isola o pool de recursos (conexões, threads, concorrência) por dependência, para que uma
dependência lenta não esgote a capacidade usada para chamar as outras. Em Lambda/ECS isso é
tipicamente limitar concorrência por integração (semáforo, pool de conexão dedicado) — sem
isolamento, uma dependência degradada derruba chamadas para dependências saudáveis só por
disputa de recurso.

## Checklist de aceite

- [ ] Toda chamada remota tem timeout explícito, menor que o budget do chamador
- [ ] Erros retryable e não-retryable estão classificados por integração, não por default genérico
- [ ] Retry usa backoff exponencial com jitter e tem limite de tentativas ou deadline total
- [ ] A operação sob retry é idempotente (ou o retry está desabilitado até que seja)
- [ ] Dependência crítica e compartilhada tem circuit breaker e/ou bulkhead, não só retry
