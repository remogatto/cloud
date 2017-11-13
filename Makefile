.PHONY: test

test:
	docker-compose -f test/docker/docker-compose.yml up -d
	go test
