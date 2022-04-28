ngrok:
	ngrok start interview --config ngrok.yml

build:
	docker-compose build

run:
	docker-compose up --build
