# SQL Race

[Race condition](https://en.wikipedia.org/wiki/Race_condition) is a well known kind of behavior, which usually ends up being a bug. And if you work with databases, the probability of stumbling across such behavior is not zero at all. Perhaps there is a misconception that this bug happens only in some strange low level code that bearded programmers know how to write. Sadly, it happens in high level code as well. But when it comes to databases, there is a common idea that transactions exist to defeat race conditions. However, using transactions alone is not enough.

If you still believe transactions alone defeat race conditions, try this demo.

Compile the testing app using [gb](http://getgb.io) tool by running: `gb build`.

Alternatively you can use the native go tool: `GOPATH=$PWD:$PWD/vendor go install sqlrace`

(run from repository root dir, it will not work with `go get`).

By default it connects to the DB under the `root` user **without a password** and connects to `127.0.0.1:3306`, uses `test` database. If you have a different testing environment setup, edit the connection string in `src/sqlrace/main.go`.

Usage example:

```
[nsf @ sqlrace]$ ./bin/sqlrace -h
Usage of ./bin/sqlrace:
  -g int
        number of goroutines (default 4)
  -m string
        decrement method: naive, transaction, locked, atomic (default "naive")
  -n int
        number of iterations per goroutine (default 1024)
```
```
[nsf @ sqlrace]$ ./bin/sqlrace -m naive
INFO[0000] Initial counter state: 4096
INFO[0000] Number of decrements per goroutine: 1024
INFO[0000] Number of goroutines: 4
INFO[0000] Method: naive
INFO[0000] Method description:
SELECT * FROM table;
UPDATE table SET ...;
INFO[0000] Result: 2864
```
```
[nsf @ sqlrace]$ ./bin/sqlrace -m transaction
INFO[0000] Initial counter state: 4096
INFO[0000] Number of decrements per goroutine: 1024
INFO[0000] Number of goroutines: 4
INFO[0000] Method: transaction
INFO[0000] Method description:
START TRANSACTION;
SELECT * FROM table;
UPDATE table SET ...;
COMMIT;
INFO[0000] Result: 2712
```
```
[nsf @ sqlrace]$ ./bin/sqlrace -m locked
INFO[0000] Initial counter state: 4096
INFO[0000] Number of decrements per goroutine: 1024
INFO[0000] Number of goroutines: 4
INFO[0000] Method: locked
INFO[0000] Method description:
START TRANSACTION;
SELECT * FROM table FOR UPDATE;
UPDATE table SET ...;
COMMIT;
INFO[0000] Result: 0
```
```
[nsf @ sqlrace]$ ./bin/sqlrace -m atomic
INFO[0000] Initial counter state: 4096
INFO[0000] Number of decrements per goroutine: 1024
INFO[0000] Number of goroutines: 4
INFO[0000] Method: atomic
INFO[0000] Method description:
UPDATE table SET counter = counter - 1;
INFO[0000] Result: 0
```

## Conclusion

As you can see transactions alone **do not defeat race conditions**, however some people still believe in this nonsense. If you're one of them, please stop.

Btw, atomic one is the fastest one.

## Notes

The code was tested on a typical linux machine.

The go version used is `1.5`.

The database used is `MariaDB 10.0.21`.

My testing database runs on `/dev/shm`, this may affect the end result, but doesn't affect correctness.

Third-party packages are provided for repeatability of the build, they come with their own licenses. All the code in `src` directory is under public domain.
