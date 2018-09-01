# sqlcheck

```
go get github.com/napsy/sqlcheck
```

This package can check the validity of simple SQL statements by:

```go
if err := sqlcheck.NewCheck("SELECT * from names WHERE > 20;") {
	return fmt.Errorf("verifying SQL statement: %v", err)
}
```