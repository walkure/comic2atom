FROM golang:1.19.4-alpine3.17 as builder

WORKDIR /app
COPY . /app/
RUN apk update && apk add --no-cache ca-certificates tzdata && update-ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && apk del tzdata
RUN go mod download
RUN addgroup -g 6128 -S nonroot && adduser -u 6128 -S nonroot -G nonroot
RUN CGO_ENABLED=0 GOOS=linux go build

FROM scratch as runner

WORKDIR /app
COPY --from=builder /etc/passwd /etc/group /etc/localtime /etc/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER nonroot
COPY --chown=nonroot:nonroot --from=builder  /app/comic2atom /app/

ENTRYPOINT ["/app/comic2atom"]
