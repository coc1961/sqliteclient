#!/bin/sh


 go run cmd/sqliteclient/main.go -d cmd/sqliteclient/testdata/test.db  -q 'select * from datatables_demo;'

