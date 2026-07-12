# Teste e revisão de acessibilidade

Objetivo: verificar com ferramentas e navegação real que a UI é operável e compreensível por
teclado e tecnologia assistiva — não apenas que o markup "parece" correto na leitura estática.

## Navegação só por teclado

Desconecte o mouse (ou simplesmente não o use) e percorra o fluxo completo:

| Tecla | Deve fazer |
| ----- | ---------- |
| `Tab` / `Shift+Tab` | Move o foco para o próximo/anterior elemento interativo, em ordem lógica |
| `Enter` | Ativa botão, link, submete formulário |
| `Espaço` | Ativa botão, marca checkbox |
| `Esc` | Fecha modal/menu/popover aberto |
| Setas | Move seleção dentro de um widget composto (tabs, menu, radio group) |

Falhas comuns: foco "sequestrado" (nunca sai de um componente), foco invisível (sem
`:focus-visible`), elemento clicável que o Tab nunca alcança.

## Leitores de tela

- **macOS — VoiceOver**: `Cmd+F5` para ativar; `VO+Right/Left` (Ctrl+Option por padrão) para
  navegar por elemento; `VO+U` abre o rotor para navegar por headings/links/landmarks.
- **Windows — NVDA**: gratuito; `Insert+Down` lê em modo contínuo; `H` salta entre headings,
  `Tab` entre campos, `D` entre landmarks.

Ouça se: o nome do controle é anunciado, o estado (`expandido`, `selecionado`, `inválido`) é
anunciado, e a ordem de leitura corresponde à ordem visual.

## Ferramentas automáticas

- **axe DevTools** (extensão de browser) — roda um scan da página renderizada e aponta violações
  com referência à regra WCAG.
- **Lighthouse** (aba Accessibility no Chrome DevTools) — score rápido + lista de problemas.

Aviso importante: **ferramentas automáticas pegam ~30–40% dos problemas de acessibilidade** —
coisas como contraste, `alt` ausente, `label` ausente. O resto (ordem de foco lógica, texto de
erro que faz sentido, se o widget de fato funciona com leitor de tela) exige teste manual. Score
100 no axe/Lighthouse não significa "acessível", significa "sem os problemas que automação detecta".

## Contraste de cor

Ratio mínimo AA:

| Conteúdo | Ratio mínimo |
| -------- | ------------- |
| Texto normal (< 18px, ou < 14px bold) | 4.5:1 |
| Texto grande (≥ 18px, ou ≥ 14px bold) | 3:1 |
| Componentes de UI (borda de input, ícone informativo) | 3:1 |

Verifique com o color picker do DevTools (mostra o ratio direto) ou axe DevTools, que já reporta
falhas de contraste na página renderizada.

## Checklist de revisão de PR de a11y

- [ ] Fluxo completo percorrido só com teclado, sem mouse
- [ ] Foco visível em todo elemento interativo (`:focus-visible` não removido sem substituto)
- [ ] Scan automático (axe/Lighthouse) roda sem violação nova introduzida pelo PR
- [ ] Contraste checado nos elementos novos/alterados (4.5:1 texto normal, 3:1 texto grande/UI)
- [ ] Pelo menos um teste manual com leitor de tela (VoiceOver ou NVDA) no fluxo principal alterado
- [ ] Se algo ficou pendente por falta de acesso a browser real, isso está dito explicitamente na revisão, não omitido

## Checklist de aceite

- [ ] Teclado, leitor de tela e ferramenta automática usados juntos — nenhum sozinho é suficiente
- [ ] Ratios de contraste verificados com número, não "parece ok"
- [ ] Revisão documenta o que foi testado e o que ficou pendente por limitação de ambiente
