# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY clients ./clients/
COPY cmd ./cmd/
COPY internal ./internal/

RUN apk --no-cache add gcc g++ sqlite
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/api ./cmd/api/main.go

FROM alpine:latest
COPY --from=build /bin/api /bin/api
RUN apk --no-cache add ca-certificates sqlite
CMD ["/bin/api"]
