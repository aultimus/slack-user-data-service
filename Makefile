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
	docker-compose -f integrationtest/docker-compose.yml up --build --exit-code-from integrationtest
