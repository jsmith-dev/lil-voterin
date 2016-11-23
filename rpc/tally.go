package core

func GetTally() (*ResultGetTally, error) {
	tally := voter.GetTally()
	return &ResultGetTally{tally}, nil
}
