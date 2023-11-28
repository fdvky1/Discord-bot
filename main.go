package main

import (
	"context"
	"fmt"
	"os"

	"github.com/fdvky1/Discord-bot/core"
	"github.com/fdvky1/Discord-bot/embed"
	"github.com/fdvky1/Discord-bot/repo"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	supa "github.com/nedpals/supabase-go"
	"github.com/uptrace/bun"
	"github.com/zekrotja/ken"
	"golang.org/x/exp/slices"
)

var bunDB *bun.DB

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	core.Clients = make(map[string]*ken.Ken)
	core.WSClients = core.WebSocketClients{
		Clients: make(map[string]*websocket.Conn),
	}
	core.Supabase = supa.CreateClient(
		os.Getenv("SUPABASE_URL"),
		os.Getenv("SUPABASE_KEY"),
	)
	bunDB = core.NewPostgresDB()
	_ = repo.NewDisabledCmdRepository(bunDB)
	_ = repo.NewNoteRepository(bunDB)
}

func AuthMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if len(token) == 0 {
		return c.Status(403).SendString("Authorization token is required")
	}
	ctx := context.Background()
	user, err := core.Supabase.Auth.User(ctx, token)
	if err != nil || user == nil {
		return c.Status(401).SendString("Unauthorized")
	}

	c.Locals("User-Id", user.ID)

	var response []struct {
		BotToken string `json:"bot_token,omitempty"`
	}
	err = core.Supabase.DB.From("users").Select("*").Eq("id", user.ID).Execute(&response)
	if err != nil {
		panic(err)
	}

	if len(response) > 0 {
		c.Locals("Bot-Token", response[0].BotToken)
	}
	return c.Next()
}

func main() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowHeaders:     "Authorization, Origin, Content-Type, Accept, Content-Length, Accept-Language, Accept-Encoding, Connection, Access-Control-Allow-Origin",
		AllowOrigins:     "*",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	app.Get("/commands", func(c *fiber.Ctx) error {
		cmds := embed.List()
		return c.JSON(map[string][]embed.Cmd{
			"commands": cmds,
		})
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		token := c.Query("token")
		ctx := context.Background()
		user, err := core.Supabase.Auth.User(ctx, token)
		if err != nil || user == nil {
			c.Close()
		}

		core.WSClients.Clients[user.ID] = c

		defer func() {
			delete(core.WSClients.Clients, user.ID)
			c.Close()
		}()

		for {
			var msg core.Msg
			err := c.ReadJSON(&msg)
			if err != nil {
				fmt.Println("Error:", err)
				break
			}
		}
	}))

	app.Use(AuthMiddleware)

	app.Get("/start", func(c *fiber.Ctx) error {

		if c.Locals("Bot-Token") == nil || len(c.Locals("Bot-Token").(string)) == 0 {
			core.SendLog(c.Locals("User-Id").(string), "Please set the bot token on the setting")
			return c.Status(400).SendString("Please set the bot token on the setting")
		}

		session := core.Clients[c.Locals("User-Id").(string)]
		if session != nil {
			session.Session().Close()
		}

		if _, err := core.Connect(core.ConnectParams{
			Id:    c.Locals("User-Id").(string),
			Token: "Bot " + c.Locals("Bot-Token").(string),
		}); err != nil {
			return c.Status(500).SendString("Internal server error")
		}
		return c.SendStatus(200)
	})

	app.Get("/stop", func(c *fiber.Ctx) error {
		session := core.Clients[c.Locals("User-Id").(string)]

		if session == nil {
			return c.Status(400).SendString("Bot hasnt started")
		}

		if err := session.Session().Close(); err != nil {
			return c.Status(500).SendString("Internal server error")
		}

		return c.SendStatus(200)
	})

	app.Get("/status", func(c *fiber.Ctx) error {
		session := core.Clients[c.Locals("User-Id").(string)]

		return c.JSON(map[string]bool{
			"status": session != nil,
		})
	})

	app.Get("/disabled", func(c *fiber.Ctx) error {
		r, _ := repo.DisabledCmdRepository.GetDisabledCmd(c.Locals("User-Id").(string))
		result := []string{}
		for _, v := range r {
			result = append(result, v.Cmd)
		}
		return c.JSON(map[string][]string{
			"commands": result,
		})
	})

	app.Post("/disable", func(c *fiber.Ctx) error {
		payload := struct {
			Commands []string `json:"commands"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}
		cmds := embed.List()
		for _, cmd := range payload.Commands {
			isAvailable := slices.IndexFunc(cmds, func(v embed.Cmd) bool { return v.Name == cmd })
			if isAvailable > -1 {
				err := repo.DisabledCmdRepository.DisableCmd(c.Locals("User-Id").(string), cmd)
				if err != nil {
					return err
				}
			}
		}
		return c.SendStatus(200)
	})

	app.Post("/enable", func(c *fiber.Ctx) error {
		payload := struct {
			Commands []string `json:"commands"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}
		cmds := embed.List()
		for _, cmd := range payload.Commands {
			isAvailable := slices.IndexFunc(cmds, func(v embed.Cmd) bool { return v.Name == cmd })
			if isAvailable > -1 {
				err := repo.DisabledCmdRepository.EnableCmd(c.Locals("User-Id").(string), cmd)
				if err != nil {
					return err
				}
			}
		}
		return c.SendStatus(200)
	})

	err := app.Listen(":3000")
	if err != nil {
		fmt.Println("Error:", err)
	}
}
