.PHONY: build build-frontend build-backend run clean test deps build-prod build-all

# Build everything (frontend + backend)
build: build-frontend build-backend

# Build frontend
build-frontend:
	cd web && npm install && npm run build

# Build backend with embedded frontend
build-backend:
	go build -o imgbed cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Clean build artifacts
clean:
	rm -f imgbed imgbed-*
	rm -rf web/dist web/node_modules
	rm -rf uploads/*

# Run tests
test:
	go test -v ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build for production (with optimizations)
build-prod: build-frontend
	CGO_ENABLED=0 go build -ldflags="-s -w" -o imgbed cmd/server/main.go

# Build for multiple platforms
build-all: build-frontend
	GOOS=linux GOARCH=amd64 go build -o imgbed-linux-amd64 cmd/server/main.go
	GOOS=darwin GOARCH=amd64 go build -o imgbed-darwin-amd64 cmd/server/main.go
	GOOS=windows GOARCH=amd64 go build -o imgbed-windows-amd64.exe cmd/server/main.go
