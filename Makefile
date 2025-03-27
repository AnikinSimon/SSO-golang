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

.PHONY: gen

gen:
	protoc -I api/proto api/proto/sso/sso.proto -I api/proto/google/api/  \
	--go_out=./streaming/go/ \
	--go_opt paths=source_relative \
	--go-grpc_out=./streaming/go/ \
	--go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=./streaming/go/ \
    --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt paths=source_relative \
    --grpc-gateway_opt generate_unbound_methods=true \
    --openapiv2_out ./api/rest/