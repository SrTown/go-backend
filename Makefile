include .env
export

# Reemplazar postgresql:// por cockroachdb://
COCKROACH_URI_MIGRATE := $(subst postgresql://,cockroachdb://,$(COCKROACH_URI))

migrate-up:
	migrate -database "$(COCKROACH_URI_MIGRATE)" -path db/migrations up

migrate-down:
	migrate -database "$(COCKROACH_URI_MIGRATE)" -path db/migrations down

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)

migrate-force:
	migrate -database "$(COCKROACH_URI_MIGRATE)" -path db/migrations force $(version)