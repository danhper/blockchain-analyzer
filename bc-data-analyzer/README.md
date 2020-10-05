# bc-data-analyzer

This is a set of Python script to analyze the data produced by `blockchain-analyzer`.

## Installation

This repository can be installed directly from GitHub using the following command:

```
pip install 'git+https://github.com/danhper/blockchain-analyzer#subdirector
y=bc-data-analyzer'
```

or after a git clone, by running

```
cd bc-data-analyzer
pip install .
```

## Usage

The entrypoint is the CLI command `bc-data-analyzer`
It takes as input a file outputted by the `bulk-process` command of `blockchain-analyzer`
and can be use to plot or produce LaTeX tables.

For example, to plot the chart area of the distribution of actions, the following
command can be used:

```
bc-data-analyzer -b eos plot-actions-over-time /path/to/eos-results.json
```

or to generate a table with the top senders:

```
bc-data-analyzer -b tezos generate-table -n top-senders /path/to/tezos-results.json
```

More information on the different options can be obtained by using the help command

```
bc-data-analyzer -h
```
