# Learnings

## harness
- [0.5] antes de adicionar um check ao `plugin-smoke` → medir o falso-positivo dele contra o repo inteiro antes de escrever; se o sinal for ruído, descartar o check  (evidence: 3 checks medidos antes de escrever em 2026-07-14 — o de "referência a rc-* inexistente" foi descartado por dar 8 FPs (nomes de produto/script); a versão ingênua do "dangling asset" acusou 56 problemas, dos quais 48 eram âncoras `#secao` e docs vendorizados; seen 1)

## verification
- [0.7] ao editar skill/doc/manifest/workflow neste repo (markdown+scripts, sem build) → rodar `node scripts/plugin-smoke.mjs`; o gate já encapsula os greps de link pendurado, skill órfã, resíduo de CLI e fóssil de toolchain  (evidence: ciclo edit→gate repetido ~50× em 2026-07-12..14, 23× como o tool call imediatamente seguinte a um Edit; nunca contradito; updated 2026-07-14; seen 2)
- [0.5] depois de `git push` → conferir o CI (`gh run list`) antes de dizer "pronto" — gate local verde **não** é evidência de CI verde  (evidence: em 2026-07-14 três pushes foram declarados verdes só com o gate local enquanto o CI falhava em "Set up Go: go.mod not found"; e dois "sucessos" eram job pulado pelo paths-filter, não job passando; seen 1)
