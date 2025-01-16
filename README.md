# auth-service

mirgate:
> go run ./cmd/migrator --dsn="postgres://{{DB_USER}}:{{DB_PASSWORD}}@{{DB_HOST}}:{{DB_PORT}}/{{DB_NAME}}?sslmode=disable" --migrations-path=./migrations