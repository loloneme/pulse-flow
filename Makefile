.PHONY: docker-up doker-down docker-logs docker0logs-app dockre-restart docker-clean

docker-up:
	@echo "Starting services..."
	docker-compose up -d --build
	@echo "Services started successfully!"

docker-down: 
	@echo "Stopping services..."
	docker-compose down
	@echo "Services stopped successfully!"

docker-logs: 
	docker-compose logs -f

docker-logs-app:
	docker-compose logs -f app

docker-restart: 
	@echo "Restarting services..."
	docker-compose restart
	@echo "Services restarted successfully!"

docker-clean:
	@echo "Cleaning up..."
	docker-compose down -v --rmi local
	@echo "Cleanup completed!"

test:
	@echo "Running tests..."
	go test -v ./...

build:
	@echo "Building application..."
	go build -o bin/pulse cmd/main/main.go
	@echo "Build completed! Binary: bin/pulse"
