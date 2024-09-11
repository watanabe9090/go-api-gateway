build:
	go build -o bin/go-api-gateway

dev:
	nodemon -L --exec go run main.go props.yaml --signal SIGTERM

test:
	go test

run: 
	go run main.go props.yaml

docker-build:
	docker build . -t go-api-gateway
