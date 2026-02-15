.PHONY: build test dev clean frontend-build frontend-install

# Build the single binary with embedded frontend.
build: frontend-build
	cp -r frontend/dist/* backend/cmd/server/dist/
	cd backend && go build -o ../dojo6 ./cmd/server

# Run all tests (Go + frontend).
test:
	cd backend && go test ./...
	cd frontend && npm test

# Start frontend dev server and Go backend concurrently.
# The Vite dev server proxies /api requests to the Go backend.
dev:
	@cd backend && go run ./cmd/server & GO_PID=$$!; \
	trap "kill $$GO_PID 2>/dev/null" EXIT; \
	cd frontend && npm run dev; \
	wait

# Install frontend dependencies.
frontend-install:
	cd frontend && npm install

# Build the frontend for production.
frontend-build: frontend-install
	cd frontend && npm run build

# Remove build artifacts.
clean:
	rm -f dojo6
	rm -rf frontend/dist
	rm -rf backend/cmd/server/dist/assets
