LINT_EXCLUDES := vendor/

.PHONY: check
check: fmt vet lint

.PHONY: fmt
fmt:
	@echo "============= Go fmt ==============="
	@test -z "$$(gofmt -s -l . 2>&1 | grep -vE "$(LINT_EXCLUDES)" | tee /dev/stderr)"

.PHONY: vet
vet:
	@echo "============= Go vet ==============="
	@test -z "$$(go list ./... | xargs go vet | tee /dev/stderr)"

.PHONY: lint
lint:
	@echo "============= Go lint ==============="
	@go get github.com/golang/lint/golint
	@test -z "$$(go list ./... | xargs -L1 "$(GOPATH)"/bin/golint | tee /dev/stderr)"



