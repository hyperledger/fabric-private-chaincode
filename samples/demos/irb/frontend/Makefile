TOP = ../../../..
include $(TOP)/build.mk

install:
	cd demo && npm install

run:
	cd demo && npm run serve

.PHONY: docker
docker:
	$(DOCKER) pull node:lts-alpine
	$(DOCKER) build -f Dockerfile -t irb-frontend .

.PHONY: run
docker-run:
	$(DOCKER) run --name irb-frontend -p 8080:8080 -d irb-frontend

.PHONY: stop
docker-stop:
	$(DOCKER) stop irb-frontend;
	$(DOCKER) rm irb-frontend;