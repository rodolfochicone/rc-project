# Design de schema

Objetivo: um schema que impede dado inválido de existir e que a query certa consegue ler
sem gambiarra. Ordem de prioridade: **normalizar → tipar corretamente → constraint →
índice pensado no design**. Desnormalize só depois de medir (ver `references/query-optimization.md`).

## Fase 1 — Normalização

- **1NF**: cada coluna é atômica (sem lista/CSV numa coluna; use tabela filha ou array/JSON tipado).
- **2NF**: nenhuma coluna não-chave depende só de parte de uma chave composta.
- **3NF**: nenhuma coluna não-chave depende de outra coluna não-chave (sem dado derivado guardado que não seja recalculado — exceto desnormalização deliberada).

Erro comum: guardar `nome_cliente` na tabela de pedidos em vez de só `cliente_id` — duplica
dado que muda (renomear cliente vira update em N linhas) e pode divergir do original.

**Desnormalizar deliberadamente** é válido quando: o dado é histórico e não deve mudar
(preço do item no momento da venda, não o preço atual do produto), ou o `JOIN` de leitura é
hot path medido e a duplicação é mais barata que recalcular. Documente a decisão perto da
coluna (constraint/comentário), não deixe implícita.

## Fase 2 — Tipos corretos

| Em vez de | Use | Por quê |
| --------- | --- | ------- |
| `text`/`varchar` para tudo | tipo específico (`int`, `boolean`, `date`, enum/`CHECK`) | perde validação e espaço; `text` livre não impede lixo |
| `timestamp` sem timezone | `timestamptz` | `timestamp` sem tz é ambíguo entre servidores/fusos |
| `float`/`double` para dinheiro | `numeric(p,s)` | ponto flutuante binário arredonda valor monetário |
| `varchar(255)` por hábito | `text` (Postgres não penaliza) ou o tamanho real do domínio | `255` é herança do MySQL antigo, não um limite real |

```sql
preco numeric(10,2) NOT NULL,
criado_em timestamptz NOT NULL DEFAULT now(),
status text NOT NULL CHECK (status IN ('pendente', 'pago', 'cancelado'))
```

## Fase 3 — Constraints como documentação executável

Toda regra de negócio que o schema pode garantir deve estar no schema — não só na aplicação
(código novo, migration manual ou acesso direto ao banco também precisam respeitar a regra).

| Constraint | Garante |
| ---------- | ------- |
| `PRIMARY KEY` | identidade única e não nula da linha |
| `FOREIGN KEY` | a referência existe na tabela pai (evita órfão) |
| `UNIQUE` | não duplicar um valor que deve ser único (email, slug) |
| `CHECK` | invariante de valor (`preco >= 0`, `status IN (...)`) |
| `NOT NULL` | campo obrigatório — o mais barato e mais esquecido |

```sql
CREATE TABLE pedidos (
  id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  cliente_id bigint NOT NULL REFERENCES clientes(id),
  total numeric(10,2) NOT NULL CHECK (total >= 0),
  status text NOT NULL DEFAULT 'pendente',
  criado_em timestamptz NOT NULL DEFAULT now()
);
```
(DDL de exemplo — recomendação, não execute sem autorização; ver `references/safety.md`.)

## Fase 4 — Chave natural vs surrogate

- **Surrogate** (`id` serial/identity/UUID): padrão seguro — estável mesmo se o dado "natural"
  mudar (email pode ser editado; CPF pode ter erro de digitação corrigido).
- **Natural** (ex.: código ISO de país, sigla de UF): ok quando o valor é imutável por definição
  e curto — evita um `JOIN` extra só para resolver o nome.
- Nunca use dado editável pelo usuário (email, username) como chave referenciada por outras
  tabelas — cascateia problema quando precisar corrigir.

## Fase 5 — Índices pensados no design

Toda `FOREIGN KEY` costuma precisar de índice na coluna filha (Postgres não cria automático,
diferente da PK) — sem ele, `DELETE`/`UPDATE` na tabela pai faz `Seq Scan` na filha para checar
a referência. `UNIQUE` já cria índice.

```sql
CREATE INDEX idx_pedidos_cliente_id ON pedidos (cliente_id);
```

## Fase 6 — Convenção de nomenclatura

Escolha uma convenção e mantenha em toda a base — não misture: `snake_case` para
tabela/coluna, singular ou plural consistente para nome de tabela, `tabela_id` para FK
(`cliente_id`, não `id_cliente` num schema que já usa o padrão contrário), timestamp com
sufixo (`criado_em`, `atualizado_em`).

## Checklist de aceite

- [ ] Nenhuma coluna guarda lista/CSV que deveria ser tabela filha
- [ ] Dinheiro em `numeric`, timestamp em `timestamptz`, enum/`CHECK` em vez de `text` livre para valor fixo
- [ ] Toda `FK` tem índice na coluna filha; toda regra de negócio verificável tem `CHECK`/`NOT NULL`/`UNIQUE`
- [ ] Chave referenciada por outra tabela é surrogate ou natural comprovadamente imutável
- [ ] Desnormalização (se houver) está documentada e justificada por medição, não por atalho
- [ ] Nomenclatura segue a convenção já usada no schema existente
