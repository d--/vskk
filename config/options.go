package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
)

type Options struct {
	ValheimServerExeLocation string     `json:"valheim_server_exe_location"`
	ValheimServerOpts        ServerOpts `json:"valheim_server_opts"`
	TimeoutMinutes           int64      `json:"timeout_minutes"`
	SteamWebApiKey           string     `json:"steam_web_api_key"`
	DiscordAuthBotToken      string     `json:"discord_auth_bot_token"`
	DiscordChannelName       string     `json:"discord_channel_name"`
	CommandPhrase            string     `json:"command_phrase"`
	RandomStartMessages      []string   `json:"random_start_messages"`
}

type ServerOpts struct {
	Name            string `json:"name"`
	Port            string `json:"port"`
	World           string `json:"world"`
	Password        string `json:"password"`
	SaveDirFullPath string `json:"save_dir_full_path"`
}

func Load(cfgFilePath string) (Options, error) {
	opts := Options{}

	cfgFileBytes, err := ioutil.ReadFile(filepath.Clean(cfgFilePath))
	if err != nil {
		err = fmt.Errorf("error reading config file: %w", err)
		return Options{}, err
	}

	err = json.Unmarshal(cfgFileBytes, &opts)
	if err != nil {
		err = fmt.Errorf("error ummarshaling config file: %w", err)
		return Options{}, err
	}

	lowerName := strings.ToLower(opts.ValheimServerOpts.Name)
	lowerPass := strings.ToLower(opts.ValheimServerOpts.Password)
	if strings.Contains(lowerName, lowerPass) {
		err = fmt.Errorf("server name cannot contain password")
		return Options{}, err
	}

	if len(lowerPass) < 5 {
		err = fmt.Errorf("server password must be at least 5 characters")
		return Options{}, err
	}

	optsValue := reflect.ValueOf(opts)
	optsType := reflect.TypeOf(opts)
	for i := 0; i < optsValue.NumField(); i++ {
		field := optsValue.Field(i)
		fieldName := optsType.Field(i).Name
		switch t := field.Interface().(type) {
		case string:
			if strings.TrimSpace(t) == "" {
				err = fmt.Errorf("option field was blank: %s", fieldName)
				return Options{}, err
			}
		}
	}

	serverOptsValue := reflect.ValueOf(opts.ValheimServerOpts)
	serverOptsType := reflect.TypeOf(opts.ValheimServerOpts)
	for i := 0; i < serverOptsValue.NumField(); i++ {
		field := serverOptsValue.Field(i)
		fieldName := serverOptsType.Field(i).Name
		switch t := field.Interface().(type) {
		case string:
			if strings.TrimSpace(t) == "" {
				err = fmt.Errorf("server option field was blank: %s", fieldName)
				return Options{}, err
			}
		}
	}

	if opts.TimeoutMinutes < 0 {
		err = fmt.Errorf("timeout cannot be less than zero",)
		return Options{}, err
	}

	if len(opts.RandomStartMessages) == 0 {
		err = fmt.Errorf("random start messages option was empty",)
		return Options{}, err
	}

	return opts, nil
}
