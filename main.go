package main

import (
	"deveui-gen-cli/deveui"
	"fmt"
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
					fmt.Printf("generating %d devEUIs\n", batchSize)
					_, err = deveui.BatchRequest(int(batchSize))
					return err
				},
			},
			//{
			//	Name:    "complete",
			//	Aliases: []string{"c"},
			//	Usage:   "complete a task on the list",
			//	Action: func(cCtx *cli.Context) error {
			//		fmt.Println("completed task: ", cCtx.Args().First())
			//		return nil
			//	},
			//},
			//{
			//	Name:    "template",
			//	Aliases: []string{"t"},
			//	Usage:   "options for task templates",
			//	Subcommands: []*cli.Command{
			//		{
			//			Name:  "add",
			//			Usage: "add a new template",
			//			Action: func(cCtx *cli.Context) error {
			//				fmt.Println("new task template: ", cCtx.Args().First())
			//				return nil
			//			},
			//		},
			//		{
			//			Name:  "remove",
			//			Usage: "remove an existing template",
			//			Action: func(cCtx *cli.Context) error {
			//				fmt.Println("removed task template: ", cCtx.Args().First())
			//				return nil
			//			},
			//		},
			//	},
			//},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
