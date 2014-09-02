all: clean build

clean:
	rm -rf bin/

build:
	gom install && gom build -o bin/logbook src/*

install:
	mkdir -p /opt/logbook
	cp -r bin /opt/logbook
	cp -r logbook.conf.sample /opt/logbook
	cp -r logbook.conf.sample /opt/logbook/logbook.conf
