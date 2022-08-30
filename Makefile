# Main Makefile

GOFMT_FILES?=$$(find . -name '*.go')

DEFAULT: fmt

fmt:
	gofmt -w $(GOFMT_FILES)
	goimports -w $(GOFMT_FILES)

clean:
	@ABS_INSTALL_TO=$(INSTALL_DIR) sh -c "'$(CURDIR)/scripts/clean.sh'"

.NOTPARALLEL:

.PHONY: fmt \
	clean