all: clean build

clean:
	rm -rf bin/

build:
	gom install && gom build -o bin/logbook src/*

install:
	mkdir -p /opt/logbook
	cp -r bin /opt/logbook
	cp -r logbook.yml.sample /opt/logbook
	cp -r logbook.yml.sample /opt/logbook/logbook.yml

test:
	gom exec ginkgo src/
