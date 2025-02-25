.DEFAULT_GOAL := goapp
APPS=server client
BIN_DIR=bin

.PHONY: all clean mkdir

all: clean $(APPS)

$(APPS): mkdir
	go build -o $(BIN_DIR)/$@ ./cmd/$@

mkdir:
	$@ -p $(BIN_DIR)

clean:
	go clean
	rm -f $(BIN_DIR)/*