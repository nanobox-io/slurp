## Sample Usage Demo
This demo has [Hoarder](github.com/nanopack/hoarder) running locally

### create some thing
```sh
dd if=/dev/urandom of=thing bs=1M count=10
```

#### tell slurp about new build
```sh
curl -k -H "X-AUTH-TOKEN: secret" https://localhost:1566/stages -d '{"new-id": "test"}'
```

#### sync files
```sh
rsync -v --delete -aR . -e 'ssh -p 1567' test@127.0.0.1:test/
```

#### commit update (done syncing)
```sh
curl -k -H "X-AUTH-TOKEN: secret" https://localhost:1566/stages/test -X PUT
```

#### delete temp build dir
```sh
curl -k -H "X-AUTH-TOKEN: secret" https://localhost:1566/stages/test -X DELETE
```

#### *get build from hoarder
```sh
curl localhost:7410/blobs/test | tar -zxf -
du -h thing
# 10M thing
```

### make changes
```sh
echo '{things"}' > file
```

#### tell slurp about new build
```sh
curl -k -H "X-AUTH-TOKEN: secret" https://localhost:1566/stages -d '{"old-id": "test", "new-id": "test2"}'
```

#### sync files
```sh
rsync --delete -aR . -e 'ssh -p 1567' test2@127.0.0.1:this-location-really-doesnt-matter
```

#### commit update (done syncing)
```sh
curl -k -H "X-AUTH-TOKEN: secret" https://localhost:1566/stages/test2 -X PUT
```

#### *get build from hoarder
```sh
curl localhost:7410/blobs/test2 | tar -zxf -
ls
# file thing
```
- *new directory
