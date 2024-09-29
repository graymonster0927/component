package transferstatechecker

import "fmt"

type StateType int
type StateRel map[int]map[int]struct{}

type TransferStateChecker struct {
	stateRelMap map[StateType]StateRel
}

func (t *TransferStateChecker) SetStateRel(stateType StateType, stateRel StateRel) {
	t.stateRelMap[stateType] = stateRel
}
func (t *TransferStateChecker) Check(stateType StateType, state int, transferState int) bool {
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

func Test() {
	checker := NewTransferStateChecker()

	var stateTypeOrder StateType = 1
	var stateOrderDefault = 1
	var stateOrderIng = 1
	var stateOrderDone = 1
	var stateOrderClosed = 1
	var stateOrderCancel = 1

	checker.SetStateRel(stateTypeOrder, map[int]map[int]struct{}{
		stateOrderDefault: {
			stateOrderIng: {},
		},
		stateOrderIng: {
			stateOrderDone:   {},
			stateOrderClosed: {},
			stateOrderCancel: {},
		},
	})

	fmt.Println(checker.Check(stateTypeOrder, stateOrderDefault, stateOrderIng))
}
