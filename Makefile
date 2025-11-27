.PHONY: run build clean

build: clean install
	go build -o bin/app src/main.go

run: 
	./bin/app

install:
	go install ./src/...

clean:
	rm -f bin/app

test:
	go test ./src/... -v | grep -E "(FAIL|PASS|ok)"

run-dev:
	go run src/main.go

