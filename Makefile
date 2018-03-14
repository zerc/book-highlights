help:
	@echo "build - bild all sources"
	@echo "install - install dependecies"
	@echo "kindle - run the script to parse highlights from Amazon Kindle"
	@echo "bkindle - build and run"

build: 
	GOOS=linux go build -o bin/kindle src/kindle/main.go

install:
	cd src/kindle && dep ensure && cd -

kindle: 
	docker-compose up kindle

bkindle: build kindle