# Generate proto and swagger files
gen:
	rm pkg/api/*
	rm swagger/*
	protoc --proto_path=api/proto api/proto/*.proto \
		   --go_out=pkg/api \
   		   --go_opt=paths=source_relative \
   		   --go-grpc_out=pkg/api \
   		   --go-grpc_opt=paths=source_relative \
   		   --openapiv2_out=swagger

# Start server side
docker-start:
	docker-compose -f deployments/docker-compose.yml up $(filter-out $@,$(MAKECMDGOALS))

server:
	docker-compose -f deployments/docker-compose.yml up

# Stop server
stop:
	docker-compose -f deployments/docker-compose.yml down

# Migration create
migrate-create:
	docker-compose -f deployments/docker-compose.yml run --rm goph_keeper_goose_service create $(filter-out $@,$(MAKECMDGOALS)) sql

# Migration up step
migrate-up:
	docker-compose -f deployments/docker-compose.yml run --rm goph_keeper_goose_service up

# Migration down step
migrate-down:
	docker-compose -f deployments/docker-compose.yml run --rm goph_keeper_goose_service redo

# Run evans service client
grpc-client-service:
	evans -r -p 3200

client:
	go run cmd/gophkeeper/client/main.go

.PHONY: clean gen server client test cert