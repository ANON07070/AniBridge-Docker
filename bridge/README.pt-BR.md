<p align="center">
  <b>🇧🇷 Português</b> ·
  <a href="README.md">🇺🇸 English</a>
</p>

# AniBridge — Go Bridge

Adaptador em Go para o [`pkg/goanime`](https://github.com/alvarorichard/Goanime).
Expõe uma interface JSON via CLI, consumida pelo backend Python
(FastAPI) através de subprocess.

## Arquitetura

```
FastAPI (backend)
    ↓
chamada subprocess (JSON in/out)
    ↓
Go Bridge (este programa)
    ↓
pkg/goanime
```

O contrato de envelope JSON (`ok`/`data`/`error`) foi escolhido para
que esse transporte possa futuramente virar um servidor HTTP sem
alterar nada do lado Python — o backend só chama
`run_bridge_command()` e nunca precisa saber que por trás existe um
subprocess.

## Pré-requisitos

- **Go 1.26.5+** — o `pkg/goanime` exige essa versão.
  - Verificar com: `go version`
  - Se estiver desatualizado, atualizar em https://go.dev/dl

## Compilando

### Windows

```powershell
cd anibridge\bridge
go mod tidy
go build -o goanime-bridge.exe .
```

### Linux/macOS

```bash
cd anibridge/bridge
go mod tidy
go build -o goanime-bridge .
chmod +x goanime-bridge
```

Cross-compilando um binário Linux a partir do Windows (útil quando a
máquina de destino não tem RAM suficiente pra compilar a árvore de
dependências do `pkg/goanime` sozinha):

```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o goanime-bridge-linux-amd64 .
```

## Comandos

### `search <query> [--source SOURCE]`

Busca em todas as fontes, ou numa fonte específica se `--source` for
passado. `--source` aceita o que o `pkg/goanime` expõe publicamente:
`AllAnime` ou `AnimeFire` (também aceita o valor bruto retornado por
uma busca anterior, ex: `Animefire.io` — normalizado internamente).

```bash
./goanime-bridge search "Naruto"
./goanime-bridge search "Naruto" --source AllAnime
```

```json
{
  "ok": true,
  "data": [
    {
      "name": "Naruto",
      "url": "...",
      "imageUrl": "...",
      "source": "AllAnime",
      "anilistId": 20,
      "malId": 457,
      "details": {
        "id": 20,
        "description": "...",
        "genres": ["Action", "Shounen"],
        "averageScore": 80,
        "episodes": 220,
        "status": "FINISHED"
      }
    }
  ]
}
```

### `episodes <anime-url> --source SOURCE` (obrigatório)

Lista os episódios de um anime. Só funciona pra fontes com valor
público em `types.Source` (`AllAnime`, `AnimeFire`) — Goyabu e
SuperFlix aparecem na busca mas não são suportados aqui no
`pkg/goanime` v1.8.6.

```bash
./goanime-bridge episodes "https://animefire.io/animes/naruto..." --source Animefire.io
```

```json
{
  "ok": true,
  "data": [
    {
      "number": "1",
      "num": 1,
      "url": "...",
      "aired": "",
      "duration": 0,
      "isFiller": false,
      "isRecap": false,
      "synopsis": "",
      "title": { "romaji": "...", "english": "...", "japanese": "..." }
    }
  ]
}
```

### `stream <anime-url> --source SOURCE [--episode-url URL] [--episode-number N]`

Resolve a URL de streaming de um episódio. `--episode-number` é usado
pro AllAnime, `--episode-url` pro AnimeFire (bate com o que
`GetEpisodeStreamURL` de fato lê internamente pra cada fonte).

Pra AnimeFire, isso também resolve um hop intermediário extra que a
própria biblioteca não trata: a URL que o `pkg/goanime` retorna pra
essa fonte não é o arquivo de vídeo final — é um endpoint que devolve
um JSON pequeno com opções de qualidade
(`{"data":[{"src","label"}]}`), usado pelo próprio JS do frontend do
AnimeFire. O bridge busca esse endpoint, escolhe a melhor qualidade
disponível, e retorna o link `.mp4` real.

```bash
./goanime-bridge stream "ANIME_ID" --source AllAnime --episode-number 1
./goanime-bridge stream "IGNORADO" --source Animefire.io --episode-url "https://animefire.io/video/..."
```

```json
{
  "ok": true,
  "data": {
    "streamUrl": "https://.../720p.mp4?token=...",
    "metadata": { "source": "animefire", "resolvedQuality": "720p" }
  }
}
```

## Envelope JSON

Sucesso:
```json
{ "ok": true, "data": ... }
```

Erro:
```json
{ "ok": false, "error": "descrição" }
```

## Detalhes de implementação relevantes

- **Parsing de `--source` é manual**, não usa `flag.FlagSet`: o pacote
  `flag` do Go para de interpretar flags no primeiro argumento
  posicional, então `search "query" --source X` ignorava a flag
  silenciosamente. Corrigido com parsing manual, independente de
  posição.
- **`enetx/http` fixado em v1.0.29**: a v1.0.26 do `enetx/http2`
  (dependência do `pkg/goanime`) tem um arquivo com build tag
  `//go:build go1.27` que referencia um campo só adicionado no
  `enetx/http` a partir da v1.0.29. O `pkg/goanime` v1.8.6 fixa a
  v1.0.28, o que quebra a build em Go 1.27+.
- **`CGO_ENABLED=0`**: o `pkg/goanime` arrasta transitivamente o
  sistema de histórico do CLI original baseado em SQLite
  (`internal/tracking`, via `go-sqlite3`, cgo), mesmo esse projeto
  nunca usando isso. Compilar essa dependência já causou falhas de
  memória em máquinas com pouca RAM. Desligar o cgo exclui isso da
  build por completo.
