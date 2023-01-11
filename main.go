package main

import (
	"deveui-gen-cli/deveuigen"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"strconv"
)

func main() {
	app := &cli.App{
		Name:      "euigen",
		Usage:     "A cli application to generate a new batch of N devEUIs",
		UsageText: "euigen [-d] N",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "discard",
				Usage:   "Discard last incomplete run if present, instead of resuming",
				Value:   false,
				Aliases: []string{"d"},
			},
		},
		Action: func(cCtx *cli.Context) error {
			var batchSize int64
			var err error
			if cCtx.Args().Get(0) == "" {
				return fmt.Errorf("Please provide a valid positive integer for batch size")
			}
			batchSize, err = strconv.ParseInt(cCtx.Args().Get(0), 10, 64)
			if err != nil {
				panic(err)
			}
			if batchSize <= 0 {
				return fmt.Errorf("Please provide a valid positive integer for batch size")
			}
			_, err = deveuigen.CreateDevEUIs(int(batchSize), !cCtx.Bool("discard"))
			return err
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
