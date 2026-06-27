# rc — Guia de Comandos

Referência rápida dos principais comandos do rc, organizada por etapa.

> **Pré-requisitos**
>
> - O binário é `bin/rc` (use `make install` para deixá-lo no `PATH` como `rc`).
> - Rode os comandos **dentro do diretório do workspace** (o projeto que o rc gerencia).
> - O **daemon precisa estar ativo** — ele atende em `localhost:2323`. Suba com `make dev`
>   (daemon em foreground + proxy da UI web) ou `rc daemon start --foreground`.
> - As skills `/es-*` rodam **dentro do seu agente de IA** (Claude Code, Codex, etc.).
>   Instale-as com `rc setup`.

---

## 1. Pipeline ideia → código (skills `/es-*`, dentro do agente)

| Comando                      | O que faz                                                      |
| ---------------------------- | -------------------------------------------------------------- |
| `/rc-idea-factory <nome>`    | (opcional) Amadurece uma ideia crua → spec pesquisada          |
| `/rc-create-prd <nome>`      | Brainstorm + pesquisa → **PRD** de negócio                     |
| `/rc-create-techspec <nome>` | PRD → **TechSpec** (arquitetura, APIs, modelos de dados)       |
| `/rc-create-tasks <nome>`    | PRD + Spec → **tasks** executáveis                             |
| `/rc-review-round <nome>`    | Revisão de código por IA → gera issues                         |
| `/rc-fix-reviews <nome>`     | Corrige os issues de review                                    |
| `/rc-final-verify`           | Exige evidência de verificação antes de declarar "concluído"   |
| `/rc-workflow-memory`        | Memória de contexto entre tasks de um workflow                 |
| `/rc-impl-peer-review`       | Peer review cross-LLM (Opus) de uma implementação              |
| `/rc-spec-peer-review`       | Peer review de uma TechSpec                                    |
| `/rc-git [ticket]`           | Cria branch, push e abre PR com confirmação em cada passo      |
| `/rc-jira`                   | Cria, lê, comenta e transiciona issues do Jira via MCP oficial |

---

## 2. Executar / validar tasks (CLI)

```bash
rc tasks run <nome> --ide claude              # executa as tasks pendentes pelo daemon
rc tasks run <nome> --ide claude --dry-run    # prévia dos prompts, sem executar
rc tasks validate --name <nome>               # valida metadados das tasks (schema v2)
```

Flags úteis do `run`: `--skip-validation`, `--force` (segue após falha de validação em ambiente não-interativo).

---

## 3. Acompanhar runs (CLI)

```bash
rc runs attach <runId>     # abre a TUI interativa de um run
rc runs watch  <runId>     # stream textual de um run
rc runs purge              # limpa artefatos de runs terminados (retenção configurada)
```

---

## 4. Reviews (CLI)

```bash
rc reviews fetch <nome> --provider coderabbit --pr 42   # importa feedback externo
rc reviews fix   <nome> --ide claude --concurrent 2 --batch-size 3
rc reviews list  <nome>     # resumo do último round
rc reviews show  <nome>     # round + linhas de issues
rc reviews watch <nome>     # run de review contínuo
```

---

## 5. Workspaces & sincronização

```bash
rc workspaces register <path>     # cadastra um workspace explicitamente
rc workspaces resolve  <path>     # resolve ou registra lazy (igual ao onboarding da UI)
rc workspaces list                # lista os registrados
rc workspaces show <id>           # detalhes de um
rc workspaces unregister <id>     # remove (se não houver runs ativos)
rc sync                           # reconcilia .rc/tasks → dashboard (global.db)
```

> Na UI, a shell é **single-workspace-per-tab**. Com 2+ workspaces registrados, aparece o
> botão **"Switch workspace"** na sidebar para alternar qual workspace esta aba usa.

---

## 6. Execução ad-hoc (sem montar PRD)

```bash
rc exec "seu prompt aqui" --ide claude       # 1 prompt pelo pipeline, headless
rc exec "..." --ide claude --tui             # com TUI interativa
rc exec --prompt-file ./p.md --ide claude    # prompt vindo de arquivo
echo "prompt" | rc exec --ide claude          # prompt via stdin
```

Modos de saída: texto (padrão), `--json` (JSONL enxuto), `raw-json` (stream completo). Use `--verbose` para logs operacionais.

---

## 7. Setup / daemon / utilitários

```bash
rc setup                       # instala as skills es-* no seu agente
rc setup --all                 # instala tudo em todos os agentes detectados
rc add skill rc-git --agent claude --yes   # instala uma única skill num agente
rc install                     # lista os recursos instaláveis
rc install --rtk               # instala só o rtk (runtime toolkit), sem o setup completo
rc install --headroom          # instala o headroom (AI toolkit) via pipx/pip
rc install --rtk --guide       # mostra o tutorial de primeiros passos do recurso (sem instalar)
rc daemon start --foreground   # sobe o daemon (make dev = isto + proxy da UI web)
rc daemon status               # estado do daemon
rc agents                      # descobre e inspeciona agentes reutilizáveis
rc ext list                    # lista extensões
rc ext enable <ext>            # habilita uma extensão (ex.: rc-idea-factory)
rc ext install ...             # instala extensão de um repo/remote
rc archive                     # arquiva workflows concluídos em .rc/tasks/_archived/
rc migrate                     # converte artefatos legados para frontmatter
rc --help                      # ajuda completa
```

---

## Fluxo típico de ponta a ponta

```text
/rc-create-prd minha-feature
  → /rc-create-techspec minha-feature
  → /rc-create-tasks minha-feature
  → rc tasks run minha-feature --ide claude
  → rc runs watch <runId>
  → /rc-review-round minha-feature
  → rc reviews fix minha-feature --ide claude
  → rc sync          # o dashboard web atualiza
```

---

## Divisão de responsabilidades

- **Terminal / agente de IA** = onde você **cria e executa** (skills `/es-*` + comandos `rc`).
- **Dashboard web** (`localhost:2323`) = onde você **acompanha** workflows, runs, reviews e memory.
- **Artefatos** vivem em `.rc/tasks/<nome>/` dentro do workspace (markdown versionável).
