---
name: rc-sql
description: Guias de banco de dados relacional para otimização de queries e design de schema. Use ao escrever ou revisar queries (índices, planos de execução com EXPLAIN, problemas N+1, joins, paginação), modelar schema (normalização, tipos, constraints, chaves) ou avaliar performance de banco. Aplica análise read-only (só SELECT/EXPLAIN) por padrão. Carrega o guia certo por tarefa a partir de references/. Não use para migrations específicas de um repositório, administração de infraestrutura de banco, ou bancos não-relacionais (NoSQL, grafos, documentos).
user-invocable: true
model: sonnet
effort: medium
---

# Banco de dados (SQL) — guias por tarefa

Guias práticos de SQL. Leia o guia da tarefa em `references/` antes de agir — cada um traz
o checklist, os comandos de verificação e os critérios de aceite. Sempre leia o schema e os
índices reais (via `\d`, `information_schema` ou o DDL do repo) antes de recomendar qualquer
mudança; não presuma a estrutura. Os exemplos usam PostgreSQL, mas os princípios valem para
qualquer banco relacional.

## Roteamento

| Tarefa | Guia |
| ------ | ---- |
| Otimização de queries (EXPLAIN, índices, N+1, paginação) | `references/query-optimization.md` |
| Design de schema (normalização, tipos, constraints, chaves) | `references/schema-design.md` |
| Segurança e execução (transações, locks, migrations, produção) | `references/safety.md` |

## Princípios sempre válidos

- **Meça com EXPLAIN antes de otimizar — não adivinhe.** Nenhum ajuste de índice ou query se justifica sem o plano de execução real na frente.
- **O índice certo bate a query esperta.** Antes de reescrever uma query complexa, verifique se falta um índice.
- **Read-only por padrão (Rule 9).** Só execute `SELECT`/`EXPLAIN` para análise. `INSERT`/`UPDATE`/`DELETE`/DDL são recomendados no relatório — nunca executados sem autorização explícita do usuário.
- **O schema é o alicerce.** Corrija o schema antes de contornar sua falha em código (validação duplicada, joins gambiarra, etc.).

## Error Handling

- Sem acesso ao banco: analise o schema/DDL e as queries do repositório estaticamente e diga explicitamente que o `EXPLAIN` real e a verificação de índices ficaram pendentes.
- Nunca rode `INSERT`/`UPDATE`/`DELETE`/DDL nem qualquer comando além de `SELECT`/`EXPLAIN` sem autorização explícita — mesmo que a mudança pareça trivial ou reversível.
