FROM node:20-alpine AS frontend
WORKDIR /app

COPY Wrestling/package*.json ./
RUN npm install --no-audit --no-fund

COPY Wrestling/ ./

RUN sed -i "s|background-image:url('/wrestling-japan.png');||g" src/styles.css

RUN npm run build

FROM golang:1.23-alpine AS backend
WORKDIR /src/server-go

COPY Wrestling/server-go/go.mod ./
COPY Wrestling/server-go/ ./
RUN go mod tidy

RUN CGO_ENABLED=0 \
    go build -trimpath \
    -ldflags="-s -w" \
    -o /out/wrestling \
    ./cmd/server

FROM alpine:3.20

WORKDIR /app

COPY --from=backend /out/wrestling ./wrestling
COPY --from=frontend /app/build ./build
COPY Wrestling/public ./public
COPY Wrestling/openapi.yaml ./openapi.yaml

ENV PORT=8080
ENV WEB_DIR=/app/build
ENV PUBLIC_DIR=/app/public
ENV OPENAPI_FILE=/app/openapi.yaml

EXPOSE 8080

CMD ["./wrestling"]