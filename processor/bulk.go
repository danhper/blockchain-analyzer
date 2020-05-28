package processor

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/danhper/blockchain-analyzer/core"
)

type Aggregator interface {
	AddBlock(block core.Block)
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
	By       string
	Detailed bool
}

type groupActionsOverTimeParams struct {
	By       string
	Duration string
}

type countTransactionsOverTimeParams struct {
	Duration string
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
			property, err := core.GetActionProperty(params.By)
			if err != nil {
				return err
			}
			aggregator = core.NewGroupedActions(property, params.Detailed)

		case "count-transactions":
			aggregator = core.NewTransactionCounter()

		case "count-transactions-over-time":
			var params countTransactionsOverTimeParams
			if err := json.Unmarshal(rawProcessor.Params, &params); err != nil {
				return err
			}
			duration, err := time.ParseDuration(params.Duration)
			if err != nil {
				return err
			}
			aggregator = core.NewTimeGroupedTransactionCount(duration)

		case "group-actions-over-time":
			var params groupActionsOverTimeParams
			if err := json.Unmarshal(rawProcessor.Params, &params); err != nil {
				return err
			}
			duration, err := time.ParseDuration(params.Duration)
			if err != nil {
				return err
			}
			property, err := core.GetActionProperty(params.By)
			if err != nil {
				return err
			}
			aggregator = core.NewTimeGroupedActions(duration, property)

		default:
			return fmt.Errorf("unknown processor %s", rawProcessor.Name)
		}
		processor := NewProcessor(rawProcessor.Name, aggregator)
		c.Processors = append(c.Processors, processor)
	}
	return nil
}

func RunBulkActions(blockchain core.Blockchain, config BulkConfig) (map[string]interface{}, error) {
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
		processorResults[processor.Name] = processor.Aggregator
	}
	result["Results"] = processorResults
	return result, nil
}
