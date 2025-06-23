BINARY_NAME=goedu-theta
MAIN_PKG=./cmd/server

.PHONY: build run clean

build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PKG)

run: build
	./bin/$(BINARY_NAME)

clean:
	rm -f bin/$(BINARY_NAME)