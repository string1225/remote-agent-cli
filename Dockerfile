FROM golang:1.26.3-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /remote-server ./cmd/server && go run ./tools/build

FROM alpine:3.23
RUN apk add --no-cache ca-certificates && addgroup -S agent && adduser -S -G agent agent && mkdir /data && chown agent:agent /data
WORKDIR /app
COPY --from=build /remote-server ./remote-server
COPY --from=build /src/dist ./dist
USER agent
ENV LISTEN_ADDR=0.0.0.0:8080 DATABASE_PATH=/data/remote-agent.db DOWNLOADS_DIR=/app/dist
EXPOSE 8080
ENTRYPOINT ["/app/remote-server"]
