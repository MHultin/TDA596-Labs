FROM golang:latest as builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -o ./main

FROM alpine
WORKDIR /app
COPY --from=builder /build/main ./main
ENV PORT=8080
ENTRYPOINT ./main $PORT