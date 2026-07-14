---
title: Quando o loop autônomo vale a pena neste repo
scope: decision
key: loop-calibration
tags: [rc-loop, roadmap, autonomia, custo]
source: rc-memory (distilled 2026-07-14)
created: 2026-07-14
updated: 2026-07-14
---

O `/rc-loop` rodou pela primeira vez aqui em 2026-07-14 (2 fases, roadmap esgotado, gate verde).
O que ficou calibrado:

**O backlog tem que vir do sensor, não da imaginação.** A `rc-roadmap` proíbe o loop de inventar
intenção, e este repo vive com backlog vazio — todo slug em `.rc/tasks/` completo. O roadmap só
existiu porque o check `dangling asset` produziu 8 achados objetivos, cada um com critério de
pronto verificável (ver [[gate-sensor-over-patch]]). Sem sensor produzindo trabalho, a pergunta 4
do readiness ("backlog grande o bastante?") responde **não**, e o certo é ficar no `/rc-pipe`.

**Custo medido:** 110k tokens, 34 tool calls, ~18 min de relógio para 5 correções mecânicas que
sairiam em ~5 min na mão. O loop não economizou tempo — pagou em *rigor* (cada agente leu o
conteúdo real antes de decidir se movia o arquivo ou apagava a promessa) e em dogfood. Para 1-2
fases triviais, **fluxo humano é mais rápido e mais seguro**; o loop se paga em migração/
build-out grande.

**O formato empurra de volta, e ele tem razão:** o invariante 1 do `roadmap-format.md` (fase é
épico, não task) rejeitou uma proposta de 5 fases de uma task cada. Fundir em 2 fases foi o certo.

**Cerimônia que não cabe:** o `rc-loop` manda planejar cada fase com `rc-create-tasks`. Para fase
mecânica já totalmente especificada pela saída do gate, isso é overhead — foi pulado
deliberadamente, e o loop funcionou. Se a fase for de verdade um épico, não pule.
