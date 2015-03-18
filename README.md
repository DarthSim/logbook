# Logbook
[![Build Status](https://travis-ci.org/DarthSim/logbook.svg?branch=master)](https://travis-ci.org/DarthSim/logbook)

The simplest logs collector for your applications. No GUI, no graphs, no analytics. Just simple HTTP API containing two commands - `put` and `get`.

### Why Logbook?
__It's better than plain text files.__ You can store your logs in text files but searching and filtering data will bring you a lot of pain. Imagine you need some logs from the previous month. Scared? Me too. And you can just forget about showing filtered logs in your admin panel.

__It's easier than big logs collectors.__ You can use big powerful log collectors like Graylog2 or Logstash, but this means you will need MongoDB, ElasticSearch and so on. Looks like overhead, doesn't it?

## Installation
You will need Go, [Gom](https://github.com/mattn/gom) and SQLite3 to build the project.

#### Build without copying to `/opt`

```bash
make
cp logbook.conf.sample logbook.conf

# Launch Logbook
bin/logbook
```

#### Build and copy to `/opt`

```bash
make && make install

# Launch Logbook
/opt/logbook/bin/logbook
```

#### Configuration

You can specify the path to the config file using `--config` key:

```bash
/opt/logbook/bin/logbook --config /etc/logbook/logbook.conf
```

## Usage
#### Authentication
Every request to Logbook should contain basic HTTP authentication. You can find and change username and password in the config file.

#### Save log message
To save log message you need to send POST request to `/{application}/put` with the following params:

* __level__ - level of the log message
* __message__ - log message
* __tags (optional)__ - list of tags. Can be either array or string separated by comma

Example:

```bash
curl --user user:password -d "level=3&message=Lorem ipsum dolor&tags[]=tag1&tags[]=tag2&tags[]=tag3" 127.0.0.1:11610/testapp/put
```

```json
{
  "id": 1,
  "application": "testapp",
  "message": "Lorem ipsum dolor",
  "level": 3,
  "tags": ["tag1", "tag2", "tag3"],
  "created_at": "2014-08-29T20:12:07.062186202+07:00"
}
```

#### Get log messages
To get log messages you need to send GET request to `/{application}/get` with the following params:

* __level__ - minimum level of log messages
* __start_time__ - search log messages after the given datetime
* __end_time__ - search log messages before the given datetime
* __tags (optional)__ - array of tags. Can be either array or string separated by comma
* __page (optional)__ - results page. Logbook returns 100 results per page. Default page number is 1

Example:

```bash
curl --user user:password "127.0.0.1:11610/testapp/get?level=3&start_time=2014-08-01&end_time=2014-08-31&tags=tag1,tag2"
```

```json
[
  {
    "id": 1,
    "application": "testapp",
    "message": "Lorem ipsum dolor",
    "level": 3,
    "tags": ["tag1", "tag2", "tag3"],
    "created_at": "2014-08-28T18:12:07.062186202+07:00"
  },
  {
    "id": 2,
    "application": "testapp",
    "message": "Sit amet",
    "level": 4,
    "tags": ["tag1", "tag2"],
    "created_at": "2014-08-29T20:01:05.062186202+07:00"
  }
]
```

## Author

Sergey Aleksandrovich
