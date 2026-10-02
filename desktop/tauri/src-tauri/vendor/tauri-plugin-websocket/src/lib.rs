// Copyright 2019-2023 Tauri Programme within The Commons Conservancy
// SPDX-License-Identifier: Apache-2.0
// SPDX-License-Identifier: MIT

//! Open a WebSocket connection using a Rust client in JS.

#![doc(
    html_logo_url = "https://github.com/tauri-apps/tauri/raw/dev/app-icon.png",
    html_favicon_url = "https://github.com/tauri-apps/tauri/raw/dev/app-icon.png"
)]

use futures_util::{stream::SplitSink, SinkExt, StreamExt};
use http::header::{HeaderName, HeaderValue};
use serde::{ser::Serializer, Deserialize, Serialize};
use tauri::{
    ipc::Channel,
    plugin::{Builder as PluginBuilder, TauriPlugin},
    Manager, Runtime, State, Window,
};
use tokio::{net::TcpStream, sync::Mutex};
use tokio_tungstenite::connect_async_with_config;
use tokio_tungstenite::{
    tungstenite::{
        client::IntoClientRequest,
        protocol::{CloseFrame as ProtocolCloseFrame, WebSocketConfig},
        Message,
    },
    MaybeTlsStream, WebSocketStream,
};

use std::collections::HashMap;
use std::str::FromStr;

type Id = u32;
type WebSocket = WebSocketStream<MaybeTlsStream<TcpStream>>;
type WebSocketWriter = SplitSink<WebSocket, Message>;
type Result<T> = std::result::Result<T, Error>;

fn local_database_url(raw: &str) -> Result<String> {
    let mut url = url::Url::parse(raw).map_err(|_| Error::InvalidEndpoint)?;
    if url.scheme() != "ws"
        || !matches!(url.host_str(), Some("127.0.0.1" | "localhost"))
        || url.port() != Some(817)
        || url.path() != "/api/database/v1"
        || !url.username().is_empty()
        || url.password().is_some()
        || url.fragment().is_some()
    {
        return Err(Error::InvalidEndpoint);
    }
    // Preserve localhost callers without depending on host-file/DNS resolution.
    url.set_host(Some("127.0.0.1")).map_err(|_| Error::InvalidEndpoint)?;
    Ok(url.into())
}

fn local_connection_limits() -> WebSocketConfig {
    WebSocketConfig::default()
        .read_buffer_size(64 * 1024)
        .write_buffer_size(64 * 1024)
        .max_write_buffer_size(8 * 1024 * 1024)
        .max_message_size(Some(8 * 1024 * 1024))
        .max_frame_size(Some(8 * 1024 * 1024))
        .accept_unmasked_frames(false)
}

#[derive(Debug, thiserror::Error)]
enum Error {
    #[error("native WebSocket permits only the local Portmaster database API")]
    InvalidEndpoint,
    #[error("native WebSocket does not permit caller-supplied headers")]
    InvalidHeaders,
    #[error("native WebSocket connection limit reached")]
    ConnectionLimit,
    #[error("native WebSocket operation timed out")]
    Timeout,
    #[error("native WebSocket message exceeds 8 MiB")]
    MessageLimit,
    #[error(transparent)]
    Websocket(#[from] tokio_tungstenite::tungstenite::Error),
    #[error("connection not found for the given id: {0}")]
    ConnectionNotFound(Id),
    #[error(transparent)]
    InvalidHeaderValue(#[from] tokio_tungstenite::tungstenite::http::header::InvalidHeaderValue),
    #[error(transparent)]
    InvalidHeaderName(#[from] tokio_tungstenite::tungstenite::http::header::InvalidHeaderName),
}

impl Serialize for Error {
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serializer.serialize_str(self.to_string().as_str())
    }
}

struct ConnectionManager(Mutex<HashMap<Id, WebSocketWriter>>, std::sync::Arc<tokio::sync::Semaphore>);

impl Default for ConnectionManager {
    fn default() -> Self {
        Self(Mutex::new(HashMap::new()), std::sync::Arc::new(tokio::sync::Semaphore::new(8)))
    }
}


#[derive(Deserialize)]
#[serde(untagged, rename_all = "camelCase")]
enum Max {
    None,
    Number(usize),
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
pub(crate) struct ConnectionConfig {
    pub read_buffer_size: Option<usize>,
    pub write_buffer_size: Option<usize>,
    pub max_write_buffer_size: Option<usize>,
    pub max_message_size: Option<Max>,
    pub max_frame_size: Option<Max>,
    #[serde(default)]
    pub accept_unmasked_frames: bool,
    pub headers: Option<Vec<(String, String)>>,
}

impl From<ConnectionConfig> for WebSocketConfig {
    fn from(config: ConnectionConfig) -> Self {
        let mut builder =
            WebSocketConfig::default().accept_unmasked_frames(config.accept_unmasked_frames);

        if let Some(read_buffer_size) = config.read_buffer_size {
            builder = builder.read_buffer_size(read_buffer_size)
        }

        if let Some(write_buffer_size) = config.write_buffer_size {
            builder = builder.write_buffer_size(write_buffer_size)
        }

        if let Some(max_write_buffer_size) = config.max_write_buffer_size {
            builder = builder.max_write_buffer_size(max_write_buffer_size)
        }

        if let Some(max_message_size) = config.max_message_size {
            let max_size = match max_message_size {
                Max::None => Option::None,
                Max::Number(n) => Some(n),
            };
            builder = builder.max_message_size(max_size);
        }

        if let Some(max_frame_size) = config.max_frame_size {
            let max_size = match max_frame_size {
                Max::None => Option::None,
                Max::Number(n) => Some(n),
            };
            builder = builder.max_frame_size(max_size);
        }

        builder
    }
}

#[derive(Deserialize, Serialize)]
struct CloseFrame {
    pub code: u16,
    pub reason: String,
}

#[derive(Deserialize, Serialize)]
#[serde(tag = "type", content = "data")]
enum WebSocketMessage {
    Text(String),
    Binary(Vec<u8>),
    Ping(Vec<u8>),
    Pong(Vec<u8>),
    Close(Option<CloseFrame>),
}

#[tauri::command]
async fn connect<R: Runtime>(
    window: Window<R>,
    url: String,
    on_message: Channel<serde_json::Value>,
    config: Option<ConnectionConfig>,
) -> Result<Id> {
    let url = local_database_url(&url)?;
    let connection_slot = window.state::<ConnectionManager>().1.clone()
        .try_acquire_owned().map_err(|_| Error::ConnectionLimit)?;
    let id = rand::random();
    let mut request = url.into_client_request()?;

    if let Some(headers) = config.as_ref().and_then(|c| c.headers.as_ref()) {
        if !headers.is_empty() {
            return Err(Error::InvalidHeaders);
        }
        for (k, v) in headers {
            let header_name = HeaderName::from_str(k.as_str())?;
            let header_value = HeaderValue::from_str(v.as_str())?;
            request.headers_mut().insert(header_name, header_value);
        }
    }

    let (ws_stream, _) = tokio::time::timeout(std::time::Duration::from_secs(10),
        connect_async_with_config(request, Some(local_connection_limits()), false))
        .await.map_err(|_| Error::Timeout)??;

    tauri::async_runtime::spawn(async move {
        let _connection_slot = connection_slot;
        let (write, read) = ws_stream.split();
        let manager = window.state::<ConnectionManager>();
        manager.0.lock().await.insert(id, write);
        let cleanup_window = window.clone();
        read.for_each(move |message| {
            let window_ = window.clone();
            let on_message_ = on_message.clone();
            async move {
                if let Ok(Message::Close(_)) = message {
                    let manager = window_.state::<ConnectionManager>();
                    manager.0.lock().await.remove(&id);
                }

                let response = match message {
                    Ok(Message::Text(t)) => {
                        serde_json::to_value(WebSocketMessage::Text(t.to_string())).unwrap()
                    }
                    Ok(Message::Binary(t)) => {
                        serde_json::to_value(WebSocketMessage::Binary(t.to_vec())).unwrap()
                    }
                    Ok(Message::Ping(t)) => {
                        serde_json::to_value(WebSocketMessage::Ping(t.to_vec())).unwrap()
                    }
                    Ok(Message::Pong(t)) => {
                        serde_json::to_value(WebSocketMessage::Pong(t.to_vec())).unwrap()
                    }
                    Ok(Message::Close(t)) => {
                        serde_json::to_value(WebSocketMessage::Close(t.map(|v| CloseFrame {
                            code: v.code.into(),
                            reason: v.reason.to_string(),
                        })))
                        .unwrap()
                    }
                    Ok(Message::Frame(_)) => serde_json::Value::Null, // This value can't be recieved.
                    Err(e) => serde_json::to_value(Error::from(e)).unwrap(),
                };

                let _ = on_message_.send(response);
            }
        })
        .await;
        // Errors/EOF also close a session; retain neither the writer nor its slot.
        cleanup_window.state::<ConnectionManager>().0.lock().await.remove(&id);
    });

    Ok(id)
}

#[cfg(test)]
mod community_tests {
    use super::*;

    #[test]
    fn rejects_non_database_targets_and_url_confusion() {
        for raw in [
            "wss://127.0.0.1:817/api/database/v1",
            "ws://127.0.0.1:818/api/database/v1",
            "ws://192.0.2.1:817/api/database/v1",
            "ws://127.0.0.1:817.attacker.test/api/database/v1",
            "ws://127.0.0.1:817@attacker.test/api/database/v1",
            "ws://attacker@127.0.0.1:817/api/database/v1",
            "ws://127.0.0.1:817/api/database/v1/../../private",
            "ws://127.0.0.1:817/api/database/v1#fragment",
            "file:///etc/passwd",
        ] {
            assert!(local_database_url(raw).is_err(), "accepted {raw}");
        }
        assert_eq!(local_database_url("ws://localhost:817/api/database/v1").unwrap(), "ws://127.0.0.1:817/api/database/v1");
        assert!(local_database_url("ws://127.0.0.1:817/api/database/v1").is_ok());
    }

    #[test]
    fn native_frame_buffers_remain_bounded() {
        let limits = local_connection_limits();
        assert_eq!(limits.max_message_size, Some(8 * 1024 * 1024));
        assert_eq!(limits.max_frame_size, Some(8 * 1024 * 1024));
        assert!(!limits.accept_unmasked_frames);
    }

    #[test]
    fn concurrent_sessions_are_capped_and_slots_recover() {
        let manager = ConnectionManager::default();
        let mut slots = Vec::new();
        for _ in 0..8 {
            slots.push(manager.1.clone().try_acquire_owned().unwrap());
        }
        assert!(manager.1.clone().try_acquire_owned().is_err());
        slots.pop();
        assert!(manager.1.clone().try_acquire_owned().is_ok());
    }
}

#[tauri::command]
async fn send(
    manager: State<'_, ConnectionManager>,
    id: Id,
    message: WebSocketMessage,
) -> Result<()> {
    let length = match &message {
        WebSocketMessage::Text(value) => value.len(),
        WebSocketMessage::Binary(value) | WebSocketMessage::Ping(value) | WebSocketMessage::Pong(value) => value.len(),
        WebSocketMessage::Close(frame) => frame.as_ref().map_or(0, |value| value.reason.len()),
    };
    if length > 8 * 1024 * 1024 {
        return Err(Error::MessageLimit);
    }
    if let Some(write) = manager.0.lock().await.get_mut(&id) {
        tokio::time::timeout(std::time::Duration::from_secs(10), write
            .send(match message {
                WebSocketMessage::Text(t) => Message::Text(t.into()),
                WebSocketMessage::Binary(t) => Message::Binary(t.into()),
                WebSocketMessage::Ping(t) => Message::Ping(t.into()),
                WebSocketMessage::Pong(t) => Message::Pong(t.into()),
                WebSocketMessage::Close(t) => Message::Close(t.map(|v| ProtocolCloseFrame {
                    code: v.code.into(),
                    reason: v.reason.into(),
                })),
            }))
            .await.map_err(|_| Error::Timeout)??;
        Ok(())
    } else {
        Err(Error::ConnectionNotFound(id))
    }
}

pub fn init<R: Runtime>() -> TauriPlugin<R> {
    Builder::default().build()
}

#[derive(Default)]
pub struct Builder {}

impl Builder {
    pub fn new() -> Self {
        Self {}
    }

    pub fn build<R: Runtime>(self) -> TauriPlugin<R> {
        PluginBuilder::new("websocket")
            .invoke_handler(tauri::generate_handler![connect, send])
            .setup(|app, _api| {
                app.manage(ConnectionManager::default());
                Ok(())
            })
            .build()
    }
}
