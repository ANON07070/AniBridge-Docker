#!/usr/bin/env bash
#
# Compila o Go Bridge, prepara o ambiente Python e sobe o servidor
# FastAPI (que também serve o frontend).
#
# Uso:
#   ./run.sh
#
# Variáveis opcionais:
#   HOST=127.0.0.1 PORT=9000 ./run.sh
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BRIDGE_DIR="$SCRIPT_DIR/bridge"
BACKEND_DIR="$SCRIPT_DIR/backend"

HOST="${HOST:-0.0.0.0}"
PORT="${PORT:-8000}"

# --- Pré-requisitos ---

if ! command -v go >/dev/null 2>&1; then
  echo "Erro: Go não encontrado no PATH." >&2
  echo "Instale Go 1.26.5 ou mais recente antes de continuar: https://go.dev/dl/" >&2
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "Erro: python3 não encontrado no PATH." >&2
  echo "Instale Python 3.10 ou mais recente antes de continuar." >&2
  exit 1
fi

# --- Go Bridge ---

echo "==> Compilando o Go Bridge..."
cd "$BRIDGE_DIR"
go mod tidy
go build -o goanime-bridge .

export BRIDGE_PATH="$BRIDGE_DIR/goanime-bridge"
echo "    Bridge compilado em: $BRIDGE_PATH"

# --- Ambiente Python ---

echo "==> Preparando ambiente Python..."
cd "$BACKEND_DIR"

if [ ! -d ".venv" ]; then
  python3 -m venv .venv
fi

# shellcheck disable=SC1091
source .venv/bin/activate

pip install --quiet --upgrade pip
pip install --quiet -r requirements.txt

# --- Servidor ---

echo "==> Subindo servidor em http://$HOST:$PORT (Ctrl+C para parar)"
echo "    Acesse pelo navegador em http://localhost:$PORT (ou pelo IP da"
echo "    máquina, se estiver acessando de outro dispositivo na rede)."
echo ""

exec uvicorn main:app --host "$HOST" --port "$PORT"
