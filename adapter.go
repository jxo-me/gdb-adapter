package gdbadapter

import (
	"context"
	"errors"
	"fmt"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"runtime"
)

const (
	defaultTableName = "sys_casbin_rule"
	flushEvery       = 1000
)

type CasbinRule struct {
	ID    uint   `orm:"id" json:"id"`
	PType string `orm:"p_type" json:"p_type"`
	V0    string `orm:"v0" json:"v0"`
	V1    string `orm:"v1" json:"v1"`
	V2    string `orm:"v2" json:"v2"`
	V3    string `orm:"v3" json:"v3"`
	V4    string `orm:"v4" json:"v4"`
	V5    string `orm:"v5" json:"v5"`
	V6    string `orm:"v6" json:"v6"`
	V7    string `orm:"v7" json:"v7"`
}

func (CasbinRule) TableName() string {
	return "sys_casbin_rule"
}

func (c *CasbinRule) queryString() (interface{}, []interface{}) {
	queryArgs := []interface{}{c.PType}

	queryStr := "p_type = ?"
	if c.V0 != "" {
		queryStr += " and v0 = ?"
		queryArgs = append(queryArgs, c.V0)
	}
	if c.V1 != "" {
		queryStr += " and v1 = ?"
		queryArgs = append(queryArgs, c.V1)
	}
	if c.V2 != "" {
		queryStr += " and v2 = ?"
		queryArgs = append(queryArgs, c.V2)
	}
	if c.V3 != "" {
		queryStr += " and v3 = ?"
		queryArgs = append(queryArgs, c.V3)
	}
	if c.V4 != "" {
		queryStr += " and v4 = ?"
		queryArgs = append(queryArgs, c.V4)
	}
	if c.V5 != "" {
		queryStr += " and v5 = ?"
		queryArgs = append(queryArgs, c.V5)
	}
	if c.V6 != "" {
		queryStr += " and v6 = ?"
		queryArgs = append(queryArgs, c.V6)
	}
	if c.V7 != "" {
		queryStr += " and v7 = ?"
		queryArgs = append(queryArgs, c.V7)
	}

	return queryStr, queryArgs
}

func (c *CasbinRule) toStringPolicy() []string {
	policy := make([]string, 0)
	if c.PType != "" {
		policy = append(policy, c.PType)
	}
	if c.V0 != "" {
		policy = append(policy, c.V0)
	}
	if c.V1 != "" {
		policy = append(policy, c.V1)
	}
	if c.V2 != "" {
		policy = append(policy, c.V2)
	}
	if c.V3 != "" {
		policy = append(policy, c.V3)
	}
	if c.V4 != "" {
		policy = append(policy, c.V4)
	}
	if c.V5 != "" {
		policy = append(policy, c.V5)
	}
	if c.V6 != "" {
		policy = append(policy, c.V6)
	}
	if c.V7 != "" {
		policy = append(policy, c.V7)
	}
	return policy
}

type Filter struct {
	PType []string
	V0    []string
	V1    []string
	V2    []string
	V3    []string
	V4    []string
	V5    []string
	V6    []string
	V7    []string
}

// Adapter represents the Gorm adapter for policy store.
type Adapter struct {
	dbGroupName string
	tableName   string
	db          gdb.DB
	ctx         context.Context
	isFiltered  bool
}

// finalizer is the destructor for Adapter.
func finalizer(a *Adapter) {
	a.db = nil
}

// NewAdapter is the constructor for Adapter.
func NewAdapter(ctx context.Context, groupName string) (*Adapter, error) {
	a := &Adapter{}
	a.dbGroupName = groupName
	a.tableName = defaultTableName
	a.ctx = ctx
	// Open the DB, create it if not existed.
	err := a.open()
	if err != nil {
		return nil, err
	}

	// Call the destructor when the object is released.
	runtime.SetFinalizer(a, finalizer)

	return a, nil
}

func (a *Adapter) open() error {
	a.db = g.DB(a.dbGroupName)
	return a.createTable()
}

func (a *Adapter) close() error {
	a.db = nil
	return nil
}

// getTableInstance return the dynamic table name
func (a *Adapter) getTableInstance() *CasbinRule {
	return &CasbinRule{}
}

// HasTable determine whether the table name exists in the database.
func (a *Adapter) HasTable(name string) (bool, error) {
	tableList, err := a.db.Tables(a.ctx)
	if err != nil {
		return false, err
	}
	for _, table := range tableList {
		if table == name {
			return true, nil
		}
	}
	return false, nil
}

func (a *Adapter) createTable() error {
	if exists, _ := a.HasTable(a.tableName); exists {
		return nil
	}
	_, err := a.db.Exec(a.ctx, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (`id` bigint unsigned NOT NULL AUTO_INCREMENT,`p_type` VARCHAR(100),`v0` VARCHAR(100),`v1` VARCHAR(100),`v2` VARCHAR(100),`v3` VARCHAR(100),`v4` VARCHAR(100),`v5` VARCHAR(100), `v6` VARCHAR(25), `v7` VARCHAR(25),PRIMARY KEY (`id`),UNIQUE KEY `idx_%s` (`p_type`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`,`v6`,`v7`))", a.tableName, a.tableName))
	return err
}

func (a *Adapter) dropTable() error {
	_, err := a.db.Exec(a.ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", a.tableName))
	return err
}

func (a *Adapter) truncateTable() error {
	_, err := a.db.Exec(a.ctx, fmt.Sprintf("TRUNCATE TABLE %s", a.tableName))
	return err
}

func loadPolicyLine(line CasbinRule, model model.Model) {
	var p = []string{line.PType,
		line.V0, line.V1, line.V2,
		line.V3, line.V4, line.V5,
		line.V6, line.V7}
	index := len(p) - 1
	for p[index] == "" {
		index--
	}
	index += 1
	p = p[:index]

	persist.LoadPolicyArray(p, model)
}

// LoadPolicy loads policy from database.
func (a *Adapter) LoadPolicy(model model.Model) error {
	var lines []CasbinRule
	if err := a.db.Model(a.tableName).Order("id").Scan(&lines); err != nil {
		return err
	}
	for _, line := range lines {
		loadPolicyLine(line, model)
	}

	return nil
}

// LoadFilteredPolicy loads only policy rules that match the filter.
func (a *Adapter) LoadFilteredPolicy(model model.Model, filter interface{}) error {
	var lines []CasbinRule

	filterValue, ok := filter.(Filter)
	if !ok {
		return errors.New("invalid filter type")
	}
	db := a.db.Model(a.tableName).Safe().Ctx(a.ctx)
	if len(filterValue.PType) > 0 {
		db = db.WhereIn("p_type", filterValue.PType)
	}
	if len(filterValue.V0) > 0 {
		db = db.WhereIn("v0", filterValue.V0)
	}
	if len(filterValue.V1) > 0 {
		db = db.WhereIn("v1", filterValue.V1)
	}
	if len(filterValue.V2) > 0 {
		db = db.WhereIn("v2", filterValue.V2)
	}
	if len(filterValue.V3) > 0 {
		db = db.WhereIn("v3", filterValue.V3)
	}
	if len(filterValue.V4) > 0 {
		db = db.WhereIn("v4", filterValue.V4)
	}
	if len(filterValue.V5) > 0 {
		db = db.WhereIn("v5", filterValue.V5)
	}
	if len(filterValue.V6) > 0 {
		db = db.WhereIn("v6", filterValue.V6)
	}
	if len(filterValue.V7) > 0 {
		db = db.WhereIn("v7", filterValue.V7)
	}
	if err := db.Order("id").Scan(&lines); err != nil {
		return err
	}

	for _, line := range lines {
		loadPolicyLine(line, model)
	}
	a.isFiltered = true

	return nil
}

// IsFiltered returns true if the loaded policy has been filtered.
func (a *Adapter) IsFiltered() bool {
	return a.isFiltered
}

func (a *Adapter) savePolicyLine(ptype string, rule []string) CasbinRule {
	line := a.getTableInstance()

	line.PType = ptype
	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}
	if len(rule) > 6 {
		line.V6 = rule[6]
	}
	if len(rule) > 7 {
		line.V7 = rule[7]
	}

	return *line
}

// SavePolicy saves policy to database.
func (a *Adapter) SavePolicy(model model.Model) error {
	err := a.truncateTable()
	if err != nil {
		return err
	}
	var lines []CasbinRule
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			lines = append(lines, a.savePolicyLine(ptype, rule))
			if len(lines) > flushEvery {
				_, err = a.db.Model(a.tableName).Data(&lines).Insert()
				if err != nil {
					return err
				}
				lines = nil
			}
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			lines = append(lines, a.savePolicyLine(ptype, rule))
			if len(lines) > flushEvery {
				_, err = a.db.Model(a.tableName).Data(&lines).Insert()
				if err != nil {
					return err
				}
				lines = nil
			}
		}
	}

	if len(lines) > 0 {
		_, err = a.db.Model(a.tableName).Data(&lines).Insert()
		if err != nil {
			return err
		}
	}

	return nil
}

// AddPolicy adds a policy rule to the store.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	line := a.savePolicyLine(ptype, rule)
	_, err := a.db.Model(a.tableName).Data(&line).Insert()
	return err
}

// RemovePolicy removes a policy rule from the store.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	tx, err := a.db.Begin(a.ctx)
	if err != nil {
		panic(err)
	}
	line := a.savePolicyLine(ptype, rule)
	err = a.rawDelete(tx, line)
	return err
}

// AddPolicies adds multiple policy rules to the store.
func (a *Adapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	var lines []CasbinRule
	for _, rule := range rules {
		lines = append(lines, a.savePolicyLine(ptype, rule))
	}
	if len(lines) > 0 {
		_, err := a.db.Model(a.tableName).Data(&lines).Insert()
		if err != nil {
			return err
		}
	}
	return nil
}

// RemovePolicies removes multiple policy rules from the store.
func (a *Adapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	return a.db.Transaction(a.ctx, func(ctx context.Context, tx *gdb.TX) error {
		for _, rule := range rules {
			line := a.savePolicyLine(ptype, rule)
			if err := a.rawDelete(tx, line); err != nil {
				return err
			}
		}
		return nil
	})
}

// RemoveFilteredPolicy removes policy rules that match the filter from the store.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	tx, err := a.db.Begin(a.ctx)
	if err != nil {
		panic(err)
	}
	line := a.getTableInstance()

	line.PType = ptype
	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		line.V0 = fieldValues[0-fieldIndex]
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		line.V1 = fieldValues[1-fieldIndex]
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		line.V2 = fieldValues[2-fieldIndex]
	}
	if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) {
		line.V3 = fieldValues[3-fieldIndex]
	}
	if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) {
		line.V4 = fieldValues[4-fieldIndex]
	}
	if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) {
		line.V5 = fieldValues[5-fieldIndex]
	}
	if fieldIndex <= 6 && 6 < fieldIndex+len(fieldValues) {
		line.V6 = fieldValues[6-fieldIndex]
	}
	if fieldIndex <= 7 && 7 < fieldIndex+len(fieldValues) {
		line.V7 = fieldValues[7-fieldIndex]
	}
	err = a.rawDelete(tx, *line)
	return err
}

func (a *Adapter) rawDelete(tx *gdb.TX, line CasbinRule) error {
	db := tx.Model(a.tableName).Safe()
	condition := gdb.Map{"p_type": line.PType}
	if line.V0 != "" {
		condition["v0"] = line.V0
	}
	if line.V1 != "" {
		condition["v1"] = line.V1
	}
	if line.V2 != "" {
		condition["v2"] = line.V2
	}
	if line.V3 != "" {
		condition["v3"] = line.V3
	}
	if line.V4 != "" {
		condition["v4"] = line.V4
	}
	if line.V5 != "" {
		condition["v5"] = line.V5
	}
	if line.V6 != "" {
		condition["v6"] = line.V6
	}
	if line.V7 != "" {
		condition["v7"] = line.V7
	}
	if _, err := db.Delete(condition); err != nil {
		return tx.Rollback()
	}
	return tx.Commit()
}

// UpdatePolicy updates a new policy rule to DB.
func (a *Adapter) UpdatePolicy(sec string, ptype string, oldRule, newPolicy []string) error {
	oldLine := a.savePolicyLine(ptype, oldRule)
	newLine := a.savePolicyLine(ptype, newPolicy)
	_, err := a.db.Model(a.tableName).Where(&oldLine).Data(newLine).OmitEmpty().Update()
	if err != nil {
		return err
	}
	return nil
}

func (a *Adapter) UpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	oldPolicies := make([]CasbinRule, 0, len(oldRules))
	newPolicies := make([]CasbinRule, 0, len(oldRules))
	for _, oldRule := range oldRules {
		oldPolicies = append(oldPolicies, a.savePolicyLine(ptype, oldRule))
	}
	for _, newRule := range newRules {
		newPolicies = append(newPolicies, a.savePolicyLine(ptype, newRule))
	}
	tx, err := a.db.Begin(a.ctx)
	if err != nil {
		panic(err)
	}
	for i := range oldPolicies {
		if _, err = tx.Model(a.tableName).Where(&oldPolicies[i]).Data(newPolicies[i]).OmitEmpty().Update(); err != nil {
			return tx.Rollback()
		}
	}
	return tx.Commit()
}

func (a *Adapter) UpdateFilteredPolicies(sec string, ptype string, newPolicies [][]string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	// UpdateFilteredPolicies deletes old rules and adds new rules.
	line := a.getTableInstance()

	line.PType = ptype
	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		line.V0 = fieldValues[0-fieldIndex]
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		line.V1 = fieldValues[1-fieldIndex]
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		line.V2 = fieldValues[2-fieldIndex]
	}
	if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) {
		line.V3 = fieldValues[3-fieldIndex]
	}
	if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) {
		line.V4 = fieldValues[4-fieldIndex]
	}
	if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) {
		line.V5 = fieldValues[5-fieldIndex]
	}
	if fieldIndex <= 6 && 6 < fieldIndex+len(fieldValues) {
		line.V6 = fieldValues[6-fieldIndex]
	}
	if fieldIndex <= 7 && 7 < fieldIndex+len(fieldValues) {
		line.V7 = fieldValues[7-fieldIndex]
	}

	newP := make([]CasbinRule, 0, len(newPolicies))
	oldP := make([]CasbinRule, 0)
	for _, newRule := range newPolicies {
		newP = append(newP, a.savePolicyLine(ptype, newRule))
	}
	tx, err := a.db.Begin(a.ctx)
	if err != nil {
		panic(err)
	}

	for i := range newP {
		str, args := line.queryString()
		if err = tx.Model(a.tableName).Where(str, args...).Scan(&oldP); err != nil {
			err = tx.Rollback()
			if err != nil {
				return nil, err
			}
			return nil, err
		}
		if _, err = tx.Model(a.tableName).Where(str, args...).Delete([]CasbinRule{}); err != nil {
			err = tx.Rollback()
			if err != nil {
				return nil, err
			}
			return nil, err
		}
		if _, err = tx.Model(a.tableName).Data(&newP[i]).Insert(); err != nil {
			err = tx.Rollback()
			if err != nil {
				return nil, err
			}
			return nil, err
		}
	}

	// return deleted rulues
	oldPolicies := make([][]string, 0)
	for _, v := range oldP {
		oldPolicy := v.toStringPolicy()
		oldPolicies = append(oldPolicies, oldPolicy)
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return oldPolicies, err
}
