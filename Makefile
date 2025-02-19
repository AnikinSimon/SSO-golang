.SILENT:

build:
	docker build -t grpcserver:0.4 .

run: build
	docker compose -f ./docker-compose.yaml up -d

stop:
	docker compose -f ./docker-compose.yaml down