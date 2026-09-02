"""
Aplicação FastAPI - AniBridge Backend.

Etapa 3: expõe GET /api/search.
Etapa 4: serve o frontend estático (index.html, css, js) na mesma
origem, para o fetch() do navegador não esbarrar em CORS.
"""

import os

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles

from routes.episodes import router as episodes_router
from routes.search import router as search_router
from routes.stream import router as stream_router

app = FastAPI(title="AniBridge Backend")

app.include_router(search_router)
app.include_router(episodes_router)
app.include_router(stream_router)

# Precisa vir DEPOIS do include_router: o mount em "/" captura qualquer
# path não reconhecido antes, então se estivesse antes do router,
# /api/search seria interceptado pelo StaticFiles em vez de chegar na rota.
FRONTEND_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "frontend")
app.mount("/", StaticFiles(directory=FRONTEND_DIR, html=True), name="frontend")
