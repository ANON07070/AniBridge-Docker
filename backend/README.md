<p align="center">
  <a href="README.pt-BR.md">🇧🇷 Português</a> ·
  <b>🇺🇸 English</b>
</p>

# AniBridge — Backend

FastAPI application: exposes the HTTP API, serves the static frontend,
and talks to the Go Bridge through a subprocess.

## Structure

```
backend/
├── main.py                  # FastAPI app, mounts routers + static frontend
├── requirements.txt
├── routes/
│   ├── search.py            # GET /api/search
│   ├── episodes.py          # GET /api/episodes
│   └── stream.py            # GET /api/stream, GET /api/stream/proxy
└── services/
    └── bridge.py            # subprocess runner for the Go Bridge
```

## Configuration

`BRIDGE_PATH` (required, unless running via Docker): absolute path to
the compiled Go Bridge binary.

```bash
export BRIDGE_PATH=/path/to/bridge/goanime-bridge   # Linux/macOS
$env:BRIDGE_PATH = "C:\path\to\bridge\goanime-bridge.exe"  # Windows
```

When running via the Docker image, this is already set inside the
container — no manual configuration needed.

## Endpoints

### `GET /api/search?q=<query>&source=<optional>`

Calls the bridge's `search` command. `source` is optional (all sources
are searched when omitted).

- `200`: array of anime results.
- `400`: bridge error (e.g. invalid source).
- `422`: missing/empty `q` (validated by FastAPI itself, before ever
  calling the bridge).

### `GET /api/episodes?url=<anime-url>&source=<source>`

Both parameters are required — unlike search, `GetAnimeEpisodes()`
needs a concrete source, it can't search across all of them.

- `200`: array of episodes.
- `400`: bridge error, including sources without episode support
  (Goyabu, SuperFlix).

### `GET /api/stream?animeUrl=...&source=...&episodeUrl=...&episodeNumber=...`

Resolves the streaming URL for an episode. `episodeUrl` is required
for AnimeFire, `episodeNumber` for AllAnime.

- `200`: `{"streamUrl": ..., "metadata": {...}}`.
- `400`: bridge error.

### `GET /api/stream/proxy?url=...&source=...&referer=...`

Proxies the video bytes from the CDN, instead of letting the browser
fetch it directly.

**Why this exists**: some CDNs (e.g. `lightspeedst.net`, used by
AnimeFire) require a specific `Referer` header to authorize
token-signed requests — confirmed in the original GoAnime CLI's source
(`internal/player/scraper.go`): *"lightspeedst.net (AnimeFire CDN)
requires Referer: https://animefire.io to authorise token-signed
requests. Without it, mpv/yt-dlp get HTTP 401 Unauthorized while the
browser (which sends the referer automatically) plays the same URL
fine."* A browser's own `<video src>` can't forge a cross-origin
Referer, so this proxy sends the correct one server-side.

**Header resolution** (`resolve_proxy_headers` in `routes/stream.py`):
prioritizes an explicit `referer` param (when the bridge's `metadata`
already provides one dynamically — e.g. AllAnime's `"direct"` stream
type does) over a small fixed table keyed by `source`. Today that
table only has an entry for AnimeFire; other sources aren't added
until there's a concrete reason to.

Supports `Range` requests (essential for seeking/buffering) and
mirrors the upstream's status code (including `206 Partial Content`)
and relevant headers.

## Running locally (without Docker)

```bash
pip install -r requirements.txt
export BRIDGE_PATH=/path/to/goanime-bridge
uvicorn main:app --reload --port 8000
```

## Design notes

- **Transport-layer isolation**: `services/bridge.py` is the only file
  that knows the Go Bridge is invoked via subprocess. Everything else
  just calls `run_bridge_command()` and gets back a dict — this could
  be swapped for an HTTP call later without touching route code.
- **Explicit UTF-8 decoding**: `subprocess.run(..., text=True)` without
  an explicit `encoding` decodes stdout/stderr using the OS's default
  encoding, which on Windows is typically `cp1252` ("charmap"), not
  UTF-8. Since the Go Bridge always writes UTF-8 JSON, any title with a
  multi-byte character (accents, native Japanese names, em dashes)
  crashed with `UnicodeDecodeError` on Windows. Fixed with an explicit
  `encoding="utf-8"`.
