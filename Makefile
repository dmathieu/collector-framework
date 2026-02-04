GOMODULES := $(shell find . -mindepth 2 \
							 -type f \
							 -name "go.mod" \
							 -not -path "./internal/tools/*" \
							 -exec dirname {} \; | sort )

.PHONY: for-all
for-all:
	@set -e; for dir in $(GOMODULES); do \
		(cd "$${dir}" && \
			echo "running $${CMD} in $${dir}" && \
		$${CMD} ); \
	done

.PHONY: test
test:
	@$(MAKE) for-all CMD="go test ./..."

.PHONY: tidy
tidy:
	@$(MAKE) for-all CMD="go mod tidy"

.PHONY: lint
lint:
	@$(MAKE) for-all CMD="golangci-lint run"

.PHONY: generate
generate:  $(BUILDER) ## Generate the collector code
	go tool go.opentelemetry.io/collector/cmd/builder --verbose --config manifest.yaml
	cd tmp/collector && go mod tidy

.PHONY: run
run: generate
	@cd tmp/collector && \
		./collector-framework --config ../../config.yaml
