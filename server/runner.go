package server

import (
	"bufio"
	"context"
	"fmt"
	"github.com/d--/vskk/config"
	"github.com/d--/vskk/discord"
	"github.com/d--/vskk/steam"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Runner struct {
	ExeLocation    string
	Options        config.ServerOpts
	Timeout        time.Duration
	SteamApi       steam.API
	RequestPlayers chan discord.BotContext
	cancel         context.CancelFunc
	done           chan int
}

func outputScan(ctx context.Context, s *bufio.Scanner, out chan string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if s.Scan() {
				out <- s.Text()
			} else {
				continue
			}
		}
	}
}

func (r *Runner) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *Runner) Done() chan int {
	return r.done
}

func (r *Runner) Start(parentCtx context.Context, botContext discord.BotContext) error {
	var ctx context.Context
	ctx, r.cancel = context.WithCancel(parentCtx)

	r.done = make(chan int)

	cmd, scanner, err := buildCmdAndOutputScanner(ctx, r.ExeLocation, r.Options)
	if err != nil {
		fmt.Println("error building command and output scanner for server:", err)
		return err
	}

	go func() {
		err = cmd.Start()
		if err != nil {
			fmt.Println("error starting server:", err)
			r.cancel()
			return
		}

		timer := time.Now()
		players := make(map[string]bool)
		lastIdSocketClosed := ""

		valheimOutput := make(chan string)
		outputScanCtx, outputScanCancel := context.WithCancel(ctx)
		go outputScan(outputScanCtx, scanner, valheimOutput)

		sleepTime := 2 * time.Millisecond
		maxSleep := 2 * time.Second
		for {
			select {
			case <-ctx.Done():
				err := cmd.Process.Kill()
				if err != nil {
					fmt.Println("failed to send kill signal:", err)
					continue
				}
				_, err = cmd.Process.Wait()
				if err != nil {
					fmt.Println("failed waiting for process to die:", err)
					continue
				}

				fmt.Println("server is shutting down...")
				outputScanCancel()
				r.done <- 0
				return
			case line := <-valheimOutput:
				sleepTime = 2 * time.Millisecond
				switch {
				case strings.HasPrefix(line, "(Filename"):
					continue
				case strings.TrimSpace(line) == "":
					continue
				case strings.Contains(line, "Got handshake from client"):
					fmt.Printf("VALHEIM: %s\n", line)
					fields := strings.Fields(line)
					id := fields[len(fields)-1]
					players[id] = true

					name, err := r.SteamApi.GetPlayerName(id)
					if err != nil {
						fmt.Println("failed to get player name:", err)
						name = "unknown"
					}

					fmt.Printf("connected: %s (%s)\n", id, name)
					timer = time.Now() // reset timeout
				case strings.Contains(line, "Closing socket"):
					fmt.Printf("VALHEIM: %s\n", line)
					fields := strings.Fields(line)
					id := fields[len(fields)-1]
					lastIdSocketClosed = id
				case strings.Contains(line, "k_ESteamNetworkingConnectionState_ClosedByPeer"):
					fmt.Printf("VALHEIM: %s\n", line)
					if _, ok := players[lastIdSocketClosed]; ok {
						fmt.Printf("player disconnected: %s\n", lastIdSocketClosed)
						timer = time.Now() // reset timeout
						delete(players, lastIdSocketClosed)
					}
				case strings.Contains(line, "Registering lobby"):
					fmt.Printf("VALHEIM: %s\n", line)
					ip, err := GetPublicIP()
					if err != nil {
						fmt.Println("could not get server public ip")
						break
					}

					message := fmt.Sprintf("Server is up!  Address:\n%s:%s", ip, r.Options.Port)
					err = botContext.SendMessage(message)
					if err != nil {
						fmt.Println("failed to send message:", err)
					}
				default:
					fmt.Printf("VALHEIM: %s\n", line)
				}
			case <-r.RequestPlayers:
				message := fmt.Sprintf("Total players: %d\n", len(players))
				for k := range players {
					name, err := r.SteamApi.GetPlayerName(k)
					if err != nil {
						fmt.Println("failed to get player name:", err)
						name = "unknown"
					}
					message += fmt.Sprintf("- %s\n", name)
				}

				err = botContext.SendMessage(message)
				if err != nil {
					fmt.Println("failed to send message:", err)
				}
			default:
				if len(players) == 0 && time.Since(timer) > r.Timeout {
					err := cmd.Process.Kill()
					if err != nil {
						fmt.Println("failed to send kill signal:", err)
						continue
					}
					_, err = cmd.Process.Wait()
					if err != nil {
						fmt.Println("failed waiting for process to die:", err)
						continue
					}
					fmt.Println("server is shutting down...")
					message := fmt.Sprintf("Looks like nobody is on.  Server is shutting down.")
					err = botContext.SendMessage(message)
					if err != nil {
						fmt.Println("failed to send message:", err)
					}
					outputScanCancel()
					r.done <- 0
					return
				} else {
					sleepTime *= 2
					if sleepTime > maxSleep {
						sleepTime = maxSleep
					}
					time.Sleep(sleepTime)
				}
			}
		}
	}()

	return nil
}

func buildCmdAndOutputScanner(ctx context.Context, exeLocation string, options config.ServerOpts) (*exec.Cmd, *bufio.Scanner, error) {
	cmd := exec.CommandContext(ctx, filepath.Clean(exeLocation),
		"-nographics",
		"-batchmode",
		"-name", options.Name,
		"-port", options.Port,
		"-world", options.World,
		"-password", options.Password,
		"-savedir", options.SaveDirFullPath)
	cmd.Env = append(cmd.Env, "SteamAppId=892970")
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	return cmd, bufio.NewScanner(reader), nil
}
