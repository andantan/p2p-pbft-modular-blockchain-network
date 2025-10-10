mod peer;
mod handler;
mod dns;

use std::net::SocketAddr;
use dns::DNS;


#[tokio::main]
async fn main() {
    let addr = SocketAddr::from(([0u8;4], 6550));

    if let Err(e) = DNS::new_dns_server(addr).run().await {
        eprintln!("occurred fatal error: {}", e);
    }
}
