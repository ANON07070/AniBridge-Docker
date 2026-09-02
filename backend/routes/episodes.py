"""
Rota de listagem de episódios de um anime.

Chama run_bridge_command() (services/bridge.py) e traduz o envelope do
bridge para respostas HTTP.
"""

from fastapi import APIRouter, HTTPException, Query

from services.bridge import run_bridge_command

router = APIRouter()


@router.get("/api/episodes")
def episodes(
    url: str = Query(
        ..., min_length=1, description="URL do anime, retornada por /api/search"
    ),
    source: str = Query(
        ...,
        min_length=1,
        description=(
            "Fonte do anime, exatamente como retornada por /api/search "
            "(ex: AllAnime, Animefire.io). Diferente de /api/search, aqui "
            "é obrigatória: GetAnimeEpisodes() da biblioteca exige uma "
            "fonte concreta."
        ),
    ),
):
    """
    Lista os episódios de um anime via Go Bridge.

    - Sucesso: retorna a lista de episódios (result["data"]).
    - Falha: HTTP 400 com a mensagem de erro do bridge. Isso inclui o
      caso de a fonte não ser suportada para episódios (Goyabu e
      SuperFlix não têm suporte a isso no pkg/goanime v1.8.6 — ver
      commands.Episodes no bridge para detalhes).
    """
    result = run_bridge_command("episodes", url, "--source", source)

    if not result.get("ok"):
        raise HTTPException(
            status_code=400,
            detail=result.get("error", "erro desconhecido do bridge"),
        )

    return result.get("data", [])
