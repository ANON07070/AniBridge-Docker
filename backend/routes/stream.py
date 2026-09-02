"""
Rotas de streaming de um episódio.

- GET /api/stream: chama run_bridge_command() (services/bridge.py) e
  traduz o envelope do bridge para respostas HTTP.
- GET /api/stream/proxy: repassa os bytes do vídeo em si, buscando no
  CDN a partir do servidor em vez de deixar o navegador buscar direto.
"""

import httpx
from fastapi import APIRouter, HTTPException, Query, Request
from fastapi.responses import StreamingResponse

from services.bridge import run_bridge_command

router = APIRouter()

# Mesmo User-Agent hardcoded que o CLI original do GoAnime usa
# especificamente porque o CDN do AnimeFire (lightspeedst.net) rejeita
# com 401 o User-Agent padrão de clientes HTTP como Go/httpx (verificado
# em internal/player/download.go, constante downloadUserAgent).
PROXY_USER_AGENT = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

# Headers fixos conhecidos por fonte, usados como fallback só quando o
# metadata retornado por /api/stream não já trouxer um "referer"
# dinamicamente (algumas fontes, como o AllAnime em streams do tipo
# "direct", expõem isso via metadata; outras, como o AnimeFire, exigem
# um valor fixo conhecido só por leitura do código-fonte do GoAnime).
#
# Hoje só AnimeFire está aqui — é a única fonte com player implementado.
# Fontes sem stream funcionando ainda (AllAnime, Goyabu, SuperFlix) não
# entram nessa tabela até que haja pedido explícito pra trabalhar nelas.
FIXED_SOURCE_REFERER = {
    "animefire": "https://animefire.io/",
    "animefire.io": "https://animefire.io/",
}


def resolve_proxy_headers(source: str, referer: str) -> dict:
    """
    Monta os headers usados para buscar o vídeo no CDN.

    Prioridade: um "referer" explícito (vindo do metadata que
    /api/stream já retornou, quando a própria API informa isso
    dinamicamente) sobre a tabela fixa por fonte. Isso evita manter
    conhecimento hardcoded duplicado para fontes que já se
    autodescrevem via metadata.
    """
    headers = {"User-Agent": PROXY_USER_AGENT}

    resolved_referer = referer or FIXED_SOURCE_REFERER.get((source or "").lower())
    if resolved_referer:
        headers["Referer"] = resolved_referer

    return headers

# Sem timeout de leitura: um stream de vídeo pode demorar mais que o
# timeout padrão do httpx (5s) só pra começar a responder, e a conexão
# fica aberta durante toda a reprodução/seek.
PROXY_TIMEOUT = httpx.Timeout(connect=15.0, read=None, write=None, pool=None)


@router.get("/api/stream")
def stream(
    animeUrl: str = Query(
        ..., min_length=1, description="URL do anime, retornada por /api/search"
    ),
    source: str = Query(
        ...,
        min_length=1,
        description="Fonte do anime, exatamente como retornada por /api/search",
    ),
    episodeUrl: str = Query(
        "",
        description=(
            "URL do episódio, retornada por /api/episodes. "
            "Necessária para AnimeFire."
        ),
    ),
    episodeNumber: str = Query(
        "",
        description=(
            "Número do episódio, retornado por /api/episodes. "
            "Necessário para AllAnime."
        ),
    ),
):
    """
    Obtém a URL de streaming de um episódio via Go Bridge.

    - Sucesso: retorna {"streamUrl": ..., "metadata": {...}}.
    - Falha: HTTP 400 com a mensagem de erro do bridge (fonte não
      suportada, episódio/anime não encontrado, etc).
    """
    args = [animeUrl, "--source", source]
    if episodeUrl:
        args.extend(["--episode-url", episodeUrl])
    if episodeNumber:
        args.extend(["--episode-number", episodeNumber])

    result = run_bridge_command("stream", *args)

    if not result.get("ok"):
        raise HTTPException(
            status_code=400,
            detail=result.get("error", "erro desconhecido do bridge"),
        )

    return result.get("data", {})


@router.get("/api/stream/proxy")
async def stream_proxy(
    request: Request,
    url: str = Query(..., min_length=1),
    source: str = Query(
        "",
        description=(
            "Fonte do anime (ex: Animefire.io). Usada para decidir os "
            "headers padrão quando 'referer' não é informado."
        ),
    ),
    referer: str = Query(
        "",
        description=(
            "Referer a usar na requisição ao CDN, quando conhecido "
            "dinamicamente (ex: vindo do metadata de /api/stream). "
            "Se ausente, cai para uma tabela fixa baseada em 'source'."
        ),
    ),
):
    """
    Faz proxy dos bytes de um stream de vídeo, em vez do navegador
    buscar direto no CDN.

    Necessário porque alguns CDNs (ex: lightspeedst.net, usado pelo
    AnimeFire) exigem um Referer específico para autorizar requisições
    com token assinado — sem ele, respondem 401 mesmo com o token
    correto (verificado em internal/player/scraper.go do GoAnime
    original). Buscando aqui no servidor, com o Referer certo, evitamos
    que o navegador precise (e não consiga) fingir essa origem.

    Repassa o header Range (essencial pro <video> do navegador
    conseguir buscar/avançar no vídeo) e espelha status/headers
    relevantes da resposta do CDN (incluindo 206 Partial Content).
    """
    forward_headers = resolve_proxy_headers(source, referer)
    range_header = request.headers.get("range")
    if range_header:
        forward_headers["Range"] = range_header

    client = httpx.AsyncClient(follow_redirects=True, timeout=PROXY_TIMEOUT)

    try:
        upstream_request = client.build_request("GET", url, headers=forward_headers)
        upstream = await client.send(upstream_request, stream=True)
    except httpx.HTTPError as e:
        await client.aclose()
        raise HTTPException(
            status_code=502, detail=f"falha ao conectar no stream: {e}"
        )

    if upstream.status_code >= 400:
        await upstream.aclose()
        await client.aclose()
        raise HTTPException(
            status_code=upstream.status_code,
            detail=f"upstream retornou {upstream.status_code} ao buscar o stream",
        )

    async def body_iterator():
        try:
            async for chunk in upstream.aiter_bytes():
                yield chunk
        finally:
            await upstream.aclose()
            await client.aclose()

    response_headers = {}
    for header_name in ("content-type", "content-length", "content-range"):
        if header_name in upstream.headers:
            response_headers[header_name] = upstream.headers[header_name]
    # Sempre sinaliza suporte a range pro navegador, mesmo se por algum
    # motivo o upstream não mandar esse header.
    response_headers["accept-ranges"] = "bytes"

    return StreamingResponse(
        body_iterator(),
        status_code=upstream.status_code,
        headers=response_headers,
    )
