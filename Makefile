ARCH ?= amd64

default: run

run:
	@set -a && [ -f .env ] && . ./.env; set +a; go run .

build:
	rm -rf app
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o app -ldflags "-s -w" -trimpath .