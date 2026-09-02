"""
Rota de busca de anime.

Chama run_bridge_command() (services/bridge.py) e traduz o envelope
{"ok": ..., "data"/"error": ...} do bridge para respostas HTTP.
"""

from typing import Optional

from fastapi import APIRouter, HTTPException, Query

from services.bridge import run_bridge_command

router = APIRouter()


@router.get("/api/search")
def search(
    q: str = Query(..., min_length=1, description="Termo de busca do anime"),
    source: Optional[str] = Query(
        None, description="Fonte específica (ex: AllAnime, AnimeFire)"
    ),
):
    """
    Executa uma busca de anime via Go Bridge.

    - Sucesso: retorna a lista de resultados (result["data"]).
    - Falha: retorna HTTP 400 com a mensagem de erro do bridge, seja ela
      originada de uma source inválida, erro de scraping, ou falha de
      transporte (bridge não encontrado, timeout, etc). Diferenciar
      esses casos com status codes distintos fica para uma etapa futura,
      se necessário.
    """
    args = [q]
    if source:
        args.extend(["--source", source])

    result = run_bridge_command("search", *args)

    if not result.get("ok"):
        raise HTTPException(
            status_code=400,
            detail=result.get("error", "erro desconhecido do bridge"),
        )

    return result.get("data", [])
