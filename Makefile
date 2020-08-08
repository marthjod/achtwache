packages ?= $(shell go list ./... | grep -v /vendor/ | grep -v /tests)

.PHONY: all
all: vet lint errcheck staticcheck test

.PHONY: vet
vet:
	go vet $(packages)

.PHONY: lint
lint:
	status=0; for package in $(packages); do ${GOBIN}/golint -set_exit_status $$package || status=1; done; exit $$status

.PHONY: errcheck
errcheck:
	status=0; for package in $(packages); do errcheck -ignoretests $$package || status=1; done; exit $$status

.PHONY: staticcheck
staticcheck:
	status=0; for package in $(packages); do staticcheck $$package || status=1; done; exit $$status

.PHONY: test
test:
	status=0; for package in $(packages); do go test -cover -coverprofile $$GOPATH/src/$$package/coverage.out $$package || status=1; done; exit $$status