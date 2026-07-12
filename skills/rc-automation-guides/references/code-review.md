# Skill: /code-review

Checklist completo de revisao de codigo baseado nos padroes do projeto.

## O que fazer ao invocar este skill

1. Analise o diff atual (arquivos alterados na branch).
2. Leia os arquivos modificados por completo para entender o contexto.
3. Execute cada secao do checklist abaixo.
4. Reporte o resultado com issues encontradas e sugestoes.

---

## 1. Arquitetura e camadas

- [ ] Regras de importacao respeitadas? (`domain/` nao importa do projeto, `usecases/` so de `domain/`, etc.)
- [ ] Sem dependencias cruzadas entre `server/`, `lambda/` e `sqs/`?
- [ ] Logica de negocio esta em usecases — nunca em controllers, handlers ou consumers?
- [ ] Usecase recebe dependencias via constructor injection?
- [ ] Nenhuma instanciacao concreta dentro de usecases?

## 2. Nomenclatura

- [ ] Classes em PascalCase com sufixo correto (`Service`, `Controller`, `Repository`, `Factory`, `Interface`)?
- [ ] Arquivos em PascalCase?
- [ ] Variaveis de ambiente em `SCREAMING_SNAKE_CASE`?
- [ ] Tabelas/colunas em `snake_case`, plural?

## 3. TypeScript

- [ ] Sem uso de `any` explicito? (usar `unknown` quando necessario)
- [ ] Contratos de repositorio definidos como `interface` (nunca `class` com throw)?
- [ ] Named exports (sem default exports, exceto routers Express)?
- [ ] Sem path aliases — caminhos relativos explicitos?

## 4. Logging

- [ ] Zero `console.log`, `console.error`, `console.warn`, `console.debug`?
- [ ] Logger injetado via construtor (`ILogger`)?
- [ ] Correlation ID definido no inicio do request?
- [ ] Nenhum dado sensivel logado?
- [ ] Niveis de log apropriados? (`info` para operacoes, `error` com objeto de erro, `debug` para diagnostico)

## 5. Validacao

- [ ] Inputs validados com Zod na fronteira (controller/handler)?
- [ ] Erro de validacao retorna `400` com contrato padrao?
- [ ] IDs como `uuid()`, strings com limites, enums com `z.enum()`?

## 6. Testes

- [ ] Todo novo usecase tem teste co-localizado (`NomeService.test.ts`)?
- [ ] Testes usam `infra: 'mock-file'` via Factory?
- [ ] Testes passam? (`pnpm test`)

## 7. DI e dependencies

- [ ] Novos servicos adicionados em `src/interfaces/server/dependencies.ts`?
- [ ] Novos servicos adicionados em `src/interfaces/lambda/dependencies.ts`?
- [ ] Ambos sincronizados?

## 8. API (se endpoint novo/alterado)

- [ ] Rota no plural, kebab-case, sem trailing slash?
- [ ] Codigos HTTP corretos (nunca `200` com body de erro)?
- [ ] Resposta segue contrato padrao (recurso unico sem envelope, colecao com `data` + `pagination`)?
- [ ] OpenAPI atualizado?

## 9. Banco de dados (se schema alterado)

- [ ] Campos obrigatorios presentes (`id UUID`, `created_at`, `updated_at`, `deleted_at`)?
- [ ] Soft delete como padrao? Queries filtram `deleted_at IS NULL`?
- [ ] Indices criados para FKs e colunas de WHERE/ORDER BY?
- [ ] Valores monetarios com `NUMERIC(15, 2)` (nunca `FLOAT`)?

## 10. Seguranca

- [ ] Queries parametrizadas (sem concatenacao de SQL)?
- [ ] Nenhum secret em codigo-fonte?
- [ ] Stack traces nao expostos em respostas?
- [ ] Headers de contexto lidos corretamente (nunca do cliente externo)?

---

## Resultado

Reporte ao usuario por categoria:

```
CODE REVIEW - Resultado
========================
Arquitetura:      [OK / N issues]
Nomenclatura:     [OK / N issues]
TypeScript:       [OK / N issues]
Logging:          [OK / N issues]
Validacao:        [OK / N issues]
Testes:           [OK / N issues]
DI:               [OK / N issues]
API:              [OK / N/A]
Banco de dados:   [OK / N/A]
Seguranca:        [OK / N issues]
```

Para cada issue, inclua:
- Arquivo e linha
- Descricao do problema
- Sugestao de correcao
- Severidade (critico / importante / sugestao)
