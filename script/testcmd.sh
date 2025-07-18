#!/bin/bash

service nginx restart
go test -p=1 ./...

