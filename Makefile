.PHONY: build/release

build/release:
	GOOS=darwin GOARCH=amd64 go build -o cslack main.go 
	tar czf cslack-darwin-amd64.tar.gz cslack
	rm -f cslack
	GOOS=linux GOARCH=amd64 go build -o cslack main.go 
	tar czf cslack-linux-amd64.tar.gz cslack
	rm -f cslack
