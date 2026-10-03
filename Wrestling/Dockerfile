FROM node:20-alpine AS frontend
WORKDIR /app
COPY package*.json ./
RUN npm install --no-audit --no-fund
COPY public ./public
COPY src ./src
COPY scripts ./scripts
RUN npm run build

FROM golang:1.23-alpine AS backend
WORKDIR /src
COPY server-go/go.mod ./server-go/
RUN cd server-go && go mod download
COPY server-go ./server-go
RUN cd server-go && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/wrestling ./cmd/server

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=backend /out/wrestling ./wrestling
COPY --from=frontend /app/build ./build
COPY public/images ./public/images
COPY openapi.yaml ./openapi.yaml
ENV OPENAPI_FILE=/app/openapi.yaml PUBLIC_DIR=/app/public WEB_DIR=/app/build PORT=8080 \
    CORS_ORIGINS=http://localhost:8080 \
    DB_MAX_CONNS=20 DB_MIN_CONNS=2
USER app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=5 CMD wget -qO- http://127.0.0.1:8080/api/ready || exit 1
CMD ["./wrestling"]
