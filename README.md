# sqliteclient Test Client (no production)

> Show query result in console as table, database and query are params

## Use

```sh
sqliteclient -d cmd/sqliteclient/testdata/test.db -q 'select * from datatables_demo;'

```
```sh
sqliteclient -d cmd/sqliteclient/testdata/test.db -f query.sql

```

## Install

 go get -u github.com/coc1961/sqliteclient/...

## Screenshot

![alt](doc/screen.png)