.PHONY: install build run

install:
	go mod download

build:
	go build -a -race -v -o apario-contribution .

run:
	time ./apario-contribution -dir tmp -file importable/$(filter-out $@$(MAKECMDGOALS)) -limit 999 -buffer 999666333