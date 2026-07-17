# VideoDB — integração opcional (SaaS pago)

VideoDB é um serviço de "percepção, memória e ação" para vídeo/áudio: ingestão, indexação
(fala, cena, semântica), busca com timestamps, edição por timeline e geração de assets.

> **Aviso de lock-in.** Depende de API paga e SDK Python. Antes de usar, confirme que o objetivo
> **não** é alcançável com ffmpeg local (`ffmpeg.md`) — que resolve corte, transcode, legenda e
> extração sem custo nem dependência externa. Só recorra ao VideoDB quando precisar do que ele
> tem de único: **busca semântica dentro do vídeo** e **indexação de fala/cena**.

## Quando faz sentido

| Precisa de | ffmpeg resolve? | Use VideoDB? |
| ---------- | --------------- | ------------ |
| Cortar/transcodificar/legendar/extrair | Sim | Não |
| Buscar "onde alguém fala sobre X" no vídeo | Não | Sim |
| Indexar cenas/fala e gerar clips por busca | Não | Sim |
| Edição por timeline gerenciada na nuvem | Parcial (local) | Talvez |

## Setup

```bash
pip install videodb
export VIDEO_DB_API_KEY="..."   # obtida em console.videodb.io
```

Sem a chave, a integração não roda — isso não é bug; é pré-requisito do SaaS.

## Operações centrais (referência de API)

```python
import videodb

conn = videodb.connect()                      # usa VIDEO_DB_API_KEY do ambiente
coll = conn.get_collection()

# Ingest: arquivo local, URL pública ou YouTube
video = coll.upload(url="https://www.youtube.com/watch?v=...")

# Indexar para busca
video.index_spoken_words()                    # transcrição / fala
video.index_scenes()                          # cenas visuais

# Buscar → resultados com timestamps + clip navegável
result = video.search("trecho sobre performance")
for shot in result.get_shots():
    print(shot.start, shot.end, shot.text)
stream_url = result.play()                    # link de stream do trecho

# Stream do vídeo completo
print(video.generate_stream())
```

A API evolui rápido. **Não confie de memória** — antes de escrever código real, confirme a
assinatura atual via o MCP Context7 ou a doc oficial (docs.videodb.io).

## Checklist antes de adotar

- [ ] O objetivo exige busca semântica/indexação (não é só corte/transcode que o ffmpeg faz)
- [ ] Custo do SaaS e dependência externa foram aceitos explicitamente
- [ ] `VIDEO_DB_API_KEY` provisionada fora do código (env/secret, nunca commitada)
- [ ] Assinaturas de API confirmadas na doc atual, não presumidas
