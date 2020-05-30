package core

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/danhper/structomap"
)

type ActionProperty int

const (
	ActionName ActionProperty = iota
	ActionSender
	ActionReceiver
)

const (
	maxTopLevelResults = 1000
	maxNestedResults   = 50
)

func GetActionProperty(name string) (ActionProperty, error) {
	switch name {
	case "name":
		return ActionName, nil
	case "sender":
		return ActionSender, nil
	case "receiver":
		return ActionReceiver, nil
	default:
		return ActionName, fmt.Errorf("no property %s for actions", name)
	}
}

func (p ActionProperty) String() string {
	switch p {
	case ActionName:
		return "name"
	case ActionSender:
		return "sender"
	case ActionReceiver:
		return "receiver"
	default:
		panic(fmt.Errorf("no such action property"))
	}
}

func (c *ActionProperty) UnmarshalJSON(data []byte) (err error) {
	var rawProperty string
	if err = json.Unmarshal(data, &rawProperty); err != nil {
		return err
	}
	*c, err = GetActionProperty(rawProperty)
	return err
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	var rawDuration string
	if err = json.Unmarshal(b, &rawDuration); err != nil {
		return err
	}
	d.Duration, err = time.ParseDuration(rawDuration)
	return err
}

type ActionsCount struct {
	Actions     map[string]uint64
	UniqueCount uint64
	TotalCount  uint64
}

func NewActionsCount() *ActionsCount {
	return &ActionsCount{
		Actions: make(map[string]uint64),
	}
}

func (a *ActionsCount) Increment(key string) {
	a.TotalCount++
	if _, ok := a.Actions[key]; !ok {
		a.UniqueCount++
	}
	a.Actions[key] += 1
}

func (a *ActionsCount) Get(key string) uint64 {
	return a.Actions[key]
}

func (a *ActionsCount) Merge(other *ActionsCount) {
	for key, value := range other.Actions {
		a.Actions[key] += value
	}
}

type NamedCount struct {
	Name  string
	Count uint64
}

var actionsCountSerializer = structomap.New().
	PickFunc(func(actions interface{}) interface{} {
		var results []NamedCount
		for name, count := range actions.(map[string]uint64) {
			results = append(results, NamedCount{Name: name, Count: count})
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].Count > results[j].Count
		})
		if len(results) > maxNestedResults {
			results = results[:maxNestedResults]
		}
		return results
	}, "Actions").
	Pick("UniqueCount", "TotalCount")

func (a *ActionsCount) MarshalJSON() ([]byte, error) {
	return json.Marshal(actionsCountSerializer.Transform(a))
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

type TimeGroupedActions struct {
	Actions   map[time.Time]*GroupedActions
	Duration  time.Duration
	GroupedBy ActionProperty
}

func NewTimeGroupedActions(duration time.Duration, by ActionProperty) *TimeGroupedActions {
	return &TimeGroupedActions{
		Actions:   make(map[time.Time]*GroupedActions),
		Duration:  duration,
		GroupedBy: by,
	}
}

func (g *TimeGroupedActions) AddBlock(block Block) {
	group := block.Time().Truncate(g.Duration)
	if _, ok := g.Actions[group]; !ok {
		g.Actions[group] = NewGroupedActions(g.GroupedBy, false)
	}
	g.Actions[group].AddBlock(block)
}

func (g *TimeGroupedActions) Result() interface{} {
	return g
}

type TimeGroupedTransactionCount struct {
	TransactionCounts map[time.Time]int
	GroupedBy         time.Duration
}

func NewTimeGroupedTransactionCount(duration time.Duration) *TimeGroupedTransactionCount {
	return &TimeGroupedTransactionCount{
		TransactionCounts: make(map[time.Time]int),
		GroupedBy:         duration,
	}
}

func (g *TimeGroupedTransactionCount) AddBlock(block Block) {
	group := block.Time().Truncate(g.GroupedBy)
	if _, ok := g.TransactionCounts[group]; !ok {
		g.TransactionCounts[group] = 0
	}
	g.TransactionCounts[group] += block.TransactionsCount()
}

func (g *TimeGroupedTransactionCount) Result() interface{} {
	return g
}

type ActionGroup struct {
	Name      string
	Count     uint64
	Names     *ActionsCount
	Senders   *ActionsCount
	Receivers *ActionsCount
}

var actionGroupSerializer = structomap.New().
	Pick("Name", "Count").
	PickIf(func(a interface{}) bool {
		return a.(*ActionGroup).Names.TotalCount > 0
	}, "Names", "Senders", "Receivers")

func (a *ActionGroup) MarshalJSON() ([]byte, error) {
	return json.Marshal(actionGroupSerializer.Transform(a))
}

func NewActionGroup(name string) *ActionGroup {
	return &ActionGroup{
		Name:      name,
		Count:     0,
		Names:     NewActionsCount(),
		Senders:   NewActionsCount(),
		Receivers: NewActionsCount(),
	}
}

type GroupedActions struct {
	Actions        map[string]*ActionGroup
	GroupedBy      string
	BlocksCount    uint64
	ActionsCount   uint64
	actionProperty ActionProperty
	detailed       bool
}

var groupedActionsSerializer = structomap.New().
	PickFunc(func(actions interface{}) interface{} {
		var results []*ActionGroup
		for _, action := range actions.(map[string]*ActionGroup) {
			results = append(results, action)
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].Count > results[j].Count
		})
		if len(results) > maxTopLevelResults {
			results = results[:maxTopLevelResults]
		}
		return results
	}, "Actions").
	Pick("GroupedBy", "BlocksCount", "ActionsCount")

func (g *GroupedActions) MarshalJSON() ([]byte, error) {
	return json.Marshal(groupedActionsSerializer.Transform(g))
}

func (g *GroupedActions) Get(key string) *ActionGroup {
	return g.Actions[key]
}

func (g *GroupedActions) GetCount(key string) uint64 {
	group := g.Get(key)
	if group == nil {
		return 0
	}
	return group.Count
}

func NewGroupedActions(by ActionProperty, detailed bool) *GroupedActions {
	actions := make(map[string]*ActionGroup)
	return &GroupedActions{
		Actions:        actions,
		GroupedBy:      by.String(),
		BlocksCount:    0,
		ActionsCount:   0,
		actionProperty: by,
		detailed:       detailed,
	}
}

func (g *GroupedActions) getActionKey(action Action) string {
	switch g.actionProperty {
	case ActionName:
		return action.Name()
	case ActionSender:
		return action.Sender()
	case ActionReceiver:
		return action.Receiver()
	default:
		panic(fmt.Errorf("no such property %d", g.actionProperty))
	}
}

func (g *GroupedActions) AddBlock(block Block) {
	g.BlocksCount += 1
	for _, action := range block.ListActions() {
		g.ActionsCount += 1
		key := g.getActionKey(action)
		if key == "" {
			continue
		}
		actionGroup, ok := g.Actions[key]
		if !ok {
			actionGroup = NewActionGroup(key)
			g.Actions[key] = actionGroup
		}
		actionGroup.Count += 1
		if g.detailed {
			actionGroup.Names.Increment(action.Name())
			actionGroup.Senders.Increment(action.Sender())
			actionGroup.Receivers.Increment(action.Receiver())
		}
	}
}

func (g *GroupedActions) Result() interface{} {
	return g
}

type TransactionCounter int

func NewTransactionCounter() *TransactionCounter {
	value := 0
	return (*TransactionCounter)(&value)
}

func (t *TransactionCounter) AddBlock(block Block) {
	*t += (TransactionCounter)(block.TransactionsCount())
}

func (t *TransactionCounter) Result() interface{} {
	return t
}

type MissingBlocks struct {
	Start uint64
	End   uint64
	Seen  map[uint64]bool
}

func NewMissingBlocks(start, end uint64) *MissingBlocks {
	return &MissingBlocks{
		Start: start,
		End:   end,
		Seen:  make(map[uint64]bool),
	}
}

func (t *MissingBlocks) AddBlock(block Block) {
	t.Seen[block.Number()] = true
}

func (t *MissingBlocks) Compute() []uint64 {
	missing := make([]uint64, 0)
	for blockNumber := t.Start; blockNumber <= t.End; blockNumber++ {
		if _, ok := t.Seen[blockNumber]; !ok {
			missing = append(missing, blockNumber)
		}
	}
	return missing
}

func (t *MissingBlocks) Result() interface{} {
	return t.Compute()
}
