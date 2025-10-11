use axum::extract::State;
use axum::http::StatusCode;
use axum::Json;
use serde::Serialize;

use crate::peer::{PeerInfo, PeerStore};

// POST /register
pub async fn post_register_handler(
    State(store): State<PeerStore>, Json(payload): Json<PeerInfo>
) -> StatusCode {
    store.register(payload).await;
    StatusCode::OK
}

// POST /heartbeat
pub async fn post_heartbeat_handler(
    State(store): State<PeerStore>, Json(payload): Json<PeerInfo>
) -> StatusCode {
    if store.heartbeat(&payload).await {
        StatusCode::OK
    } else {
        StatusCode::NOT_FOUND
    }
}

// POST /deregister
pub async fn post_deregister_handler(
    State(store): State<PeerStore>, Json(payload): Json<PeerInfo>
) -> StatusCode {
    if store.deregister(&payload).await.is_some() {
        StatusCode::OK
    } else {
        StatusCode::NOT_FOUND
    }
}

#[derive(Serialize)]
pub struct GetPeersResponse {
    peers: Vec<PeerInfo>,
}

// GET /peers
pub async fn get_peers_handler(State(store): State<PeerStore>) -> Json<GetPeersResponse> {
    let peers = store.get_peers().await;

    Json(GetPeersResponse { peers })
}

// GET /validators
pub async fn get_validators_handler(State(store): State<PeerStore>) -> Json<GetPeersResponse> {
    let validator_set = store.get_validator_set().await;

    Json(GetPeersResponse { peers: validator_set })
}