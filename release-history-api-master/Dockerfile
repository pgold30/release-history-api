FROM golang:1.12-alpine3.9 as builder

RUN apk --no-cache add git

RUN mkdir -p /src/release-history-api
WORKDIR /src/release-history-api

COPY . /src/release-history-api
RUN go build

FROM alpine:3.9

WORKDIR /app/

COPY --from=builder /src/release-history-api/release-history-api .
CMD ["./release-history-api"]
