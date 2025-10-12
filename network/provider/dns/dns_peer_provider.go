package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/andantan/modular-blockchain/network/provider"
	"io"
	"net/http"
)

type DnsPeerProvider struct {
	dnsServerAddr string
	dnsUri        string
}

func NewDnsPeerProvider(dnsServerAddr string) *DnsPeerProvider {
	return &DnsPeerProvider{
		dnsServerAddr: dnsServerAddr,
		dnsUri:        fmt.Sprintf("http://%s", dnsServerAddr),
	}
}

func (p *DnsPeerProvider) DiscoverPeers() ([]*provider.PeerInfo, error) {
	return p.getPeersFromEndpoint("/peers")
}

func (p *DnsPeerProvider) GetValidators() ([]*provider.PeerInfo, error) {
	return p.getPeersFromEndpoint("/validators")
}

func (p *DnsPeerProvider) Register(our *provider.PeerInfo) error {
	return p.postJSON("/register", our)
}

func (p *DnsPeerProvider) Deregister(our *provider.PeerInfo) error {
	return p.postJSON("/deregister", our)
}

func (p *DnsPeerProvider) Heartbeat(our *provider.PeerInfo) error {
	return p.postJSON("/heartbeat", our)
}

func (p *DnsPeerProvider) postJSON(endpoint string, data any) error {
	url := p.dnsUri + endpoint

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request to %s failed with status: %s", endpoint, resp.Status)
	}

	return nil
}

type PeerSliceResponse struct {
	Peers []*provider.PeerInfo `json:"peers"`
}

func (p *DnsPeerProvider) getPeersFromEndpoint(endpoint string) ([]*provider.PeerInfo, error) {
	url := p.dnsUri + endpoint

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s failed with status: %s", endpoint, resp.Status)
	}

	var peerResp PeerSliceResponse
	if err = json.NewDecoder(resp.Body).Decode(&peerResp); err != nil {
		return nil, err
	}

	return peerResp.Peers, nil
}
