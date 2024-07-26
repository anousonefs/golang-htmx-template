migrateup:
	migrate -path migrations -database "postgres://user:pwd@localhost:5432/db_name?sslmode=disable" -verbose up

migratedown:
	migrate -path migrations -database "postgres://user:pwd@localhost:5432/db_name?sslmode=disable" -verbose down

test:
	go test -v -cover ./...

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/anousonefs/golang-htmx-template/db/sqlc Store

.PHONY: tailwind-watch
tailwind-watch:
	./tailwindcss -i ./static/css/input.css -o ./static/css/style.css --watch

.PHONY: tailwind-build
tailwind-build:
	./tailwindcss -i ./static/css/input.css -o ./static/css/style.min.css --minify

.PHONY: templ-generate
templ-generate:
	templ generate

.PHONY: templ-watch
templ-watch:
	templ generate --watch
	
.PHONY: dev
dev:
	go build -o ./tmp/$(APP_NAME) ./$(APP_NAME)/main.go && air

.PHONY: build
build:
	make tailwind-build
	make templ-generate
	go build -ldflags "-X main.Environment=production" -o ./bin/$(APP_NAME) ./cmd/$(APP_NAME)/main.go

.PHONY: vet
vet:
	go vet ./...

.PHONY: staticcheck
staticcheck:
	staticcheck ./...

.PHONY: test
test:
	  go test -race -v -timeout 30s ./...


