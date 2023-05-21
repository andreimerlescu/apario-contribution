.PHONY: install build run dbuild drun

install:
	go mod download

build:
	go build -a -race -v -o apario-contribution .

run:
	time ./apario-contribution -dir tmp -file importable/$(filter-out $@$(MAKECMDGOALS)) -limit 999 -buffer 999666333

dbuild:
	docker build -t apario-contribution .

drun: dbuild
	docker run apario-contribution $(filter-out $@$(MAKECMDGOALS))
