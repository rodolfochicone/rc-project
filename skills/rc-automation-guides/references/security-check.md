# Skill: /security-check

Revisao de seguranca por PR, baseada em `agent-docs/SECURITY.md` e OWASP Top 10.

## O que fazer ao invocar este skill

1. Leia `agent-docs/SECURITY.md` para contexto completo.
2. Analise o diff atual (arquivos alterados na branch).
3. Execute o checklist abaixo item por item.
4. Reporte o resultado ao usuario.

---

## Checklist de seguranca

### 1. Validacao de inputs

- [ ] Todos os inputs de controllers/handlers estao validados com Zod?
- [ ] IDs validados com `uuid()`?
- [ ] Strings com `min(1)` e `max(N)` para evitar payloads gigantes?
- [ ] Enums validados com `z.enum([...])` (nunca string aberta)?
- [ ] Datas validadas com `z.string().datetime()` (ISO 8601)?
- [ ] Arrays com `.min()` e `.max()` quando aplicavel?
- [ ] Erros de validacao retornam `400` com contrato padrao?

### 2. Injection (A03)

- [ ] Queries SQL sao parametrizadas? (nunca concatenar string com input do usuario)
- [ ] Nenhum `eval()`, `new Function()`, ou construcao dinamica de codigo?
- [ ] URLs nao sao construidas a partir de input do usuario sem validacao?

### 3. Protecao de dados sensiveis

- [ ] Nenhuma senha, token ou dado de cartao e logado?
- [ ] Nenhum dado pessoal alem do necessario para debug (LGPD)?
- [ ] Payloads completos de request nao sao logados em `info` (apenas em `debug`)?
- [ ] Stack traces nao sao expostos em respostas de producao?
- [ ] Campos internos de infra (IDs de banco, nomes de tabela) nao expostos na resposta?

### 4. Controle de acesso (A01)

- [ ] Dados filtrados por `partner` quando envolvem dados multi-tenant?
- [ ] Headers de contexto (`partner`, `product`, `user`) nunca aceitos diretamente do cliente externo?
- [ ] A aplicacao confia nos headers injetados pelo API Gateway Authorizer?

### 5. Secrets e configuracao (A05)

- [ ] Nenhum secret hardcoded em codigo-fonte ou comentarios?
- [ ] `.env` no `.gitignore`?
- [ ] Variaveis de ambiente nomeadas em `SCREAMING_SNAKE_CASE`?
- [ ] IAM roles com least privilege (nenhuma Lambda com `*`)?

### 6. Integridade de dados (A04/A08)

- [ ] Eventos publicados **apos** persistencia (nunca antes)?
- [ ] Consumers sao idempotentes?
- [ ] Body de mensagens SQS validado antes de processar?

### 7. Dependencias (A06)

- [ ] Nenhuma dependencia nova com vulnerabilidade conhecida?
- [ ] `@types/*` em `devDependencies` (nunca em `dependencies`)?

---

## Resultado

Reporte ao usuario:

```
SECURITY CHECK - Resultado
===========================
Validacao de inputs:    [OK / ISSUE - descreva]
Injection:              [OK / ISSUE - descreva]
Dados sensiveis:        [OK / ISSUE - descreva]
Controle de acesso:     [OK / ISSUE - descreva]
Secrets:                [OK / ISSUE - descreva]
Integridade de dados:   [OK / ISSUE - descreva]
Dependencias:           [OK / ISSUE - descreva]
```

Se encontrar issues, descreva o problema, a localizacao no codigo e a correcao recomendada.
