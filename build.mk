include $(TOP)/config.mk

# optionlly allow local overriding defaults
-include $(TOP)/config.override.mk

.PHONY: all
#all: build check test integration
all: build test

.PHONY: build
#.PHONY: check
.PHONY: test
#.PHONY: integration
.PHONY: clean

.PHONY: docker
.PHONY: docker-run
.PHONY: docker-stop
.PHONY: docker-clean
