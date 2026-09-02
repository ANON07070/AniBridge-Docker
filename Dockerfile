# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------
# Etapa 1: compila o Go Bridge
#
# Usa uma imagem separada só pra compilação. Essa imagem inteira (com
# o toolchain do Go) NÃO vai pra imagem final — só o binário resultante
# é copiado na etapa 2. Isso mantém a imagem final pequena.
#
# CGO_ENABLED=0: pkg/goanime arrasta transitivamente uma dependência de
# histórico/tracking do CLI original (go-sqlite3, via cgo) que este
# projeto nunca usa. Compilar isso exige um arquivo C de ~250 mil
# linhas, o que já travou builds em máquinas com pouca RAM (confirmado
# em testes reais). Desligando cgo, esse código nem entra na build.
# ---------------------------------------------------------------------
FROM golang:1.27-alpine AS bridge-builder

WORKDIR /build

COPY bridge/ ./

RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o goanime-bridge .

# ---------------------------------------------------------------------
# Etapa 2: imagem final — só Python, sem toolchain Go
# ---------------------------------------------------------------------
FROM python:3.12-slim

WORKDIR /app

# Bridge já compilado (só o binário)
COPY --from=bridge-builder /build/goanime-bridge /app/bridge/goanime-bridge

# Dependências Python primeiro, pra aproveitar cache do Docker quando só
# o código da aplicação mudar (sem mexer em requirements.txt)
COPY backend/requirements.txt /app/backend/requirements.txt
RUN pip install --no-cache-dir -r /app/backend/requirements.txt

# Código da aplicação
COPY backend/ /app/backend/
COPY frontend/ /app/frontend/

# Sempre o mesmo caminho dentro do container — ninguém precisa mais
# definir essa variável manualmente, diferente de rodar fora do Docker.
ENV BRIDGE_PATH=/app/bridge/goanime-bridge

WORKDIR /app/backend

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
