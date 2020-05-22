# blockchain-analyzer

CLI tool to fetch and analyze transactions data from multiple blockchains.

Currently supported blockchains:

* [Tezos](https://tezos.com/)
* [EOS](https://eos.io/)
* [XRP](https://ripple.com/xrp/)

## Build

Go needs to be installed. The tool can then be installed by running

```
go get github.com/danhper/blockchain-analyzer/cmd/blockchain-fetcher
```

## Usage

### Fetching data

The `fetch` command can be used to fetch the data:

```
blockchain-analyzer fetch -b BLOCKCHAIN -o OUTPUT_FILE --start START_BLOCK --end END_BLOCK

# examples
blockchain-analyzer fetch -b eos -o eos-blocks.jsonl --start 500000 --end 600000
```

The `check` command can then be used to check the fetched data.

```
blockchain-analyzer check -b eos -p 'eos-blocks*.jsonl' -o missing.jsonl --start 500000 --end 600000
```


### Analyzing data

Check the tool help for more information

```
blockchain-analyzer -h
```
