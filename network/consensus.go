package network

type ConsensusEngine interface {
	Start()
	HandleMessage(ConsensusMessage) error
	Stop()
}
