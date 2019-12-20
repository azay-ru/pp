#! /bin/sh

go install -ldflags "-X github.com/azay-ru/pp/app.version=$(date +%Y%d%m)"
