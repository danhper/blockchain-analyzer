package main

import (
	"fmt"
	"log"
	"os"
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

func addActionPropertyFlag(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:    "property",
		Aliases: []string{"p"},
		Value:   "name",
		Usage:   "Property to use for actions",
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

func addBlockchainFlag(flags []cli.Flag) []cli.Flag {
	return append(flags,
		&cli.StringFlag{
			Name:     "blockchain",
			Aliases:  []string{"b"},
			Value:    "",
			Usage:    "Blockchain to use",
			Required: true,
		},
	)
}

func blockchainFromCLI(c *cli.Context) (core.Blockchain, error) {
	switch c.String("blockchain") {
	case "xrp":
		return xrp.New(), nil
	case "eos":
		return eos.New(), nil
	case "tezos":
		return tezos.New(), nil
	default:
		cli.ShowSubcommandHelp(c)
		return nil, cli.NewExitError("wrong blockchain argument. valid: 'xrp', 'eos', 'tezos'", 1)
	}
}

func makeAction(f func(*cli.Context, core.Blockchain) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		blockchain, err := blockchainFromCLI(c)
		if err != nil {
			return err
		}
		return f(c, blockchain)
	}
}

func main() {
	app := &cli.App{
		Usage: "Tool to fetch and analyze blockchain transactions",
		Flags: addBlockchainFlag(nil),
		Commands: []*cli.Command{
			{
				Name:  "fetch",
				Flags: addFetchFlags(nil),
				Usage: "Fetches blockchain data",
				Action: makeAction(func(c *cli.Context, blockchain core.Blockchain) error {
					return blockchain.FetchData(c.String("output"), c.Uint64("start"), c.Uint64("end"))
				}),
			},
			{
				Name:  "check",
				Flags: addPatternFlag(addFetchFlags(nil)),
				Usage: "Checks for missing blocks in data",
				Action: makeAction(func(c *cli.Context, blockchain core.Blockchain) error {
					return processor.OutputAllMissingBlockNumbers(
						blockchain, c.String("pattern"), c.String("output"),
						c.Uint64("start"), c.Uint64("end"))
				}),
			},
			{
				Name:  "count-transactions",
				Flags: addPatternFlag(addRangeFlags(nil, false)),
				Usage: "Count the number of transactions in the data",
				Action: makeAction(func(c *cli.Context, blockchain core.Blockchain) error {
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
				Name:  "count-actions",
				Flags: addPatternFlag(addOutputFlag(addRangeFlags(nil, false))),
				Usage: "Count and groups the number of \"actions\" in the data",
				Action: makeAction(func(c *cli.Context, blockchain core.Blockchain) error {
					counts, err := processor.CountActions(
						blockchain, c.String("pattern"),
						c.Uint64("start"), c.Uint64("end"))
					if err != nil {
						return err
					}
					return core.Persist(counts, c.String("output"))
				}),
			},
			{
				Name: "count-actions-per-time",
				Flags: addGroupDurationFlag(
					addPatternFlag(addOutputFlag(addRangeFlags(nil, false)))),
				Usage: "Count and groups per time the number of \"actions\" in the data",
				Action: makeAction(func(c *cli.Context, blockchain core.Blockchain) error {
					duration, err := time.ParseDuration(c.String("duration"))
					if err != nil {
						return err
					}
					actionProperty, err := core.GetActionProperty(c.String("property"))
					if err != nil {
						return err
					}
					counts, err := processor.CountActionsPerTime(
						blockchain, c.String("pattern"),
						c.Uint64("start"), c.Uint64("end"),
						duration, actionProperty)
					if err != nil {
						return err
					}
					return core.Persist(counts, c.String("output"))
				}),
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
