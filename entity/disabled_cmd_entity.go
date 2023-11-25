package entity

import "github.com/uptrace/bun"

type DisabledCmdEntity struct {
	bun.BaseModel `bun:"table:disabled_cmd"`

	Id  string `bun:"id,type:uuid"`
	Cmd string `bun:"command_name,type:text"`
}
