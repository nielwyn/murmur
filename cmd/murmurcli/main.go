package main

import (
	"fmt"
	"os"

	"github.com/nielwyn/murmur/internal/config"
)

type state struct {
	cfg    *config.Config
	client *apiClient
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(command) error
}

func (c *commands) register(name string, handler func(command) error) {
	c.handlers[name] = handler
}

func (c *commands) run(cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(cmd)
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

	s := &state{
		cfg:    &cfg,
		client: newAPIClient(cfg.APIURL, cfg.AuthToken),
	}

	cmds := commands{handlers: map[string]func(command) error{}}

	cmds.register("register", s.handleRegister)
	cmds.register("login", s.handleLogin)

	cmds.register("feeds", s.requireAuth(s.handleListFeeds))
	cmds.register("addfeed", s.requireAuth(s.handleCreateFeed))
	cmds.register("following", s.requireAuth(s.handleListFollowing))
	cmds.register("follow", s.requireAuth(s.handleFollowFeed))
	cmds.register("unfollow", s.requireAuth(s.handleUnfollowFeed))

	cmd := command{name: os.Args[1], args: os.Args[2:]}
	if err := cmds.run(cmd); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
