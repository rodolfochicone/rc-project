# Otimização de queries

Objetivo: fazer a query certa rodar rápido, sem adivinhar. Ordem de prioridade:
**EXPLAIN → índice → reescrita da query → desnormalização**. Não reescreva uma query
antes de ver o plano de execução real.

## Fase 1 — Ler o plano de execução

`EXPLAIN` mostra o plano estimado; `EXPLAIN (ANALYZE, BUFFERS)` executa a query de verdade
e mostra custo, linhas e I/O reais. Use `ANALYZE` só em query de leitura — em `UPDATE`/`DELETE`
ele executa o comando (envolva em transação e faça `ROLLBACK`, ou peça autorização antes).

```sql
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM pedidos WHERE cliente_id = 123;
```

O que olhar:
- **`Seq Scan` em tabela grande** com filtro seletivo → falta índice na coluna do `WHERE`/`JOIN`.
- **`rows` estimado vs `rows` real (ANALYZE)** muito diferentes → estatísticas desatualizadas (`ANALYZE tabela;` como recomendação, não execute sem autorização).
- **`cost=inicial..total`**: o custo total é o que o planner usa para escolher o plano; compare planos, não valores absolutos.
- **`Index Scan` vs `Index Only Scan`**: o segundo não toca a tabela (usa só o índice) — mais rápido quando aplicável.

## Fase 2 — Índices

- **B-tree (padrão)**: bom para `=`, `<`, `>`, `BETWEEN`, `ORDER BY`. É o índice certo na maioria dos casos.
- **Composto**: a ordem das colunas importa. `(cliente_id, criado_em)` serve para `WHERE cliente_id = ? ORDER BY criado_em`, mas não ajuda um filtro só por `criado_em`. Coloque a coluna de igualdade antes da de range/ordenação.
- **Covering / `INCLUDE`**: adicione colunas só para permitir `Index Only Scan` sem inflar a chave do índice.
  ```sql
  CREATE INDEX idx_pedidos_cliente ON pedidos (cliente_id) INCLUDE (status, total);
  ```
- **Parcial**: indexa só o subconjunto relevante, mais barato que um índice completo.
  ```sql
  CREATE INDEX idx_pedidos_pendentes ON pedidos (criado_em) WHERE status = 'pendente';
  ```
- Índice demais custa em escrita (todo `INSERT`/`UPDATE` mantém todos os índices) — não crie índice para query que não existe.

## Fase 3 — Sargability (a query precisa deixar o índice ser usado)

Função na coluna do `WHERE` impede o uso do índice — a função deve estar do outro lado da comparação.

```sql
-- ruim: não usa índice em criado_em
WHERE DATE(criado_em) = '2026-01-01'

-- bom: sargable
WHERE criado_em >= '2026-01-01' AND criado_em < '2026-01-02'
```

O mesmo vale para `LOWER(coluna) = ...` (crie índice de expressão `ON tabela (LOWER(coluna))`
se a busca case-insensitive for frequente) e para `coluna::text = '123'` em coluna `int`.

## Fase 4 — N+1

Sintoma: uma query por linha de um resultado anterior (comum em ORM: buscar pedidos, depois
buscar o cliente de cada pedido em loop). Detecção: log de queries mostra N queries idênticas
com parâmetro variando, ou o tempo total escala linearmente com o tamanho do resultado.

Correção — trocar N queries por 1 `JOIN` ou 1 query batelada com `IN`:

```sql
-- N+1: 1 query de pedidos + N queries de cliente
SELECT * FROM clientes WHERE id = $1;  -- repetida N vezes

-- corrigido: 1 join
SELECT p.*, c.nome
FROM pedidos p
JOIN clientes c ON c.id = p.cliente_id
WHERE p.criado_em >= '2026-01-01';

-- ou 1 batch quando o join não cabe na modelagem do código
SELECT * FROM clientes WHERE id = ANY($1::int[]);
```

## Fase 5 — Paginação

`OFFSET` grande é caro: o banco ainda percorre e descarta todas as linhas anteriores.
Paginação por cursor (keyset) usa o último valor visto e escala O(1) por página.

```sql
-- OFFSET: degrada conforme a página cresce
SELECT * FROM pedidos ORDER BY id LIMIT 20 OFFSET 10000;

-- keyset: usa o índice, custo constante
SELECT * FROM pedidos WHERE id > $ultimo_id ORDER BY id LIMIT 20;
```

Use `OFFSET` só para paginação pequena e sem necessidade de "ir para a página N" em volume alto.

## Outras práticas

- **Evite `SELECT *`**: traz colunas que não serão usadas (custo de I/O e rede) e quebra
  `Index Only Scan`. Liste as colunas necessárias.
- **Desnormalização** é válida quando o `JOIN` de leitura é o hot path medido e a duplicação
  controlada (contador, campo calculado) é mais barata que recalcular — só depois de medir,
  nunca como primeira opção (ver `references/schema-design.md`).

## Checklist de aceite

- [ ] Query lenta tem `EXPLAIN (ANALYZE, BUFFERS)` real, não suposição
- [ ] `Seq Scan` em tabela grande com filtro seletivo foi resolvido com índice ou justificado
- [ ] Coluna do `WHERE`/`JOIN` não tem função aplicada (sargable) ou tem índice de expressão
- [ ] N+1 identificado e resolvido com join/batch, não com mais uma query por item
- [ ] Paginação de volume alto usa keyset, não `OFFSET` alto
- [ ] Nenhum comando além de `SELECT`/`EXPLAIN` foi executado sem autorização
