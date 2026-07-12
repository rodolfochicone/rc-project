---
name: rc-seo
description: Guias de SEO para auditoria e otimização de sites e conteúdo. Use ao fazer auditoria técnica (meta tags, sitemap, robots.txt, dados estruturados schema.org, canonical, Core Web Vitals, indexação), otimização on-page e de conteúdo (pesquisa de palavra-chave, intenção de busca, títulos/headings/copy, brief de conteúdo, cobertura semântica) ou SEO programático (geração de páginas em escala, templates de landing, internal linking, gestão de slugs). Carrega o guia certo por tarefa a partir de references/. Não use para mídia paga/SEM, copywriting sem objetivo de ranqueamento, ou análise de backlinks/off-page.
user-invocable: true
model: sonnet
effort: medium
---

# SEO — guias por tarefa

Guias práticos de SEO. Leia o guia da tarefa em `references/` antes de agir — cada um traz
o checklist, os comandos de verificação e os critérios de aceite. Ao auditar um projeto real,
combine com Grep/Glob para inspecionar o HTML/templates gerados; não presuma o markup.

## Roteamento

| Tarefa | Guia |
| ------ | ---- |
| Auditoria técnica (crawl, index, tags, structured data, CWV) | `references/technical.md` |
| Otimização on-page e de conteúdo (keyword, intent, copy) | `references/on-page.md` |
| SEO programático / páginas em escala | `references/programmatic.md` |

## Princípios sempre válidos

- **Uma intenção por página.** Cada URL responde a uma intenção de busca clara. Duas intenções → duas páginas.
- **Índice antes de rank.** Uma página não ranqueia se não for rastreada e indexada. Resolva crawl/index primeiro.
- **Conteúdo > truque.** Nenhuma tag conserta conteúdo que não satisfaz a intenção. Otimização técnica é piso, não teto.
- **Meça, não adivinhe.** Verifique com comandos/ferramentas reais (curl, o HTML renderizado, Search Console) antes de afirmar um problema.

## Error Handling

- Sem acesso ao site em produção: audite os templates/HTML no repo e diga explicitamente que a verificação é estática (sem CWV de campo nem status de indexação real).
- Se o site depende de renderização client-side (SPA), sinalize que o crawler pode ver HTML vazio — verifique o HTML servido, não só o DOM.
