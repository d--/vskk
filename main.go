package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/d--/vskk/config"
	"github.com/d--/vskk/discord"
	"github.com/d--/vskk/server"
	"github.com/d--/vskk/steam"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Valheim Server Knock Knock Bot
// This is a wrapper that will allow people in Discord to spin up the dedicated server on demand.
func main() {
	cfgFilePath := flag.String("c", "config.json", "config file path")
	flag.Parse()

	opts, err := config.Load(*cfgFilePath)
	if err != nil {
		fmt.Println("error loading config:", err)
		return
	}

	dg, err := discordgo.New("Bot " + opts.DiscordAuthBotToken)
	if err != nil {
		fmt.Println("error creating Discord session:", err)
		return
	}

	commands := discord.NewBotCommands()
	commands.Add("start")
	commands.Add("players")

	dg.AddHandler(discord.CreateMessageHandler(opts.DiscordChannelName, opts.CommandPhrase, commands))
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection:", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	go bot(ctx, opts, commands)

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	fmt.Println("Bot is running.  Press CTRL-C to exit.")
	<-chanSignal
	cancel()
	err = dg.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func bot(ctx context.Context, opts config.Options, commands *discord.BotCommands) {
	const (
		serverStarted = iota
		serverStopped
	)
	state := serverStopped

	chanStartServer, _ := commands.Get("start")
	chanRequestPlayers, _ := commands.Get("players")

	runner := &server.Runner{
		ExeLocation: opts.ValheimServerExeLocation,
		Options: opts.ValheimServerOpts,
		Timeout: time.Duration(opts.TimeoutMinutes) * time.Minute,
		RequestPlayers: chanRequestPlayers,
		SteamApi: steam.API{Key: opts.SteamWebApiKey},
	}
	for {
		select {
		case <-ctx.Done():
			runner.Stop()
			return
		case <-runner.Done():
			state = serverStopped
		case botContext := <-chanStartServer:
			if state == serverStopped {
				err := runner.Start(ctx, botContext)
				if err != nil {
					fmt.Println("server runner failed to start:", err)
					continue
				}

				randomMsg := opts.RandomStartMessages[rand.Int() % len(opts.RandomStartMessages)]
				message := fmt.Sprintf("%s  Spinning up...", randomMsg)
				message = fmt.Sprintf("%s  (This could take a few minutes.)", message)
				err = botContext.SendMessage(message)
				if err != nil {
					fmt.Println("failed to send message:", err)
				}
				state = serverStarted
			} else {
				message := fmt.Sprintf("Server is already started.")
				err := botContext.SendMessage(message)
				if err != nil {
					fmt.Println("failed to send message:", err)
				}
			}
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
