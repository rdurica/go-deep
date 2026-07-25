.PHONY: help lesson race solutions verify project mirror fmt vet lint check all

L ?= 01
LESSON := $(shell printf 'lesson-%02d' $$((10#$(L))) 2>/dev/null || echo lesson-$(L))
LESSON_DIR := lessons/$(LESSON)

help:
	@echo "make lesson L=07     - test cvičení jedné lekce"
	@echo "make race L=44       - test cvičení s -race"
	@echo "make solutions       - všechna referenční řešení musí projít"
	@echo "make project P=01    - test projektu p01-*"
	@echo "make mirror          - přegeneruj testy v solutions/ z exercise/"
	@echo "make fmt             - gofmt -w celého repa"
	@echo "make vet             - go vet ./..."
	@echo "make lint            - golangci-lint run (pokud je nainstalovaný)"
	@echo "make check           - fmt kontrola + vet + solutions + projekty"

lesson:
	@test -d $(LESSON_DIR)/exercise || (echo "chybí $(LESSON_DIR)/exercise"; exit 1)
	cd $(LESSON_DIR)/exercise && go test -count=1 .

race:
	@test -d $(LESSON_DIR)/exercise || (echo "chybí $(LESSON_DIR)/exercise"; exit 1)
	cd $(LESSON_DIR)/exercise && go test -race -count=1 .

solutions:
	@failed=0; \
	for d in lessons/lesson-*/solutions; do \
	  if ls "$$d"/*_test.go >/dev/null 2>&1; then \
	    (cd "$$d" && go test -count=1 .) || failed=1; \
	  fi; \
	done; \
	exit $$failed

verify: check

project:
	@test -n "$(P)" || (echo "použití: make project P=01"; exit 1)
	@dir=$$(ls -d projects/p$(P)-* 2>/dev/null | head -1); \
	test -n "$$dir" || (echo "projekt p$(P) nenalezen"; exit 1); \
	cd "$$dir" && go test -count=1 ./...

mirror:
	@./scripts/mirror_tests.sh

fmt:
	@gofmt -w .

vet:
	@go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 \
		&& golangci-lint run ./... \
		|| echo "golangci-lint není nainstalovaný, přeskakuji (viz docs/tooling.md)"

check:
	@echo "==> gofmt"
	@unformatted=$$(gofmt -l . | grep -v '^$$' || true); \
	 if [ -n "$$unformatted" ]; then echo "nenaformátováno:"; echo "$$unformatted"; exit 1; fi
	@echo "==> go vet"
	@go vet ./...
	@echo "==> referenční řešení"
	@$(MAKE) --no-print-directory solutions
	@echo "==> stuby cvičení musí padat"
	@failed=0; \
	for d in lessons/lesson-*/exercise; do \
	  if (cd "$$d" && go test -count=1 . >/dev/null 2>&1); then \
	    echo "$$d prochází, ale má padat na nedokončeném TODO"; failed=1; \
	  fi; \
	done; \
	exit $$failed
	@echo "==> projekty"
	@go test -count=1 ./projects/...
	@echo "==> index"
	@python3 scripts/generate_index.py
	@echo "OK"
