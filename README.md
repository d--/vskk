# vskk (Valheim Server Knock-Knock Bot)

This is a Discord bot that will spin up your Valheim dedicated server in response to a command phrase.

- Responds to two commands by default:
```text
!valheim start   - starts the server
!valheim players - lists the number (and Steam names) of active players
```

Other little features:
- Notifies the channel on start-up and echoes its public IP
- Stops after timeout when there are no more active players
- Filters some junk out of the logs

## Requirements
- Windows
- Go 1.15 or above
- Valheim Dedicated Server
- Discord "Server" and Channel
- Steam Web API Key (https://steamcommunity.com/dev/apikey)
- Discord Auth Bot Token (https://discord.com/developers)

## Build
```shell
go install
```

## Run
```
%GOPATH%\bin\vskk.exe -c config.json
```

- Make sure that your configured port (default 2456) and the two above it are forwarded in your routing configuration.
- The server password must be less than five characters.
