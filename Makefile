#GOLANGCI_LINT_VERSION := "v1.43.0" # Optional configuration to pinpoint golangci-lint version.

# The head of Makefile determines location of dev-go to include standard targets.
GO ?= go
export GO111MODULE = on

ifneq "$(GOFLAGS)" ""
  $(info GOFLAGS: ${GOFLAGS})
endif

ifneq "$(wildcard ./vendor )" ""
  $(info Using vendor)
  modVendor =  -mod=vendor
  ifeq (,$(findstring -mod,$(GOFLAGS)))
      export GOFLAGS := ${GOFLAGS} ${modVendor}
  endif
  ifneq "$(wildcard ./vendor/github.com/bool64/dev)" ""
  	DEVGO_PATH := ./vendor/github.com/bool64/dev
  endif
endif

ifeq ($(DEVGO_PATH),)
	DEVGO_PATH := $(shell GO111MODULE=on $(GO) list ${modVendor} -f '{{.Dir}}' -m github.com/bool64/dev)
	ifeq ($(DEVGO_PATH),)
    	$(info Module github.com/bool64/dev not found, downloading.)
    	DEVGO_PATH := $(shell export GO111MODULE=on && $(GO) get github.com/bool64/dev && $(GO) list -f '{{.Dir}}' -m github.com/bool64/dev)
	endif
endif

-include $(DEVGO_PATH)/makefiles/main.mk
-include $(DEVGO_PATH)/makefiles/build.mk
-include $(DEVGO_PATH)/makefiles/lint.mk
-include $(DEVGO_PATH)/makefiles/test-unit.mk
-include $(DEVGO_PATH)/makefiles/test-integration.mk
-include $(DEVGO_PATH)/makefiles/bench.mk
-include $(DEVGO_PATH)/makefiles/reset-ci.mk

# Add your custom targets here.
BUILD_PKG = ./cmd/qonto
BUILD_LDFLAGS="-s -w"
INTEGRATION_TEST_TARGET = -allure -coverpkg ./internal/... integration_test.go

APP_PATH = $(shell pwd)
APP_SCRIPTS = $(APP_PATH)/resources/app/scripts
SRC_PROTO_PATH = $(APP_PATH)/resources/proto
GO_PROTO_PATH = $(APP_PATH)/pkg/proto
SWAGGER_PATH = $(APP_PATH)/resources/swagger

-include $(APP_PATH)/resources/app/makefiles/database.mk
-include $(APP_PATH)/resources/app/makefiles/dep.mk
-include $(APP_PATH)/resources/app/makefiles/protoc.mk

## Run tests
test: test-unit test-integration

## Generate code from proto file(s)
proto-gen-code: protoc-cli
	protoc --proto_path=$(SRC_PROTO_PATH) $(SRC_PROTO_PATH)/*.proto  --go_opt=paths=source_relative --go_out=:$(GO_PROTO_PATH) --go-grpc_opt=paths=source_relative --go-grpc_out=:$(GO_PROTO_PATH) --grpc-gateway_opt=paths=source_relative --grpc-gateway_out=:$(GO_PROTO_PATH) --openapiv2_out=:$(SWAGGER_PATH)
	@cat $(SWAGGER_PATH)/service.swagger.json | jq del\(.paths[][].responses.'"default"'\) > $(SWAGGER_PATH)/service.swagger.json.tmp
	@mv $(SWAGGER_PATH)/service.swagger.json.tmp $(SWAGGER_PATH)/service.swagger.json
	@cat $(SWAGGER_PATH)/service.swagger.json | jq del\(.paths.'"/v1/transfer/bulk"'.post.responses.'"200"'\) > $(SWAGGER_PATH)/service.swagger.json.tmp
	@mv $(SWAGGER_PATH)/service.swagger.json.tmp $(SWAGGER_PATH)/service.swagger.json
