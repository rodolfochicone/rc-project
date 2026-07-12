# Skill: /db-migration

Guia para criar migracoes e alteracoes de schema no banco de dados, seguindo `agent-docs/DATA-MODELING.md` e `agent-docs/DATABASE.md`.

## O que fazer ao invocar este skill

1. Leia os documentos de contexto:
   - `agent-docs/DATA-MODELING.md` — schema, soft delete, migracoes, tipos recomendados
   - `agent-docs/DATABASE.md` — nomenclatura de tabelas, colunas, ENUMs, chaves

2. Pergunte ao usuario:
   - Qual a operacao? (criar tabela, adicionar coluna, criar indice, alterar tipo, etc.)
   - Qual a tabela/entidade envolvida?
   - Quais campos e tipos?

3. Execute o checklist abaixo.

---

## Checklist de criacao de tabela

### Passo 1 — Nomenclatura

Valide a nomenclatura antes de escrever SQL:
- Tabela: `snake_case`, plural (`orders`, `deal_stages`)
- Colunas: `snake_case` (`created_at`, `partner_id`)
- PK: sempre `id`
- FK: `<entidade_singular>_id` (`order_id`, `customer_id`)
- Indices: `idx_<tabela>_<coluna(s)>`
- Constraints unique: `uq_<tabela>_<coluna(s)>`
- Booleanos: prefixo `is_` ou `has_`
- Nunca prefixos como `tbl_`, `tb_`, `t_`

### Passo 2 — Campos obrigatorios

Toda tabela DEVE ter:

```sql
id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
created_at TIMESTAMPTZ NOT NULL    DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL    DEFAULT NOW(),
deleted_at TIMESTAMPTZ NULL
```

- `id`: UUID, nunca sequencial
- Datas: `TIMESTAMPTZ`, nunca `TIMESTAMP`
- `deleted_at`: presente em toda tabela (soft delete e o padrao)

### Passo 3 — Tipos corretos

| Dado | Tipo PostgreSQL |
|---|---|
| Identificadores | `UUID` |
| Textos curtos | `VARCHAR(N)` com limite explicito |
| Textos longos | `TEXT` |
| Valores monetarios | `NUMERIC(15, 2)` — nunca `FLOAT` |
| Datas com hora | `TIMESTAMPTZ` |
| Apenas data | `DATE` |
| Booleanos | `BOOLEAN` |
| Enumeracoes | `VARCHAR` + `CHECK` ou tipo `ENUM` — valores em `UPPER_SNAKE_CASE` |
| JSON estruturado | `JSONB` |

### Passo 4 — Indices

Crie indice para:
- Toda FK (`<entidade>_id`)
- Colunas usadas em `WHERE` frequentes
- Colunas usadas em `ORDER BY` de listagens paginadas
- Indice parcial para soft delete: `WHERE deleted_at IS NULL`

```sql
CREATE INDEX idx_orders_customer_id ON orders (customer_id);
CREATE INDEX idx_orders_created_at  ON orders (created_at DESC);
CREATE INDEX idx_orders_deleted_at  ON orders (deleted_at) WHERE deleted_at IS NULL;
```

### Passo 5 — Tabelas de juncao (many-to-many)

Ambos os nomes no plural, unidos por `_`, em ordem alfabetica:
```
contacts_deals    -- contacts <-> deals (c < d)
contacts_persons  -- contacts <-> persons (c < p)
```

### Passo 6 — Arquivo de migracao

```
migrations/
├── 1700000000000_create-orders.sql
├── 1700000000001_add-status-to-orders.sql
```

- Nome em `kebab-case` com timestamp como prefixo
- Descritivo do que a migracao faz — nunca nomes genericos
- Uma responsabilidade por arquivo
- Toda migracao deve ter rollback (`down`) documentado

### Passo 7 — Soft delete nas queries

Toda query de leitura deve filtrar `WHERE deleted_at IS NULL`:

```sql
-- Soft delete
UPDATE orders SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- Leitura (sempre filtrar ativos)
SELECT * FROM orders WHERE customer_id = $1 AND deleted_at IS NULL;
```

---

## Regras importantes

- Migracoes sao **imutaveis** apos executadas em producao
- Nunca `DROP COLUMN` ou `DROP TABLE` diretamente em producao
- `ADD COLUMN ... DEFAULT NULL` para evitar lock em tabelas grandes
- PostgreSQL e a fonte de verdade. OpenSearch e projecao para leitura
- IDs nunca sequenciais — sempre UUID com `gen_random_uuid()`

---

## Entidade no domain

Apos criar a migracao, crie a entidade correspondente em `src/domain/<contexto>/`:
- Interface TypeScript representando a entidade
- `<Nome>RepositoryInterface` com os metodos necessarios
- `<Nome>RepositoryFactory` para selecao de implementacao

Siga o padrao em `agent-docs/ARCHITECTURE.md` secao 5.
