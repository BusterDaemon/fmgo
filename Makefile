REQUIRED_FILES = (main.go cmd/app/app.go internal/*/*.go)


build REQUIRED_FILES:
	mkdir build
	go build -o build/fmgo.bin