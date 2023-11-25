package core

import (
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

	k.Session().AddHandlerOnce(func(s *discordgo.Session, evt *discordgo.Connect) {
		Clients[params.Id] = k
		SendLog(params.Id, "Bot is connected to Discord.")
	})

	k.Session().AddHandlerOnce(func(s *discordgo.Session, evt *discordgo.Disconnect) {
		delete(Clients, params.Id)
		k.Unregister()
		SendLog(params.Id, "Bot is disconnected from Discord.")
	})

	// defer k.Unregister()
	err = session.Open()
	if err != nil {
		return nil, err
	}

	return k, nil
}