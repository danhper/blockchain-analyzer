package core

import "encoding/json"

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

func (a *ActionsCount) Persist(outputFile string) error {
	file, err := CreateFile(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	return encoder.Encode(a.actions)
}
