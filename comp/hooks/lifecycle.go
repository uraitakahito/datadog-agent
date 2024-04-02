package fxhelper

import "context"

type lcHookFunc func(context.Context) error

type Lifecycle struct {
	OnStart lcHookFunc
	OnStop  lcHookFunc
}
