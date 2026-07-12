# SEO técnico — auditoria

Objetivo: garantir que as páginas certas sejam rastreadas, indexadas e entendidas pelos buscadores.
Ordem de prioridade: **crawl → index → markup → performance**. Não otimize markup de uma página que não indexa.

## Fase 1 — Crawl e indexação

1. **robots.txt** (`/robots.txt`): confirme que não bloqueia recursos essenciais nem as páginas que devem indexar.
   Erro comum: `Disallow: /` em ambiente que foi pra produção sem remover.
2. **sitemap.xml**: existe, está referenciado no robots.txt (`Sitemap: https://.../sitemap.xml`), lista apenas URLs 200 canônicas (sem redirects, sem noindex, sem 404).
3. **Meta robots / X-Robots-Tag**: páginas que devem indexar não têm `<meta name="robots" content="noindex">` nem header `X-Robots-Tag: noindex`.
4. **Status HTTP**: páginas canônicas retornam 200; redirects são 301 (permanente) e não encadeados (A→B→C vira A→C).

Verificação:
```bash
curl -sI https://exemplo.com/pagina | grep -iE "^(HTTP|location|x-robots-tag)"
curl -s https://exemplo.com/robots.txt
```

## Fase 2 — Tags de indexação por página

| Tag | Regra |
| --- | ----- |
| `<title>` | Único por página, 50–60 chars, keyword principal à esquerda |
| `<meta name="description">` | 140–160 chars, descreve a intenção, chama o clique (não é fator de rank direto, mas afeta CTR) |
| `<link rel="canonical">` | Aponta para a versão preferida (absoluta, https, sem parâmetros de tracking); auto-referente na página canônica |
| `<html lang>` / `hreflang` | Idioma correto; `hreflang` recíproco e com `x-default` se houver multi-idioma |
| Open Graph / Twitter Card | `og:title`, `og:description`, `og:image` (não é rank, mas controla o preview em compartilhamento) |

Erros comuns: canonical apontando para home em todas as páginas; múltiplos `<title>`; canonical relativo.

```bash
curl -s https://exemplo.com/pagina \
  | grep -ioE '<title>[^<]*</title>|<link[^>]*canonical[^>]*>|<meta[^>]*(description|robots)[^>]*>'
```

## Fase 3 — Dados estruturados (schema.org / JSON-LD)

Use JSON-LD no `<head>` ou fim do `<body>`. Tipos mais comuns: `Article`, `Product` (+ `Offer`, `AggregateRating`),
`BreadcrumbList`, `FAQPage`, `Organization`, `LocalBusiness`.

```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "Article",
  "headline": "Título",
  "datePublished": "2026-01-15",
  "author": {"@type": "Person", "name": "Autor"}
}
</script>
```

Regras: o schema deve refletir o conteúdo **visível** na página (marcar dado ausente é violação);
valide no Rich Results Test do Google e no validator.schema.org.

## Fase 4 — Core Web Vitals (performance)

| Métrica | Bom | O que ataca |
| ------- | --- | ----------- |
| LCP (maior conteúdo) | < 2,5s | otimizar imagem hero, preload de recurso crítico, TTFB do servidor |
| CLS (deslocamento) | < 0,1 | reservar `width`/`height` em imagens e ads, evitar injeção de conteúdo acima do fold |
| INP (interatividade) | < 200ms | reduzir JS de long tasks, quebrar hidratação pesada |

Meça com Lighthouse (lab) e Search Console → Core Web Vitals (campo/CrUX). Priorize dados de campo.

## Checklist de aceite

- [ ] robots.txt não bloqueia páginas indexáveis; sitemap referenciado e só com URLs 200 canônicas
- [ ] Cada página: `<title>` único, description, canonical auto-referente correto
- [ ] JSON-LD válido e coerente com o conteúdo visível
- [ ] Sem redirects encadeados; canônicas retornam 200
- [ ] CWV medido (lab + campo quando houver acesso); LCP/CLS/INP dentro do alvo ou plano de ataque listado
