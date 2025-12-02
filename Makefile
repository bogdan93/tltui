.PHONY: run build clean

build: clean install-go-deps
	go build -o bin/tltui src/main.go
	sudo cp ./bin/tltui /usr/local/bin/

run: 
	./bin/tltui

install-go-deps:
	go install ./src/...

clean:
	rm -f bin/tltui

test:
	go test ./src/... -v | grep -E "(FAIL|PASS|ok)"

run-dev:
	go run src/main.go

