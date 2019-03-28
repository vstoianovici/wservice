BUILD=go build -ldflags="-s -w"
BUILDPATH=./cmd

all: build

build:
	@cd $(BUILDPATH); $(BUILD) -v -o ./wService

test:
	@go test -v; cd $(BUILDPATH); go test -v

clean:
	@rm -f $(BUILDPATH)/wService

.PHONY: all build clean test



