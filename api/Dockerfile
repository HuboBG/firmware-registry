FROM golang:1.23.0-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/firmware-registry-api ./cmd/firmware-registry

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/firmware-registry-api /app/
COPY migrations /app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/firmware-registry-api"]
