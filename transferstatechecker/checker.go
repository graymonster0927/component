package transferstatechecker

type StateType int
type State int
type StateRel map[State]map[State]struct{}

type TransferStateChecker struct {
	stateRelMap map[StateType]StateRel
}

func (t *TransferStateChecker) SetStateRel(stateType StateType, stateRel StateRel) {
	t.stateRelMap[stateType] = stateRel
}
func (t *TransferStateChecker) Check(stateType StateType, state State, transferState State) bool {
	statRel, ok := t.stateRelMap[stateType]
	if !ok {
		return false
	}

	_, ok = statRel[state][transferState]
	return ok

}

func NewTransferStateChecker() *TransferStateChecker {
	return &TransferStateChecker{
		stateRelMap: make(map[StateType]StateRel),
	}
}
