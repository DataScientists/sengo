# Setup databse

setup_db:
	./bin/init_db.sh

setup_db_test:
	./bin/init_db_test.sh

setup_db_e2e:
	./bin/init_db_e2e.sh

migrate_docker_schema:
	./bin/migrate_schema.sh
	#Start dev server

start:
	air

migrate_schema:
	docker exec -it go_app go run ./cmd/migration/main.go
test_unit:
	docker exec -it go_app gotestsum --hide-summary=skipped --format short -- -v -short -tags=unit -coverprofile=cover.out ./pkg/...

test_repository:
		docker exec -it go_app gotestsum --hide-summary=skipped  --format short -- -v -tags=unit  ./pkg/adapter/repository/...

test_e2e:
		docker exec -it go_app go test ./test/e2e/...

generate_ent:
	go generate ./ent

generate_repo_mocks:
	go generate ./pkg/usecase/repository/...

gqlgen:
	gqlgen

.PHONY: setup_db setup_db_test setup_db_e2e start migrate_schema test_e2e test_repository generate_ent generate_repo_mocks gqlgen
