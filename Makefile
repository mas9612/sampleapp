BINNAME=sampleapp

all: dep test build

build:
	go build -o sampleapp .

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINNAME)

dep:
	dep ensure
