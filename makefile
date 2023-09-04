.PHONY: all clean build test install

VERSION := 1.5.0
LDFLAGS := "-X main.Version=${VERSION}"

all: build test

clean:
	rm ./ratproxy

build: main.go config.go lambda.go
	go build -o ratproxy -ldflags ${LDFLAGS} *.go

test:
	go test *.go

install:
	go install -ldflags ${LDFLAGS}
