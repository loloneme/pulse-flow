.PHONY: docker-up doker-down docker-logs docker0logs-app dockre-restart docker-clean

docker-elk-up:
	@echo "Starting ELK services..."
	docker-compose -f elk/docker-compose.yml up -d --build
	@echo "ELK services started successfully!"

docker-app-up:
	@echo "Starting app services..."
	docker-compose up -d --build
	@echo "App services started successfully!"

docker-up: docker-elk-up docker-app-up
	@echo "All services started successfully!"

docker-down: 
	@echo "Stopping ELK services..."
	docker-compose -f elk/docker-compose.yml down
	@echo "Stopping app services..."
	docker-compose down 
	@echo "All services stopped successfully!"

docker-logs: 
	docker-compose -f elk/docker-compose.yml logs -f
	docker-compose logs -f

docker-logs-app:
	docker-compose -f elk/docker-compose.yml logs -f app
	docker-compose logs -f app

docker-restart: 
	@echo "Restarting services..."
	docker-compose -f elk/docker-compose.yml restart
	docker-compose restart
	@echo "Services restarted successfully!"

docker-clean:
	@echo "Cleaning up..."
	docker-compose -f elk/docker-compose.yml down -v --rmi local
	docker-compose down -v --rmi local
	@echo "Cleanup completed!"

test:
	@echo "Running tests..."
	go test -v ./...

build:
	@echo "Building application..."
	go build -o bin/pulse cmd/main/main.go
	@echo "Build completed! Binary: bin/pulse"
