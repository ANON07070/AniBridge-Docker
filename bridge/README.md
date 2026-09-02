<p align="center">
  <a href="README.pt-BR.md">🇧🇷 Português</a> ·
  <b>🇺🇸 English</b>
</p>

# AniBridge — Go Bridge

Go adapter for [`pkg/goanime`](https://github.com/alvarorichard/Goanime).
Exposes a JSON interface over a CLI, consumed by the Python backend
(FastAPI) via subprocess.

## Architecture

```
FastAPI (backend)
    ↓
subprocess call (JSON in/out)
    ↓
Go Bridge (this program)
    ↓
pkg/goanime
```

The JSON envelope contract (`ok`/`data`/`error`) was chosen so this
transport could later be swapped for an HTTP server without changing
anything on the Python side — the backend only calls
`run_bridge_command()` and never needs to know it's using a subprocess.

## Requirements

- **Go 1.26.5+** — `pkg/goanime` requires it.
  - Check with: `go version`
  - If older, update at https://go.dev/dl

## Building

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

Cross-compiling a Linux binary from Windows (useful when the target
machine has too little RAM to compile `pkg/goanime`'s dependency tree
itself):

```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o goanime-bridge-linux-amd64 .
```

## Commands

### `search <query> [--source SOURCE]`

Searches across all sources, or a specific one when `--source` is
given. `--source` accepts what `pkg/goanime` exposes publicly:
`AllAnime` or `AnimeFire` (also matches the raw value returned by a
previous search, e.g. `Animefire.io` — normalized internally).

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

### `episodes <anime-url> --source SOURCE` (required)

Lists episodes for an anime. Only works for sources with a public
`types.Source` value (`AllAnime`, `AnimeFire`) — Goyabu and SuperFlix
show up in search but aren't supported here in `pkg/goanime` v1.8.6.

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

Resolves the streaming URL for an episode. `--episode-number` is used
for AllAnime, `--episode-url` for AnimeFire (matches what
`GetEpisodeStreamURL` actually reads internally per source).

For AnimeFire, this also resolves an extra intermediate hop the
library itself doesn't handle: the URL `pkg/goanime` returns for this
source isn't the final video file — it's an endpoint that returns a
small JSON with quality options (`{"data":[{"src","label"}]}`), used
by AnimeFire's own frontend JS. The bridge fetches that endpoint,
picks the best available quality, and returns the real `.mp4` link.

```bash
./goanime-bridge stream "ANIME_ID" --source AllAnime --episode-number 1
./goanime-bridge stream "IGNORED" --source Animefire.io --episode-url "https://animefire.io/video/..."
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

## JSON envelope

Success:
```json
{ "ok": true, "data": ... }
```

Error:
```json
{ "ok": false, "error": "description" }
```

## Notable implementation details

- **`--source` parsing is manual**, not `flag.FlagSet`: Go's `flag`
  package stops parsing at the first positional argument, so
  `search "query" --source X` silently ignored the flag. Fixed by
  parsing arguments manually, position-independent.
- **`enetx/http` pinned to v1.0.29**: v1.0.26 of `enetx/http2` (a
  `pkg/goanime` dependency) has a file gated by `//go:build go1.27`
  referencing a field only added in `enetx/http` v1.0.29. `pkg/goanime`
  v1.8.6 pins v1.0.28, which breaks the build under Go 1.27+.
- **`CGO_ENABLED=0`**: `pkg/goanime` transitively pulls in the original
  CLI's SQLite-backed history feature (`internal/tracking`, via
  `go-sqlite3`, cgo), even though this project never uses it. Building
  that dependency has caused out-of-memory failures on machines with
  limited RAM. Disabling cgo excludes it from the build entirely.
