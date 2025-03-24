use actix::{Actor, Addr, AsyncContext, Handler, Message, StreamHandler};
use actix_web::{web, App, HttpRequest, HttpResponse, HttpServer};
use actix_web_actors::ws;
use log::info;
use once_cell::sync::Lazy;
use serde_json::Value;
use serde_urlencoded;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

// Global shared store for streamers
static STREAMERS: Lazy<Arc<Mutex<HashMap<String, Addr<StreamerWebSocket>>>>> =
    Lazy::new(|| Arc::new(Mutex::new(HashMap::new())));

// Actix messages for updating watchers
#[derive(Message)]
#[rtype(result = "()")]
struct AddWatcher {
    watcher_id: String,
    addr: Addr<WatcherWebSocket>,
}

#[derive(Message)]
#[rtype(result = "()")]
struct RemoveWatcher {
    watcher_id: String,
}

#[derive(Message)]
#[rtype(result = "()")]
struct GetWatchers;

#[derive(Message)]
#[rtype(result = "()")]
struct BroadcastMessage {
    message: String,
}

#[derive(Clone)]
struct StreamerWebSocket {
    name: String,
    watchers: HashMap<String, Addr<WatcherWebSocket>>,
}

impl Actor for StreamerWebSocket {
    type Context = ws::WebsocketContext<Self>;

    fn started(&mut self, ctx: &mut Self::Context) {
        let mut store = STREAMERS.lock().unwrap();
        store.insert(self.name.clone(), ctx.address()); // Store the actor address instead of cloning
        info!("📡 Streamer '{}' added to global store", self.name);
    }

    fn stopped(&mut self, _: &mut Self::Context) {
        let mut store = STREAMERS.lock().unwrap();
        store.remove(&self.name);
        info!(
            "❌ Streamer '{}' disconnected, removing all watchers",
            self.name
        );
    }
}

// Handle adding a watcher
impl Handler<AddWatcher> for StreamerWebSocket {
    type Result = ();

    fn handle(&mut self, msg: AddWatcher, _: &mut Self::Context) {
        self.watchers.insert(msg.watcher_id.clone(), msg.addr);
        info!(
            "🙌 Watcher '{}' added to Streamer '{}'",
            msg.watcher_id, self.name
        );
    }
}

// Handle removing a watcher
impl Handler<RemoveWatcher> for StreamerWebSocket {
    type Result = ();

    fn handle(&mut self, msg: RemoveWatcher, _: &mut Self::Context) {
        self.watchers.remove(&msg.watcher_id);
        info!(
            "❌ Watcher '{}' removed from Streamer '{}'",
            msg.watcher_id, self.name
        );
    }
}

// Handle broadcast messages in StreamerWebSocket
impl Handler<BroadcastMessage> for StreamerWebSocket {
    type Result = ();

    fn handle(&mut self, msg: BroadcastMessage, _: &mut Self::Context) {
        info!("📢 Streamer '{}' broadcasting: {}", self.name, msg.message);

        for (_, watcher_addr) in &self.watchers {
            watcher_addr.do_send(BroadcastMessage {
                message: msg.message.clone(),
            });
        }
    }
}

// WebSocket Stream Handler for messages
impl StreamHandler<Result<ws::Message, ws::ProtocolError>> for StreamerWebSocket {
    fn handle(&mut self, msg: Result<ws::Message, ws::ProtocolError>, ctx: &mut Self::Context) {
        if let Ok(ws::Message::Text(text)) = msg {
            info!("📡 Streamer '{}' received message: {}", self.name, text);

            match serde_json::from_str::<Value>(&text) {
                Ok(json) => {
                    if let Some(command) = json.get("command").and_then(|c| c.as_str()) {
                        match command {
                            "list" => {
                                let watchers_list: Vec<String> =
                                    self.watchers.keys().cloned().collect();
                                let watchers_json = serde_json::to_string(&watchers_list).unwrap();
                                ctx.text(watchers_json);
                            }
                            "whois" => {
                                let response = format!(r#"{{ "name": "{}" }}"#, self.name);
                                ctx.text(response);
                            }
                            "broadcast" => {
                                if let Some(message) = json.get("message").and_then(|m| m.as_str())
                                {
                                    info!(
                                        "📢 Streamer '{}' is broadcasting: {}",
                                        self.name, message
                                    );
                                    for watcher_addr in self.watchers.values() {
                                        watcher_addr.do_send(BroadcastMessage {
                                            message: message.to_string(),
                                        });
                                    }
                                }
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

// WebSocket Actor for Watchers
struct WatcherWebSocket {
    watcher_id: String,
    streamer_id: String,
}

impl Actor for WatcherWebSocket {
    type Context = ws::WebsocketContext<Self>;

    fn started(&mut self, ctx: &mut Self::Context) {
        let store = STREAMERS.lock().unwrap();
        if let Some(streamer) = store.get(&self.streamer_id) {
            let watcher_addr = ctx.address(); // Get the correct watcher address
            streamer.do_send(AddWatcher {
                watcher_id: self.watcher_id.clone(),
                addr: watcher_addr,
            });
            info!(
                "🙌 Watcher '{}' connected to streamer '{}'",
                self.watcher_id, self.streamer_id
            );
            ctx.text(format!(
                "Connected as Watcher: {} to Streamer: {}",
                self.watcher_id, self.streamer_id
            ));
        }
    }

    fn stopped(&mut self, _: &mut Self::Context) {
        let store = STREAMERS.lock().unwrap();
        if let Some(streamer) = store.get(&self.streamer_id) {
            streamer.do_send(RemoveWatcher {
                watcher_id: self.watcher_id.clone(),
            });
            info!(
                "❌ Watcher '{}' disconnected from streamer '{}'",
                self.watcher_id, self.streamer_id
            );
        }
    }
}

// Implement StreamHandler for WatcherWebSocket
impl StreamHandler<Result<ws::Message, ws::ProtocolError>> for WatcherWebSocket {
    fn handle(&mut self, msg: Result<ws::Message, ws::ProtocolError>, ctx: &mut Self::Context) {
        if let Ok(ws::Message::Text(text)) = msg {
            info!(
                "💬 Watcher '{}' received message: {}",
                self.watcher_id, text
            );

            match serde_json::from_str::<Value>(&text) {
                Ok(json) => {
                    if let Some(command) = json.get("command").and_then(|c| c.as_str()) {
                        match command {
                            "broadcast" => {
                                if let Some(message) = json.get("message").and_then(|m| m.as_str())
                                {
                                    info!(
                                        "📢 Watcher '{}' is broadcasting: {}",
                                        self.watcher_id, message
                                    );
                                    let store = STREAMERS.lock().unwrap();
                                    if let Some(streamer) = store.get(&self.streamer_id) {
                                        streamer.do_send(BroadcastMessage {
                                            message: message.to_string(),
                                        });
                                    }
                                }
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

// Handle broadcast in watcher
impl Handler<BroadcastMessage> for WatcherWebSocket {
    type Result = ();

    fn handle(&mut self, msg: BroadcastMessage, ctx: &mut Self::Context) {
        info!(
            "📢 Watcher '{}' received broadcast: {}",
            self.watcher_id, msg.message
        );
        ctx.text(msg.message);
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
            watchers: HashMap::new(),
        },
        &req,
        stream,
    )
    .map_err(|e| e.into())
}

// WebSocket handler for watchers
async fn watcher_ws(
    req: HttpRequest,
    stream: web::Payload,
) -> Result<HttpResponse, actix_web::Error> {
    let query_string = req.query_string();
    let params: HashMap<String, String> =
        serde_urlencoded::from_str(query_string).unwrap_or_default();

    let watcher_id = match params.get("id") {
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

    // Check if the streamer exists
    let store = STREAMERS.lock().unwrap();
    if !store.contains_key(&streamer_id) {
        info!(
            "❌ Connection rejected: streamer '{}' does not exist",
            streamer_id
        );
        return Ok(actix_web::HttpResponse::BadRequest().body("Streamer does not exist"));
    }

    info!(
        "🙌 Watcher '{}' connecting to Streamer '{}'",
        watcher_id, streamer_id
    );

    // Start the WebSocket session for the watcher
    ws::start(
        WatcherWebSocket {
            watcher_id,
            streamer_id,
        },
        &req,
        stream,
    )
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();
    info!("🚀 Server is starting at ws://127.0.0.1:8080");

    HttpServer::new(move || {
        App::new()
            .route("/streamer", web::get().to(streamer_ws))
            .route("/watcher", web::get().to(watcher_ws))
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await
}
