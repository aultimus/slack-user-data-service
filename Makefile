ngrok:
	ngrok start interview --config ngrok.yml

build:
	docker-compose build

run:
	./env.sh
	docker-compose up --build
