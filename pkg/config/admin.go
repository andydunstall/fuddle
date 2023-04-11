package config

import (
	"go.uber.org/zap/zapcore"
)

type Admin struct {
	BindAddr string
	BindPort int
}

func (c *Admin) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("bind-addr", c.BindAddr)
	e.AddInt("bind-port", c.BindPort)
	return nil
}

func DefaultAdminConfig() *Admin {
	return &Admin{
		BindAddr: "0.0.0.0",
		BindPort: 8112,
	}
}
