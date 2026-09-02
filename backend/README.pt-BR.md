<p align="center">
  <b>🇧🇷 Português</b> ·
  <a href="README.md">🇺🇸 English</a>
</p>

# AniBridge — Backend

Aplicação FastAPI: expõe a API HTTP, serve o frontend estático, e
conversa com o Go Bridge via subprocess.

## Estrutura

```
backend/
├── main.py                  # app FastAPI, monta routers + frontend estático
├── requirements.txt
├── routes/
│   ├── search.py            # GET /api/search
│   ├── episodes.py          # GET /api/episodes
│   └── stream.py            # GET /api/stream, GET /api/stream/proxy
└── services/
    └── bridge.py            # subprocess runner pro Go Bridge
```

## Configuração

`BRIDGE_PATH` (obrigatória, exceto rodando via Docker): caminho
absoluto pro binário compilado do Go Bridge.

```bash
export BRIDGE_PATH=/caminho/para/bridge/goanime-bridge   # Linux/macOS
$env:BRIDGE_PATH = "C:\caminho\para\bridge\goanime-bridge.exe"  # Windows
```

Rodando via imagem Docker, isso já vem definido dentro do container —
sem configuração manual necessária.

## Endpoints

### `GET /api/search?q=<query>&source=<opcional>`

Chama o comando `search` do bridge. `source` é opcional (busca em
todas as fontes se omitido).

- `200`: array de resultados de anime.
- `400`: erro do bridge (ex: source inválida).
- `422`: `q` ausente/vazio (validado pelo próprio FastAPI, antes de
  sequer chamar o bridge).

### `GET /api/episodes?url=<anime-url>&source=<source>`

Os dois parâmetros são obrigatórios — diferente da busca,
`GetAnimeEpisodes()` exige uma fonte concreta, não dá pra buscar em
todas de uma vez.

- `200`: array de episódios.
- `400`: erro do bridge, incluindo fontes sem suporte a episódios
  (Goyabu, SuperFlix).

### `GET /api/stream?animeUrl=...&source=...&episodeUrl=...&episodeNumber=...`

Resolve a URL de streaming de um episódio. `episodeUrl` é obrigatório
pra AnimeFire, `episodeNumber` pra AllAnime.

- `200`: `{"streamUrl": ..., "metadata": {...}}`.
- `400`: erro do bridge.

### `GET /api/stream/proxy?url=...&source=...&referer=...`

Faz proxy dos bytes do vídeo vindos do CDN, em vez de deixar o
navegador buscar direto.

**Por que isso existe**: alguns CDNs (ex: `lightspeedst.net`, usado
pelo AnimeFire) exigem um `Referer` específico pra autorizar
requisições com token assinado — confirmado no código-fonte do CLI
original do GoAnime (`internal/player/scraper.go`): *"lightspeedst.net
(AnimeFire CDN) requires Referer: https://animefire.io to authorise
token-signed requests. Without it, mpv/yt-dlp get HTTP 401 Unauthorized
while the browser (which sends the referer automatically) plays the
same URL fine."* Um `<video src>` do navegador não consegue forjar um
Referer de origem cruzada sozinho, então esse proxy manda o certo pelo
lado do servidor.

**Resolução de headers** (`resolve_proxy_headers` em
`routes/stream.py`): prioriza um parâmetro `referer` explícito (quando
o `metadata` do bridge já traz um dinamicamente — ex: o tipo de stream
`"direct"` do AllAnime traz) sobre uma tabela fixa pequena indexada por
`source`. Hoje essa tabela só tem entrada pro AnimeFire; outras fontes
não entram até que haja motivo concreto.

Suporta requisições `Range` (essencial pra seek/buffering) e espelha o
status code do upstream (incluindo `206 Partial Content`) e os headers
relevantes.

## Rodando localmente (sem Docker)

```bash
pip install -r requirements.txt
export BRIDGE_PATH=/caminho/para/goanime-bridge
uvicorn main:app --reload --port 8000
```

## Notas de design

- **Isolação da camada de transporte**: `services/bridge.py` é o único
  arquivo que sabe que o Go Bridge é executado via subprocess. Todo o
  resto só chama `run_bridge_command()` e recebe um dict de volta —
  isso poderia virar uma chamada HTTP depois sem tocar no código das
  rotas.
- **Decodificação UTF-8 explícita**: `subprocess.run(..., text=True)`
  sem `encoding` explícito decodifica stdout/stderr usando o encoding
  padrão do sistema operacional, que no Windows normalmente é `cp1252`
  ("charmap"), não UTF-8. Como o Go Bridge sempre escreve JSON em
  UTF-8, qualquer título com caractere multi-byte (acentos, nomes
  nativos em japonês, travessões) quebrava com `UnicodeDecodeError` no
  Windows. Corrigido com `encoding="utf-8"` explícito.
