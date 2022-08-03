Gorm Adapter
====

Gdb Adapter is the goframe orm adapter for Casbin.

## Installation

    go get github.com/jxo-me/gdb-adapter

## Simple Example

```go
package main

import (
	"context"
	"github.com/casbin/casbin/v2"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	gdbadapter "github.com/jxo-me/gdb-adapter"
)

func main() {
	// Initialize a gdb adapter and use it in a Casbin enforcer:
	// The adapter will use the MySQL database named "casbin".
	// If it doesn't exist, the adapter will create it automatically.
	ctx := context.Background()
	a, _ := gdbadapter.NewAdapter(ctx, gdb.DefaultGroupName) // Your driver and data source.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)

	// Load the policy from DB.
	_ = e.LoadPolicy()

	// Check the permission.
	e.Enforce("alice", "data1", "read")

	// Modify the policy.
	// e.AddPolicy(...)
	// e.RemovePolicy(...)

	// Save the policy back to DB.
	_ = e.SavePolicy()
}
```

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
