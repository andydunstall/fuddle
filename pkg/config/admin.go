package config

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

type Admin struct {
	BindAddr string
	BindPort int

	AdvAddr string
	AdvPort int
}

func (c *Admin) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("bind-addr", c.BindAddr)
	e.AddInt("bind-port", c.BindPort)
	e.AddString("adv-addr", c.AdvAddr)
	e.AddInt("adv-port", c.AdvPort)
	return nil
}

func DefaultAdminConfig() *Admin {
	return &Admin{
		BindAddr: "0.0.0.0",
		BindPort: 8112,
		AdvAddr:  "",
		AdvPort:  8112,
	}
}

func (c *Admin) JoinBindAddr() string {
	return fmt.Sprintf("%s:%d", c.BindAddr, c.BindPort)
}

func (c *Admin) JoinAdvAddr() string {
	return fmt.Sprintf("%s:%d", c.AdvAddr, c.AdvPort)
}
