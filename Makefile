.PHONY: all clean prepare_rocksdb prepare build install test
.SILENT: prepare_rocksdb

current_dir          = $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
rocksdb_ver          = 4.1
rocksdb_repo         = https://github.com/facebook/rocksdb
rocksdb_default_path = $(current_dir)/rocksdb

ROCKSDB_PATH         ?= $(rocksdb_default_path)
ROCKSDB_INCLUDE_PATH ?= $(ROCKSDB_PATH)/include
ROCKSDB_LIB_PATH     ?= $(ROCKSDB_PATH)

all: clean build

clean:
	rm -rf bin/

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

prepare: CFLAGS += -I$(abspath $(ROCKSDB_INCLUDE_PATH))
prepare: LDFLAGS += $(abspath $(ROCKSDB_LIB_PATH))/librocksdb.a
prepare: LDFLAGS += -lstdc++ -lm -lz -lbz2 -lsnappy
prepare: LDFLAGS += -Wl,-undefined -Wl,dynamic_lookup
prepare: prepare_rocksdb
	cd $(current_dir)
	CGO_CFLAGS="$(CFLAGS)" CGO_LDFLAGS="$(LDFLAGS)" gom $(GOM_INSTALL_FLAGS) install

build: prepare
	cd $(current_dir)
	gom build -o bin/logbook src/*

install:
	cd $(current_dir)
	mkdir -p /opt/logbook
	cp -r bin /opt/logbook
	cp -r logbook.yml.sample /opt/logbook
	cp -r logbook.yml.sample /opt/logbook/logbook.yml

test:
	cd $(current_dir)
	gom exec ginkgo src/
