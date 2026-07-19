package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nielwyn/murmur/internal/config"
	"github.com/nielwyn/murmur/internal/database"
	"github.com/nielwyn/murmur/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
	svc *service.Service
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(s, cmd)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: murmurcli <command> [args...]")
		os.Exit(1)
	}

	cfg, err := config.Read()
	if err != nil {
		fmt.Println("error reading config:", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DBUrl)
	if err != nil {
		fmt.Println("error connecting to database:", err)
		os.Exit(1)
	}
	defer pool.Close()

	db := database.New(pool)
	s := &state{
		cfg: &cfg,
		db:  db,
		svc: service.New(db),
	}

	cmds := commands{handlers: map[string]func(*state, command) error{}}
	cmds.register("register", handlerRegister)
	cmds.register("login", handlerLogin)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))
	cmds.register("agg", handlerAgg)

	cmd := command{name: os.Args[1], args: os.Args[2:]}
	if err := cmds.run(s, cmd); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
