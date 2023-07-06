package gdbadapter

import (
	"context"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/util"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func testGetPolicy(t *testing.T, e *casbin.Enforcer, res [][]string) {
	myRes := e.GetPolicy()
	log.Print("Policy: ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy: ", myRes, ", supposed to be ", res)
	}
}

func testGetPolicyWithoutOrder(t *testing.T, e *casbin.Enforcer, res [][]string) {
	myRes := e.GetPolicy()
	log.Print("Policy: ", myRes)

	if !arrayEqualsWithoutOrder(myRes, res) {
		t.Error("Policy: ", myRes, ", supposed to be ", res)
	}
}

func arrayEqualsWithoutOrder(a [][]string, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}

	mapA := make(map[int]string)
	mapB := make(map[int]string)
	order := make(map[int]struct{})
	l := len(a)

	for i := 0; i < l; i++ {
		mapA[i] = util.ArrayToString(a[i])
		mapB[i] = util.ArrayToString(b[i])
	}

	for i := 0; i < l; i++ {
		for j := 0; j < l; j++ {
			if _, ok := order[j]; ok {
				if j == l-1 {
					return false
				} else {
					continue
				}
			}
			if mapA[i] == mapB[j] {
				order[j] = struct{}{}
				break
			} else if j == l-1 {
				return false
			}
		}
	}
	return true
}

func initPolicy(t *testing.T, a *Adapter) {
	// Because the DB is empty at first,
	// so we need to load the policy from the file adapter (.CSV) first.
	e, err := casbin.NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		panic(err)
	}

	// This is a trick to save the current policy to the DB.
	// We can't call e.SavePolicy() because the adapter in the enforcer is still the file adapter.
	// The current policy means the policy in the Casbin enforcer (aka in memory).
	err = a.SavePolicy(e.GetModel())
	if err != nil {
		panic(err)
	}

	// Clear the current policy.
	e.ClearPolicy()
	testGetPolicy(t, e, [][]string{})

	// Load the policy from DB.
	err = a.LoadPolicy(e.GetModel())
	if err != nil {
		panic(err)
	}
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func cleanPolicy(ctx context.Context, a *Adapter) {
	// Clear the current policy.
	if a.tableName != "" {
		_, _ = a.db.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s", a.tableName))
	}
}

func testSaveLoad(t *testing.T, a *Adapter) {
	// Initialize some policy in DB.
	initPolicy(t, a)
	// Note: you don't need to look at the above code
	// if you already have a working DB with policy inside.

	// Now the DB has policy, so we can provide a normal use case.
	// Create an adapter and an enforcer.
	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func initAdapter(t *testing.T, ctx context.Context, driverName string) *Adapter {
	// Create an adapter
	a, err := NewAdapter(ctx, driverName)
	if err != nil {
		panic(err)
	}
	listSql := []string{
		fmt.Sprintf("INSERT INTO `%s` (`id`, `p_type`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`, `v6`, `v7`) VALUES ('1', 'p', 'alice', 'data1', 'read', '', '', '', '', '')", a.tableName),
		fmt.Sprintf("INSERT INTO `%s` (`id`, `p_type`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`, `v6`, `v7`) VALUES ('2', 'p', 'bob', 'data2', 'write', '', '', '', '', '')", a.tableName),
		fmt.Sprintf("INSERT INTO `%s` (`id`, `p_type`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`, `v6`, `v7`) VALUES ('3', 'p', 'data2_admin', 'data2', 'read', '', '', '', '', '')", a.tableName),
		fmt.Sprintf("INSERT INTO `%s` (`id`, `p_type`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`, `v6`, `v7`) VALUES ('4', 'p', 'data2_admin', 'data2', 'write', '', '', '', '', '')", a.tableName),
		fmt.Sprintf("INSERT INTO `%s` (`id`, `p_type`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`, `v6`, `v7`) VALUES ('5', 'g', 'alice', 'data2_admin', '', '', '', '', '', '')", a.tableName),
		fmt.Sprintf("INSERT INTO `%s` (`id`, `p_type`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`, `v6`, `v7`) VALUES ('6', 'p', 'jack', 'data1', 'read', 'col3', 'col4', 'col5', 'col6', 'col7')", a.tableName),
		fmt.Sprintf("INSERT INTO `%s` (`id`, `p_type`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`, `v6`, `v7`) VALUES ('7', 'p', 'jack2', 'data1', 'read', 'col3', 'col4', 'col5', 'col6', 'col7')", a.tableName),
	}
	for _, sql := range listSql {
		_, _ = a.db.Exec(ctx, sql)
	}
	// Initialize some policy in DB.
	initPolicy(t, a)
	// Now the DB has policy, so we can provide a normal use case.
	// Note: you don't need to look at the above code
	// if you already have a working DB with policy inside.

	return a
}

func TestNilField(t *testing.T) {
	ctx := context.Background()
	a, err := NewAdapter(ctx, gdb.DefaultGroupName)
	assert.Nil(t, err)

	e, err := casbin.NewEnforcer("examples/rbac_model.conf", a)
	assert.Nil(t, err)
	e.EnableAutoSave(false)

	ok, err := e.AddPolicy("", "data1", "write")
	assert.Nil(t, err)
	_ = e.SavePolicy()
	assert.Nil(t, e.LoadPolicy())

	ok, err = e.Enforce("", "data1", "write")
	assert.Nil(t, err)
	assert.Equal(t, ok, true)
}

func testAutoSave(t *testing.T, a *Adapter) {

	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
	// AutoSave is enabled by default.
	// Now we disable it.
	e.EnableAutoSave(false)

	// Because AutoSave is disabled, the policy change only affects the policy in Casbin enforcer,
	// it doesn't affect the policy in the storage.
	e.AddPolicy("alice", "data1", "write")
	// Reload the policy from the storage to see the effect.
	e.LoadPolicy()
	// This is still the original policy.
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Now we enable the AutoSave.
	e.EnableAutoSave(true)

	// Because AutoSave is enabled, the policy change not only affects the policy in Casbin enforcer,
	// but also affects the policy in the storage.
	e.AddPolicy("alice", "data1", "write")
	// Reload the policy from the storage to see the effect.
	e.LoadPolicy()
	// The policy has a new rule: {"alice", "data1", "write"}.
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"alice", "data1", "write"}})

	// Remove the added rule.
	e.RemovePolicy("alice", "data1", "write")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Remove "data2_admin" related policy rules via a filter.
	// Two rules: {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"} are deleted.
	e.RemoveFilteredPolicy(0, "data2_admin")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}})
}

func testFilteredPolicy(t *testing.T, a *Adapter) {
	// NewEnforcer() without an adapter will not auto load the policy
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf")
	// Now set the adapter
	e.SetAdapter(a)

	// Load only alice's policies
	assert.Nil(t, e.LoadFilteredPolicy(Filter{V0: []string{"alice"}}))
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}})

	// Load only bob's policies
	assert.Nil(t, e.LoadFilteredPolicy(Filter{V0: []string{"bob"}}))
	testGetPolicy(t, e, [][]string{{"bob", "data2", "write"}})

	// Load policies for data2_admin
	assert.Nil(t, e.LoadFilteredPolicy(Filter{V0: []string{"data2_admin"}}))
	testGetPolicy(t, e, [][]string{{"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Load policies for alice and bob
	assert.Nil(t, e.LoadFilteredPolicy(Filter{V0: []string{"alice", "bob"}}))
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}})
}

func testUpdatePolicy(t *testing.T, a *Adapter) {
	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)

	e.EnableAutoSave(true)
	e.UpdatePolicy([]string{"alice", "data1", "read"}, []string{"alice", "data1", "write"})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "write"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func testUpdatePolicies(t *testing.T, a *Adapter) {
	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)

	e.EnableAutoSave(true)
	e.UpdatePolicies([][]string{
		{"alice", "data1", "write"},
		{"bob", "data2", "write"},
	}, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "read"},
	})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "read"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})
}

func testUpdateFilteredPolicies(t *testing.T, a *Adapter) {
	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)

	e.EnableAutoSave(true)
	e.UpdateFilteredPolicies([][]string{{"alice", "data1", "write"}}, 0, "alice", "data1", "read")
	e.UpdateFilteredPolicies([][]string{{"bob", "data2", "read"}}, 0, "bob", "data2", "write")
	e.LoadPolicy()
	testGetPolicyWithoutOrder(t, e, [][]string{{"alice", "data1", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"bob", "data2", "read"}})
}

func TestAdapters(t *testing.T) {
	ctx := context.Background()
	a := initAdapter(t, ctx, gdb.DefaultGroupName)
	testAutoSave(t, a)
	testSaveLoad(t, a)
	testFilteredPolicy(t, a)
	testUpdatePolicy(t, a)
	testUpdatePolicies(t, a)
	testUpdateFilteredPolicies(t, a)
	cleanPolicy(ctx, a)
}

func TestAddPolicies(t *testing.T) {
	ctx := context.Background()
	a := initAdapter(t, ctx, gdb.DefaultGroupName)
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
	_, err := e.AddPolicies([][]string{
		{"jack", "data1", "read"},
		{"jack2", "data1", "read"},
	})
	if err != nil {
		t.Error("AddPolicies error:", err)
	}
	err = e.LoadPolicy()
	if err != nil {
		t.Error("LoadPolicy error:", err)
	}

	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
		{"jack", "data1", "read"},
		{"jack2", "data1", "read"},
	})
	cleanPolicy(ctx, a)
}

func TestAddPoliciesFullColumn(t *testing.T) {
	ctx := context.Background()
	a := initAdapter(t, ctx, gdb.DefaultGroupName)
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
	_, err := e.AddPolicies([][]string{
		{"jack", "data1", "read", "col3", "col4", "col5", "col6", "col7"},
		{"jack2", "data1", "read", "col3", "col4", "col5", "col6", "col7"},
	})
	if err != nil {
		t.Error("AddPolicies error:", err)
	}
	err = a.LoadPolicy(e.GetModel())
	if err != nil {
		t.Error("LoadPolicy error:", err)
	}
	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
		{"jack", "data1", "read", "col3", "col4", "col5", "col6", "col7"},
		{"jack2", "data1", "read", "col3", "col4", "col5", "col6", "col7"},
	})
	cleanPolicy(ctx, a)
}
