FROM golang:1.23-alpine as builder
RUN update-ca-certificates
WORKDIR app/
COPY go.mod .
ENV GO111MODULE=on
RUN go mod download && go mod verify
COPY . .
ENV GOCACHE=/root/.cache/go-build
RUN GOOS=linux go build -a -installsuffix cgo -ldflags '-w -extldflags "-static"' -o /app .
FROM alpine:latest
WORKDIR /usr/local/bin
COPY --from=builder /app /usr/local/bin/app
ENTRYPOINT ["./app"]
