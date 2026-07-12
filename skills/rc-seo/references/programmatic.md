# SEO programático — páginas em escala

Objetivo: gerar muitas páginas (centenas a milhares) a partir de um dataset + template, cada uma
mirando uma keyword de cauda longa, **sem cair em conteúdo raso ou duplicado**. O risco central do
SEO programático é escala de lixo: 10.000 páginas finas machucam o site inteiro.

## Quando usar

Padrão `[modificador] + [entidade]` com demanda real e dado estruturado por trás:
- "restaurantes em {cidade}", "{produto} vs {produto}", "salário de {cargo} em {cidade}", "como converter {unidade} para {unidade}".

Só vale se cada página tiver **dado único e útil** — não apenas o nome trocado num texto boilerplate.

## Fase 1 — Dataset e keywords

1. Modele o dataset: uma linha = uma página. Cada linha precisa dos dados que tornam a página útil e distinta.
2. Valide demanda: existe volume de busca real para o padrão? Amostre algumas combinações antes de gerar tudo.
3. Descarte combinações sem dado suficiente (ex.: cidade sem nenhum restaurante cadastrado → não gere a página).

## Fase 2 — Template

- Um template por tipo de página. Partes fixas (estrutura) + slots dinâmicos (dados da linha).
- **Conteúdo único por página**: além dos campos, gere seções derivadas do dado (estatísticas, comparações,
  listas) para que o texto varie de verdade, não só o `{nome}`.
- Inclua o mesmo rigor on-page: `<title>`, description, H1, JSON-LD por página (ver `technical.md` e `on-page.md`).

## Fase 3 — Evitar os três venenos

| Veneno | Sintoma | Defesa |
| ------ | ------- | ------ |
| Thin content | página só com o boilerplate + 1 variável | mínimo de dado real por página; senão `noindex` ou não gere |
| Conteúdo duplicado | páginas quase idênticas | canonical correto; conteúdo derivado que varia; consolidar quase-duplicatas |
| Index bloat | milhares de páginas fracas indexadas | indexe só as que têm dado suficiente; `noindex` no resto; podar as que não performam |

Regra: prefira **menos páginas fortes** a muitas fracas. Cada URL indexada deve merecer existir.

## Fase 4 — Links internos e sitemap em escala

- **Internal linking programático**: cada página linka para vizinhas relevantes (mesma categoria, cidade próxima,
  produto relacionado) via regras — não deixe páginas órfãs.
- **Páginas hub**: crie índices/categorias que apontam para os clusters (ex.: "restaurantes por cidade") para
  distribuir autoridade e dar caminho de crawl.
- **Sitemap**: gere programaticamente; divida em múltiplos sitemaps (< 50.000 URLs cada) com sitemap index; inclua só URLs indexáveis.

## Fase 5 — Qualidade e monitoramento

- Rollout gradual: publique um lote, meça indexação e performance no Search Console antes de escalar.
- Métrica-chave: % de páginas geradas que efetivamente indexam e recebem cliques. Baixa taxa → o padrão é raso, corrija antes de gerar mais.
- Poda contínua: páginas que nunca indexam/rankeiam viram `noindex` ou são removidas.

## Checklist de aceite

- [ ] Cada página tem dado único suficiente (não só variável trocada em boilerplate)
- [ ] Combinações sem dado não geram página indexável
- [ ] Title/description/H1/JSON-LD por página, canonical correto
- [ ] Internal linking por regras + páginas hub; sem órfãs
- [ ] Sitemap(s) gerado(s), < 50k URLs cada, só indexáveis
- [ ] Rollout gradual com métrica de indexação/cliques antes de escalar
