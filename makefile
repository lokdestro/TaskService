build:
	go build -o app main.go

run:
	go run main.go

clean:
	rm -f app

test_cover:
	go test -cover ./...

docker_up:
	 docker compose up -d --build

docker_down:
	docker compose down