package entity

type StatusSet map[GroupParticipantStatus]bool

func newStatusSet(statuses ...GroupParticipantStatus) StatusSet {
	set := make(StatusSet, len(statuses))
	for _, status := range statuses {
		set[status] = true
	}

	return set
}

type StatusMatrix map[GroupParticipantStatus]StatusSet

func (mx StatusMatrix) IsCorrectTransit(from, to GroupParticipantStatus) bool {
	set, ok := mx[from]
	if !ok {
		return false
	}

	return set[to]
}

//nolint:exhaustive // these aren't enum switch statements
var (
	MxActionOnSomeone = StatusMatrix{
		JoinedStatus: newStatusSet(KickedStatus),
		KickedStatus: newStatusSet(JoinedStatus),
	}
	MxActionOnOneself = StatusMatrix{
		JoinedStatus: newStatusSet(LeftStatus),
		LeftStatus:   newStatusSet(JoinedStatus),
	}
)
