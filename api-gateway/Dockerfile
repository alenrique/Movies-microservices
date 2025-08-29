# Arquivo: api-gateway/Dockerfile

# --- ESTÁGIO 1: Build ---
FROM golang:1.24-alpine AS builder
RUN apk update && apk upgrade --no-cache

WORKDIR /app

COPY go.mod go.work ./
RUN go work sync

# Copia o código-fonte específico do gateway
COPY api-gateway/ ./api-gateway

# Copia os arquivos .proto e os gerados
COPY proto/ ./proto

# Compila o nosso gateway
WORKDIR /app/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api-gateway-bin .

# --- ESTÁGIO 2: Final ---
FROM alpine:3.19
RUN apk update && apk upgrade --no-cache

WORKDIR /app

# Copiamos apenas o executável do gateway
COPY --from=builder /app/api-gateway-bin .

# O gateway não precisa do movies.json

# Expomos a porta 8080, que é a porta do nosso servidor HTTP.
EXPOSE 8080

# O comando para iniciar o container.
CMD ["./api-gateway-bin"]