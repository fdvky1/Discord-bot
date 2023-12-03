package repo

import (
	"context"
	"time"

	"github.com/fdvky1/Discord-bot/entity"
	"github.com/uptrace/bun"
)

var ActiveRepository *activeRepository

type activeRepository struct {
	DB *bun.DB
}

func NewActiveRepository(DB *bun.DB) *activeRepository {
	ActiveRepository = &activeRepository{DB: DB}
	return ActiveRepository
}

func (repository activeRepository) IsActive(id string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	var result entity.ActiveEntity
	if err := repository.DB.NewSelect().Model(&result).Where("id = ? ", id).Scan(ctx); err != nil {
		return false, err
	}

	return result.Active, nil
}

func (repository activeRepository) Update(id string, active bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if _, err := repository.DB.NewRaw("INSERT INTO active VALUES(?, ?) ON CONFLICT(id) DO UPDATE SET active = ?", id, active, active).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (repository activeRepository) ActiveUser() ([]string, error) {
	var ids []string

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if err := repository.DB.NewSelect().NewRaw("SELECT id FROM active WHERE active = true").Scan(ctx, &ids); err != nil {
		return ids, err
	}

	return ids, nil
}
