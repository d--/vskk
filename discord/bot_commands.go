package discord

type BotCommands struct {
	commandChannels map[string]chan BotContext
}

func NewBotCommands() *BotCommands {
	return &BotCommands{
		commandChannels: make(map[string]chan BotContext),
	}
}

func (c *BotCommands) Add(name string) {
	c.commandChannels[name] = make(chan BotContext)
}

func (c *BotCommands) Get(name string) (chan BotContext, bool) {
	cdc, ok := c.commandChannels[name]
	return cdc, ok
}
