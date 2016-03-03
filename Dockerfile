FROM debian:jessie
MAINTAINER Sergey Aleksandrovich <darthsim@gmail.com>

EXPOSE 11610

COPY . /app
WORKDIR /app

RUN buildDeps='g++ gcc libc6-dev make ca-certificates curl git libgflags-dev' \
  && runDeps='libsnappy-dev zlib1g-dev libbz2-dev' \
  && goDownloadUrl="https://golang.org/dl/go1.6.linux-amd64.tar.gz" \
  && goDownloadSha256="5470eac05d273c74ff8bac7bef5bad0b5abbd1c4052efbdbc8db45332e836b0b" \
  && apt-get update \
  && apt-get install -y --no-install-recommends $buildDeps $runDeps \
  && curl -fsSL "$goDownloadUrl" -o golang.tar.gz \
  && echo "$goDownloadSha256  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
  && PATH="/usr/local/go/bin:$PATH" GOPATH="/go" make \
  && apt-get purge -y --auto-remove $buildDeps \
  && rm -rf /var/lib/apt/lists/* \
  && rm -rf rocksdb \
  && rm -rf golang.tar.gz /usr/local/go

RUN cp logbook.sample.conf /logbook.conf \
  && mkdir /data \
  && ln -s /data /app/db

ENTRYPOINT ["/app/bin/logbook"]
CMD ["--config", "/logbook.conf"]
