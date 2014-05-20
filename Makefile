GIT_REV := $(shell git rev-parse HEAD)

all:
	echo "package main\nvar Revision = \"$(GIT_REV)\"\n" > revision.go
	go build
