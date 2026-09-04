FROM golang:1.26.6-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/financial-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/financial-api /financial-api

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/financial-api"]
