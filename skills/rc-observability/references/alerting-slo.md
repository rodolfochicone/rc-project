# Alertas e SLO

Objetivo: transformar métricas em sinais que alguém precisa agir sobre — nem mais, nem menos.
Ordem de prioridade: **defina o que "bom" significa (SLI/SLO) → derive o alerta do orçamento
que está sendo gasto → só depois monte o dashboard**. Um alerta sem SLO por trás é um chute
sobre o que é "normal".

## SLI, SLO e error budget

| Termo | Definição | Exemplo |
| ----- | --------- | ------- |
| **SLI** (indicator) | Métrica que mede a experiência real do usuário/consumidor | % de mensagens do SQS processadas com sucesso em até 5 min |
| **SLO** (objective) | Meta numérica para o SLI, num período | 99,5% das mensagens processadas em até 5 min, medido em 30 dias |
| **Error budget** | O quanto pode falhar sem violar o SLO | 0,5% de 30 dias = ~3h36min de "folga" para falhar |

Como derivar um SLO:

1. Escolha o SLI a partir do que o **consumidor** da automação sente — não o que é fácil de
   medir. Para uma Lambda disparada por EventBridge, "invocou" não é SLI; "processou o evento
   com sucesso dentro do SLA de negócio" é.
2. Olhe o histórico real (semanas, não um dia bom) antes de fixar o número. Um SLO mais frouxo
   que a realidade atual é inútil; mais rígido que a infraestrutura suporta gera alerta constante.
3. O error budget dita a política de risco: budget gasto → prioridade em estabilidade acima de
   feature nova; budget sobrando → espaço para mudança mais arriscada.

## Alertas acionáveis

Um alerta só existe se alguém for **fazer algo** ao recebê-lo. Regras:

- **Sintoma, não causa.** Alerte em "taxa de erro do consumer > X%" ou "DLQ recebendo mensagens",
  não em "CPU da task ECS > 80%" — CPU alta pode ser normal sob carga e não dizer nada sobre
  impacto real.
- **Todo alerta tem runbook.** Se a pessoa que recebe o alerta às 3h não sabe o próximo passo em
  30 segundos, falta um link para o runbook no próprio alerta (não "procure na wiki depois").
- **Baseado em orçamento, não em limiar arbitrário.** Prefira "queimando error budget rápido
  demais para bater o SLO do mês" a um número fixo escolhido no escuro.
- **Evite alert fatigue.** Todo alerta que dispara e é ignorado/silenciado repetidamente é
  sinal de que o limiar está errado ou o alerta não deveria existir — corrija ou apague, não
  acumule.
- **Severidade proporcional à ação exigida.** Página quem está de plantão só para o que não
  espera até o próximo dia útil; o resto vai para um canal assíncrono (ticket, Slack não urgente).

### Bom vs. mau alerta

| Mau alerta | Por quê | Melhor versão |
| ---------- | ------- | -------------- |
| "Lambda X invocada" | Não diz se algo está errado | "Taxa de erro da Lambda X > 2% em 5 min" |
| "Memória da task ECS em 85%" | Utilização alta pode ser normal | "Task ECS reiniciando por OOM (restart count > N em 10 min)" |
| "Fila SQS não está vazia" | Fila com mensagem é o estado normal | "Idade da mensagem mais antiga na fila > SLA de processamento" |
| "Erro 5xx apareceu no log" | Um erro isolado pode ser transiente | "Taxa de 5xx sustentada acima do budget por N minutos" |

## Dashboards úteis

- **RED por serviço**: um painel por serviço/Lambda/consumer com rate, error rate e latência
  (p50/p90/p99) lado a lado — é o primeiro lugar que se olha ao investigar "algo está lento/quebrado".
- **Visão de frota**: para quem opera muitas automações (várias Lambdas/ECS/filas), um painel
  agregado que mostra qual serviço está fora do SLO agora, ordenado por severidade — não uma
  lista de 40 gráficos individuais que ninguém rola até o fim.
- **Saturação de recursos compartilhados**: profundidade de fila, connection pool, concurrency
  limit de Lambda — o que satura primeiro quando o tráfego sobe.

Dashboard não substitui alerta: dashboard é para investigação ativa; alerta é para avisar sem
que alguém precise estar olhando.

## Checklist de aceite

- [ ] SLI escolhido reflete a experiência de quem consome o serviço, não uma métrica de infra por conveniência
- [ ] SLO tem número e período definidos, calibrado pelo histórico real
- [ ] Todo alerta é sintoma (impacto ao usuário/consumidor), não causa interna
- [ ] Todo alerta crítico tem runbook linkado; nenhum alerta sem próxima ação clara
- [ ] Sem alertas conhecidos por serem ignorados/silenciados de forma recorrente
- [ ] Existe um dashboard RED por serviço e uma visão agregada de frota (se houver mais de um serviço)
