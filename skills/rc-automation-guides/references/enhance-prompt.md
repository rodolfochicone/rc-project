# Skill: /enhance-prompt

Melhora e reestrutura o prompt do usuario antes de executar a tarefa. O objetivo e transformar qualquer instrucao em um prompt otimizado em markdown que maximize a eficiencia do Claude Code.

## Uso

```
/enhance-prompt <seu prompt original aqui>
```

O argumento `$ARGUMENTS` contem o prompt original do usuario.

---

## Instrucoes

Voce recebeu o seguinte prompt do usuario:

> $ARGUMENTS

### Passo 1 — Analise do prompt original

Analise o prompt original e identifique:

- **Objetivo principal**: qual e a tarefa central solicitada?
- **Contexto implicito**: que informacoes estao subentendidas?
- **Ambiguidades**: ha partes vagas que precisam ser clarificadas?
- **Escopo**: qual o tamanho e complexidade da tarefa?

### Passo 2 — Reestruturacao em markdown otimizado

Reescreva o prompt seguindo esta estrutura markdown:

```markdown
## Objetivo
[Descricao clara e direta do que precisa ser feito]

## Contexto
[Informacoes relevantes sobre o estado atual, arquivos envolvidos, dependencias]

## Requisitos
- [Requisito 1]
- [Requisito 2]
- [...]

## Criterios de aceite
- [ ] [Criterio 1]
- [ ] [Criterio 2]
- [ ] [...]

## Restricoes (se aplicavel)
- [Restricao 1 — ex: nao alterar arquivo X, manter compatibilidade com Y]
```

### Passo 3 — Apresentacao e execucao

1. Exiba o prompt melhorado ao usuario dentro de um bloco markdown
2. Pergunte: **"Prompt melhorado acima. Deseja executar assim, ajustar algo, ou redirecionar para um comando especifico (ex: `/new-feature`, `/api-endpoint`)?"**
3. Se o usuario confirmar (ou responder "sim", "ok", "vai", "execute", "roda"), execute a tarefa diretamente usando o prompt melhorado
4. Se o usuario indicar um comando (ex: "roda com /new-feature", "usa /api-endpoint"), execute o comando indicado passando o prompt melhorado como argumento
5. Se o usuario pedir ajustes, aplique as modificacoes e apresente novamente

### Regras de melhoria

- **Seja especifico**: substitua termos vagos por acoes concretas
- **Adicione contexto do projeto**: se o prompt envolve codigo, mencione os caminhos e camadas relevantes (domain, usecases, infrastructure, interfaces)
- **Inclua criterios de aceite**: transforme expectativas implicitas em checklist verificavel
- **Respeite os padroes**: reforce regras do CLAUDE.md relevantes a tarefa (nomenclatura, clean architecture, testes, etc.)
- **Mantenha a intencao original**: nunca altere o que o usuario quer, apenas melhore como ele pede
- **Lingua**: mantenha a mesma lingua do prompt original (portugues ou ingles)
- **Concisao**: o prompt melhorado deve ser mais claro, nao necessariamente mais longo
