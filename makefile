rootdir = $(realpath .)

bin/maildev: src/*
	GOPATH=$(rootdir) \
	go build -o bin/maildev main
