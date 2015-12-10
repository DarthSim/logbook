.PHONY: all clean prepare_rocksdb prepare build install test
.SILENT: prepare_rocksdb

current_dir          := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
rocksdb_ver          := 4.1
rocksdb_repo         := https://github.com/facebook/rocksdb
rocksdb_default_path := $(current_dir)/rocksdb

ROCKSDB_PATH         ?= $(rocksdb_default_path)
ROCKSDB_INCLUDE_PATH ?= $(ROCKSDB_PATH)/include
ROCKSDB_LIB_PATH     ?= $(ROCKSDB_PATH)

uname_S := $(shell sh -c 'uname -s 2>/dev/null || echo not')

CFLAGS += -I$(abspath $(ROCKSDB_INCLUDE_PATH))
LDFLAGS += $(abspath $(ROCKSDB_LIB_PATH))/librocksdb.a
LDFLAGS += -lstdc++ -lm -lz -lbz2 -lsnappy

vendor := $(current_dir)/_vendor
goenv  := GOPATH="$(vendor):$(GOPATH)" CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)"

ifeq ($(uname_S),Darwin)
	LDFLAGS += -Wl,-undefined -Wl,dynamic_lookup
else
	LDFLAGS += -Wl,-unresolved-symbols=ignore-all
endif
ifeq ($(uname_S),Linux)
	LDFLAGS += -lrt
endif

all: clean build

clean:
	rm -rf $(current_dir)/bin

prepare_rocksdb:
	if [ "$(ROCKSDB_PATH)" = "$(rocksdb_default_path)" ]; then \
		if [ ! -d $(ROCKSDB_PATH) ]; then \
			git clone $(rocksdb_repo) $(ROCKSDB_PATH) --single-branch --branch=v$(rocksdb_ver) --depth=1; \
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
	mkdir -p /opt/logbook
	cp -r bin /opt/logbook
	cp -r logbook.yml.sample /opt/logbook
	cp -r logbook.yml.sample /opt/logbook/logbook.yml

test: prepare_rocksdb
	cd $(current_dir)
	$(goenv) ginkgo

vendorize:
	cd $(current_dir)
	rm -rf $(vendor)
	GOPATH=$(vendor) go get -d
	find $(vendor) -name ".git" -type d | xargs rm -rf
