# Segurança e execução em banco de dados

Objetivo: nunca causar incidente ao analisar ou recomendar mudança de banco. Regra central
(Rule 9 do projeto): **acesso a banco é read-only por padrão.**

## Regra read-only (Rule 9)

- Execute **apenas** `SELECT` e `EXPLAIN` para investigar schema, dado e plano de execução.
- `INSERT`/`UPDATE`/`DELETE`/DDL (`CREATE`/`ALTER`/`DROP`) são **recomendados no relatório**,
  nunca executados — mesmo que pareçam triviais, reversíveis ou "só num ambiente de teste".
- Autorização explícita do usuário é obrigatória antes de qualquer comando fora de
  `SELECT`/`EXPLAIN`. Sem essa autorização, entregue o comando pronto e pare.
- `EXPLAIN ANALYZE` executa a query de verdade — use só em `SELECT`. Em `UPDATE`/`DELETE`,
  ou evite, ou envolva em transação com `ROLLBACK` explícito e só com autorização.

## Transações e isolamento

- Toda escrita multi-linha deve estar em transação (`BEGIN` ... `COMMIT`) para ser atômica —
  mas isso é uma recomendação de design, não uma execução sua.
- Nível de isolamento padrão do Postgres é `READ COMMITTED`: cada statement vê os commits
  anteriores, mas não protege contra leitura não repetível dentro da mesma transação.
  `SERIALIZABLE` protege mais, custa mais (retry em conflito).
- **Transação longa (long-running tx)** trava `VACUUM` de limpar linhas mortas (bloat) e
  pode segurar lock além do necessário. Nunca abra transação e espere input externo (rede,
  humano) antes de fechar.

## Locks e deadlock

- `UPDATE`/`DELETE` tomam lock de linha; DDL (`ALTER TABLE`) toma lock de tabela — pode
  bloquear leituras/escritas concorrentes se não usar a forma `CONCURRENTLY`/sem rewrite.
- **Deadlock**: duas transações travam linhas em ordem oposta (A trava linha 1 depois 2;
  B trava linha 2 depois 1) e cada uma espera a outra. Evite: sempre trave recursos na
  mesma ordem (ex.: por `id` crescente) entre transações concorrentes.
- Query de diagnóstico (read-only):
  ```sql
  SELECT pid, state, wait_event_type, query
  FROM pg_stat_activity
  WHERE state != 'idle';
  ```

## Migrations seguras (recomendação — não execução)

- **Backfill em lotes**: nunca `UPDATE tabela_grande SET ...` sem `WHERE` limitando o lote
  (trava a tabela toda e cresce o WAL). Recomende faixas (`WHERE id BETWEEN $a AND $b`) em
  loop com commit entre lotes.
- **`CREATE INDEX CONCURRENTLY`**: evita lock exclusivo de escrita durante a criação do
  índice — recomende sempre essa forma em tabela com tráfego de produção, nunca a forma sem
  `CONCURRENTLY`.
- **Adicionar coluna `NOT NULL`**: em Postgres recente, `ALTER TABLE ... ADD COLUMN x int
  NOT NULL DEFAULT 0` não trava a tabela (default é aplicado sem rewrite) — mas `NOT NULL`
  sem `DEFAULT` numa tabela com dado existente falha. Recomende adicionar com `DEFAULT`, ou
  em duas etapas: adicionar nullable, backfillar em lote, só então aplicar `NOT NULL`.
- **Expand/contract**: para renomear ou trocar tipo de coluna em uso, recomende expandir
  primeiro (nova coluna, escrever nas duas, backfill), migrar os leitores, e só então
  contrair (remover a coluna antiga) — nunca trocar em um passo só se houver leitor em
  produção dependendo do nome/tipo antigo.

## Evitar operação perigosa em produção

- Sem `WHERE`, `UPDATE`/`DELETE` afeta a tabela inteira — sempre confira o `WHERE` com um
  `SELECT` equivalente antes de recomendar a escrita.
- Full table scan em tabela grande (`Seq Scan` sem filtro seletivo) em horário de pico
  compete por I/O com o tráfego real — recomende rodar em horário de baixo uso ou com
  `LIMIT`/lote.
- Nunca recomende `DROP`/`TRUNCATE` sem confirmar que existe backup/restore point.

## Checklist de aceite

- [ ] Nenhum comando fora de `SELECT`/`EXPLAIN` foi executado sem autorização explícita
- [ ] Toda recomendação de escrita/DDL foi entregue como comando pronto, não aplicada
- [ ] Migration recomendada usa `CONCURRENTLY`/lote/expand-contract quando a tabela tem tráfego
- [ ] `UPDATE`/`DELETE` recomendado tem `WHERE` verificado por `SELECT` equivalente antes
- [ ] Risco de lock/deadlock/transação longa foi avaliado antes de recomendar a escrita
