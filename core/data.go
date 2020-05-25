package core

import (
	"encoding/json"
	"fmt"
	"time"
)

type ActionProperty int

const (
	ActionName ActionProperty = iota
	ActionSender
)

func GetActionProperty(name string) (ActionProperty, error) {
	switch name {
	case "name":
		return ActionName, nil
	case "sender":
		return ActionSender, nil
	default:
		return ActionName, fmt.Errorf("no property %s for actions", name)
	}
}

type ActionsCount struct {
	actions map[string]uint64
}

func NewActionsCount() *ActionsCount {
	return &ActionsCount{
		actions: make(map[string]uint64),
	}
}

func (a *ActionsCount) Increment(action string) {
	a.actions[action] += 1
}

func (a *ActionsCount) Get(action string) uint64 {
	return a.actions[action]
}

func (a *ActionsCount) Merge(other *ActionsCount) {
	for key, value := range other.actions {
		a.actions[key] += value
	}
}

func (a *ActionsCount) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.actions)
}

func Persist(entity interface{}, outputFile string) error {
	file, err := CreateFile(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	return encoder.Encode(entity)
}

type GroupedActions struct {
	Actions   map[time.Time]*ActionsCount
	GroupedBy time.Duration
}

func NewGroupedActions(duration time.Duration) *GroupedActions {
	return &GroupedActions{
		Actions:   make(map[time.Time]*ActionsCount),
		GroupedBy: duration,
	}
}

func (g *GroupedActions) AddActions(timestamp time.Time, actions *ActionsCount) {
	group := timestamp.Truncate(g.GroupedBy)
	if _, ok := g.Actions[group]; !ok {
		g.Actions[group] = NewActionsCount()
	}
	g.Actions[group].Merge(actions)
}

type GroupedTransactionCount struct {
	TransactionCounts map[time.Time]int
	GroupedBy         time.Duration
}

func NewGroupedTransactionCount(duration time.Duration) *GroupedTransactionCount {
	return &GroupedTransactionCount{
		TransactionCounts: make(map[time.Time]int),
		GroupedBy:         duration,
	}
}

func (g *GroupedTransactionCount) AddBlock(block Block) {
	group := block.Time().Truncate(g.GroupedBy)
	if _, ok := g.TransactionCounts[group]; !ok {
		g.TransactionCounts[group] = 0
	}
	g.TransactionCounts[group] += block.TransactionsCount()
}
