.PHONY: dev dev-api dev-fe build prod deps migrate

CONFIG ?= configs/config.example.yaml

deps:
	cd backend && go mod tidy
	cd frontend && npm install
	pip3 install -r tools/requirements.txt || python3 -m pip install -r tools/requirements.txt

dev-api:
	cd backend && go run ./cmd/server -config ../$(CONFIG) -mode api

dev-fe:
	cd frontend && npm run dev

dev:
	@echo "Run 'make dev-api' and 'make dev-fe' in separate terminals"

build-fe:
	cd frontend && npm run build
	rm -rf backend/internal/embed/dist/*
	cp -r frontend/dist/* backend/internal/embed/dist/

build:
	$(MAKE) build-fe
	cd backend && go build -o ../bin/oc2-manager ./cmd/server

prod:
	$(MAKE) build
	./bin/oc2-manager -config $(CONFIG) -mode public

migrate:
	@echo "Migrations run automatically on server start"
