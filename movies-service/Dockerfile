# --- ESTÁGIO 1: Build ---
FROM golang:1.24-alpine AS builder
# Atualize todos os pacotes do Alpine Linux para as versões mais recentes
RUN apk update && apk upgrade --no-cache

# Define o diretório de trabalho dentro do container.
WORKDIR /app

# Copia os arquivos de gerenciamento de módulos primeiro.
# Isso aproveita o cache do Docker: se esses arquivos não mudarem,
# o 'go mod download' não será executado novamente.
COPY go.mod go.work ./
RUN go work sync

# Copia o código-fonte do nosso serviço.
COPY movies-service/ ./movies-service

# Copia os arquivos .proto e os gerados para que a compilação funcione.
COPY proto/ ./proto

# Compila o nosso aplicativo.
# CGO_ENABLED=0 cria um binário estático (não depende de libs do sistema).
# GOOS=linux garante que o executável seja para Linux (o sistema do container).
WORKDIR /app/movies-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/movies-service-bin .

# --- ESTÁGIO 2: Final ---
# Começamos com uma imagem Alpine limpa e minúscula, sem nada de Go.
FROM alpine:3.19

WORKDIR /app

# Atualize todos os pacotes do Alpine Linux para as versões mais recentes
RUN apk update && apk upgrade --no-cache

# Copiamos APENAS o executável compilado do estágio 'builder'.
COPY --from=builder /app/movies-service-bin .

# Copiamos o movies.json, que é necessário para a lógica de 'seed'.
COPY --from=builder /app/movies-service/movies.json .

# Expomos a porta 50051, que é a porta que nosso servidor gRPC escuta.
EXPOSE 50051

# O comando que será executado quando o container iniciar.
CMD ["./movies-service-bin"]