package entity

import "github.com/uptrace/bun"

type ActiveEntity struct {
	bun.BaseModel `bun:"table:active"`

	Id     string `bun:"id,pk,type:uuid"`
	Active bool   `bun:"active"`
}
