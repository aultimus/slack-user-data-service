ngrok:
	ngrok start interview --config ngrok.yml

build:
	docker-compose build

down:
	docker-compose down --volumes --remove-orphans

run:
	./env.sh
	docker-compose up --build
