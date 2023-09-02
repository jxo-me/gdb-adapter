package gdbadapter

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/os/glog"
	"sync/atomic"
)

// Logger is the implementation for a Logger using golang log.
type Logger struct {
	Enable int32
	Ctx    context.Context
	Log    *glog.Logger
}

// EnableLog controls whether print the message.
func (l *Logger) EnableLog(enable bool) {
	i := 0
	if enable {
		i = 1
	}
	atomic.StoreInt32(&(l.Enable), int32(i))
}

// IsEnabled returns if logger is enabled.
func (l *Logger) IsEnabled() bool {
	return atomic.LoadInt32(&(l.Enable)) != 0
}

// LogModel log info related to model.
func (l *Logger) LogModel(model [][]string) {
	var str string
	for i := range model {
		for j := range model[i] {
			str += " " + model[i][j]
		}
		str += "\n"
	}
	l.Log.Info(l.Ctx, str)
}

// LogEnforce log info related to enforce.
func (l *Logger) LogEnforce(matcher string, request []interface{}, result bool, explains [][]string) {
	l.Log.Info(l.Ctx, map[string]interface{}{
		"matcher":  matcher,
		"request":  request,
		"result":   result,
		"explains": explains,
	})
}

// LogRole log info related to role.
func (l *Logger) LogRole(roles []string) {
	l.Log.Info(l.Ctx, map[string]interface{}{
		"roles": roles,
	})
}

// LogPolicy log info related to policy.
func (l *Logger) LogPolicy(policy map[string][][]string) {
	data := make(map[string]interface{}, len(policy))
	for k := range policy {
		data[k] = policy[k]
	}
	l.Log.Info(l.Ctx, data)
}

func (l *Logger) LogError(err error, msg ...string) {
	l.Log.Errorf(l.Ctx, fmt.Sprintf("error: %s, msg: %s", err.Error(), msg))
}
