module github.com/mdouchement/smsc3

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/fiorix/go-smpp v0.0.0-20181129163705-6dbf72b9bcea
	github.com/goburrow/cache v0.1.1-0.20200221222329-5a7015e20557
	github.com/mdouchement/basex v0.0.0-20200802103314-a4f42a0e6590
	github.com/mdouchement/logger v0.0.0-20200719134033-63d0eda5567d
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
)

replace github.com/fiorix/go-smpp => github.com/mdouchement/go-smpp v0.0.0-20200830150802-c576c3524752
