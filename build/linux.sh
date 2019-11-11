#! /bin/sh

now=$(date +%Y%d%m)
go install -ldflags "-X main.version=$now"
