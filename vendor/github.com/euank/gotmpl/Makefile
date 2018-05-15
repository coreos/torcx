GO_FILES = $(shell find . -type f -name '*.go')

all: ./bin/gotmpl

release: clean
	docker build -f ./Dockerfile.build -t euank/gotmpl:build .
	docker run -v "$(shell pwd)/bin:/go/src/github.com/euank/gotmpl/bin" euank/gotmpl:build

./bin/gotmpl: $(GO_FILES)
	go build -o ./bin/gotmpl ./cmd/gotmpl

test: ./bin/gotmpl
	go test -v ./...

clean:
	rm -f ./bin/gotmpl

