# blockchain-fetcher

Simple CLI tool to fetch transactions data from multiple blockchains.

Currently supported blockchains:

* [Tezos](https://tezos.com/)
* [EOS](https://eos.io/)
* [XRP](https://ripple.com/xrp/)

## Build

Go needs to be installed. The tool can then be installed by running

```
go get github.com/danhper/blockchain-data-fetcher/cmd/blockchain-fetcher
```

## Usage

The `fetch` command can be used to fetch the data:

```
blockchain-fetcher fetch -b BLOCKCHAIN -o OUTPUT_FILE --start START_BLOCK --end END_BLOCK

# examples
blockchain-fetcher fetch -b eos -o eos-blocks.jsonl --start 500000 --end 600000
```

The `check` command can then be used to check the fetched data.

```
blockchain-fetcher check -b eos -p 'eos-blocks*.jsonl' -o missing.jsonl --start 500000 --end 600000
```
