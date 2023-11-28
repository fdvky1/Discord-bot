package general

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/fdvky1/Discord-bot/embed"
	"github.com/fdvky1/Discord-bot/entity"
	"github.com/fdvky1/Discord-bot/repo"
	"github.com/zekrotja/ken"
)

var noteOptions = [][]string{
	{"list", "List"},
	{"add", "Add"},
	{"remove", "Remove"},
}

type NoteCommand struct{}

var (
	_ ken.SlashCommand        = (*NoteCommand)(nil)
	_ ken.DmCapable           = (*NoteCommand)(nil)
	_ ken.AutocompleteCommand = (*NoteCommand)(nil)
)

func (c *NoteCommand) Name() string {
	return "note"
}

func (c *NoteCommand) Description() string {
	return "Save your notes"
}

func (c *NoteCommand) Version() string {
	return "1.0.0"
}

func (c *NoteCommand) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (c *NoteCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:         discordgo.ApplicationCommandOptionString,
			Name:         "options",
			Required:     true,
			Description:  "Choose option",
			Autocomplete: true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "key",
			Required:    false,
			Description: "Enter the key",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "note",
			Required:    false,
			Description: "Enter the note",
		},
	}
}

func (c *NoteCommand) IsDmCapable() bool {
	return false
}

func (c *NoteCommand) Autocomplete(ctx *ken.AutocompleteContext) ([]*discordgo.ApplicationCommandOptionChoice, error) {
	input, ok := ctx.GetInput("options")

	if !ok {
		return nil, nil
	}

	choises := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(noteOptions))
	input = strings.ToLower(input)

	for _, lang := range noteOptions {
		if strings.HasPrefix(lang[1], input) {
			choises = append(choises, &discordgo.ApplicationCommandOptionChoice{
				Name:  lang[0],
				Value: lang[0],
			})
		}
	}

	return choises, nil
}

func (c *NoteCommand) Run(ctx ken.Context) (err error) {
	opt := ctx.Options().GetByName("options").StringValue()
	UserId := ctx.Get("User-Id").(string)
	if opt == "list" {
		notes, _ := repo.NoteRepository.GetAll(UserId, ctx.GetEvent().GuildID)
		text := "Notes:\n"
		for _, note := range notes {
			text += fmt.Sprintf("> %s\n", note.Key)
		}
		return ctx.RespondMessage(text)
	} else {
		key, ok := ctx.Options().GetByNameOptional("key")
		if opt == "add" && ok {
			note, ok := ctx.Options().GetByNameOptional("note")
			if ok {
				err := repo.NoteRepository.PutNote(entity.NoteEntity{
					Id:      UserId,
					GuildId: ctx.GetEvent().GuildID,
					Key:     key.StringValue(),
					Value:   note.StringValue(),
				})
				if err != nil {
					return ctx.RespondMessage(fmt.Sprintf("Error %v", err))
				}
				return ctx.RespondMessage(fmt.Sprintf("The note with key \"%s\" was saved successfully", key.StringValue()))
			}
		} else if ok {
			err := repo.NoteRepository.RemoveNote(UserId, ctx.GetEvent().GuildID, key.StringValue())
			if err != nil {
				return ctx.RespondMessage(fmt.Sprintf("Error %v", err))
			}
			return ctx.RespondMessage(fmt.Sprintf("The note with key \"%s\" was deleted successfully", key.StringValue()))
		}
	}
	return ctx.RespondMessage("Invalid")
}

func init() {
	embed.Cmds["note"] = &NoteCommand{}
}
