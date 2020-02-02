[![Build Status](https://travis-ci.org/nanobox-io/slurp.svg)](https://travis-ci.org/nanobox-io/slurp)
[![GoDoc](https://godoc.org/github.com/nanobox-io/slurp?status.svg)](https://godoc.org/github.com/nanobox-io/slurp)

# Slurp
Intermediary to the stored build/blob, used specifically to speed up publishing nanobox builds.

## Quickstart:
```sh
# Once hoarder is running, slurp can be quickly started by running:
slurp -b /tmp/build

# register a new build
curl -k https://localhost:1566/stages -d '{"new-id": "test"}'
# sync up your build (current directory)
rsync -v --delete -aR . -e 'ssh -p 1567' test@127.0.0.1:test
# tell slurp you are done syncing
curl -k https://localhost:1566/stages/test -X PUT
# Congratulations!
```
**Part II:**
```sh
# after modifying your code, register a new build
curl -k https://localhost:1566/stages -d '{"old-id": "test", "new-id": "test2"}'
# sync up your build (current directory)
rsync -v --delete -aR . -e 'ssh -p 1567' test2@127.0.0.1:this-location-really-doesnt-matter
# tell slurp you are done syncing
curl -k https://localhost:1566/stages/test2 -X PUT
# Congratulations!
```

## Usage:

### As a Server
To start slurp as a server, run:

`slurp`

An optional config file can also be passed on startup:

`slurp -c /path/to/config.json`

>config.json
```json
{
  "api-token": "secret",
  "api-address": "https://127.0.0.1:1566",
  "build-dir": "/var/db/slurp/build/",
  "insecure": true,
  "log-level": "info",
  "ssh-addr": "127.0.0.1:1567",
  "ssh-host": "/var/db/slurp/slurp_rsa",
  "store-addr": "hoarders://127.0.0.1:7410",
  "store-token": ""
}
```

`slurp -h` will show usage and a list of commands:

```
slurp - build intermediary

Usage:
  slurp [flags]

Flags:
  -a, --api-address="https://127.0.0.1:1566": Listen uri for the API (scheme defaults to https)
  -t, --api-token="secret": Token for API Access
  -b, --build-dir="/var/db/slurp/build/": Build staging directory
  -c, --config-file="": Configuration file to load
  -i, --insecure[=true]: Disable tls certificate verification when connecting to storage
  -l, --log-level="info": Log level to output [fatal|error|info|debug|trace]
  -s, --ssh-addr="127.0.0.1:1567": Address ssh server will listen on (ip:port combo)
  -k, --ssh-host="/var/db/slurp/slurp_rsa": SSH host (private) key file
  -S, --store-addr="hoarders://127.0.0.1:7410": Storage host address
  -T, --store-token="": Storage auth token
  -v, --version[=false]: Print version info and exit
```

## API:

| Method | Route | Description | Payload | Output |
| --- | --- | --- | --- | --- |
| **POST** | /stages | Stage a new build | json stage object | json auth object |
| **PUT** | /stages/:id | Commit a new build | nil | success/err message |
| **DELETE** | /stages/:id | Delete a build | nil | success/err message |
- Commit will clean up the staged build *after* pushing it to storage
- Delete will clean up the staged build *without* pushing it to storage

## Data types:

### Stage
json:
```json
{
  "old-id": "abc123",
  "new-id": "def456"
}
```
Fields:
- **old-id**: ID (in storage) of build to update
- **new-id**: ID for the new build (required)

### Auth
json:
```json
{
  "secret": "def456"
}
```
Fields:
- **secret**: Contains the username to ssh with (ID of new build)

## Todo
- rebuild auth user list on reboot
- routinely clean up undeleted builds

## Changelog
- v0.0.4 (July 26, 2016)
  - Explicitly define protocols
  - Unchanged builds keep the same md5 in hoarder

- v0.0.3 (July 22, 2016)
  - Stream data to/from hoarder, lowering memory footprint

- v0.0.2 (May 4, 2016)
  - Create config option (`store-ssl`) to enable cert verification to storage backend
  - Add `-v` and `--version` as a command

- v0.0.1 (April 25, 2016)
  - slurp is born

[![nanobox oss logo](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
