package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fdvky1/Discord-bot/core"
	"github.com/fdvky1/Discord-bot/embed"
	"github.com/fdvky1/Discord-bot/repo"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	supa "github.com/nedpals/supabase-go"
	"github.com/uptrace/bun"
)

var bunDB *bun.DB

type resp struct {
	Message string `json:"message"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	core.WSClients = core.WebSocketClients{
		Clients: make(map[string]*websocket.Conn),
	}
	core.Supabase = supa.CreateClient(
		os.Getenv("SUPABASE_URL"),
		os.Getenv("SUPABASE_KEY"),
	)
	bunDB = core.NewPostgresDB()
	repo.NewDisabledCmdRepository(bunDB)
	repo.NewNoteRepository(bunDB)
	repo.NewActiveRepository(bunDB)

	go func() {
		ids, err := repo.ActiveRepository.ActiveUser()
		if err == nil {
			for _, id := range ids {
				var result []struct {
					BotToken string `json:"bot_token,omitempty"`
				}
				err := core.Supabase.DB.From("users").Select("*").Eq("id", id).Execute(&result)
				if err == nil && result[0].BotToken != "" {
					s, err := core.StartBot(id, "Bot "+result[0].BotToken)
					if err == nil {
						core.Clients[id] = s
					}
				}
			}
		}
	}()

}

func main() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	done := make(chan bool, 1)

	go handleSignal(sc, done)

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

		for c != nil {
			_, _, err := c.ReadMessage()
			if err != nil {
				return // Calls the deferred function, i.e. closes the connection on error
			}
		}
	}))

	app.Use(authMiddleware)

	app.Get("/start", func(c *fiber.Ctx) error {
		userId := c.Locals("User-Id").(string)
		var result []struct {
			BotToken string `json:"bot_token,omitempty"`
		}

		err := core.Supabase.DB.From("users").Select("*").Eq("id", userId).Execute(&result)
		if err != nil || len(result) == 0 {
			return c.Status(404).JSON(resp{
				Message: "User not found",
			})
		}

		if result[0].BotToken == "" {
			core.SendLog(userId, "Please set the bot token on the setting")
			return c.Status(400).JSON(resp{
				Message: "Please set the bot token on the setting",
			})
		}

		session := core.Clients[userId]
		if session != nil {
			delete(core.Clients, userId)
			if err := session.Session().Close(); err != nil {
				return c.Status(500).JSON(resp{
					Message: "Failed restarting bot",
				})
			}
		}

		s, err := core.StartBot(userId, "Bot "+result[0].BotToken)
		if err != nil {
			return c.Status(500).JSON(resp{
				Message: "Failed to start bot, please check bot token",
			})
		}
		core.Clients[userId] = s
		return c.SendStatus(200)
	})

	app.Get("/stop", func(c *fiber.Ctx) error {
		userId := c.Locals("User-Id").(string)
		s := core.Clients[userId]

		if s == nil {
			return c.Status(400).JSON(resp{
				Message: "Bot hasnt started",
			})
		}

		delete(core.Clients, userId)
		repo.ActiveRepository.Update(userId, false)
		if err := s.Session().Close(); err != nil {
			return c.Status(500).JSON(resp{
				Message: "Failed to start bot, please check bot token",
			})
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
		result, _ := repo.DisabledCmdRepository.GetDisabledCmd(c.Locals("User-Id").(string))
		return c.JSON(map[string][]string{
			"commands": result,
		})
	})

	app.Post("/disable", func(c *fiber.Ctx) error {
		var payload struct {
			Commands []string `json:"commands"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}
		for _, cmd := range payload.Commands {
			if embed.Cmds[cmd] != nil {
				err := repo.DisabledCmdRepository.DisableCmd(c.Locals("User-Id").(string), cmd)
				if err != nil {
					return err
				}
			}
		}
		return c.SendStatus(200)
	})

	app.Post("/enable", func(c *fiber.Ctx) error {
		var payload struct {
			Commands []string `json:"commands"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}
		for _, cmd := range payload.Commands {
			if embed.Cmds[cmd] != nil {
				err := repo.DisabledCmdRepository.EnableCmd(c.Locals("User-Id").(string), cmd)
				if err != nil {
					return err
				}
			}
		}
		return c.SendStatus(200)
	})

	go func() {
		if err := app.Listen(":" + os.Getenv("PORT")); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
			done <- true
		}
	}()

	<-done
}

func authMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if len(token) == 0 {
		return c.Status(403).JSON(resp{
			Message: "Authorization token is required",
		})
	}
	ctx := context.Background()
	user, err := core.Supabase.Auth.User(ctx, token)
	if err != nil || user == nil {
		return c.Status(401).JSON(resp{
			Message: "Unauthorized",
		})
	}

	c.Locals("User-Id", user.ID)

	return c.Next()
}

func handleSignal(sigChan <-chan os.Signal, doneChan chan<- bool) {
	sig := <-sigChan
	fmt.Printf("Received signal: %v\n", sig)

	closeAllSession()

	doneChan <- true
}

func closeAllSession() {
	for id, s := range core.Clients {
		fmt.Printf("Closing: %s\n", id)
		delete(core.Clients, id)
		s.Session().Close()
	}
}
