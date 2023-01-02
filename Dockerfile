FROM golang:1.19-alpine AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/go-domestia

FROM scratch
COPY --from=builder /out/go-domestia /go-domestia

ENTRYPOINT ["/go-domestia"]
