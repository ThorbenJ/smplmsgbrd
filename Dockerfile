FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /smplmsgbrd .

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
COPY --from=builder /smplmsgbrd /usr/local/bin/smplmsgbrd
USER app
EXPOSE 8080
ENTRYPOINT ["smplmsgbrd"]
