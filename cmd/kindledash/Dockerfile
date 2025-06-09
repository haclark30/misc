# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

RUN mkdir clients
RUN mkdir cmd
RUN mkdir internal

COPY clients ./clients/
COPY cmd ./cmd/
COPY internal ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/kindledash ./cmd/kindledash/*.go

FROM alpine:latest
RUN apk update && apk add bash && apk --no-cache add tzdata
RUN mkdir .fitbit_token
COPY --from=build /bin/kindledash /bin/kindledash
CMD ["/bin/kindledash"]
