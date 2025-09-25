FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .
FROM gcr.io/distroless/base-debian12
WORKDIR /app
ENV PORT=8082
EXPOSE 8082
COPY --from=builder /app/server /app/server
ENTRYPOINT ["/app/server"]