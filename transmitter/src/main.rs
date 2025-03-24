use actix::{Actor, StreamHandler};
use actix_web::{web, App, HttpRequest, HttpResponse, HttpServer};
use actix_web_actors::ws;
use log::info;
use once_cell::sync::Lazy;
use serde_json::Value;
use serde_urlencoded;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

// Global shared store for streamers
static STREAMERS: Lazy<Arc<Mutex<HashMap<String, StreamerWebSocket>>>> =
    Lazy::new(|| Arc::new(Mutex::new(HashMap::new())));

// WebSocket Actor for Streamers
#[derive(Clone)]
struct StreamerWebSocket {
    name: String,
    supporters: Vec<String>,
}

impl Actor for StreamerWebSocket {
    type Context = ws::WebsocketContext<Self>;

    fn started(&mut self, _: &mut Self::Context) {
        let mut store = STREAMERS.lock().unwrap();
        store.insert(self.name.clone(), self.clone());
        info!("📡 Streamer '{}' added to global store", self.name);
    }

    fn stopped(&mut self, _: &mut Self::Context) {
        let mut store = STREAMERS.lock().unwrap();
        store.remove(&self.name);
        info!(
            "❌ Streamer '{}' disconnected, removing supporters: {:?}",
            self.name, self.supporters
        );
    }
}

impl StreamerWebSocket {
    fn add_supporter(&mut self, supporter_id: String) {
        if !self.supporters.contains(&supporter_id) {
            self.supporters.push(supporter_id.clone());
            info!(
                "🙌 Supporter '{}' added to Streamer '{}'",
                supporter_id, self.name
            );
        }
    }

    fn remove_supporter(&mut self, supporter_id: &String) {
        if let Some(pos) = self.supporters.iter().position(|x| x == supporter_id) {
            self.supporters.remove(pos);
            info!(
                "❌ Supporter '{}' removed from Streamer '{}'",
                supporter_id, self.name
            );
        }
    }
}

impl StreamHandler<Result<ws::Message, ws::ProtocolError>> for StreamerWebSocket {
    fn handle(&mut self, msg: Result<ws::Message, ws::ProtocolError>, ctx: &mut Self::Context) {
        if let Ok(ws::Message::Text(text)) = msg {
            info!("📡 Streamer '{}' received message: {}", self.name, text);

            match serde_json::from_str::<Value>(&text) {
                Ok(json) => {
                    if let Some(command) = json.get("command").and_then(|c| c.as_str()) {
                        match command {
                            "list" => {
                                let supporters_json =
                                    serde_json::to_string(&self.supporters).unwrap();
                                ctx.text(supporters_json);
                            }
                            "whois" => {
                                let response = format!(r#"{{ "name": "{}" }}"#, self.name);
                                ctx.text(response);
                            }
                            _ => {
                                ctx.text(r#"{"error": "Unknown command"}"#);
                            }
                        }
                    } else {
                        ctx.text(r#"{"error": "Invalid command format"}"#);
                    }
                }
                Err(_) => {
                    ctx.text(r#"{"error": "Invalid JSON"}"#);
                }
            }
        }
    }
}

// WebSocket Actor for Supporters
struct SupporterWebSocket {
    supporter_id: String,
    streamer_id: String,
}

impl Actor for SupporterWebSocket {
    type Context = ws::WebsocketContext<Self>;

    fn started(&mut self, ctx: &mut Self::Context) {
        let mut store = STREAMERS.lock().unwrap();
        if let Some(streamer) = store.get_mut(&self.streamer_id) {
            streamer.add_supporter(self.supporter_id.clone());
            info!(
                "🙌 Supporter '{}' connected to streamer '{}'",
                self.supporter_id, self.streamer_id
            );
            ctx.text(format!(
                "Connected as Supporter: {} to Streamer: {}",
                self.supporter_id, self.streamer_id
            ));
        }
    }

    fn stopped(&mut self, _: &mut Self::Context) {
        let mut store = STREAMERS.lock().unwrap();
        if let Some(streamer) = store.get_mut(&self.streamer_id) {
            streamer.remove_supporter(&self.supporter_id);
            info!(
                "❌ Supporter '{}' disconnected from streamer '{}'",
                self.supporter_id, self.streamer_id
            );
        }
    }
}

impl StreamHandler<Result<ws::Message, ws::ProtocolError>> for SupporterWebSocket {
    fn handle(&mut self, msg: Result<ws::Message, ws::ProtocolError>, ctx: &mut Self::Context) {
        if let Ok(ws::Message::Text(text)) = msg {
            info!("💬 Supporter received message: {}", text);
            ctx.text(format!("Supporter: {}", text));
        }
    }
}

// WebSocket handler for streamers
async fn streamer_ws(
    req: HttpRequest,
    stream: web::Payload,
) -> Result<HttpResponse, actix_web::Error> {
    let query_string = req.query_string();
    let params: HashMap<String, String> =
        serde_urlencoded::from_str(query_string).unwrap_or_default();

    let streamer_id = match params.get("id") {
        Some(id) if !id.is_empty() => id.clone(),
        _ => {
            info!("❌ Connection rejected: missing 'id' query parameter");
            return Ok(actix_web::HttpResponse::BadRequest().body("Missing 'id' query parameter"));
        }
    };

    info!("🎥 New streamer '{}' connected", streamer_id);
    ws::start(
        StreamerWebSocket {
            name: streamer_id,
            supporters: Vec::new(),
        },
        &req,
        stream,
    )
    .map_err(|e| e.into())
}

// WebSocket handler for supporters
async fn supporter_ws(
    req: HttpRequest,
    stream: web::Payload,
) -> Result<HttpResponse, actix_web::Error> {
    let query_string = req.query_string();
    let params: HashMap<String, String> =
        serde_urlencoded::from_str(query_string).unwrap_or_default();

    let supporter_id = match params.get("id") {
        Some(id) if !id.is_empty() => id.clone(),
        _ => {
            info!("❌ Connection rejected: missing 'id' query parameter");
            return Ok(actix_web::HttpResponse::BadRequest().body("Missing 'id' query parameter"));
        }
    };

    let streamer_id = match params.get("streamer_id") {
        Some(id) if !id.is_empty() => id.clone(),
        _ => {
            info!("❌ Connection rejected: missing 'streamer_id' query parameter");
            return Ok(
                actix_web::HttpResponse::BadRequest().body("Missing 'streamer_id' query parameter")
            );
        }
    };

    // Check if the streamer exists and add supporter
    let mut store = STREAMERS.lock().unwrap();
    if let Some(streamer) = store.get_mut(&streamer_id) {
        streamer.add_supporter(supporter_id.clone());
    } else {
        info!(
            "❌ Connection rejected: streamer '{}' does not exist",
            streamer_id
        );
        return Ok(actix_web::HttpResponse::BadRequest().body("Streamer does not exist"));
    }

    info!(
        "🙌 Supporter '{}' is connecting to Streamer '{}'",
        supporter_id, streamer_id
    );

    ws::start(
        SupporterWebSocket {
            supporter_id,
            streamer_id,
        },
        &req,
        stream,
    )
    .map_err(|e| e.into())
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();
    info!("🚀 Server is starting at ws://127.0.0.1:8080");

    HttpServer::new(move || {
        App::new()
            .route("/streamer", web::get().to(streamer_ws))
            .route("/supporter", web::get().to(supporter_ws))
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await
}
