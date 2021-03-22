package main

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

var logger zerolog.Logger

func init() {
}

func globalSetup(c *cli.Context) error {
	// Setup logger
	logLevel := c.String("log-level")
	var zeroLogLevel zerolog.Level

	switch strings.ToLower(logLevel) {
	case "trace":
		zeroLogLevel = zerolog.TraceLevel
	case "debug":
		zeroLogLevel = zerolog.DebugLevel
	case "info":
		zeroLogLevel = zerolog.InfoLevel
	case "warn":
		zeroLogLevel = zerolog.WarnLevel
	case "error":
		zeroLogLevel = zerolog.ErrorLevel
	default:
		zeroLogLevel = zerolog.InfoLevel
	}

	console := zerolog.ConsoleWriter{Out: os.Stdout}

	logger = zerolog.New(console).Level(zeroLogLevel).With().Timestamp().Logger()
	logger.Debug().Str("LogLevel", logLevel).Msg("Set log level")
	return nil
}

func main() {
	app := &cli.App{
		Name:        "catbox",
		Description: "Utility command of catbox",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				EnvVars: []string{"CATBOX_LOG_LEVEL"},
				Usage:   "LogLevel [trace|debug|info|warn|error]",
			},
		},
		Commands: []*cli.Command{
			newAPICommand(),
		},
		Before: globalSetup,
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error().Err(err).Msg("Failed")
		os.Exit(1)
	}
}
