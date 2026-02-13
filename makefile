kindledash:
	watchexec -r -e go --wrap-process session -- "go run ./cmd/kindledash/*.go test"

api:
	watchexec -r -e go --wrap-process session -- "go run ./cmd/api/*.go"

build:
	go build -o bin/kindledash ./cmd/kindledash/*.go
	go build -o bin/api ./cmd/api/*.go
