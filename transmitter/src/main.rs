use actix::ActorFutureExt;
use actix::ContextFutureSpawner;
use actix::{Actor, Addr, AsyncContext, Handler, Message, StreamHandler, WrapFuture};
use actix_web::{web, App, HttpRequest, HttpResponse, HttpServer};
use actix_web_actors::ws;
use log::info;
use once_cell::sync::Lazy;
use serde_json::Value;
use serde_urlencoded;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

// Global shared store for rooms

static ROOMS: Lazy<Arc<Mutex<HashMap<String, Addr<RoomActor>>>>> =
    Lazy::new(|| Arc::new(Mutex::new(HashMap::new())));

// Actix messages for managing members
#[derive(Message)]
#[rtype(result = "()")]
struct AddMember {
    member_id: String,
    addr: Addr<MemberWebSocket>,
}

#[derive(Message)]
#[rtype(result = "()")]
struct RemoveMember {
    member_id: String,
}

#[derive(Message)]
#[rtype(result = "()")]
struct BroadcastMessage {
    message: String,
}

// Define a custom message for closing WebSocket connections
#[derive(Message)]
#[rtype(result = "()")]
struct CloseConnection;

// Remove duplicate `GetMembers` struct
#[derive(Message)]
#[rtype(result = "Vec<String>")]
struct GetMembers;

// Room actor to manage members
#[derive(Clone)]
struct RoomActor {
    room_id: String,
    members: HashMap<String, Addr<MemberWebSocket>>,
}

impl Actor for RoomActor {
    type Context = actix::Context<Self>; // Use regular Actix context

    fn started(&mut self, ctx: &mut Self::Context) {
        let mut store = ROOMS.lock().unwrap();
        store.insert(self.room_id.clone(), ctx.address()); // Store the actor address
        info!("📡 Room '{}' created", self.room_id);
    }

    fn stopped(&mut self, _: &mut Self::Context) {
        let mut store = ROOMS.lock().unwrap();
        store.remove(&self.room_id);
        info!("❌ Room '{}' removed", self.room_id);
    }
}

// Implement GetMembers handler in RoomActor
impl Handler<GetMembers> for RoomActor {
    type Result = Vec<String>;

    fn handle(&mut self, _: GetMembers, _: &mut Self::Context) -> Self::Result {
        self.members.keys().cloned().collect()
    }
}

// Handle adding a member
impl Handler<AddMember> for RoomActor {
    type Result = ();

    fn handle(&mut self, msg: AddMember, _: &mut Self::Context) {
        // Check if the member already exists
        if let Some(existing_addr) = self.members.get(&msg.member_id) {
            info!(
                "⚠️ Member '{}' already exists in Room '{}'. Replacing connection.",
                msg.member_id, self.room_id
            );

            // Send termination signal using the new CloseConnection message
            existing_addr.do_send(CloseConnection);
        }

        // Replace with the new connection
        self.members.insert(msg.member_id.clone(), msg.addr);
        info!(
            "🙌 Member '{}' added to Room '{}'",
            msg.member_id, self.room_id
        );
    }
}

// Handle removing a member
impl Handler<RemoveMember> for RoomActor {
    type Result = ();

    fn handle(&mut self, msg: RemoveMember, _: &mut Self::Context) {
        self.members.remove(&msg.member_id);
        info!(
            "❌ Member '{}' removed from Room '{}'",
            msg.member_id, self.room_id
        );
    }
}

// Handle broadcast messages in RoomActor
impl Handler<BroadcastMessage> for RoomActor {
    type Result = ();

    fn handle(&mut self, msg: BroadcastMessage, _: &mut Self::Context) {
        info!("📢 Room '{}' broadcasting: {}", self.room_id, msg.message);

        for (_, member_addr) in &self.members {
            member_addr.do_send(BroadcastMessage {
                message: msg.message.clone(),
            });
        }
    }
}

// WebSocket Stream Handler for messages
impl StreamHandler<Result<ws::Message, ws::ProtocolError>> for RoomActor {
    fn handle(&mut self, _msg: Result<ws::Message, ws::ProtocolError>, _ctx: &mut Self::Context) {
        // RoomActor should not handle WebSocket messages directly
    }
}

// WebSocket Actor for Members
struct MemberWebSocket {
    member_id: String,
    room_id: String,
}

impl Actor for MemberWebSocket {
    type Context = ws::WebsocketContext<Self>;

    fn started(&mut self, ctx: &mut Self::Context) {
        let store = ROOMS.lock().unwrap();
        if let Some(room) = store.get(&self.room_id) {
            let member_addr = ctx.address(); // Get the correct member address
            room.do_send(AddMember {
                member_id: self.member_id.clone(),
                addr: member_addr,
            });
            info!(
                "🙌 Member '{}' connected to Room '{}'",
                self.member_id, self.room_id
            );
            ctx.text(format!(
                "Connected as Member: {} to Room: {}",
                self.member_id, self.room_id
            ));
        }
    }

    fn stopped(&mut self, _: &mut Self::Context) {
        let store = ROOMS.lock().unwrap();
        if let Some(room) = store.get(&self.room_id) {
            room.do_send(RemoveMember {
                member_id: self.member_id.clone(),
            });
            info!(
                "❌ Member '{}' disconnected from Room '{}'",
                self.member_id, self.room_id
            );
        }
    }
}

// Implement the handler in MemberWebSocket
impl Handler<CloseConnection> for MemberWebSocket {
    type Result = ();

    fn handle(&mut self, _: CloseConnection, ctx: &mut Self::Context) {
        ctx.close(Some(ws::CloseReason {
            code: ws::CloseCode::Normal,
            description: Some("Replaced by new connection".to_string()),
        }));
    }
}

// Implement StreamHandler for MemberWebSocket
impl StreamHandler<Result<ws::Message, ws::ProtocolError>> for MemberWebSocket {
    fn handle(&mut self, msg: Result<ws::Message, ws::ProtocolError>, ctx: &mut Self::Context) {
        if let Ok(ws::Message::Text(text)) = msg {
            info!("💬 Member '{}' received message: {}", self.member_id, text);

            match serde_json::from_str::<Value>(&text) {
                Ok(json) => {
                    if let Some(command) = json.get("command").and_then(|c| c.as_str()) {
                        match command {
                            "list" => {
                                let store = ROOMS.lock().unwrap();
                                if let Some(room) = store.get(&self.room_id) {
                                    let addr = room.clone();
                                    addr.send(GetMembers)
                                        .into_actor(self)
                                        .then(|res, _act, ctx| {
                                            if let Ok(members) = res {
                                                let response = serde_json::to_string(&members)
                                                    .unwrap_or_else(|_| "[]".to_string());
                                                ctx.text(response);
                                            }
                                            actix::fut::ready(())
                                        })
                                        .wait(ctx);
                                }
                            }
                            "whois" => {
                                let response =
                                    format!(r#"{{ "member_id": "{}" }}"#, self.member_id);
                                ctx.text(response);
                            }
                            "broadcast" => {
                                if let Some(message) = json.get("message").and_then(|m| m.as_str())
                                {
                                    info!(
                                        "📢 Member '{}' is broadcasting: {}",
                                        self.member_id, message
                                    );
                                    let store = ROOMS.lock().unwrap();
                                    if let Some(room) = store.get(&self.room_id) {
                                        room.do_send(BroadcastMessage {
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

// Handle broadcast in member
impl Handler<BroadcastMessage> for MemberWebSocket {
    type Result = ();

    fn handle(&mut self, msg: BroadcastMessage, ctx: &mut Self::Context) {
        info!(
            "📢 Member '{}' received broadcast: {}",
            self.member_id, msg.message
        );
        ctx.text(msg.message);
    }
}

// WebSocket handler for rooms
async fn room_ws(req: HttpRequest, stream: web::Payload) -> Result<HttpResponse, actix_web::Error> {
    let query_string = req.query_string();
    let params: HashMap<String, String> =
        serde_urlencoded::from_str(query_string).unwrap_or_default();

    let room_id = match params.get("room_id") {
        Some(id) if !id.is_empty() => id.clone(),
        _ => {
            info!("❌ Connection rejected: missing 'room_id' query parameter");
            return Ok(HttpResponse::BadRequest().body("Missing 'room_id' query parameter"));
        }
    };

    let member_id = match params.get("member_id") {
        Some(id) if !id.is_empty() => id.clone(),
        _ => {
            info!("❌ Connection rejected: missing 'member_id' query parameter");
            return Ok(HttpResponse::BadRequest().body("Missing 'member_id' query parameter"));
        }
    };

    // Check if the room exists, if not create it
    let mut store = ROOMS.lock().unwrap();
    let _room_actor = store.entry(room_id.clone()).or_insert_with(|| {
        RoomActor {
            room_id: room_id.clone(),
            members: HashMap::new(),
        }
        .start() // Now correctly starts as an Actix actor
    });

    ws::start(MemberWebSocket { member_id, room_id }, &req, stream).map_err(|e| e.into())
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();
    info!("🚀 Server is starting at ws://127.0.0.1:8080");

    HttpServer::new(move || App::new().route("/room", web::get().to(room_ws)))
        .bind("127.0.0.1:8080")?
        .run()
        .await
}
