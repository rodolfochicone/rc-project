# Componentes interativos acessíveis

Objetivo: garantir que formulários, modais e widgets compostos (menus, tabs, accordion) sejam
operáveis por teclado, anunciados corretamente por leitor de tela e com foco gerenciado.

## Formulários

Toda entrada precisa de uma `<label>` associada — nunca só `placeholder` (some ao digitar, não
é lido de forma confiável por todo leitor de tela).

```html
<label for="email">E-mail</label>
<input id="email" type="email" aria-describedby="email-erro" aria-invalid="true">
<p id="email-erro" role="alert">Informe um e-mail válido.</p>
```

- `aria-invalid="true"` no campo com erro.
- `aria-describedby` liga o campo à mensagem de erro/ajuda (o leitor de tela lê a descrição
  depois do label ao focar o campo).
- Agrupamento de campos relacionados (ex.: endereço, opções de rádio de um mesmo grupo):
  `<fieldset>` + `<legend>` — a legenda é anunciada como contexto de cada campo do grupo.

```html
<fieldset>
  <legend>Forma de pagamento</legend>
  <label><input type="radio" name="pagamento" value="pix"> Pix</label>
  <label><input type="radio" name="pagamento" value="cartao"> Cartão</label>
</fieldset>
```

## Modal / dialog

```html
<div role="dialog" aria-modal="true" aria-labelledby="titulo-modal">
  <h2 id="titulo-modal">Confirmar exclusão</h2>
  ...
  <button>Cancelar</button>
</div>
```

- `role="dialog"` + `aria-modal="true"` (ou o elemento nativo `<dialog>`, que já cobre isso).
- **Focus trap**: Tab dentro do modal não escapa para o conteúdo por trás enquanto ele está aberto.
- **Retorno de foco**: ao fechar, o foco volta para o elemento que abriu o modal (não para `<body>`).
- **Esc fecha** o modal.
- `aria-labelledby` aponta para o título do modal, para o leitor de tela anunciar o nome ao abrir.

## Menus, tabs, disclosure, accordion

Esses padrões têm interação de teclado específica (setas, Home/End, Enter/Espaço) definida pelo
**ARIA Authoring Practices Guide (APG)** — não invente a interação; siga o padrão do APG para o
widget (ex.: Tabs Pattern, Menu Button Pattern, Accordion Pattern, Disclosure Pattern) e reuse a
implementação de referência antes de montar uma do zero.

Disclosure simples (mostrar/ocultar), sem widget composto:

```html
<button aria-expanded="false" aria-controls="painel-detalhes">Detalhes</button>
<div id="painel-detalhes" hidden>...</div>
```

`aria-expanded` reflete o estado atual; `hidden` (ou `aria-hidden` + display none) remove o painel
fechado da árvore de acessibilidade.

## Navegação por teclado e gestão de foco

- **Ordem de tab lógica**: segue a ordem visual/de leitura do DOM. Evite `tabindex` positivo
  (`tabindex="1"`, `"2"`...) — ele quebra a ordem natural; use apenas `tabindex="0"` (inclui no
  fluxo natural) ou `tabindex="-1"` (foco só programático).
- **`:focus-visible`**: nunca remova o outline de foco sem substituir por um indicador visível
  equivalente (`outline: none` sem alternativa é bloqueador).
- **Skip link**: primeiro elemento focável da página, permite ir direto ao `<main>` sem passar
  por todo o menu a cada carregamento.

```html
<a class="skip-link" href="#conteudo-principal">Ir para o conteúdo</a>
...
<main id="conteudo-principal">...</main>
```

- **SPA / navegação por rota**: ao trocar de "página" sem reload real, mova o foco programaticamente
  (ex.: para o `h1` da nova view) e anuncie a mudança — o leitor de tela não percebe a troca de
  rota sozinho como perceberia um load de página real.

## Quando NÃO usar ARIA

- Se existe elemento HTML nativo com o comportamento certo (`<button>`, `<details>`, `<dialog>`,
  `<select>`), use-o em vez de recriar o papel com `role`/`aria-*` em uma `<div>`.
- Não adicione `role` redundante ao elemento que já tem esse papel implícito (`<button role="button">`).
- ARIA nunca muda comportamento do navegador (foco, ativação por teclado) — só a semântica anunciada.
  Adicionar `role="button"` a uma `<div>` sem também tratar `tabindex`, `Enter`/`Espaço` e foco é
  ARIA quebrada, pior que não ter ARIA.

## Checklist de aceite

- [ ] Todo input tem `<label>` associada; erro usa `aria-invalid` + `aria-describedby`
- [ ] Modal: `role="dialog"`/`aria-modal`, focus trap, Esc fecha, foco retorna ao gatilho ao fechar
- [ ] Widgets compostos (menu/tabs/accordion) seguem o padrão de teclado do ARIA APG
- [ ] Nenhum `tabindex` positivo; `:focus-visible` sempre perceptível
- [ ] Skip link presente; foco movido/anunciado em troca de rota de SPA
- [ ] ARIA usada só onde o HTML nativo não resolve, e sempre com o comportamento de teclado correspondente implementado
