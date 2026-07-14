# Loop Engineering — explicação completa

> Análise do artigo `docs/loop.txt`. Referências no formato `loop.txt:linha`.

## Veredito em um parágrafo

O texto defende uma mudança de postura no uso de agentes de IA: parar de **prompt engineering** (dar uma instrução por vez e revisar o resultado manualmente) e passar a **loop engineering** (projetar o ciclo automatizado que prompta o agente, verifica o resultado, decide o próximo passo e continua até o trabalho passar num critério de sucesso). A frase-síntese está em `loop.txt:12` e `loop.txt:16`: enquanto você revisa e corrige manualmente, *"o humano ainda é o loop"* — loop engineering tira o humano de dentro do ciclo e o coloca projetando o ciclo.

---

## 1. A ideia central: quem é o loop?

O modelo antigo (`loop.txt:47-59`):

```
você prompta → agente responde → você revisa → você acha o erro → você corrige → repete
```

Nesse fluxo, o gargalo é humano. Você não escala além da sua própria atenção.

O modelo novo (`loop.txt:60-67`):

```
você define a meta → o loop descobre o que é preciso → o loop planeja
→ o agente executa → um verificador checa → o loop corrige falhas
→ o sistema para quando a meta é atingida
```

A distinção mais afiada do texto (`loop.txt:68-69`):
> *Prompting dá ao agente uma **instrução**. Loop engineering dá ao agente um **trabalho**.*

E o exemplo concreto da diferença de mentalidade (`loop.txt:319-322`):
- Prompt engineer: *"Escreva uma função."*
- Loop engineer: *"Escreva, teste, corrija até passar, depois resuma a mudança."*

O texto cita Boris Cherny (head do Claude Code na Anthropic) em `loop.txt:18-19`: *"Eu não prompto o Claude mais. Tenho loops rodando que promptam o Claude e descobrem o que fazer. Meu trabalho é escrever loops."*

---

## 2. Os 5 estágios do ciclo

Todo loop, independentemente do tamanho, tem o mesmo esqueleto (`loop.txt:74-82`, resumido em `loop.txt:295-300` como "Goal. Action. Check. Fix. Repeat"):

| Estágio | O que faz |
|---|---|
| **Discover** | Descobre o que é necessário para atingir a meta |
| **Plan** | Planeja o trabalho |
| **Execute** | O agente faz a tarefa |
| **Verify** | Um verificador (idealmente outro agente) checa o resultado |
| **Iterate** | Se falhou, volta ao loop; se passou, entrega |

A regra de ouro (`loop.txt:80-84`): *se passa, entrega; se falha, volta pro loop.* Não é um prompt perfeito — é um sistema que melhora a saída até ela atingir o padrão.

---

## 3. Dois tamanhos de loop

**Single-Agent Loop** (`loop.txt:87-100`) — um agente roda o ciclo inteiro (descobre, planeja, executa, verifica, corrige). É como uma pessoa reescrevendo o próprio rascunho. Bom para tarefas focadas, escopo pequeno: correção de bug, resumo de pesquisa, rascunho de conteúdo.

**Fleet Loop** (`loop.txt:101-119`) — um **orquestrador** recebe a meta, quebra em pedaços e distribui para **especialistas**, que por sua vez usam **subagentes** para tarefas estreitas. É um pequeno time rodando um projeto de ponta a ponta:

```
                 Orquestrador (dono da missão)
        ↓                 ↓                 ↓
   Research          Engineering           QA
   Specialist        Specialist        Specialist
        ↓                 ↓                 ↓
   Web             Code Writer         Test Writer
   Researcher      + Debugger          + Bug Tracker
```

---

## 4. Dois tipos de loop (a distinção prática mais importante)

O texto marca isso como *"a distinção prática mais importante"* (`loop.txt:121`).

**Open Loops** (`loop.txt:123-135`) — exploratórios. Você dá uma meta ampla e deixa o agente buscar o caminho. Poderoso (descobre coisas que você não especificou), mas caro e bagunçado: tenta caminhos demais, queima tokens, produz saída ruim rápido, deriva da meta real, fica difícil de controlar.

**Closed Loops** (`loop.txt:136-151`) — limitados. O humano projeta o caminho antes; o loop roda sozinho, mas dentro de regras claras. Tem: meta clara, passos definidos, avaliação após cada passo, condição de parada e ponto de hand-off se travar.

A recomendação prática (`loop.txt:150-151`): **comece com closed loops**; abra-os depois, quando suas verificações estiverem robustas. Closed loop é mais barato, mais confiável, saída mais limpa.

---

## 5. Os 6 blocos de construção

Conceitualmente são 5 estágios, mas na prática você precisa de 6 blocos para fazer o loop funcionar (`loop.txt:152-224`):

1. **Automations** (`:155-164`) — o "batimento cardíaco". O que dispara o loop sem você lembrar de rodar: todo dia de manhã, quando abre um PR, quando um arquivo muda, quando surge um ticket, até todos os testes passarem. *Se você ainda precisa iniciar tudo manualmente, o loop não está trabalhando de verdade.*

2. **Worktrees** (`:165-171`) — isolamento. Quando vários agentes editam código, sem separação eles colidem (dois editam o mesmo arquivo, um sobrescreve o outro). Um worktree dá a cada agente um workspace/branch limpo, permitindo paralelismo sem virar bagunça.

3. **Skills** (`:172-183`) — conhecimento reutilizável do projeto. Em vez de explicar o projeto toda vez, você escreve o contexto uma vez: visão, arquitetura, regras, passos de build, passos de teste, coisas que o agente nunca deve fazer. *Sem skills, todo loop começa "frio".*

4. **Plugins e Connectors** (`:184-196`) — acesso a ferramentas reais (GitHub, Slack, Linear, Jira, Gmail, Drive, banco, API de staging). É a diferença entre *"aqui está uma sugestão de fix"* e *"abri o PR, linkei o ticket, acompanhei o CI e postei o update"*.

5. **Subagents** (`:197-208`) — separar quem faz de quem verifica. O agente que escreveu o código é generoso demais ao revisá-lo; o que escreveu o artigo não vê as próprias seções fracas. Use agentes distintos para exploração, implementação, revisão, teste, fact-checking e resumo final. *A qualidade melhora quando o revisor não é o mesmo que fez o trabalho.*

6. **Memory** (`:209-224`) — o que permite o loop continuar entre execuções. O modelo esquece; o repo, as notas, o log do projeto e os tickets não. Pode viver em Markdown, logs, tickets do Linear, issues do GitHub, Obsidian, bancos, Claude Projects. *Sem memória, o loop começa do zero toda vez.*

---

## 6. Exemplos concretos de loop

O texto dá quatro esqueletos idênticos aplicados a domínios diferentes (`loop.txt:225-300`):

- **Coding Loop** (`:227-244`): lê VISION/ARCHITECTURE → planeja mudança → edita → roda testes → se falha, lê erro/corrige/re-testa → se passa, resume → para.
- **Research Loop** (`:245-261`): define pergunta → busca fontes → resume → verifica claims contra fontes → compara conflitos → sintetiza → para quando bate o limiar de confiança.
- **Content Loop** (`:262-278`): define tópico/público/meta → rascunho → agente crítico revisa → reescreve → pontua contra critérios → publica se passa, reescreve se falha.
- **Sales Outreach Loop** (`:279-294`): define ICP → acha leads → enriquece → qualifica → personaliza → revisão de qualidade → envia ou escala para humano.

Todos compartilham o mesmo esqueleto: **Meta → Ação → Check → Fix → Repete até pronto** (`loop.txt:295-300`).

---

## 7. O problema real: custo de tokens

Esta é a parte que o texto insiste que "ninguém fala o suficiente" (`loop.txt:20-45, 351-354`):

- Um loop de código médio: **50K–200K tokens** (`:23`)
- Um fleet loop (orquestrador + especialistas): **500K–2M tokens** (`:24`)
- Um loop diário agendado: **milhões de tokens por semana** (`:25`)

Cada retry, cada auto-correção, cada verificação, cada subagente custa tokens (`:26-29`). A tese: *loop engineering não é difícil porque a ideia é complicada — é difícil porque a maioria não pode bancar deixar agentes rodando livremente por muito tempo* (`:31-32`).

A solução apontada (`loop.txt:35-45`): **modelos baratos de contexto longo** viabilizam loops. Para rodar loops diariamente você precisa de input/output baratos, janelas de contexto grandes, tool calling, saída JSON, alta concorrência e contexto suficiente para lembrar o que aconteceu antes no loop. Sem isso, loops são "experimentos caros"; com isso, viram "workflows práticos".

---

## 8. A conclusão do texto

O "unlock" (`loop.txt:355-361`): pare de tentar escrever um prompt perfeito; comece a construir o loop que torna saídas imperfeitas melhores. A frase de fecho: **"Um loop confiável vence um prompt perfeito."**

---

## Avaliação crítica (o que o texto acerta e o que ele omite)

O documento é um artigo de opinião/marketing bem estruturado, não um paper técnico. Separando fato de retórica:

**O que sustenta bem:**
- A separação maker/checker (bloco 5) é um princípio real e verificável — viés de auto-avaliação de LLMs é documentado. É o argumento mais sólido.
- A distinção closed vs. open loop mapeia diretamente o trade-off entre autonomia e controle/custo. "Comece fechado" é conselho prudente.
- O critério de parada explícito (verify + stop condition) é o que diferencia um sistema de um agente que roda em círculos queimando tokens.

**O que o texto glosa por cima (pontos cegos):**
- **A verificação é o elo frágil.** O loop só é tão bom quanto o verificador. O texto assume que "um checker verifica o resultado" resolve, mas se o critério de sucesso não for automatizável e objetivo (ex.: testes que passam, schema válido), o verificador vira outro palpite de LLM — e você trocou um humano-no-loop por um custo maior sem ganho de confiabilidade. Em código isso funciona (testes são um oráculo real); em "conteúdo" e "sales", o "score contra critérios" é muito mais mole.
- **Números de custo sem fonte.** Os ranges de tokens (`:23-25`) são plausíveis mas não referenciados; trate como ordem de grandeza, não benchmark.
- **Loops abertos ainda são frágeis na prática** — o próprio texto admite deriva e caos, mas subvende quão difícil é fazer um closed loop *realmente* fechar sem intervenção humana em domínios sem oráculo objetivo.
- **Falta o tema de segurança/blast radius.** Um loop com connectors que "abre PR, mexe no banco, envia e-mail" sozinho precisa de guardrails — o texto não toca nisso.

---

## Conexão com este projeto (k2-ops-agent)

Este próprio repositório é um exemplo vivo de vários desses blocos:
- **Skills** (bloco 3) → o `CLAUDE.md` do projeto é exatamente o "arquivo de skill" descrito (visão, arquitetura, regras, testes, "coisas que nunca fazer").
- **Memory** (bloco 6) → o projeto já usa Postgres próprio para "sessions, memory, endpoint catalog".
- **Plugins/Connectors** (bloco 4) → os toolkits (DB, REST API, CloudWatch, GitHub) são precisamente os connectors do texto.
- **Guardrails** → o `guardrail.guarded_write` (staging executa, produção exige confirmação) é justamente o guardrail de blast radius que o artigo *não* menciona, mas que um loop com poder de escrita precisa. O projeto está à frente do texto nesse ponto.

---

## Bottom line

Loop engineering = deixar de operar agentes um prompt por vez e passar a projetar o sistema fechado (meta → descobre → planeja → executa → verifica → itera → para) que os opera sozinho, montado a partir de 6 blocos (automations, worktrees, skills, plugins, subagents, memory). O gargalo prático não é a ideia, é o custo de tokens, resolvido por modelos baratos de contexto longo. O ponto mais forte é separar quem faz de quem verifica; o ponto mais frágil, que o texto subestima, é que o loop só vale o quanto vale seu verificador.
