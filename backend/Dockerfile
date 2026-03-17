# ---- build stage ----------------------------------------------------------
FROM golang:1.26-alpine AS build

WORKDIR /src

# Copy module metadata first to maximize Docker layer reuse.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/worker ./cmd/worker

# ---- api runtime ----------------------------------------------------------
FROM alpine:3.23 AS api

RUN apk add --no-cache ca-certificates wget

WORKDIR /app
COPY --from=build /out/api /app/api

EXPOSE 8080

CMD ["/app/api"]

# ---- worker runtime -------------------------------------------------------
FROM alpine:3.23 AS worker

RUN apk add --no-cache ca-certificates wget

WORKDIR /app
COPY --from=build /out/worker /app/worker

EXPOSE 8081

CMD ["/app/worker"]
