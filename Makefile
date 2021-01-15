package_path = github.com/yossisp/csv-to-spotify/pkg
docker_image = csvtospotify/server
port = 8000

all: build
run:
	go run cmd/server/main.go
dev:
	# hot reload server for development
	nodemon cmd/server/main.go
test:
	# copy .env files to relevant source dirs for tests
	# https://github.com/joho/godotenv/issues/43#issuecomment-337364023
	cp .env pkg/client
	go test -v $(package_path)/csv
	go test -v $(package_path)/client
	rm pkg/client/.env
build:
	go build -o bin/main cmd/server/main.go 
docker-build:
	docker build --no-cache -t $(docker_image):latest .
docker-run:
	docker run --rm --env-file=.env -it -p $(port):$(port) $(docker_image)
pre-commit:
	pre-commit run --all-files