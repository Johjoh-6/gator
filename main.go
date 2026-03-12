package main

import (
	"database/sql"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"

	_ "github.com/lib/pq"
)

const userName = "Johjoh-6"

type state struct {
	config *config.Config
	db     *database.Queries
}

func main() {
	// read the config file
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}
	// init the db, state and commands
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		panic(err)
	}
	dbQueries := database.New(db)

	state := &state{config: cfg, db: dbQueries}
	cmds := &commands{}

	// register commands
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAggregator)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	// get the args from the command line
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Need at least 2 arguments")
		fmt.Println("usage: gator <command> [args...]")
		os.Exit(1)
	}
	cmd := command{name: args[1], args: args[2:]}
	err = cmds.run(state, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
