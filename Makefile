.PHONY: integrationtest

ngrok:
	ngrok start interview --config ngrok.yml

build:
	docker-compose build

down:
	docker-compose down --volumes --remove-orphans

run:
	docker-compose up --build

integrationtest:
	docker-compose --file integration-test-docker-compose.yml up --build
