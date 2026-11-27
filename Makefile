.PHONY: web build db run docker docker-up docker-down

web:
	cd web && npm install && npm run build

build: web
	mkdir -p _tmp _data
	go build -o _tmp/releasedesk-server ./cmd/releasedesk-server

db:
	docker compose up -d postgres

run: build db
	./_tmp/releasedesk-server

docker:
	docker compose build

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
