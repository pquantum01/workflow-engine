# Stage 1: Frontend build
FROM node:20-alpine AS web
WORKDIR /web
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Go build (embeds frontend dist)
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.work ./
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download
COPY backend/ ./backend/
COPY --from=web /web/dist ./backend/webroot/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" \
    -o /out/server ./backend/cmd/server

# Stage 3: Runtime
FROM alpine:3.20
RUN addgroup -S app && adduser -S -G app app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/server /server
COPY backend/migrations /migrations
COPY workflows /workflows
USER app
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:8080/healthz || exit 1
ENTRYPOINT ["/server"]
