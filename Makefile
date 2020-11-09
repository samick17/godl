MAKEFLAGS += --silent
BIN_NAME=downloader

clean:
	rm -rf ./dist

build:
	go build -ldflags "-s -w" -o ./dist/${BIN_NAME}

run:
	go run main.go

test:
	make build
	./dist/${BIN_NAME}
# 	./dist/${BIN_NAME} -version=85 -platform=linux
# 	./dist/${BIN_NAME} -version=latest -platform=linux
