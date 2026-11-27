# Pin toolchain versions here so local hosts and CI cannot drift apart.
# Frontend assets are compiled first, then embedded into the Go binary.
FROM node:22-bookworm-slim AS web
WORKDIR /src
COPY web/package.json web/package-lock.json ./web/
COPY web ./web
COPY internal/views/templates ./internal/views/templates
WORKDIR /src/web
RUN npm ci && npm run build

FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/cmd/releasedesk-server/assets/bundle.js ./cmd/releasedesk-server/assets/bundle.js
COPY --from=web /src/cmd/releasedesk-server/assets/tailwind.css ./cmd/releasedesk-server/assets/tailwind.css
ENV CGO_ENABLED=0
RUN go build -o /out/releasedesk-server ./cmd/releasedesk-server

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/releasedesk-server /app/releasedesk-server
RUN mkdir -p /app/_data
ENV RELEASEDESK_DATA_DIR=/app/_data
ENV RELEASEDESK_PORTAL_URL=http://localhost:9090
EXPOSE 8080 9090
VOLUME ["/app/_data"]
CMD ["/app/releasedesk-server"]
