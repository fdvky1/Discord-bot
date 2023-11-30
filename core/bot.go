package core

import (
	"fmt"

	_ "github.com/fdvky1/Discord-bot/cmd"
	"github.com/fdvky1/Discord-bot/embed"
	"github.com/fdvky1/Discord-bot/entity"
	"github.com/fdvky1/Discord-bot/repo"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
	"golang.org/x/exp/slices"
)

var Clients BotClient

type BotClient map[string]*ken.Ken

type ConnectParams struct {
	Id    string
	Token string
}

// fix error import cycle
type LogMiddleware struct {
	Id string
}

var (
	_ ken.MiddlewareBefore = (*LogMiddleware)(nil)
)

func (c *LogMiddleware) Before(ctx *ken.Ctx) (next bool, err error) {
	cmd := ctx.GetCommand()
	ctx.Set("User-Id", c.Id)
	SendLog(c.Id, fmt.Sprintf("Execute: %s", cmd.Name()))
	next = true
	return
}

func Connect(params ConnectParams) (*ken.Ken, error) {
	session, err := discordgo.New(params.Token)
	if err != nil {
		return nil, err
	}
	// defer session.Close()
	k, err := ken.New(session)
	if err != nil {
		return nil, err
	}

	disabledCmds, _ := repo.DisabledCmdRepository.GetDisabledCmd(params.Id)
	var availableCmds []ken.Command

	for name, v := range embed.Cmds {
		c, ok := v.(ken.Command)
		isDisabled := slices.IndexFunc(disabledCmds, func(v entity.DisabledCmdEntity) bool { return v.Cmd == name })
		if ok && isDisabled == -1 {
			// fmt.Printf("Load: %s\n", name)
			availableCmds = append(availableCmds, c)
		}
	}

	if len(availableCmds) > 0 {
		k.RegisterCommands(
			availableCmds...,
		)
	}

	k.RegisterMiddlewares(
		&LogMiddleware{
			Id: params.Id,
		},
	)

	k.Session().AddHandlerOnce(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.GuildID == "" || m.Author.ID == s.State.User.ID {
			return
		}
		noteIsDisabled := slices.IndexFunc(disabledCmds, func(v entity.DisabledCmdEntity) bool { return v.Cmd == "note" })
		if noteIsDisabled > -1 {
			return
		}
		note, _ := repo.NoteRepository.Get(params.Id, m.GuildID, m.Content)
		if note.Value != "" {
			s.ChannelMessageSend(m.ChannelID, note.Value)
		}
	})

	k.Session().AddHandlerOnce(func(s *discordgo.Session, evt *discordgo.Connect) {
		Clients[params.Id] = k
		SendLog(params.Id, "Bot is connected to Discord.")
	})

	k.Session().AddHandlerOnce(func(s *discordgo.Session, evt *discordgo.Disconnect) {
		SendLog(params.Id, "Bot is disconnected from Discord.")
		if Clients[params.Id] == nil {
			k.Unregister()
		} else {
			s.Open()
		}
	})

	err = session.Open()
	if err != nil {
		return nil, err
	}

	return k, nil
}
