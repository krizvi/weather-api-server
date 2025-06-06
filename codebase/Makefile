.PHONY: build run clean test

build:
	go build -o weather-api ./web/

run:
	go run ./web/

clean:
	rm -f weather-api

test:
	go test -v ./...

fmt:
	go fmt ./...

check:
	@if [ -z "$$OPENWEATHER_API_KEY" ]; then echo "Set OPENWEATHER_API_KEY env var"; exit 1; fi

dev: check run