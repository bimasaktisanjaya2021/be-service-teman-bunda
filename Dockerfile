  # builder image
  FROM golang:1.18.1-alpine as builder
  WORKDIR /build
  COPY . .
  RUN apk add git && CGO_ENABLED=0 GOOS=linux go build -o be-service-teman-bunda .

  # generate clean, final image for end users
  FROM alpine
  RUN apk add --no-cache curl
  RUN apk update && apk add ca-certificates && apk add tzdata && apk add git
  COPY --from=builder /build .
  ENV TZ="Asia/Makassar"
  EXPOSE 9090

  # Use an entrypoint script to handle the health check and startup process
  CMD ./be-service-teman-bunda

  # Add the health check instruction
  HEALTHCHECK --interval=30s --timeout=10s CMD curl -f https://aether-go.temanbundabelanja.com/ || kill 1