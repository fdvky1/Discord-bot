package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/fdvky1/Discord-bot/entity"

	"github.com/uptrace/bun"
)

var NoteRepository *noteRepository

type noteRepository struct {
	DB *bun.DB
}

func NewNoteRepository(DB *bun.DB) *noteRepository {
	NoteRepository = &noteRepository{DB: DB}
	return NoteRepository
}

func (repository noteRepository) PutNote(note entity.NoteEntity) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	_, err := repository.DB.NewInsert().Model(&note).Exec(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (repository noteRepository) RemoveNote(id string, guildId string, key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	_, err := repository.DB.NewRaw("DELETE FROM notes WHERE id = ? AND guild_id = ? AND key = ?", id, guildId, key).Exec(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (repository noteRepository) GetAll(id string, guildId string) ([]entity.NoteEntity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	var result []entity.NoteEntity
	if err := repository.DB.NewSelect().Model(&result).Where("id = ? AND guild_id = ?", id, guildId).Scan(ctx); err != nil {
		return result, err
	}

	return result, nil
}

func (repository noteRepository) Get(id string, guildId string, key string) (entity.NoteEntity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	var result entity.NoteEntity
	if err := repository.DB.NewSelect().Model(&result).Where("id = ? AND guild_id = ? AND key = ?", id, guildId, key).Scan(ctx); err != nil {
		return result, err
	}

	return result, nil
}
