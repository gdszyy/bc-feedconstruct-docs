FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/bffd ./cmd/bffd
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/migrate ./cmd/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/bffd /app/bin/bffd
COPY --from=build /out/migrate /app/bin/migrate
COPY migrations /app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/bin/bffd"]
