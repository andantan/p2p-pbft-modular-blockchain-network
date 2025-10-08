package pbft

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/go-kit/log"
)

type PbftProposer struct {
	logger  log.Logger
	privKey *crypto.PrivateKey
}

func NewPbftProposer(k *crypto.PrivateKey) *PbftProposer {
	return &PbftProposer{
		logger:  util.LoggerWithPrefixes("Proposer"),
		privKey: k,
	}
}

func (p *PbftProposer) Createblock(c core.Chain, mp core.VirtualMemoryPool) (*block.Block, error) {
	ph, err := c.GetCurrentHeader()

	if err != nil {
		return nil, err
	}

	txx, err := mp.Pending()

	if err != nil {
		return nil, err
	}

	bd := block.NewBody(txx)
	b, err := block.NewBlockFromPrevHeader(ph, bd)

	if err != nil {
		return nil, err
	}

	if err = b.Sign(p.privKey); err != nil {
		return nil, err
	}

	if err = b.Verify(); err != nil {
		return nil, err
	}

	h, _ := b.Hash()
	_ = p.logger.Log("msg", "new block created", "hash", h.ShortString(8), "height", b.Header.Height, "weight", b.Header.Weight)

	return b, nil
}

func (p *PbftProposer) ProposeBlock(b *block.Block) (network.Message, error) {
	sequence := b.Header.Height
	view := uint64(0)

	msg := NewPbftPrePrepareMessage(view, sequence, b, p.privKey.PublicKey())

	if err := msg.Sign(p.privKey); err != nil {
		return nil, err
	}

	return msg, nil
}
