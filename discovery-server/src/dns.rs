use std::net::SocketAddr;
use axum::Router;
use axum::routing::{get, post};
use crate::{peer, handler};

pub struct DNS {
    address: SocketAddr,
    server: Router,
    store: peer::PeerStore
}

impl DNS {
    pub fn new_dns_server(address: SocketAddr) -> Self {
        let store = peer::PeerStore::new();
        let server = Router::new()
            .route("/register", post(handler::post_register_handler))
            .route("/peers", get(handler::get_peers_handler))
            .route("/validators", get(handler::get_validators_handler))
            .route("/heartbeat", post(handler::post_heartbeat_handler))
            .route("/deregister", post(handler::post_deregister_handler))
            .with_state(store.clone());

        Self {
            address,
            server,
            store
        }
    }

    pub async fn run(&self) -> Result<(), std::io::Error> {
        println!("Peer discovery dns server running: {}", self.address);

        let store_for_pruning = self.store.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(tokio::time::Duration::from_secs(60));

            loop {
                interval.tick().await;
                store_for_pruning.prune_stale_peers(1).await;
            }
        });

        let listener = tokio::net::TcpListener::bind(self.address).await?;
        axum::serve(listener, self.server.clone()).await
    }
}