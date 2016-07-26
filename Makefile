GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
VER_NUM=latest
DOCKER_OPTIONS="--no-cache"
IMAGE_NAME=unixvoid/beacon-client:$(VER_NUM)
REDIS_DB_HOST_DIR="/tmp/"
HOST_IP=192.168.1.9
CURRENT_DIR=$(shell pwd)

all: beacon-client

dependencies:
	go get gopkg.in/gcfg.v1
	go get git.unixvoid.com/mfaltys/glogger

beacon-client: beacon-client.go
	$(GOC) beacon-client.go

run: beacon-client.go
	go run beacon-client.go

stat: beacon-client.go
	$(CGOR) $(GOC) $(GOFLAGS) beacon-client.go

install: stat
	cp beacon-client /usr/bin

docker:
	$(MAKE) stat
	mkdir stage.tmp/
	cp beacon-client stage.tmp/
	#cp auth stage.tmp/
	cp deps/Dockerfile stage.tmp/
	#cp config.gcfg stage.tmp/
	cd stage.tmp/ && \
		sudo docker build $(DOCKER_OPTIONS) -t $(IMAGE_NAME) .
	@echo "$(IMAGE_NAME) built"

dockerrun:
	sudo docker run \
		-d \
		--name beacon-client \
		--add-host dockerhost:$(HOST_IP) \
		-v $(CURRENT_DIR)/config.gcfg:/config.gcfg:ro \
		-v $(CURRENT_DIR)/auth:/auth:ro \
		mfaltys/beacon-client
	sudo docker logs -f beacon-client
	sudo docker rm beacon-client

clean:
	rm -f beacon-client
	rm -rf stage.tmp/
#CGO_ENABLED=0 go build -a -ldflags '-s' beacon-client.go
