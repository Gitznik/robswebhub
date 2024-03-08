setup-env:
	./scripts/init_postgres.sh

teardown-env:
	RUNNING_POSTGRES_CONTAINER=$(docker ps --filter 'name=postgres' --format '{{{{.ID}}') && docker kill ${RUNNING_POSTGRES_CONTAINER}

build-docker:
	docker build -t robswebhub .

run-docker:
	docker run -p 8080:8080 --env APP_ENVIRONMENT=production --env APP_APPLICATION__PORT=8080 --env DATABASE_URL="postgres://postgres:password@localhost:5432/robswebhub" robswebhub

build-and-run: build-docker run-docker

run:
	cargo watch -x run
