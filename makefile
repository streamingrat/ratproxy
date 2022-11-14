.PHONY: all clean build test install

all: build test

clean:
	rm ./ratproxy

build: main.go config.go lambda.go
	go build -o ratproxy *.go

test:
	go test *.go

install:
	go install
