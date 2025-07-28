 DOCKER_COMPOSE = docker-compose
# Docker-команды
docker-build:
	${DOCKER_COMPOSE} build

docker-up:
	${DOCKER_COMPOSE} up -d

docker-down:
	${DOCKER_COMPOSE} down

docker-logs:
	${DOCKER_COMPOSE} logs -f url-shortener

