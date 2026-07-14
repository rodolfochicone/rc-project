# Memory Index

- [Processo de release do plugin](release-process.md) — convention — bump dos 2 manifests `.claude-plugin/`, CHANGELOG, tag anotada, `gh release`, gotcha do `--latest`
- [Cross-links entre artefatos .rc/](artifact-cross-links.md) — convention — PRD⇄TechSpec⇄tasks via links md relativos; backlink de review no corpo, nunca no frontmatter
- [O gate é o plugin-smoke](gate-sensor-over-patch.md) — convention — defeito de classe vira sensor, não só patch; e o gate não vê `description` de frontmatter
- [Fósseis do fork Compozy](compozy-fork-fossils.md) — context — o de-fork reescreveu conteúdo e esqueceu nomes/config; onde procurar o próximo
- [Quando o loop autônomo vale a pena](loop-calibration.md) — decision — backlog vem do sensor, nunca da imaginação; custo medido vs. fazer na mão
- [auto-docs.yml e prompt injection](auto-docs-prompt-surface.md) — context — o que segura hoje é o gatilho `merged == true`; se mudar, redesenhar antes
