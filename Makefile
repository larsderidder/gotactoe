rice:
	go get github.com/GeertJohan/go.rice
	go get github.com/GeertJohan/go.rice/rice

build: rice
	go build -o gotactoe
	rice append --exec gotactoe

install: build
	mv gotactoe $(GOPATH)/bin
