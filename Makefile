.PHONY: all clean prepare_rocksdb prepare build install test
.SILENT: prepare_rocksdb

current_dir          := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
rocksdb_ver          := 4.5
rocksdb_repo         := https://github.com/facebook/rocksdb
rocksdb_default_path := $(current_dir)/rocksdb

ROCKSDB_PATH         ?= $(rocksdb_default_path)
ROCKSDB_INCLUDE_PATH ?= $(ROCKSDB_PATH)/include
ROCKSDB_LIB_PATH     ?= $(ROCKSDB_PATH)

uname_S := $(shell sh -c 'uname -s 2>/dev/null || echo not')

CFLAGS += -I$(abspath $(ROCKSDB_INCLUDE_PATH))
LDFLAGS += -L$(abspath $(ROCKSDB_LIB_PATH))

ifeq ($(uname_S),Darwin)
	LDFLAGS += -Wl,-undefined -Wl,dynamic_lookup
else
	LDFLAGS += -Wl,-unresolved-symbols=ignore-all
endif
ifeq ($(uname_S),Linux)
	LDFLAGS += -lrt
endif

vendor := $(current_dir)/_vendor
goenv  := GOPATH="$(vendor):$(GOPATH)" CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)"

INSTALL_PATH ?= /opt/logbook

all: clean build

clean:
	rm -rf $(current_dir)/bin

prepare_rocksdb:
	if [ "$(ROCKSDB_PATH)" = "$(rocksdb_default_path)" ]; then \
		if [ ! -d $(ROCKSDB_PATH) ]; then \
			git clone $(rocksdb_repo) $(ROCKSDB_PATH) --single-branch --branch=$(rocksdb_ver).fb --depth=1; \
		fi; \
		if [ ! -e $(ROCKSDB_PATH)/librocksdb.a ]; then \
			echo "\nMaking RocksDB...\n"; \
			cd $(ROCKSDB_PATH) && make static_lib; \
		fi; \
	fi

build: prepare_rocksdb
	cd $(current_dir)
	$(goenv) go build -o bin/logbook

install:
	cd $(current_dir)
	mkdir -p $(INSTALL_PATH)/bin
	cp -r bin/logbook $(INSTALL_PATH)/bin
	cp logbook.sample.conf $(INSTALL_PATH)/logbook.conf

test: prepare_rocksdb
	cd $(current_dir)
	$(goenv) ginkgo

vendorize:
	cd $(current_dir)
	rm -rf $(vendor)
	GOPATH=$(vendor) go get -d
	find $(vendor) -name ".git" -type d | xargs rm -rf
