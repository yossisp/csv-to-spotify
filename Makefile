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
docker-build-prod:
	docker build
      --build-arg SPOTIFY_CLIENT_ID_SECRET_BASE64=$$SPOTIFY_CLIENT_ID_SECRET_BASE64
      --build-arg MARKET=$$MARKET
      --build-arg MONGO_DB_NAME=$$MONGO_DB_NAME
      --build-arg MONGO_ATLAS_CONNECTION=$$MONGO_ATLAS_CONNECTION
      --build-arg KAFKA_BROKERS=$$KAFKA_BROKERS
      --build-arg KAFKA_USERNAME=$$KAFKA_USERNAME
      --build-arg KAFKA_PASSWORD=$$KAFKA_PASSWORD
      --build-arg KAFKA_GROUP_ID=$$KAFKA_GROUP_ID
      --build-arg KAFKA_TRACK_PROGRESS_TOPIC=$$KAFKA_TRACK_PROGRESS_TOPIC
      --build-arg INPUT_FILE_EXT=$$INPUT_FILE_EXT
      --build-arg PORT=$$PORT
      --build-arg ALLOWED_ORIGINS=$$ALLOWED_ORIGINS
      --build-arg TRACK_LOOKUP_INTERVAL=$$TRACK_LOOKUP_INTERVAL
      --build-arg TEST_REFRESH_TOKEN=$$TEST_REFRESH_TOKEN
      -f Dockerfile -t $$DOCKERHUB_CSV_TO_SPOTIFY_IMAGE .
docker-run:
	docker run --rm --env-file=.env -it -p $(port):$(port) $(docker_image)