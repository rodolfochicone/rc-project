# Memory Index

- [Processo de release do plugin](release-process.md) — convention — bump dos 2 manifests `.claude-plugin/`, CHANGELOG, tag anotada, `gh release`, gotcha do `--latest`
- [Cross-links entre artefatos .rc/](artifact-cross-links.md) — convention — PRD⇄TechSpec⇄tasks via links md relativos; backlink de review no corpo, nunca no frontmatter
- [O gate é o plugin-smoke](gate-sensor-over-patch.md) — convention — defeito de classe vira sensor, não só patch; FP medido antes de escrever (85% na versão ingênua); tier de aviso
- [Uma regra, um dono](skill-rule-ownership.md) — convention — a seção de recap duplica o Workflow; a régua de 7 degraus, e por que hook/skill/`references/` são os únicos donos que viajam com o plugin
- [Fósseis do fork Compozy](compozy-fork-fossils.md) — context — o de-fork reescreveu conteúdo e esqueceu nomes/config; o próximo está na prosa das skills e **fora do repo** (`~/.claude/`), onde nenhum gate alcança
- [Quando o loop autônomo vale a pena](loop-calibration.md) — decision — backlog vem do sensor, nunca da imaginação; custo medido vs. fazer na mão
- [auto-docs.yml e prompt injection](auto-docs-prompt-surface.md) — context — o que segura hoje é o gatilho `merged == true`; se mudar, redesenhar antes
