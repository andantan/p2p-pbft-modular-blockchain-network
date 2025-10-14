package pbft

import (
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/consensus"
	"github.com/andantan/modular-blockchain/util"
	"github.com/go-kit/log"
)

type PbftProposer struct {
	logger            log.Logger
	privKey           *crypto.PrivateKey
	chain             core.Chain
	virtualMemoryPool core.VirtualMemoryPool
}

func NewPbftProposer(k *crypto.PrivateKey, c core.Chain, mp core.VirtualMemoryPool) *PbftProposer {
	return &PbftProposer{
		logger:            util.LoggerWithPrefixes("Proposer"),
		privKey:           k,
		chain:             c,
		virtualMemoryPool: mp,
	}
}

func (p *PbftProposer) Createblock() (*block.Block, error) {
	ph, err := p.chain.GetCurrentHeader()

	if err != nil {
		return nil, err
	}

	txx, err := p.virtualMemoryPool.Pending()

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

func (p *PbftProposer) ProposeBlock(b *block.Block) (consensus.ConsensusMessage, error) {
	sequence := b.Header.Height
	view := uint64(0)

	msg := NewPbftPrePrepareMessage(view, sequence, b, p.privKey.PublicKey())

	if err := msg.Sign(p.privKey); err != nil {
		return nil, err
	}

	return msg, nil
}
