# Kannel

```sh
$ docker-compose kill && docker-compose rm -f && docker-compose up
```

```sh
# Kannels fakesmsc
$ ./send-sms.sh cm
# Golang smsc3
$ ./send-sms.sh sinch
```

- http://localhost:13000/status?password=admin

## Echo client

```sh
$ cd echo
$ go run .
```