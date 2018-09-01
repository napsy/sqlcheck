# sqlcheck

```
go get github.com/napsy/sqlcheck
```

This package can check the validity of simple SQL statements by:

```go
if err := sqlcheck.NewCheck("SELECT * from names WHERE > 20;").Verify() {
	return fmt.Errorf("verifying SQL statement: %v", err)
}
/*
OUTPUT:
verifying SQL statement: left side of ">" must be one of [identifier (0..9)]
*/
```
## What's missing

This is basic stuff, the code could generate a proper AST of the SQL statements and do more advanced checking. The current implementation is only aware of SELECT and WHERE without further nesting and operators >, < and =. Also, left and right brackets are supported.
