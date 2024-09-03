TARGET = rsb

build: $(TARGET)

SRC = $(shell find . -name "*.go")

$(TARGET): $(SRC) go.mod
	go build -o $(TARGET) cmd/$@/main.go

install:
	cp $(TARGET) $${HOME}/.local/bin

clean:
	rm -f $(TARGET)