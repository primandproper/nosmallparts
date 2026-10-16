# syntax=docker/dockerfile:1.7

# The web app, built first: the Go binary embeds it.
FROM node:22-bookworm-slim AS web
RUN npm install --global pnpm@10.33.0
WORKDIR /src
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY frontend/package.json frontend/
RUN pnpm install --frozen-lockfile --filter @nsp/web
COPY frontend/ frontend/
RUN pnpm --filter @nsp/web build

# The server, with the web app inside it. modernc's SQLite is pure Go, so the
# binary needs no cgo and runs on a static base.
FROM golang:1.27-bookworm AS server
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=web /src/frontend/dist/ internal/web/dist/
RUN CGO_ENABLED=0 ./scripts/build.sh -o /out/nosmallparts ./cmd/main

# Root, because Fly mounts the /data volume owned by root and SQLite writes
# there.
FROM gcr.io/distroless/static-debian12
COPY --from=server /out/nosmallparts /nosmallparts
COPY backend/config/production.json /etc/nosmallparts/config.json
ENV NOSMALLPARTS_CONFIG_FILEPATH=/etc/nosmallparts/config.json
EXPOSE 3001
VOLUME /data
ENTRYPOINT ["/nosmallparts"]
CMD ["serve"]
