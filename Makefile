.DEFAULT_GOAL := build

BIN_FILE=leoblog

build:
	@go build -o "${BIN_FILE}"

clean:
	go clean
	rm -f "cp.out"
	rm -f nohup.out
	rm -f "${BIN_FILE}"

test:
	go test

check:
	go test

cover:
	go test -coverprofile cp.out
	go tool cover -html=cp.out

run:
	./"${BIN_FILE}"

lint:
	golangci-lint run --enable-all
