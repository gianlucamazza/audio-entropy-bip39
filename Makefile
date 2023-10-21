# Name of the binary output
BINARY_NAME=audio-entropy-bip39

# Command to fetch dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download

# Command to build the project
.PHONY: build
build: deps
	@echo "Building..."
	go build -o dist/$(BINARY_NAME) cmd/$(BINARY_NAME)/main.go

# Command to run the project
.PHONY: run
run: build
	@echo "Running..."
	./dist/$(BINARY_NAME)

# Command to run tests (if any)
.PHONY: test
test: deps
	@echo "Testing..."
	go test -v ./...

# Command to clean up generated files
.PHONY: clean
clean:
	@echo "Cleaning..."
	go clean
	rm -f dist/$(BINARY_NAME)

# Default command if only 'make' is run
all: run
