package main

import (
	"context"
	"fmt"
	"log"
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

var supabase *supa.Client
var bunDB *bun.DB

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	core.Clients = make(map[string]*ken.Ken)
	core.WSClients = core.WebSocketClients{
		Clients: make(map[string]*websocket.Conn),
	}
	supabase = supa.CreateClient(
		os.Getenv("SUPABASE_URL"),
		os.Getenv("SUPABASE_KEY"),
	)
	bunDB = core.NewPostgresDB()
	_ = repo.NewDisabledCmdRepository(bunDB)
	// repo.DisabledCmdRepository = disabledCmdRepo
}

func main() {
	app := fiber.New()

	// app.Use(func(c *fiber.Ctx) error {
	// 	c.Set("Access-Control-Allow-Origin", "*")
	// 	return c.Next()
	// })
	app.Use(cors.New(cors.Config{
		AllowHeaders:     "Authorization, Origin, Content-Type, Accept, Content-Length, Accept-Language, Accept-Encoding, Connection, Access-Control-Allow-Origin",
		AllowOrigins:     "*",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		token := c.Query("token")
		ctx := context.Background()
		user, err := supabase.Auth.User(ctx, token)
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

	app.Get("/start", func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if len(token) == 0 {
			return c.Status(403).SendString("Authorization token is required")
		}
		ctx := context.Background()
		user, err := supabase.Auth.User(ctx, token)
		if err != nil || user == nil {
			return c.Status(401).SendString("Unauthorized")
		}

		var response []struct {
			BotToken string `json:"bot_token,omitempty"`
		}
		err = supabase.DB.From("users").Select("*").Eq("id", user.ID).Execute(&response)
		if err != nil {
			panic(err)
		}

		if len(response) == 0 || response[0].BotToken == "" {
			return c.Status(400).SendString("Please set the bot token on the setting")
		}

		if _, err := core.Connect(core.ConnectParams{
			Id:    user.ID,
			Token: "Bot " + response[0].BotToken,
		}); err != nil {
			return c.Status(500).SendString("Internal server error")
		}
		return c.SendStatus(200)
	})

	app.Get("/stop", func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if len(token) == 0 {
			return c.Status(403).SendString("Authorization token is required")
		}
		ctx := context.Background()
		user, err := supabase.Auth.User(ctx, token)
		if err != nil || user == nil {
			return c.Status(401).SendString("Unauthorized")
		}

		session := core.Clients[user.ID]

		if session == nil {
			return c.Status(400).SendString("Bot hasnt started")
		}

		if err := session.Session().Close(); err != nil {
			return c.Status(500).SendString("Internal server error")
		}

		return c.SendStatus(200)
	})

	app.Get("/status", func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if len(token) == 0 {
			return c.Status(403).SendString("Authorization token is required")
		}
		ctx := context.Background()
		user, err := supabase.Auth.User(ctx, token)
		if err != nil || user == nil {
			return c.Status(401).SendString("Unauthorized")
		}

		session := core.Clients[user.ID]

		return c.JSON(map[string]bool{
			"status": session != nil,
		})
	})

	app.Get("/commands", func(c *fiber.Ctx) error {
		cmds := embed.List()
		return c.JSON(map[string][]embed.Cmd{
			"commands": cmds,
		})
	})

	app.Get("/disabled", func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if len(token) == 0 {
			return c.Status(403).SendString("Authorization token is required")
		}
		ctx := context.Background()
		user, err := supabase.Auth.User(ctx, token)
		if err != nil || user == nil {
			return c.Status(401).SendString("Unauthorized")
		}
		r, _ := repo.DisabledCmdRepository.GetDisabledCmd(user.ID)
		result := []string{}
		for _, v := range r {
			result = append(result, v.Cmd)
		}
		return c.JSON(map[string][]string{
			"commands": result,
		})
	})

	app.Post("/disable", func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if len(token) == 0 {
			return c.Status(403).SendString("Authorization token is required")
		}
		ctx := context.Background()
		user, err := supabase.Auth.User(ctx, token)
		if err != nil || user == nil {
			return c.Status(401).SendString("Unauthorized")
		}
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
				err := repo.DisabledCmdRepository.DisableCmd(user.ID, cmd)
				if err != nil {
					return err
				}
			}
		}
		return c.SendStatus(200)
	})

	app.Post("/enable", func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if len(token) == 0 {
			return c.Status(403).SendString("Authorization token is required")
		}
		ctx := context.Background()
		user, err := supabase.Auth.User(ctx, token)
		if err != nil || user == nil {
			return c.Status(401).SendString("Unauthorized")
		}
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
				err := repo.DisabledCmdRepository.EnableCmd(user.ID, cmd)
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