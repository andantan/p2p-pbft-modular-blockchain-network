package network

type Broadcaster interface {
	Broadcast(Message) error
}
