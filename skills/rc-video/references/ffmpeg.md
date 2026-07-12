# Processamento local com ffmpeg

Receitas testadas. Pré-requisito: `ffmpeg -version`. Todas rodam local, sem serviço externo.
Regra geral: **copiar streams (`-c copy`) quando não precisa recodificar** — é instantâneo e sem perda.

## Inspecionar

```bash
ffprobe -v error -show_format -show_streams input.mp4          # metadados completos
ffprobe -v error -select_streams v:0 -show_entries stream=width,height,r_frame_rate,codec_name -of csv=p=0 input.mp4
```

## Cortar / trim

```bash
# Sem recodificar (rápido, corta no keyframe mais próximo)
ffmpeg -ss 00:00:30 -to 00:01:00 -i input.mp4 -c copy out.mp4

# Preciso ao frame (recodifica; -ss depois do -i)
ffmpeg -i input.mp4 -ss 00:00:30 -to 00:01:00 -c:v libx264 -c:a aac out.mp4
```

## Transcodificar / comprimir para web

```bash
# H.264 (compatibilidade máxima). CRF menor = melhor qualidade/maior arquivo (18–28; 23 é padrão)
ffmpeg -i input.mov -c:v libx264 -crf 23 -preset medium -c:a aac -b:a 128k -movflags +faststart out.mp4

# H.265/HEVC (arquivos ~50% menores, menos compatível)
ffmpeg -i input.mp4 -c:v libx265 -crf 28 -c:a aac out.mp4
```
`-movflags +faststart` move o índice pro início → o vídeo começa a tocar antes de baixar tudo (essencial pra web).

## Redimensionar / mudar aspect ratio

```bash
ffmpeg -i input.mp4 -vf "scale=1280:-2" out.mp4                 # largura 1280, altura automática (par)

# Vídeo vertical 9:16 (Reels/Shorts) a partir de horizontal: crop central
ffmpeg -i input.mp4 -vf "crop=ih*9/16:ih,scale=1080:1920" -c:a copy vertical.mp4
```

## Extrair áudio / frames / thumbnail

```bash
ffmpeg -i input.mp4 -vn -c:a libmp3lame -q:a 2 audio.mp3        # só áudio (mp3)
ffmpeg -i input.mp4 -vn -c:a copy audio.m4a                     # áudio sem recodificar
ffmpeg -i input.mp4 -ss 00:00:05 -frames:v 1 thumb.jpg          # 1 thumbnail no segundo 5
ffmpeg -i input.mp4 -vf fps=1 frames/frame_%04d.png            # 1 frame por segundo
```

## Legendas

```bash
# Queimar (burn-in, fica gravado no vídeo)
ffmpeg -i input.mp4 -vf "subtitles=legenda.srt" -c:a copy out.mp4

# Embutir soft sub (usuário liga/desliga; só container mkv/mp4)
ffmpeg -i input.mp4 -i legenda.srt -c copy -c:s mov_text out.mp4
```

## Concatenar

```bash
# Mesmos codec/resolução: concat demuxer (rápido, sem recodificar)
printf "file '%s'\n" a.mp4 b.mp4 c.mp4 > list.txt
ffmpeg -f concat -safe 0 -i list.txt -c copy out.mp4

# Arquivos diferentes: recodifique para um formato comum antes de concatenar
```

## GIF de qualidade (com paleta)

```bash
ffmpeg -i input.mp4 -vf "fps=12,scale=480:-1:flags=lanczos,palettegen" palette.png
ffmpeg -i input.mp4 -i palette.png -vf "fps=12,scale=480:-1:flags=lanczos,paletteuse" out.gif
```

## HLS (streaming adaptativo simples)

```bash
ffmpeg -i input.mp4 -c:v libx264 -c:a aac -hls_time 6 -hls_playlist_type vod out.m3u8
```

## Notas

- Falhou por codec ausente: cheque o build (`ffmpeg -codecs | grep <codec>`); libx265/nvenc dependem do build instalado.
- Qualidade x tamanho: ajuste `-crf` (vídeo) e `-b:a` (áudio); `-preset slower` melhora compressão ao custo de tempo.
- Nunca sobrescreva o input com o mesmo nome do output.
