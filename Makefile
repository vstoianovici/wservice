BUILD=go build -ldflags="-s -w"
BUILDPATH=./cmd

all: build

build:
	@cd $(BUILDPATH); $(BUILD) -v -o ./wService

clean:
	@rm -f $(BUILDPATH)/wService

.PHONY: all build clean



