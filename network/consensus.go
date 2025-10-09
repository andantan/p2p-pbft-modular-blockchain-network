package network

type ConsensusEngine interface {
	StartEngine()
	HandleMessage(ConsensusMessage)
	StopEngine()
}
