.PHONY: run build clean

run:
	go run src/main.go

build: clean
	go build -o bin/app src/main.go

clean:
	rm -f bin/app

test:
	go test ./src/... -v | grep -E "(FAIL|PASS|ok)"

