package core

import (
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	_ "github.com/fdvky1/Discord-bot/cmd"
	"github.com/fdvky1/Discord-bot/embed"
	"github.com/fdvky1/Discord-bot/repo"
	"github.com/zekrotja/ken"
)

var Clients map[string]*ken.Ken

func init() {
	Clients = make(map[string]*ken.Ken)
}

func StartBot(id string, token string) (*ken.Ken, error) {
	session, err := discordgo.New(token)
	if err != nil {
		return nil, err
	}

	k, err := ken.New(session)
	if err != nil {
		return nil, err
	}

	disabledCmds, _ := repo.DisabledCmdRepository.GetDisabledCmd(id)
	var cmds []ken.Command

	for name, c := range embed.Cmds {
		if !slices.Contains(disabledCmds, name) {
			cmd, ok := c.(ken.Command)
			if ok {
				cmds = append(cmds, cmd)
			}
		}
	}

	if len(cmds) > 0 {
		k.RegisterCommands(
			cmds...,
		)
	}

	k.RegisterMiddlewares(
		&logMiddleware{
			Id: id,
		},
	)

	session.AddHandler(func(s *discordgo.Session, evt *discordgo.Connect) {
		repo.ActiveRepository.Update(id, true)
		SendLog(id, "Bot is connected to Discord.")
	})

	session.AddHandler(func(s *discordgo.Session, evt *discordgo.Disconnect) {
		SendLog(id, "Bot is disconnected from Discord.")
		if Clients[id] != nil {
			s.Open()
		} else {
			k.Unregister()
		}
	})

	err = session.Open()
	return k, err
}

type logMiddleware struct {
	Id string
}

var (
	_ ken.MiddlewareBefore = (*logMiddleware)(nil)
)

func (mid *logMiddleware) Before(ctx *ken.Ctx) (next bool, err error) {
	cmd := ctx.GetCommand()
	ctx.Set("User-Id", mid.Id)
	SendLog(mid.Id, fmt.Sprintf("Execute: %s", cmd.Name()))
	next = true
	return
}
