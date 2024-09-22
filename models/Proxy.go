package models

import "github.com/uptrace/bun"

type ProxyStatus int

const (
	ProxyEnable  ProxyStatus = 1
	ProxyDisable ProxyStatus = 2
)

type Proxy struct {
	bun.BaseModel `bun:"table:contacts"`

	ID   int64  `bun:"primary_key,autoincrement"`
	Host string `bun:"host,notnull"`
	// http:// https:// socks5://
	Method   string      `bun:"method,notnull"`
	Username string      `bun:"username"`
	Password string      `bun:"password"`
	Status   ProxyStatus `bun:"status,default:1"`
}
