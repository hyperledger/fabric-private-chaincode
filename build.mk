include $(TOP)/config.mk

# optionlly allow local overriding defaults
-include $(TOP)/config.override.mk

.PHONY: all
#all: build checks test integration
all: build checks test

.PHONY: build
.PHONY: checks
.PHONY: test
#.PHONY: integration
.PHONY: clean

.PHONY: docker
.PHONY: docker-run
.PHONY: docker-stop
.PHONY: docker-clean
