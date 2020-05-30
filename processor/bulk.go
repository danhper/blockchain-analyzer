package processor

import (
	"encoding/json"
	"fmt"

	"github.com/danhper/blockchain-analyzer/core"
)

type Aggregator interface {
	AddBlock(block core.Block)
	Result() interface{}
}

type Processor struct {
	Aggregator Aggregator
	Name       string
}

func NewProcessor(name string, aggregator Aggregator) Processor {
	return Processor{
		Aggregator: aggregator,
		Name:       name,
	}
}

type groupActionsParams struct {
	By       core.ActionProperty
	Detailed bool
}

type groupActionsOverTimeParams struct {
	By       core.ActionProperty
	Duration core.Duration
}

type countTransactionsOverTimeParams struct {
	Duration core.Duration
}

type BulkConfig struct {
	Pattern       string
	StartBlock    uint64
	EndBlock      uint64
	RawProcessors []struct {
		Name   string
		Type   string
		Params json.RawMessage
	} `json:"Processors"`
	Processors []Processor `json:"-"`
}

func (c *BulkConfig) UnmarshalJSON(data []byte) error {
	type rawConfig BulkConfig
	if err := json.Unmarshal(data, (*rawConfig)(c)); err != nil {
		return err
	}
	for _, rawProcessor := range c.RawProcessors {
		var aggregator Aggregator
		switch rawProcessor.Type {
		case "group-actions":
			var params groupActionsParams
			if err := json.Unmarshal(rawProcessor.Params, &params); err != nil {
				return err
			}
			aggregator = core.NewGroupedActions(params.By, params.Detailed)

		case "count-transactions":
			aggregator = core.NewTransactionCounter()

		case "count-transactions-over-time":
			var params countTransactionsOverTimeParams
			if err := json.Unmarshal(rawProcessor.Params, &params); err != nil {
				return err
			}
			aggregator = core.NewTimeGroupedTransactionCount(params.Duration.Duration)

		case "group-actions-over-time":
			var params groupActionsOverTimeParams
			if err := json.Unmarshal(rawProcessor.Params, &params); err != nil {
				return err
			}
			aggregator = core.NewTimeGroupedActions(params.Duration.Duration, params.By)

		default:
			return fmt.Errorf("unknown processor %s", rawProcessor.Name)
		}
		processor := NewProcessor(rawProcessor.Name, aggregator)
		c.Processors = append(c.Processors, processor)
	}
	return nil
}

func RunBulkActions(blockchain core.Blockchain, config BulkConfig) (map[string]interface{}, error) {
	missingBlockProcessor := NewProcessor("MissingBlocks", core.NewMissingBlocks(config.StartBlock, config.EndBlock))
	config.Processors = append(config.Processors, missingBlockProcessor)
	blocks, err := YieldAllBlocks(config.Pattern, blockchain, config.StartBlock, config.EndBlock)
	if err != nil {
		return nil, err
	}

	for block := range blocks {
		for _, processor := range config.Processors {
			processor.Aggregator.AddBlock(block)
		}
	}

	result := make(map[string]interface{})
	result["Config"] = config
	processorResults := make(map[string]interface{})
	for _, processor := range config.Processors {
		processorResults[processor.Name] = processor.Aggregator.Result()
	}
	result["Results"] = processorResults
	return result, nil
}
