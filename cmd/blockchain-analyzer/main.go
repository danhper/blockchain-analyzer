package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/danhper/blockchain-analyzer/eos"
	"github.com/danhper/blockchain-analyzer/processor"
	"github.com/danhper/blockchain-analyzer/tezos"
	"github.com/danhper/blockchain-analyzer/xrp"
	"github.com/urfave/cli/v2"
)

func addStartFlag(flags []cli.Flag, required bool) []cli.Flag {
	return append(flags, &cli.IntFlag{
		Name:     "start",
		Aliases:  []string{"s"},
		Required: required,
		Value:    0,
		Usage:    "Start block/ledger index",
	})
}

func addEndFlag(flags []cli.Flag, required bool) []cli.Flag {
	return append(flags, &cli.IntFlag{
		Name:     "end",
		Aliases:  []string{"e"},
		Required: required,
		Value:    0,
		Usage:    "End block/ledger index",
	})
}

func addOutputFlag(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:     "output",
		Aliases:  []string{"o"},
		Usage:    "Base output filepath",
		Required: true,
	})
}

func addConfigFlag(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:     "config",
		Aliases:  []string{"c"},
		Usage:    "Configuration file",
		Required: true,
	})
}

func addActionPropertyFlag(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:  "by",
		Value: "name",
		Usage: "Property to group the actions by",
	})
}

func addRangeFlags(flags []cli.Flag, required bool) []cli.Flag {
	return addStartFlag(addEndFlag(flags, required), required)
}

func addFetchFlags(flags []cli.Flag) []cli.Flag {
	return addRangeFlags(addOutputFlag(flags), true)
}

func addPatternFlag(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:     "pattern",
		Aliases:  []string{"p"},
		Value:    "",
		Usage:    "Patterns of files to check",
		Required: true,
	})
}

func addGroupDurationFlag(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:    "duration",
		Aliases: []string{"d"},
		Value:   "6h",
		Usage:   "Duration to group by when counting",
	})
}

func addDetailedFlag(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.BoolFlag{
		Name:     "detailed",
		Usage:    "Whether to add the details about sender/receivers etc",
		Value:    false,
		Required: false,
	})
}

func addCpuProfileFlag(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:     "cpu-profile",
		Usage:    "Path where to store the CPU profile",
		Value:    "",
		Required: false,
	})
}

func makeAction(f func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		cpuProfile := c.String("cpu-profile")
		if cpuProfile != "" {
			f, err := os.Create(cpuProfile)
			if err != nil {
				return fmt.Errorf("could not create CPU profile: %s", err.Error())
			}
			defer f.Close()
			if err := pprof.StartCPUProfile(f); err != nil {
				return fmt.Errorf("could not start CPU profile: %s", err.Error())
			}
			defer pprof.StopCPUProfile()
		}

		return f(c)
	}
}

func addCommonCommands(blockchain core.Blockchain, commands []*cli.Command) []*cli.Command {
	return append(commands, []*cli.Command{
		{
			Name:  "fetch",
			Flags: addFetchFlags(nil),
			Usage: "Fetches blockchain data",
			Action: makeAction(func(c *cli.Context) error {
				return blockchain.FetchData(c.String("output"), c.Uint64("start"), c.Uint64("end"))
			}),
		},
		{
			Name:  "check",
			Flags: addPatternFlag(addFetchFlags(nil)),
			Usage: "Checks for missing blocks in data",
			Action: makeAction(func(c *cli.Context) error {
				return processor.OutputAllMissingBlockNumbers(
					blockchain, c.String("pattern"), c.String("output"),
					c.Uint64("start"), c.Uint64("end"))
			}),
		},
		{
			Name:  "count-transactions",
			Flags: addPatternFlag(addRangeFlags(nil, false)),
			Usage: "Count the number of transactions in the data",
			Action: makeAction(func(c *cli.Context) error {
				count, err := processor.CountTransactions(
					blockchain, c.String("pattern"),
					c.Uint64("start"), c.Uint64("end"))
				if err != nil {
					return err
				}
				fmt.Printf("found %d transactions\n", count)
				return nil
			}),
		},
		{
			Name: "group-actions",
			Flags: addDetailedFlag(addActionPropertyFlag(
				addPatternFlag(addOutputFlag(addRangeFlags(nil, false))))),
			Usage: "Count and groups the number of \"actions\" in the data",
			Action: makeAction(func(c *cli.Context) error {
				actionProperty, err := core.GetActionProperty(c.String("by"))
				if err != nil {
					return err
				}
				counts, err := processor.GroupActions(
					blockchain, c.String("pattern"),
					c.Uint64("start"), c.Uint64("end"),
					actionProperty, c.Bool("detailed"))
				if err != nil {
					return err
				}
				return core.Persist(counts, c.String("output"))
			}),
		},
		{
			Name: "group-actions-over-time",
			Flags: addActionPropertyFlag(addGroupDurationFlag(
				addPatternFlag(addOutputFlag(addRangeFlags(nil, false))))),
			Usage: "Count and groups per time the number of \"actions\" in the data",
			Action: makeAction(func(c *cli.Context) error {
				duration, err := time.ParseDuration(c.String("duration"))
				if err != nil {
					return err
				}
				actionProperty, err := core.GetActionProperty(c.String("by"))
				if err != nil {
					return err
				}
				counts, err := processor.CountActionsOverTime(
					blockchain, c.String("pattern"),
					c.Uint64("start"), c.Uint64("end"),
					duration, actionProperty)
				if err != nil {
					return err
				}
				return core.Persist(counts, c.String("output"))
			}),
		},
		{
			Name:  "count-transactions-over-time",
			Flags: addGroupDurationFlag(addPatternFlag(addOutputFlag(addRangeFlags(nil, false)))),
			Usage: "Count number of \"transactions\" over time in the data",
			Action: makeAction(func(c *cli.Context) error {
				duration, err := time.ParseDuration(c.String("duration"))
				if err != nil {
					return err
				}
				counts, err := processor.CountTransactionsOverTime(
					blockchain, c.String("pattern"),
					c.Uint64("start"), c.Uint64("end"), duration)
				if err != nil {
					return err
				}
				return core.Persist(counts, c.String("output"))
			}),
		},
		{
			Name:  "bulk-process",
			Flags: addConfigFlag(addOutputFlag(nil)),
			Usage: "Bulk process the data according to the given configuration file",
			Action: makeAction(func(c *cli.Context) error {
				file, err := os.Open(c.String("config"))
				if err != nil {
					return err
				}
				defer file.Close()

				var config processor.BulkConfig
				if err := json.NewDecoder(file).Decode(&config); err != nil {
					return err
				}
				result, err := processor.RunBulkActions(blockchain, config)
				if err != nil {
					return err
				}
				return core.Persist(result, c.String("output"))
			}),
		},
		{
			Name:  "export",
			Flags: addPatternFlag(addOutputFlag(addRangeFlags(nil, false))),
			Usage: "Export a subset of the fields to msgpack format for faster processing",
			Action: makeAction(func(c *cli.Context) error {
				return processor.ExportToMsgpack(blockchain, c.String("pattern"),
					c.Uint64("start"), c.Uint64("end"), c.String("output"))
			}),
		},
	}...)
}

var eosCommands []*cli.Command = []*cli.Command{
	{
		Name:  "export-transfers",
		Flags: addPatternFlag(addOutputFlag(addRangeFlags(nil, false))),
		Usage: "Export all the transfers to a CSV file",
		Action: makeAction(func(c *cli.Context) error {
			return eos.ExportTransfers(
				c.String("pattern"),
				c.Uint64("start"), c.Uint64("end"), c.String("output"))
		}),
	},
}

func main() {
	app := &cli.App{
		Usage: "Tool to fetch and analyze blockchain transactions",
		Flags: addCpuProfileFlag(nil),
		Commands: []*cli.Command{
			{
				Name:        "eos",
				Usage:       "Analyze EOS data",
				Subcommands: addCommonCommands(eos.New(), eosCommands),
			},
			{
				Name:        "tezos",
				Usage:       "Analyze Tezos data",
				Subcommands: addCommonCommands(tezos.New(), nil),
			},
			{
				Name:        "xrp",
				Usage:       "Analyze XRP data",
				Subcommands: addCommonCommands(xrp.New(), nil),
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
