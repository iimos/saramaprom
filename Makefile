export GO111MODULE := on
export GOPROXY := go-proxy.oss.wandera.net
export GONOSUMDB := github.com/wandera/*

all: check test

MAKEFLAGS += --no-print-directory

prepare:
	@echo "Downloading tools"
	@cat tools.go | grep _ | cut -f2 -d " " | xargs -tI % sh -c "go install %"

check: prepare
	@echo "Running check"
ifeq (, $(shell which golangci-lint))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.50.1
endif
	golangci-lint run
	go mod tidy

test: prepare
	@echo "Running tests"
	mkdir -p report
	go test -race -v ./... -coverprofile=report/coverage.txt | tee report/report.txt
	go-junit-report -set-exit-code < report/report.txt > report/report.xml
	gocov convert report/coverage.txt | gocov-xml > report/coverage.xml
	go mod tidy

clean:
	@echo "Running clean"
	rm -rf "report/"

.PHONY: all check test prepare
