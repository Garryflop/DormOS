.PHONY: up down clean

up:
	docker-compose up -d

down:
	docker-compose down

clean:
	docker-compose down -v
	rm -rf gen/
	find . -name "*.exe" -type f -delete
