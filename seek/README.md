automation tool for cv generation.

# program builds

go module builds two programs:

1. server - http server accepts json

2. cli - program accept standard input

# usage

`make` - builds server and cli

`make server` - builds just the server

`make cli` - builds just the cli

`make clean` - removes server and cli program binaries in root dir

---

`GEMINIAPIKEY={key xxxx} ./server`

`GEMINIAPIKEY={key xxxx} ./cli [job desciption text]`




