package entity

import "github.com/uptrace/bun"

type NoteEntity struct {
	bun.BaseModel `bun:"table:notes"`

	Id      string `bun:"id,type:uuid"`
	GuildId string `bun:"guild_id,type:text"`
	Key     string `bun:"key,type:text"`
	Value   string `bun:"value,type:text"`
}
