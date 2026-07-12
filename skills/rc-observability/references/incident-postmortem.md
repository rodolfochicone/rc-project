# Resposta a incidente e postmortem

Objetivo: conter o impacto rápido, entender a causa real (não a mais óbvia) e deixar um
registro que impede a repetição. Ordem de prioridade: **conter → comunicar → investigar →
documentar**. Não pule para o postmortem sem primeiro estabilizar o serviço.

## Resposta a incidente

### Classificação de severidade

| Severidade | Critério | Exemplo |
| ---------- | -------- | ------- |
| **SEV1** | Impacto amplo, sem workaround, perda de dado possível | Fila principal parada, mensagens acumulando na DLQ sem reprocesso |
| **SEV2** | Impacto parcial ou degradado, workaround existe | Uma automação específica falhando, resto da frota ok |
| **SEV3** | Impacto mínimo ou já contido | Erro pontual, retry resolveu, sem acúmulo |

Classifique pelo **impacto real observado**, não pela gravidade técnica do erro — um stack
trace assustador que não afeta ninguém é SEV3; um erro discreto que trava todo pedido é SEV1.

### Papéis

- **Comandante do incidente**: coordena, decide, não é necessariamente quem conserta o código.
- **Comunicação**: mantém stakeholders informados em intervalos regulares, sem que precisem perguntar.
- **Investigação**: foca em causa e mitigação; não se distrai respondendo perguntas de status.

Em incidentes pequenos (SEV3) uma pessoa acumula os três papéis; em SEV1/SEV2, separe-os —
misturar investigação com "responder todo mundo perguntando o que está acontecendo" atrasa
a resolução.

### Canal e comunicação

Abra um canal dedicado por incidente (não overload no canal geral da equipe). Primeira
mensagem: o que se sabe, impacto observado, próxima atualização em X minutos — mesmo que
"ainda investigando" seja a única informação nova.

## Análise de causa raiz

**5 whys** — pergunte "por quê" repetidamente até chegar numa causa sistêmica, não numa
causa que é só "o código tinha um bug":

```
1. Por que o pedido não foi processado? → A Lambda lançou timeout.
2. Por que deu timeout? → A chamada ao serviço de pagamento demorou 30s.
3. Por que demorou 30s? → O serviço de pagamento estava sob throttle.
4. Por que não tratamos o throttle? → Não há retry com backoff nessa chamada.
5. Por que não há retry? → O template de integração usado não inclui política de retry por padrão.
```

A causa raiz aqui não é "faltou retry nessa Lambda" (sintoma local) — é "o template usado por
toda a frota não tem retry padrão". Pare quando a próxima pergunta sair do controle do time
(ex.: "por que o serviço de pagamento externo throttla" não é mais causa raiz sua) ou quando
achar algo estrutural e corrigível.

**Evite parar na primeira causa aparente.** "Deploy quebrou" é sintoma; "por que o deploy quebrou
sem ser pego antes de produção" é a pergunta que leva à causa raiz.

## Postmortem blameless

Foco no **sistema** que permitiu a falha, nunca na pessoa que apertou o botão. Frases proibidas:
"o desenvolvedor deveria ter...", "erro humano de...". Frases corretas: "o processo permitiu
que...", "não havia guard-rail para...". Isso não é retórica — é o que faz as pessoas
reportarem incidentes sem medo, o que é o dado que você precisa para melhorar.

## Template de postmortem

```markdown
# Postmortem: <título curto do incidente>

**Data**: 2026-07-08
**Severidade**: SEV1/SEV2/SEV3
**Duração**: início HH:MM — fim HH:MM (X min)
**Autor**: <nome>

## Resumo

Uma ou duas frases: o que aconteceu, impacto, como foi resolvido.

## Impacto

- Quem/o que foi afetado (usuários, automações, dados)
- Quantidade (mensagens perdidas/atrasadas, requisições com erro, valor se aplicável)

## Timeline

| Hora | Evento |
| ---- | ------ |
| 14:02 | Alerta de taxa de erro disparado |
| 14:05 | Incidente aberto, canal criado |
| 14:20 | Causa identificada: throttle no serviço de pagamento |
| 14:35 | Mitigação aplicada (retry manual + circuit breaker) |
| 15:10 | Confirmado normalizado, incidente encerrado |

## Causa raiz

Resultado do 5 whys — a causa sistêmica, não o sintoma.

## O que funcionou

- O que ajudou a detectar/conter rápido (alerta específico, runbook, dashboard)

## O que faltou

- O que teria detectado/evitado mais rápido e não existia

## Action items

| Ação | Dono | Prazo |
| ---- | ---- | ----- |
| Adicionar retry com backoff no template de integração de pagamento | @nome | 2026-07-15 |
| Alerta de saturação na fila X | @nome | 2026-07-12 |
```

## Checklist de aceite

- [ ] Severidade classificada pelo impacto real, não pela gravidade técnica do erro
- [ ] Incidente teve canal dedicado e comunicação em intervalos regulares
- [ ] Causa raiz aplicou 5 whys até uma causa sistêmica, não parou no primeiro sintoma
- [ ] Postmortem é blameless — nenhuma frase aponta para uma pessoa
- [ ] Todo action item tem dono e prazo; nenhum item genérico sem responsável
