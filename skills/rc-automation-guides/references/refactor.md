claude # Skill: /refactor

Workflow guiado para refatorar codigo existente seguindo os padroes definidos nos `agent-docs/`.

## O que fazer ao invocar este skill

1. Leia os documentos de contexto necessarios:
   - `agent-docs/ARCHITECTURE.md` — camadas, Clean Architecture, regras de importacao
   - `agent-docs/CODING-STANDARDS.md` — nomenclatura, TypeScript patterns, logging

2. Pergunte ao usuario (se nao informado):
   - Qual modulo/arquivo/funcao deseja refatorar?
   - Qual o objetivo do refactor? (ex: melhorar legibilidade, corrigir violacao de arquitetura, extrair usecase, renomear, simplificar)
   - Ha alguma restricao ou cuidado especial? (ex: nao alterar contrato publico, manter compatibilidade)

3. Execute o checklist na ordem abaixo.

---

## Passo 1 — Analise do codigo atual

Leia e compreenda o codigo alvo antes de qualquer mudanca:

- Identifique a responsabilidade atual do modulo/classe/funcao
- Mapeie dependencias (quem importa e quem e importado)
- Identifique testes existentes que cobrem o codigo
- Identifique violacoes dos padroes do projeto (se houver)

Apresente um resumo ao usuario com:
- O que o codigo faz hoje
- Problemas ou violacoes encontrados
- Plano de refactor proposto

**Aguarde confirmacao do usuario antes de prosseguir.**

---

## Passo 2 — Validacao de arquitetura

Antes de aplicar mudancas, verifique que o plano respeita:

- Regras de importacao entre camadas:
  - `domain/` nao importa nada do projeto (apenas `@escaletech/logger`)
  - `usecases/` importa apenas de `domain/`
  - `infrastructure/` importa apenas de `domain/`
  - Sem dependencias cruzadas entre `server/`, `lambda/` e `sqs/`
- Logica de negocio permanece em usecases (nunca em controllers/handlers/consumers)
- Contratos de repositorio continuam como `interface` (nunca `class`)

Se o refactor proposto violar alguma regra, alerte o usuario e ajuste o plano.

---

## Passo 3 — Aplicar mudancas

Execute as alteracoes seguindo os padroes:

- **Nomenclatura:** PascalCase para classes/interfaces, sufixos corretos (`Service`, `Controller`, `Repository`, `Interface`, `Factory`)
- **TypeScript:** `strict: true`, sem `any` (use `unknown`), named exports, sem path aliases
- **Logging:** `ILogger` injetado via construtor, sem `console.*`
- **Validacao:** Zod na fronteira (controllers/handlers), regras de negocio em usecases
- **JSDoc:** Manter ou adicionar JSDoc nas funcoes/classes modificadas
- **DI:** Se servicos foram adicionados, renomeados ou removidos, atualizar AMBOS:
  - `src/interfaces/server/dependencies.ts`
  - `src/interfaces/lambda/dependencies.ts`

---

## Passo 4 — Atualizar testes

- Se o refactor alterou a assinatura ou comportamento de um usecase, atualize o teste co-localizado (`<NomeService>.test.ts`)
- Se um service novo foi extraido, crie teste co-localizado usando `infra: 'mock-file'` via Factory
- Se um service foi removido, remova o teste correspondente
- Nunca instancie repositorio concreto diretamente no teste

---

## Passo 5 — Verificacao automatizada

Execute os dois comandos abaixo em ordem. Se qualquer um falhar, corrija antes de prosseguir para o proximo.

### 5.1 — Build

```bash
pnpm build
```

**Criterio:** sem erros de compilacao TypeScript. Se falhar, corrija e re-execute.

### 5.2 — Testes

```bash
pnpm test
```

**Criterio:** `0 failed` no output do Vitest. Se falhar, corrija e re-execute.

---

## Resultado

Reporte ao usuario:

```
REFACTOR - Resultado
=====================
Modulo:             [nome do modulo/arquivo refatorado]
Objetivo:           [descricao do refactor]
Arquivos alterados: [lista]
pnpm build:         [OK / FALHOU]
pnpm test:          [OK (X testes) / FALHOU]
Arquitetura:        [OK / ISSUE - descreva]
DI sincronizado:    [OK / N/A]
```

### Checklist de revisao rapida

- [ ] Regras de importacao entre camadas respeitadas?
- [ ] `dependencies.ts` sincronizado em server/ e lambda/ (se aplicavel)?
- [ ] Sem `console.*`, sem `any`, sem default exports?
- [ ] JSDoc atualizado nas funcoes/classes modificadas?
- [ ] Testes co-localizados atualizados/criados?
- [ ] Nenhum comportamento alterado sem intencao?
