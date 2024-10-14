postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres12  createdb --username=root --owner=root simple_bank
dropdb:
	docker exec -it postgres12  dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up
migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1
migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down
migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1
new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)
	
gonew:
	docker run -it -v D:\Projects\simplebank:/go/src/app -w /go/src/app golang:1.20 bash
go: 
	docker exec -it go bash
sqlc:
	docker run --rm -it -v D:\Projects\simplebank:/src -w /src sqlc/sqlc generate
test:
	docker exec -it go go test -short -v -cover ./...
server:
	go run main.go
mock:
	mockgen -package mockdb  -destination db/mock/store.go github.com/techschool/simplebank/db/sqlc Store
	mockgen -package mockwk  -destination worker/mock/distributor.go github.com/techschool/simplebank/worker TaskDistributor
proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	rm -f doc/statik/statik.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
    proto/*.proto
	statik -src=./doc/swagger -dest=doc

evans:
	evans --host localhost --port 9091 -r repl
	
redis:
	docker run -d --name redis --network app-network redis:7-alpine


.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock proto