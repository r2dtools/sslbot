#!/bin/bash

service nginx restart
go test -tags="nginx common" -p=1 ./...

