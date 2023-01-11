package main

import (
	"deveui-gen-cli/deveui"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"strconv"
)

func main() {
	app := &cli.App{
		Name:  "euigen",
		Usage: "A cli application to generate new batches of devEUIs",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "resume",
				Usage:   "Resume last incomplete run if present",
				Value:   true,
				Aliases: []string{"r"},
			},
		},
		Action: func(cCtx *cli.Context) error {
			var batchSize int64
			var err error
			if cCtx.Args().Get(0) == "" {
				return fmt.Errorf("please provide a valid positive integer for batch size")
			}
			batchSize, err = strconv.ParseInt(cCtx.Args().Get(0), 10, 64)
			if err != nil {
				panic(err)
			}
			if batchSize <= 0 {
				return fmt.Errorf("please provide a valid positive integer for batch size")
			}
			_, err = deveui.CreateDevEUIs(int(batchSize), cCtx.Bool("resume"))
			return err
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
