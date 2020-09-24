BIN_NAME=downloader

clean:
	rm -rf ./dist

build:
	go build -ldflags "-s -w" -o ./dist/${BIN_NAME}

run:
	go run main.go
