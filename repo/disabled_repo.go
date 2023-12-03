package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

var DisabledCmdRepository *disabledCmdRepository

type disabledCmdRepository struct {
	DB *bun.DB
}

func NewDisabledCmdRepository(DB *bun.DB) *disabledCmdRepository {
	DisabledCmdRepository = &disabledCmdRepository{DB: DB}
	return DisabledCmdRepository
}

func (repository disabledCmdRepository) DisableCmd(id string, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	_, err := repository.DB.NewRaw("INSERT INTO disabled_cmd VALUES(?, ?)",
		id, name).Exec(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (repository disabledCmdRepository) EnableCmd(id string, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	_, err := repository.DB.NewRaw("DELETE FROM disabled_cmd WHERE id = ? AND command_name = ?",
		id, name).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (repository disabledCmdRepository) GetDisabledCmd(id string) ([]string, error) {
	cmds := make([]string, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if err := repository.DB.NewSelect().NewRaw("SELECT command_name FROM disabled_cmd WHERE id = ?", id).Scan(ctx, &cmds); err != nil {
		return cmds, err
	}

	return cmds, nil
}
