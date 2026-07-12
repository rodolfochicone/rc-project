---
name: rc-a11y
description: Guias de acessibilidade (a11y) para interfaces web seguindo WCAG 2.2 AA. Use ao construir ou revisar UI (componentes React, formulários, modais, navegação, tabelas) para garantir semântica HTML correta, uso apropriado de ARIA, navegação por teclado, gestão de foco, contraste de cor e suporte a leitores de tela. Carrega o guia certo por tarefa a partir de references/. Não use para acessibilidade de PDF/documentos, auditoria jurídica formal de conformidade, ou design visual sem relação com a11y.
user-invocable: true
model: sonnet
effort: medium
---

# Acessibilidade (a11y) — guias por tarefa

Guias práticos de acessibilidade web. Leia o guia da tarefa em `references/` antes de agir —
cada um traz o checklist, os exemplos de markup e os critérios de aceite. Ao revisar UI real,
combine com Grep/Glob para inspecionar o markup/componentes gerados; não presuma a semântica.

## Roteamento

| Tarefa | Guia |
| ------ | ---- |
| Fundamentos — HTML semântico, WCAG POUR, nível AA, alt text, lang | `references/foundations.md` |
| Componentes interativos — formulários, modal, menus/tabs, foco/teclado | `references/patterns.md` |
| Teste e revisão — teclado, leitor de tela, ferramentas automáticas, contraste | `references/testing.md` |

## Princípios sempre válidos

- **HTML semântico primeiro.** Elemento nativo (`<button>`, `<nav>`, `<label>`) resolve a maioria dos casos antes de qualquer ARIA.
- **ARIA só quando o HTML nativo não resolve.** "No ARIA is better than bad ARIA" — ARIA incorreta é pior que ausência de ARIA.
- **Tudo operável por teclado.** Se uma ação só funciona com mouse/toque, é um bloqueador de acessibilidade.
- **Contraste e foco visíveis.** Cor não é o único indicador de estado; `:focus-visible` sempre perceptível.

## Error Handling

- SPA renderizada no cliente pode esconder problemas do leitor de tela (DOM que só existe após JS rodar). Sem acesso ao browser, revise o markup/componentes estaticamente e diga explicitamente que a verificação com screen reader/axe ficou pendente.
