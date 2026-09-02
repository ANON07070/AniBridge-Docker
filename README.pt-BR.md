<p align="center">
  <img src="docs/logo.png" alt="Logo do AniBridge" width="480">
</p>

<p align="center">
  <b>🇧🇷 Português</b> ·
  <a href="README.md">🇺🇸 English</a>
</p>

# AniBridge

[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Python](https://img.shields.io/badge/Python-3.10%2B-3776AB?style=flat&logo=python&logoColor=white)](https://www.python.org/)
[![FastAPI](https://img.shields.io/badge/FastAPI-009688?style=flat&logo=fastapi&logoColor=white)](https://fastapi.tiangolo.com/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=flat&logo=docker&logoColor=white)](https://www.docker.com/)

Aplicação web self-hosted open-source para assistir anime através do navegador.
Busca, listagem de episódios e reprodução — sem precisar instalar nada
no cliente além de um navegador moderno.

## Arquitetura

```
Navegador → Frontend (HTML/CSS/JS) → FastAPI → Go Bridge → pkg/goanime
```

- **Frontend**: estático, servido pelo próprio FastAPI (sem CORS a
  configurar).
- **Backend**: FastAPI, expõe a API e chama o Go Bridge via subprocess.
- **Go Bridge**: adapta o [`pkg/goanime`](https://github.com/alvarorichard/Goanime)
  (biblioteca de scraping em Go) pra um CLI que fala JSON, já que o
  backend é em Python.

Mais detalhes de cada camada: [`bridge/README.pt-BR.md`](bridge/README.pt-BR.md),
[`backend/README.pt-BR.md`](backend/README.pt-BR.md),
[`frontend/README.pt-BR.md`](frontend/README.pt-BR.md).

## Status atual

Fluxo completo funcionando pra **AnimeFire**: busca → lista de
episódios → reprodução no navegador, incluindo seleção automática da
melhor qualidade e acesso remoto (não só localhost).

**Limitações conhecidas:**
- **Goyabu e SuperFlix**: aparecem na busca, mas não têm suporte a
  episódios/stream — `pkg/goanime` v1.8.6 não expõe esses valores em
  `types.Source`, então não há como buscar episódios pra eles nesta
  versão.
- **AllAnime**: busca e episódios funcionam, mas obter a URL de stream
  falha atualmente (timeout ao acessar uma página de referer
  necessária, `mkissa.to`). Sem prioridade no momento.
- **AnimeFire**: funcionando de ponta a ponta, incluindo uma camada de
  proxy que busca os bytes do vídeo pelo servidor (alguns CDNs exigem
  um `Referer` específico que o navegador não consegue forjar sozinho).

Projeto pausado no momento para fase de beta test.

## Como rodar

### Docker — imagem pronta (recomendado, sem compilar nada)

A imagem é compilada automaticamente pelo GitHub Actions, então não
importa se sua máquina tem pouca RAM — nada é compilado localmente
(compilar o Go Bridge localmente já travou builds em máquinas com ~4GB
de RAM):

```bash
docker pull ghcr.io/ANON07070/anibridge:latest
docker run -p 8000:8000 ghcr.io/ANON07070/anibridge:latest
```


### Docker — buildar localmente

```bash
docker compose up --build
```

**Atenção**: isso compila o Go Bridge na sua própria máquina. Se sua
máquina tiver pouca RAM, prefira a imagem pronta acima.

### Windows, sem Docker

```powershell
cd bridge
go build -o anibridge.exe .

cd ..\backend
$env:BRIDGE_PATH = "CAMINHO_COMPLETO_PARA\bridge\anibridge.exe"
pip install -r requirements.txt
uvicorn main:app --reload --port 8000
```

### Linux, sem Docker

```bash
./run.sh
```

O script compila o bridge, cria um venv Python, instala as
dependências e sobe o servidor. Aceita `HOST`/`PORT` opcionais:

```bash
PORT=9000 ./run.sh
```

Requer Go 1.26.5+ e Python 3.10+ já instalados na máquina.

### Acessando

Em todos os casos acima: `http://localhost:8000`. Pra acessar de outro
dispositivo na mesma rede, suba com `--host 0.0.0.0` (Windows) ou
`HOST=0.0.0.0 ./run.sh` (Linux), e acesse pelo IP local da máquina.

## Regras do projeto

- Usar `pkg/goanime` diretamente, sem reescrever lógica de scraping.
- Verificar o código-fonte da biblioteca antes de assumir qualquer
  comportamento — a documentação dela já se mostrou desatualizada mais
  de uma vez.
- Escopo incremental: cada funcionalidade nova é validada antes de
  seguir pra próxima.
- Nada de Redis, PostgreSQL ou autenticação por enquanto —
  simplicidade primeiro.
- Depois de publicar uma nova imagem: o pacote no GHCR nasce
  **privado** por padrão, mesmo com o repositório público — precisa
  marcar como público manualmente em Settings → Packages na primeira
  vez, senão `docker pull` de outras pessoas falha com erro de
  autenticação.

## Estrutura do projeto

```
anibridge/
├── .github/workflows/  # CI: builda e publica a imagem Docker no GHCR
├── bridge/      # Go: adapta pkg/goanime pra JSON via CLI
├── backend/     # FastAPI: API + serve o frontend
├── frontend/    # HTML/CSS/JS estático
├── Dockerfile
├── docker-compose.yml
└── run.sh       # sobe tudo de uma vez no Linux, sem Docker
```

## Créditos

Construído em cima do [`pkg/goanime`](https://github.com/alvarorichard/Goanime),
de [alvarorichard](https://github.com/alvarorichard).
