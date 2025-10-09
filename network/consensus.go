package network

type ConsensusEngine interface {
	Start()
	HandleMessage(ConsensusMessage)
	Finalize()
	Stop()
}
