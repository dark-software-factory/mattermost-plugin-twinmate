PLUGIN_ID := com.darksoftwarefactory.twinmate
VERSION := 0.1.0

.PHONY: test build-server build-web bundle clean build-mcp

test:
	go test ./...
	cd webapp && npm test

build-server:
	mkdir -p server/dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o server/dist/plugin-linux-amd64 ./server

build-web:
	cd webapp && npm run build

build-mcp:
	mkdir -p dist
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o dist/twinmate-mcp-darwin-arm64 ./cmd/twinmate-mcp
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o dist/twinmate-mcp-darwin-amd64 ./cmd/twinmate-mcp
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/twinmate-mcp-linux-amd64 ./cmd/twinmate-mcp

bundle: build-server build-web
	mkdir -p dist/plugin/server/dist dist/plugin/webapp/dist
	cp plugin.json dist/plugin/plugin.json
	cp server/dist/plugin-linux-amd64 dist/plugin/server/dist/plugin-linux-amd64
	cp webapp/dist/main.js dist/plugin/webapp/dist/main.js
	tar -C dist/plugin -czf dist/$(PLUGIN_ID)-$(VERSION).tar.gz .

clean:
	rm -f server/dist/plugin-linux-amd64
	rm -rf dist/plugin
