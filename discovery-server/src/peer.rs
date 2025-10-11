use std::collections::HashMap;
use std::sync::Arc;
use chrono::{DateTime, Utc};
use rand::prelude::SliceRandom;
use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PeerInfo {
    #[serde(rename = "address")]
    pub address: String,

    #[serde(rename = "net_addr")]
    pub net_addr: String,

    #[serde(rename = "connections")]
    pub connections: u8,

    #[serde(rename = "max_connections")]
    pub max_connections: u8,

    #[serde(rename = "height")]
    pub height: u64,

    #[serde(rename = "is_validator")]
    pub is_validator: bool,
}

#[derive(Debug, Clone)]
pub struct Peer {
    pub info: PeerInfo,
    pub last_seen: DateTime<Utc>,
}

impl Peer {
    pub fn new(info: PeerInfo) -> Self {
        Self {
            info,
            last_seen: Utc::now(),
        }
    }
}

pub type PeerMap = HashMap<String, Peer>;

#[derive(Clone)]
pub struct PeerStore {
    peers: Arc<RwLock<PeerMap>>,
}

impl PeerStore {
    pub fn new() -> Self {
        Self {
            peers: Arc::new(RwLock::new(PeerMap::new())),
        }
    }

    pub async fn register(&self, peer_info: PeerInfo) {
        let mut peer_map = self.peers.write().await;
        let peer_address = peer_info.address.clone();
        let peer_net_addr = peer_info.net_addr.clone();
        let peer = Peer::new(peer_info);

        peer_map.insert(peer_address.clone(), peer);

        println!("[Register] New peer registered: {peer_address} at {peer_net_addr}");
    }

    pub async fn get_peers(&self) -> Vec<PeerInfo> {
        let peer_map = self.peers.read().await;
        let mut infos: Vec<PeerInfo> = peer_map.values().map(|peer|
            peer.info.clone()
        ).collect();
        let count = std::cmp::min(infos.len(), 500);
        let mut rng = rand::rng();

        infos.shuffle(&mut rng);
        infos.into_iter().take(count).collect()
    }

    pub async fn get_validator_set(&self) -> Vec<PeerInfo> {
        let peer_map = self.peers.read().await;
        let infos: Vec<PeerInfo> = peer_map
            .values()
            .filter(|peer| { peer.info.is_validator })
            .map(|peer| peer.info.clone())
            .collect();

        infos
    }

    pub async fn heartbeat(&self, info: &PeerInfo) -> bool {
        let mut peer_map = self.peers.write().await;

        if let Some(peer) = peer_map.get_mut(&info.address) {
            println!("[Heartbeat] Peer status updated for: {}", info.address);
            peer.last_seen = Utc::now();
            peer.info = info.clone();

            return true;
        }

        println!("[Heartbeat] Peer not found or ID mismatch for addr: {} at {}", info.address, info.net_addr);

        false
    }

    pub async fn deregister(&self, peer_to_remove: &PeerInfo) -> Option<Peer> {
        let mut peer_map = self.peers.write().await;

        if let Some(peer) = peer_map.get(&peer_to_remove.address) {
            if peer.info.net_addr == peer_to_remove.net_addr {
                println!("[Deregister] Peer removed: {} at {}",
                         peer_to_remove.address, peer_to_remove.net_addr);
                return peer_map.remove(&peer_to_remove.address);
            }
        }

        println!("[Deregister] Peer not found or ID mismatch for addr: {} at {}",
                 peer_to_remove.address, peer_to_remove.net_addr);
        None
    }

    pub async fn prune_stale_peers(&self, threshold_minute: i64) {
        let mut peer_map = self.peers.write().await;
        let now = Utc::now();
        let stale_threshold = chrono::Duration::minutes(threshold_minute);

        peer_map.retain(|_, peer| {
            let time_since_last_seen = now.signed_duration_since(peer.last_seen);
            let is_fresh = time_since_last_seen < stale_threshold;

            if !is_fresh {
                println!("[Prune] Stale peer removed: {} at {}", peer.info.address, peer.info.net_addr);
            }

            is_fresh
        })
    }
}