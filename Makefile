REPO = $(shell git remote -v | grep '^origin\s.*(fetch)$$' | awk '{print $$2}' | sed -E 's/^.*(\/\/|@)//;s/\.git$$//' | sed 's/:/\//g')
SRC = $(shell find . -name "*.go")
VENDER_FILE = go.mod
GO = go
TARGET = rsb
COMMIT_ID ?= $(shell git rev-parse --short HEAD)
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
VERSION ?= 0.0.1
LDFLAGS += -X "$(REPO)/pkg/version.BuildTS=$(BUILT_TS)"
LDFLAGS += -X "$(REPO)/pkg/version.GitHash=$(COMMIT_ID)"
LDFLAGS += -X "$(REPO)/pkg/version.Version=$(VERSION)"
LDFLAGS += -X "$(REPO)/pkg/version.GitBranch=$(BRANCH)"


build: $(TARGET)

$(TARGET): $(SRC) $(VENDER_FILE)
	$(GO) build -ldflags '${LDFLAGS} -X "$(REPO)/pkg/version.App=$@"' -o $@$(BINARY_SUFFIX) $(REPO)/cmd/$@/

install:
	cp $(TARGET) $${HOME}/.local/bin

clean:
	rm -f $(TARGET)