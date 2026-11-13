FROM golang:1.27.4-alpine AS builder

ENV CGO_ENABLED=0
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go tool wire ./internal/server
RUN go build -mod=readonly -trimpath -ldflags="-s -w" -o /out/agentmesh-server ./cmd/server

FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S agentmesh \
    && adduser -S -G agentmesh agentmesh

WORKDIR /app

COPY --from=builder /out/agentmesh-server /usr/local/bin/agentmesh-server

USER agentmesh
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/agentmesh-server"]
CMD ["-config", "/app/deployments/server.toml"]
