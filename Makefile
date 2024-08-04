.DEFAULT_GOAL := run
.SILENT:
.PHONY:

set:
	go env -w CGO_ENABLED=1

test:
	go test -v ./...

clean:
	go clean
	go clean -cache
	rmdir /S /Q bin

tidy:
	go mod tidy

fmt: tidy
	go fmt ./...

vet: fmt
	go vet ./...

build: vet set
	@echo "Building"
	go env -w GOOS=windows
	go env -w GOARCH=amd64
	go build -tags "sqlite_foreign_keys sqlite_secure_delete" -o bin/game-streams_win64.exe game-streams/main.go
	@echo "Build complete"

run: build
	bin/game-streams_win64.exe
