FROM golang:1.23-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o imdb-meilisearch .
RUN chmod +x /app/imdb-meilisearch



FROM alpine:latest AS meilisearch-builder
RUN apk add --no-cache curl
COPY install_meilisearch.sh /install_meilisearch.sh
RUN chmod +x /install_meilisearch.sh
RUN /install_meilisearch.sh



FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y --no-install-recommends libgcc-s1 libstdc++6 curl && apt-get clean
COPY --from=go-builder /app/imdb-meilisearch /usr/local/bin/imdb-meilisearch
COPY --from=meilisearch-builder /usr/local/bin/meilisearch /usr/local/bin/meilisearch
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

