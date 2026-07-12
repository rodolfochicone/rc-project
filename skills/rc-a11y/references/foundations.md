# Fundamentos de acessibilidade

Objetivo: garantir que a estrutura base da página seja compreensível por leitores de tela,
navegação por teclado e outras tecnologias assistivas antes de qualquer componente interativo.

## HTML semântico

Landmarks — cada página deve ter no máximo um `<header>` e um `<footer>` de topo, um `<main>`
(único), e `<nav>` para blocos de navegação. Landmarks permitem que o leitor de tela salte
direto para a seção relevante.

```html
<header>...</header>
<nav aria-label="Principal">...</nav>
<main>
  <h1>Título da página</h1>
  ...
</main>
<footer>...</footer>
```

Heading (`h1`–`h6`) em hierarquia — um único `h1` por página, sem saltar níveis (h2 → h4 sem h3).
Heading não é escolhido pelo tamanho visual da fonte; é escolhido pela posição na estrutura do
documento. Use CSS para estilizar, não para simular hierarquia.

Listas (`<ul>`/`<ol>`/`<li>`) — qualquer conjunto de itens relacionados (menu, cards, resultados)
deve usar lista semântica, não `<div>`s soltas, para que o leitor de tela anuncie "lista com N itens".

Botão vs link — `<button>` dispara uma ação na página atual (abrir modal, submeter, alternar
estado); `<a href>` navega para outra URL/âncora. Nunca use `<div onClick>` para nenhum dos dois:
perde foco por teclado, `role`, e ativação por Enter/Espaço de graça.

## WCAG 2.2 — princípios POUR

| Princípio | Significado | Exemplo |
| --------- | ------------ | ------- |
| **P**erceptível | A informação existe em mais de um canal sensorial | alt text em imagem, legenda em vídeo, contraste de cor |
| **O**perável | Toda função funciona por teclado e sem limite de tempo forçado | sem trap de foco, sem timeout sem aviso |
| **C**ompreensível | Texto e comportamento são previsíveis | labels claras, erro de formulário explica o que corrigir |
| **R**obusto | Funciona em diferentes tecnologias assistivas | HTML válido, ARIA usada corretamente |

**Nível AA** é o padrão de conformidade mais citado em requisitos legais e de produto (ex.: WCAG
2.2 AA). Nível A é o mínimo aceitável isolado; AAA é o mais estrito e raramente exigido em produto
inteiro. Ao dizer "o componente é acessível", o padrão implícito é AA salvo indicação contrária.

## Texto alternativo em imagens

- Imagem com conteúdo informativo: `alt` descreve o que a imagem comunica, não repete "imagem de".
- Imagem puramente decorativa (ilustração de fundo, ícone redundante com texto ao lado):
  `alt=""` (vazio, não omitido) — isso instrui o leitor de tela a pular a imagem.
- Imagem que é link/botão sem outro texto: o `alt` descreve o destino/ação, não a imagem em si.

```html
<img src="grafico-vendas.png" alt="Vendas cresceram 20% no trimestre">
<img src="ornamento.svg" alt="">
```

## Idioma da página

`<html lang="pt-BR">` (ou o idioma correto) é obrigatório — sem ele, leitores de tela usam a
pronúncia/voz padrão errada para todo o conteúdo. Se um trecho está em outro idioma, marque
o trecho com `lang` local (`<span lang="en">...</span>`).

## Checklist de aceite

- [ ] Landmarks corretos: um `<main>`, `<nav>` com `aria-label` quando houver mais de um
- [ ] Um único `h1`; hierarquia de heading sem saltos
- [ ] Listas semânticas para conjuntos de itens; `<button>`/`<a>` usados pelo papel correto, nunca `<div onClick>`
- [ ] Toda imagem tem `alt` (descritivo ou `alt=""` se decorativa)
- [ ] `<html lang>` definido corretamente
