.PHONY: install build run dbuild drun

PROJECT = apario-contribution

install:
	go mod download

build:
	time go build -a -race -v -o $(PROJECT) .

run: build
	time ./$(PROJECT) -dir tmp -file importable/$(filter-out $@$(MAKECMDGOALS)) -limit 999 -buffer 999666333

dbuild:
	docker build -t $(PROJECT) .

drun:
	docker run $(PROJECT) -v tmp:/app/tmp -v logs:/app/logs -dir tmp -file importable/$(filter-out $@$(MAKECMDGOALS)) -limit 33 -buffer 454545 -pdfcpu 1 -gs 1 -pdftotext 1 -convert 1 -pdftoppm 1 -png2jpg 1 -resize 1 -shafile 1 -watermark 1 -darkimage 1 -filedata 3 -shastring 3 -wjsonfile 3
