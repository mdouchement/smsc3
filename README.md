# SMSC3

SMSC3 is a SMSC simulator based on SMPP3.4.

It alows to fake incoming and outgoing SMS.


## Usage

See `smsc3.go` if you want to customise the configuration via environment variables.

```sh
$ go run .
[2020-08-30 14:38:46]  INFO Listening SMPP :20001
[2020-08-30 14:38:46]  INFO Listening HTTP on :6000
```


### Example with Kannel:

1. Launch smsc3 docker container

```sh
$ docker network create kannel
$ ./run.sh
```

2. Launch Kannel

```sh
$ cd .kannel
$ docker-compose kill && docker-compose rm -f && docker-compose up
```

Test Kannel:

```sh
$ ./send-sms.sh cm
```

3. Send an incoming SMS (SMSC -> ESM)

`POST http://localhost:6000/deliver`

```json
{
    "session": "kannel-sinch",
    "from": "GOPHER",
    "to": "+33600000001",
    "message": "Hello world!",
    "-message": "Hello world! バカ"
}
```

```json
{
    "status": 200,
    "message": "OK 1U6i7TeNjcE (2)"
}
```

4. Send an outgoing SMS (ESM -> SMSC)

```sh
$ cd .kannel
$ ./send-sms.sh sinch
```


## License

**MIT**


## Contributing

All PRs are welcome.

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
