package kklogger

import (
	"testing"
)

func TestDebug(t *testing.T) {
	type args struct {
		args []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "struct", args: args{args: []interface{}{JsonMsg{
			Type: "type",
			Data: "data",
		}}}},
		{name: "simple", args: args{args: []interface{}{"simple data"}}},
		{name: "test", args: args{args: []interface{}{"type", "data"}}},
	}

	SetLogLevel("TRACE")
	SetLoggerHooks([]LoggerHook{&DefaultLoggerHook{}})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Debug(tt.args.args)
		})
	}
}
