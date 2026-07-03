package configuration

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var (
	CustomWebhookUrl   string
	DiscordWebhookUrl  string
	LogPlanesToConsole bool
	SlackWebhookUrl    string
	StateFilePath      string
	TailNumbers        []string
)

const (
	// Config IDs
	config_customWebhookUrl   string = "customWebhookUrl"
	config_discordWebhookUrl  string = "discordWebhookUrl"
	config_logPlanesToConsole string = "logPlanesToConsole"
	config_slackWebhookUrl    string = "slackWebhookUrl"
	config_stateFilePath      string = "stateFilePath"
	config_tailNumbers        string = "tailNumbers"

	// Environment variables
	env_customWebhookUrl   string = "CUSTOM_WEBHOOK_URL"
	env_discordWebHookUrl  string = "DISCORD_WEBHOOK_URL"
	env_logPlanesToConsole string = "LOG_PLANES_TO_CONSOLE"
	env_slackWebhookUrl    string = "SLACK_WEBHOOK_URL"
	env_stateFilePath      string = "STATE_FILE"
	env_tailNumbers        string = "TAIL_NUMBERS"
	env_configFile         string = "CONFIG_FILE"
)

// defaultConfigDir returns the per-user directory config.yml/config.yaml and
// state.json live in by default (e.g. ~/.config/gee-bee).
func defaultConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "gee-bee")
}

func GetConfig() {
	var err error

	viper.SetDefault(config_customWebhookUrl, "")
	viper.SetDefault(config_discordWebhookUrl, "")
	viper.SetDefault(config_logPlanesToConsole, true)
	viper.SetDefault(config_slackWebhookUrl, "")
	viper.SetDefault(config_tailNumbers, []string{"28000", "29000"})

	// Bind the Viper key to an associated environment variable name
	err = viper.BindEnv(config_customWebhookUrl, env_customWebhookUrl)
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindEnv(config_discordWebhookUrl, env_discordWebHookUrl)
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindEnv(config_logPlanesToConsole, env_logPlanesToConsole)
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindEnv(config_slackWebhookUrl, env_slackWebhookUrl)
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindEnv(config_stateFilePath, env_stateFilePath)
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindEnv(config_tailNumbers, env_tailNumbers)
	if err != nil {
		log.Fatal(err)
	}

	if path := os.Getenv(env_configFile); path != "" {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(defaultConfigDir())
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
		} else {
			log.Fatalf("fatal error in config file: %v", err)
		}
	}

	CustomWebhookUrl = viper.GetString(config_customWebhookUrl)
	DiscordWebhookUrl = viper.GetString(config_discordWebhookUrl)
	LogPlanesToConsole = viper.GetBool(config_logPlanesToConsole)
	SlackWebhookUrl = viper.GetString(config_slackWebhookUrl)

	StateFilePath = viper.GetString(config_stateFilePath)
	if StateFilePath == "" {
		StateFilePath = filepath.Join(defaultConfigDir(), "state.json")
	}

	tailNumbers := viper.GetString(config_tailNumbers)
	TailNumbers = strings.Split(strings.ToUpper(strings.ReplaceAll(tailNumbers, " ", "")), ",")
}
