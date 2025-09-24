#!/bin/bash

service apache2 restart
go test -tags="apache common" -p=1 ./...
