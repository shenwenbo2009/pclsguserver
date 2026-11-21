# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/shared/package.json ./shared/
COPY frontend/admin/package.json ./admin/
COPY frontend/market/package.json ./market/

RUN cd shared && npm install --silent
RUN cd admin && npm install --silent
RUN cd market && npm install --silent

COPY frontend/ ./

RUN cd admin && npm run build
RUN cd market && npm run build

# Stage 2: Build Go binary
FROM golang:1.27-alpine AS go-builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy frontend build output into web/ for go:embed
COPY --from=frontend-builder /app/web/admin ./web/admin
COPY --from=frontend-builder /app/web/market ./web/market

RUN CGO_ENABLED=0 GOOS=linux go build -o /pclsguserver ./cmd/server

# Stage 3: Final runtime image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=go-builder /pclsguserver /app/pclsguserver
COPY config.yaml.example /app/config.yaml.example

RUN mkdir -p /app/data /app/logs

EXPOSE 8080

ENV PCL_SERVER_DATADIR=/app/data
ENV PCL_DB_DRIVER=sqlite
ENV PCL_DB_DSN=/app/data/meta.db

ENTRYPOINT ["/app/pclsguserver"]
CMD ["-config", "/app/config.yaml"]