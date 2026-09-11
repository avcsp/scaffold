default: run

run:
	@set -a && [ -f .env ] && . ./.env; set +a; go run routes.go app.go

build:
	rm -rf app
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app -ldflags "-s -w" -trimpath routes.go app.go