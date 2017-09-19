#!/usr/bin/env bash

echo "install deps"
go get github.com/revel/cmd/revel
go get github.com/revel/revel
go get -u github.com/jinzhu/gorm
go get github.com/Gr1N/revel-gorm/app
go get github.com/lib/pq
echo "deps installed"

/go/bin/revel run app