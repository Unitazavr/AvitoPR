# docker/Dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod ./
#go.sum
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /app/bin/server ./cmd
RUN chmod +x /app/bin/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /app/bin/server /usr/local/bin/server
EXPOSE 8080
CMD ["/usr/local/bin/server"]
