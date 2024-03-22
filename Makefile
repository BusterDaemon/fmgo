SOURCES := $(wildcard *.go)

build: $(SOURCES)
	mkdir -p build
	go build -o build/fmgo.bin ./main.go

clean:
	rm -rf ./build