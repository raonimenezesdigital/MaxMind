# Estágio de Build
FROM golang:1.21-alpine as builder

WORKDIR /app

# Instala ferramentas necessárias
RUN apk add --no-cache curl tar

# Copia os arquivos
COPY . .

# Baixa dependências e compila
RUN go mod tidy
RUN go build -o main .

# ARGUMENTO DE BUILD (Sua chave entra aqui)
ARG MAXMIND_LICENSE_KEY

# Baixa o banco de dados da MaxMind
RUN curl -L "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=${MAXMIND_LICENSE_KEY}&suffix=tar.gz" -o GeoLite2-City.tar.gz \
    && tar -xzf GeoLite2-City.tar.gz \
    && mv GeoLite2-City_*/*.mmdb .

# Estágio Final (Imagem leve)
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/GeoLite2-City.mmdb .

CMD ["./main"]
