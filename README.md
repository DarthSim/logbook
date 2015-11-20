# Logbook
[![Build Status](https://travis-ci.org/DarthSim/logbook.svg?branch=master)](https://travis-ci.org/DarthSim/logbook)

The simplest logs collector for your applications. No GUI, no graphs, no analytics. Just a simple HTTP API containing two commands - `put` and `get`.

## Why Logbook?
Logbook is a great choice when you need to log some events and then show them somewhere, and if you don't want to load your primary DB with logs. Logbook allows you to save logs pretty quickly and get them back filtered by time, level and tags.

Logbook doesn't need any additional software (except Go for compilation of course).

## Installation
You need Go 1.4+ and [Gom](https://github.com/mattn/gom) to build the project.

#### Build without copying to `/opt`

```bash
make
cp logbook.yml.sample logbook.yml

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
/opt/logbook/bin/logbook --config /etc/logbook/logbook.yml
```

## Usage
#### Authentication
Every request to Logbook should contain HTTP basic auth. You can find and change username and password in the config file.

#### Save log message
To save log message you need to send POST request to `/{application}/put` with the following params:

Param      | Description
-----------|------------
level      | Level of the log message
message    | Log message
tags       | _(optional)_ String of tags separated by the comma
created_at | _(optional)_ Datetime of record (format: `YYYY-MM-DDThh:mm:ss[.sss][±hh:mm]`). Default: current time.

Example:

```bash
curl --user user:password -d "level=3&message=Lorem ipsum dolor&tags=tag1,tag2,tag3&created_at=2014-08-29T20:12:07.062+07:00" 127.0.0.1:11610/testapp/put
```

```json
{
  "message": "Lorem ipsum dolor",
  "level": 3,
  "tags": ["tag1", "tag2", "tag3"],
  "created_at": "2014-08-29T20:12:07.062+07:00"
}
```

#### Get log messages
To get log messages you need to send GET request to `/{application}/get` with the following params:

Param      | Description
-----------|------------
level      | Minimum level of log messages
start_time | Search log messages after the given DateTime.<br/>Format: `YYYY-MM-DD` or `YYYY-MM-DDThh:mm:ss[.sss][±hh:mm]`
end_time   | Search log messages before the given DateTime.<br/>Format: `YYYY-MM-DD` or `YYYY-MM-DDThh:mm:ss[.sss][±hh:mm]`
tags       | _(optional)_ String of required tags separated by comma
page       | _(optional)_ Results page. Logbook returns 100 results per page by default (you can change this number in the config file). Default page number is 1

Example:

```bash
curl --user user:password "127.0.0.1:11610/testapp/get?level=3&start_time=2014-08-01&end_time=2014-08-31&tags=tag1,tag2"
```

```json
[
  {
    "message": "Lorem ipsum dolor",
    "level": 3,
    "tags": ["tag1", "tag2", "tag3"],
    "created_at": "2014-08-28T18:12:07.062186202+07:00"
  },
  {
    "message": "Sit amet",
    "level": 4,
    "tags": ["tag1", "tag2"],
    "created_at": "2014-08-29T20:01:05.062186202+07:00"
  }
]
```

### Notes and Limitations

* Logbook is designed for a fast saving. Fetching is still fast, but not as fast as saving.
* As in any other DB, fetching records with an offset just skips first N records, which makes pagination a little bit expensive on a big page number. Keep this in mind while fetching the 1000th page.
* Logbook uses Bolt and this means that it meets Bolt's limitations:
  > Bolt uses a memory-mapped file, so the underlying operating system handles the caching of the data. Typically, the OS will cache as much of the file as it can in memory and will release memory as needed to other processes. This means that Bolt can show very high memory usage when working with large databases. However, this is expected, and the OS will release memory as needed. Bolt can handle databases much larger than the available physical RAM, provided its memory-map fits in the process virtual address space. It may be problematic on 32-bits systems.

  > The data structures in the Bolt database are memory mapped so the data file will be endian-specific. This means that you cannot copy a Bolt file from a little endian machine to a big endian machine and have it work. For most users, this is not a concern since most modern CPUs are little endian.

  > Because of the way pages are laid out on disk, Bolt cannot truncate data files and return free pages back to the disk. Instead, Bolt maintains a free list of unused pages within its data file. These free pages can be reused by later transactions. This works well for many use cases as databases generally tend to grow. However, it's important to note that deleting large chunks of data will not allow you to reclaim that space on disk.

## Author

Sergey Aleksandrovich
