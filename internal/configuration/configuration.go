package configuration

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

var (
	CustomWebhookUrl   string
	DiscordWebhookUrl  string
	FetchInterval      int
	LogPlanesToConsole bool
	SlackWebhookUrl    string
	TailNumbers        []string
)

const (
	// Defaults
	defaultFetchInterval int = 60

	// Config IDs
	config_customWebhookUrl   string = "customWebhookUrl"
	config_discordWebhookUrl  string = "discordWebhookUrl"
	config_fetchInterval      string = "fetchInterval"
	config_logPlanesToConsole string = "logPlanesToConsole"
	config_slackWebhookUrl    string = "slackWebhookUrl"
	config_tailNumbers        string = "tailNumbers"

	// Environment variables
	env_customWebhookUrl   string = "CUSTOM_WEBHOOK_URL"
	env_discordWebHookUrl  string = "DISCORD_WEBHOOK_URL"
	env_fetchInterval      string = "FETCH_INTERVAL"
	env_logPlanesToConsole string = "LOG_PLANES_TO_CONSOLE"
	env_slackWebhookUrl    string = "SLACK_WEBHOOK_URL"
	env_tailNumbers        string = "TAIL_NUMBERS"
)

func GetConfig() {
	var err error

	viper.SetDefault(config_customWebhookUrl, "")
	viper.SetDefault(config_discordWebhookUrl, "")
	viper.SetDefault(config_fetchInterval, defaultFetchInterval)
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

	err = viper.BindEnv(config_fetchInterval, env_fetchInterval)
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

	err = viper.BindEnv(config_tailNumbers, env_tailNumbers)
	if err != nil {
		log.Fatal(err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
		} else {
			panic(fmt.Errorf("fatal error in config file: %w", err))
		}
	}

	CustomWebhookUrl = viper.GetString(config_customWebhookUrl)
	DiscordWebhookUrl = viper.GetString(config_discordWebhookUrl)
	FetchInterval = viper.GetInt(config_fetchInterval)
	LogPlanesToConsole = viper.GetBool(config_logPlanesToConsole)
	SlackWebhookUrl = viper.GetString(config_slackWebhookUrl)

	tailNumbers := viper.GetString(config_tailNumbers)
	TailNumbers = strings.Split(strings.ToUpper(strings.ReplaceAll(tailNumbers, " ", "")), ",")

	if FetchInterval < 60 {
		log.Printf("Fetch interval of %ds detected. You might hit rate limits. Please consider using the default of %ds instead.", FetchInterval, defaultFetchInterval)
	}
}
