package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/m-mizutani/catbox/pkg/api"
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/urfave/cli/v2"
)

type apiConfig struct {
	AWSRegion string
	TableName string
	AssetDir  string
	Addr      string
	Port      int
}

func newAPICommand() *cli.Command {
	var config apiConfig

	return &cli.Command{
		Name: "api",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "aws-region",
				Aliases:     []string{"r"},
				EnvVars:     []string{"AWS_REGION"},
				Destination: &config.AWSRegion,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "table-name",
				Aliases:     []string{"t"},
				EnvVars:     []string{"CATBOX_TABLE_NAME"},
				Destination: &config.TableName,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "asset-dir",
				Aliases:     []string{"d"},
				EnvVars:     []string{"CATBOX_ASSET_DIR"},
				Destination: &config.AssetDir,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "Addr",
				Usage:       "server binding address",
				Aliases:     []string{"a"},
				EnvVars:     []string{"CATBOX_ADDR"},
				Destination: &config.Addr,
				Value:       "127.0.0.1",
			},
			&cli.IntFlag{
				Name:        "Port",
				Usage:       "Port number",
				Aliases:     []string{"p"},
				EnvVars:     []string{"CATBOX_PORT"},
				Destination: &config.Port,
				Value:       9080,
			},
		},
		Action: func(c *cli.Context) error {
			return apiCommand(c, config)
		},
	}
}

func apiCommand(c *cli.Context, config apiConfig) error {
	ctrl := controller.New()
	ctrl.AwsRegion = config.AWSRegion
	ctrl.TableName = config.TableName

	apiConfig := api.Config{
		AssetDir: config.AssetDir,
	}

	gin.SetMode(gin.DebugMode)
	engine := gin.Default()
	api.SetupBase(engine, ctrl, &apiConfig)
	api.SetupAssets(engine)
	api.SetupAPI(engine)

	logger.Info().Interface("config", config).Msg("Starting server...")
	if err := engine.Run(fmt.Sprintf("%s:%d", config.Addr, config.Port)); err != nil {
		logger.Error().Err(err).Interface("config", config).Msg("Server error")
	}

	return nil
}
