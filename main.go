package main

import (
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
	"github.com/whi-tw/steamodded/app/steamodded"
	"os"
	"runtime"
)

func main() {
	logLevel := log.InfoLevel
	if logLevelEnv := os.Getenv("LOG_LEVEL"); logLevelEnv != "" {
		var err error
		logLevel, err = log.ParseLevel(logLevelEnv)
		if err != nil {
			log.Fatalf("Invalid log level: %s", logLevelEnv)
		}
	}
	log.SetLevel(logLevel)

	app := &cli.App{
		Name:  "steammodded-injector",
		Usage: "A mod injector for Balatro",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Usage:   "Set the log level (debug, info, warn, error)",
				Value:   "info",
				Action: func(context *cli.Context, value string) error {
					level, err := log.ParseLevel(value)
					if err != nil {
						return err
					}
					log.SetLevel(level)
					return nil
				},
			},
		},
		Action: func(context *cli.Context) error {
			firstArg := context.Args().First()
			if firstArg != "" {
				return context.App.Command("inject").Run(context, "inject", context.Args().First())
			}
			return context.App.Command("help").Run(context)
		},
	}

	var injectCommandArgsUsage string
	switch runtime.GOOS {
	case "windows":
		injectCommandArgsUsage = "C:\\path\\to\\balatro.exe"
	default:
		injectCommandArgsUsage = "/path/to/balatro"
	}

	app.Commands = []*cli.Command{
		{
			Name:      "inject",
			Aliases:   []string{"i"},
			Usage:     "Inject the mod loader into the Balatro executable",
			ArgsUsage: fmt.Sprintf("<%s>", injectCommandArgsUsage),
			Action: func(context *cli.Context) error {
				firstArg := context.Args().First()
				if firstArg == "" {
					log.Error("No file provided to inject")
					return context.App.Command("help").Run(context, "inject")
				}

				return steamodded.DoInjection(firstArg)
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Errorf(err.Error())
	}
}
