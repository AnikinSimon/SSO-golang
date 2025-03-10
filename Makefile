.PHONY: cert
.SILENT: cert

build: cert
	docker buildx build -t grpcserver:0.4 .

run: build
	docker compose -f ./docker-compose.yaml up -d

stop:
	docker compose -f ./docker-compose.yaml down

cert:
	cd cert; ls; ./gen.sh; cd ..

test:
	go test ./tests/suite