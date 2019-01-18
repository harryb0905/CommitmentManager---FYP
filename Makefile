.PHONY: all dev clean build env-up env-down run

all: clean build env-up run

dev: build run

##### BUILD
build:
	@echo "Build ..."
	@dep ensure
	@go build
	@echo "Build done"

##### ENV
env-up:
	@echo "Start environment ..."
	@cd fixtures && docker-compose up --force-recreate -d
	@echo "Environment up"

env-down:
	@echo "Stop environment ..."
	@cd fixtures && docker-compose down
	@echo "Environment down"

##### RUN
run:
	@echo "Start app ..."
	@./scc300-network

##### CLEAN
clean: env-down
	@echo "Clean up ..."
	@rm -rf /tmp/scc300-network-* scc300-network
	@docker rm -f -v `docker ps -a --no-trunc | grep "scc300-network" | cut -d ' ' -f 1` 2>/dev/null || true
	@docker rmi `docker images --no-trunc | grep "scc300-network" | cut -d ' ' -f 1` 2>/dev/null || true
	@echo "Clean up done"
