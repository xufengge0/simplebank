postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres12  createdb --username=root --owner=root simple_bank
dropdb:
	docker exec -it postgres12  dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up
migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

gonew:
	docker run -it -v D:\Projects\simplebank:/go/src/app -w /go/src/app golang:1.20 bash
go: 
	docker exec -it go bash

sqlc:
	docker run --rm -it -v D:\Projects\simplebank:/src -w /src sqlc/sqlc generate
test:
	docker exec -it go go test -v -cover ./...

.PHONY: postgres createdb dropdb migrateup migratedown