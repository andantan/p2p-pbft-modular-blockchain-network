package tcp

//func TestTCPNode_ConnectAndAccept(t *testing.T) {
//	nodeANetAddr := ":4000"
//	nodeBNetAddr := ":5000"
//
//	privKeyA, _ := crypto.GeneratePrivateKey()
//	msgChA := make(chan network.RawMessage)
//	newPeerChA := make(chan *TcpPeer)
//	delPeerChA := make(chan *TcpPeer)
//	stopChA := make(chan struct{})
//	nodeA := NewTcpNode(privKeyA, nodeANetAddr, msgChA, newPeerChA, delPeerChA, stopChA)
//
//	privKeyB, _ := crypto.GeneratePrivateKey()
//	msgChB := make(chan network.RawMessage)
//	newPeerChB := make(chan *TcpPeer)
//	delPeerChB := make(chan *TcpPeer)
//	stopChB := make(chan struct{})
//	nodeB := NewTcpNode(privKeyB, nodeBNetAddr, msgChB, newPeerChB, delPeerChB, stopChB)
//
//	go nodeA.Listen()
//
//	t.Cleanup(func() {
//		nodeA.Close()
//	})
//
//	// delay for nodeA starting
//	time.Sleep(100 * time.Millisecond)
//
//	assert.NoError(t, nodeB.Connect(nodeANetAddr))
//
//	select {
//	case peer := <-nodeA.newPeerCh:
//		assert.NotNil(t, peer)
//		assert.Equal(t, nodeB.address, peer.Address())
//		assert.Equal(t, nodeB.pubKey, *peer.PublicKey())
//		assert.Equal(t, nodeBNetAddr, peer.NetAddr())
//
//	case <-time.After(2 * time.Second):
//		t.Fatal("Node A did not receive peer from B")
//	}
//
//	select {
//	case peer := <-nodeB.newPeerCh:
//		assert.NotNil(t, peer)
//		assert.Equal(t, nodeA.address, peer.Address())
//		assert.Equal(t, nodeA.pubKey, *peer.PublicKey())
//		assert.Equal(t, nodeANetAddr, peer.NetAddr())
//
//	case <-time.After(2 * time.Second):
//		t.Fatal("Node B did not receive peer from A")
//	}
//}
//
//func TestTCPNode_Stop(t *testing.T) {
//	nodeANetAddr := ":6000"
//	privKey, _ := crypto.GeneratePrivateKey()
//	imcomingMsgCh := make(chan network.RawMessage)
//	newPeerCh := make(chan *TcpPeer)
//	delPeerCh := make(chan *TcpPeer)
//	stopCh := make(chan struct{})
//
//	defer close(imcomingMsgCh)
//	defer close(newPeerCh)
//	defer close(delPeerCh)
//
//	node := NewTcpNode(privKey, nodeANetAddr, imcomingMsgCh, newPeerCh, delPeerCh, stopCh)
//
//	go node.Listen()
//
//	// delay for node starting
//	time.Sleep(100 * time.Millisecond)
//
//	close(stopCh)
//	node.Close()
//
//	// delay for node stopping
//	time.Sleep(100 * time.Millisecond)
//
//	ln, err := net.Listen("tcp", nodeANetAddr)
//
//	// if listener closed normally, err must be nil
//	assert.NoError(t, err)
//	if err == nil {
//		_ = ln.Close()
//	}
//}
//
//func TestTCPNode_TieBreaking(t *testing.T) {
//	privKeyA, err := crypto.GeneratePrivateKey()
//	assert.NoError(t, err)
//	nodeANetAddr := ":7000"
//	msgChA := make(chan network.RawMessage)
//	newPeerChA := make(chan *TcpPeer)
//	delPeerChA := make(chan *TcpPeer)
//	stopChA := make(chan struct{})
//	nodeA := NewTcpNode(privKeyA, nodeANetAddr, msgChA, newPeerChA, delPeerChA, stopChA)
//
//	privKeyB, err := crypto.GeneratePrivateKey()
//	assert.NoError(t, err)
//	nodeBNetAddr := ":8000"
//	msgChB := make(chan network.RawMessage)
//	newPeerChB := make(chan *TcpPeer)
//	delPeerChB := make(chan *TcpPeer)
//	stopChB := make(chan struct{})
//	nodeB := NewTcpNode(privKeyB, nodeBNetAddr, msgChB, newPeerChB, delPeerChB, stopChB)
//
//	go nodeA.Listen()
//	t.Cleanup(func() {
//		nodeA.Close()
//	})
//	go nodeB.Listen()
//	t.Cleanup(func() {
//		nodeB.Close()
//	})
//	// delay for nodeA, B starting
//	time.Sleep(100 * time.Millisecond)
//
//	wg := new(sync.WaitGroup)
//	wg.Add(2)
//
//	go func() {
//		defer wg.Done()
//		_ = nodeA.Connect(nodeBNetAddr)
//	}()
//
//	go func() {
//		defer wg.Done()
//		_ = nodeB.Connect(nodeANetAddr)
//	}()
//
//	wg.Wait()
//
//	// delay for nodeA, B handshaking
//	time.Sleep(200 * time.Millisecond)
//
//	go func() {
//		for range nodeA.newPeerCh {
//		}
//		for range nodeB.newPeerCh {
//		}
//	}()
//
//	assert.Equal(t, nodeA.peerSet.Len(), 1)
//	assert.Equal(t, nodeB.peerSet.Len(), 1)
//}
