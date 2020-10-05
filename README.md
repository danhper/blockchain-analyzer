# blockchain-analyzer

[![CircleCI](https://circleci.com/gh/danhper/blockchain-analyzer.svg?style=svg)](https://circleci.com/gh/danhper/blockchain-analyzer)

CLI tool to fetch and analyze transactions data from several blockchains.

Currently supported blockchains:

- [Tezos](https://tezos.com/)
- [EOS](https://eos.io/)
- [XRP](https://ripple.com/xrp/)

## Installation

### Static binaries

We provide static binaries for Windows, macOS and Linux with each [release](https://github.com/danhper/blockchain-analyzer/releases)

### From source

Go needs to be installed. The tool can then be installed by running

```
go get github.com/danhper/blockchain-analyzer/cmd/blockchain-analyzer
```

## Usage

### Fetching data

The `fetch` command can be used to fetch the data:

```
blockchain-analyzer BLOCKCHAIN fetch -o OUTPUT_FILE --start START_BLOCK --end END_BLOCK

# for example from 500,000 to 699,999 inclusive:
blockchain-analyzer eos fetch -o eos-blocks.jsonl.gz --start 500000 --end 699999
```

The data has the following format:

- One block per line, including transactions, formatted in JSON. Documentation of block format can be found in each chain documentation
  - [EOS](https://developers.eos.io/manuals/eos/latest/nodeos/plugins/chain_api_plugin/api-reference/index#operation/get_block)
  - [Tezos](https://tezos.gitlab.io/api/rpc.html#get-block-id)
  - [XRP](https://xrpl.org/ledger.html)
- Grouped in files of 100,000 blocks each, suffixed by the block range (e.g. `eos-blocks-500000--599999.jsonl` and `eos-blocks-600000--699999.jsonl` for the above)
- Gziped if the `.gz` extension is added to the output file name (recommended)

The `check` command can then be used to check the fetched data. It will ensure that all the block from `--start` to `--end` exist in the given files, and output the missing blocks into `missing.jsonl` if any.

```
blockchain-analyzer eos check -p 'eos-blocks*.jsonl.gz' -o missing.jsonl --start 500000 --end 699999
```

### Analyzing data

The simplest way to analyze the data is to provide a configuration file about what to analyze and run the tool with the following command.

```
blockchain-analyzer <tezos|eos|xrp> bulk-process -c config.json -o tmp/results.json
```

Configuration files used for [our paper](https://arxiv.org/abs/2003.02693) can be found in the [config](./config) directory.

The tool's help also contains information about what other commands can be used

```plain
$ ./build/blockchain-analyzer -h
NAME:
   blockchain-analyzer - Tool to fetch and analyze blockchain transactions

USAGE:
   blockchain-analyzer [global options] command [command options] [arguments...]

COMMANDS:
   eos      Analyze EOS data
   tezos    Analyze Tezos data
   xrp      Analyze XRP data
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --cpu-profile value  Path where to store the CPU profile
   --help, -h           show help (default: false)

# the following is also available for xrp and tezos
$ ./build/blockchain-analyzer eos -h
NAME:
   blockchain-analyzer eos - Analyze EOS data

USAGE:
   blockchain-analyzer eos command [command options] [arguments...]

COMMANDS:
   export-transfers              Export all the transfers to a CSV file
   fetch                         Fetches blockchain data
   check                         Checks for missing blocks in data
   count-transactions            Count the number of transactions in the data
   group-actions                 Count and groups the number of "actions" in the data
   group-actions-over-time       Count and groups per time the number of "actions" in the data
   count-transactions-over-time  Count number of "transactions" over time in the data
   bulk-process                  Bulk process the data according to the given configuration file
   export                        Export a subset of the fields to msgpack format for faster processing
   help, h                       Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

### Interpreting results

We provide Python scripts to plot and generate table out of the data from the analysis.
Please check the [bc-data-analyzer](./bc-data-analyzer) directory for more information.

## Dataset

All the data used in our paper mentioned below can be downloaded from the following link:

https://imperialcollegelondon.box.com/s/jijwo76e2pxlbkuzzt1yjz0z3niqz7yy

This includes data from October 1, 2019 to April 30, 2020 for EOS, Tezos and XRP, which corresponds to the following blocks:

| Blockchain | Start block | End block |
| ---------- | ----------: | --------: |
| EOS        |    82152667 | 118286375 |
| XRP        |    50399027 |  55152991 |
| Tezos      |      630709 |    932530 |

## Supporting other blockchains

Although the framework currently only supports EOS, Tezos and XRP, it has been designed to easily support other blockchains.
The three following interfaces need to be implemented in order to do so:

```go
type Blockchain interface {
	FetchData(filepath string, start, end uint64) error
	ParseBlock(rawLine []byte) (Block, error)
	EmptyBlock() Block
}

type Block interface {
	Number() uint64
	TransactionsCount() int
	Time() time.Time
	ListActions() []Action
}

type Action interface {
	Sender() string
	Receiver() string
	Name() string
}
```

We also provide a utilities to make methods such as `FetchData` easier to implement.
[Existing implementations](https://github.com/danhper/blockchain-analyzer/blob/master/tezos/tezos.go) can be used as a point of reference for how a new blockchain can be supported.

## Academic work

This tool has originally been created to analyze data for the following paper: [Revisiting Transactional Statistics of High-scalability Blockchain](https://arxiv.org/abs/2003.02693), to be presented at [IMC'20](https://conferences.sigcomm.org/imc/2020/accepted/).  
If you are using this for academic work, we would be thankful if you could cite it.

```
@misc{perez2020revisiting,
    title={Revisiting Transactional Statistics of High-scalability Blockchain},
    author={Daniel Perez and Jiahua Xu and Benjamin Livshits},
    year={2020},
    eprint={2003.02693},
    archivePrefix={arXiv},
    primaryClass={cs.CR}
}
```
