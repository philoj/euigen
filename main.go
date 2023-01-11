package main

import (
	"deveui-gen-cli/deveui"
	"log"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "euiGen",
		Usage: "A cli application to generate new batches of devEUIs",
		Commands: []*cli.Command{
			{
				Name:    "generate",
				Aliases: []string{"gen"},
				Usage:   "generate a batch of N devEUIs",
				Action: func(cCtx *cli.Context) error {
					var batchSize int64 = 10
					var err error
					if cCtx.Args().Get(0) != "" {
						batchSize, err = strconv.ParseInt(cCtx.Args().Get(0), 10, 64)
						if err != nil {
							panic(err)
						}
					}
					_, err = deveui.CreateDevEUIs(int(batchSize))
					return err
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
