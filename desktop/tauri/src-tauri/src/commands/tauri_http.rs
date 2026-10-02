use reqwest::{header::HeaderName, Client, Method, Response};
use serde::{Deserialize, Serialize};
use std::time::Duration;
use tauri::State;
use url::Url;

const MAX_REQUEST_BYTES: usize = 8 * 1024 * 1024;
const MAX_RESPONSE_BYTES: usize = 64 * 1024 * 1024;

/// Native HTTP exists only to reach the local Portmaster API. Do not let
/// webview content turn it into an unrestricted browser-origin bypass.
pub fn create_http_client() -> Client {
    Client::builder()
        .pool_max_idle_per_host(10)
        .cookie_store(true)
        .redirect(reqwest::redirect::Policy::none())
        .no_proxy()
        .connect_timeout(Duration::from_secs(5))
        .timeout(Duration::from_secs(30))
        .user_agent("Portmaster Community UI")
        .build()
        .expect("failed to build HTTP client")
}

fn local_api_url(raw: &str) -> Result<Url, String> {
    let mut url = Url::parse(raw).map_err(|_| "invalid API URL".to_string())?;
    if url.scheme() != "http"
        || !matches!(url.host_str(), Some("127.0.0.1" | "localhost"))
        || url.port() != Some(817)
        || !url.username().is_empty()
        || url.password().is_some()
        || url.fragment().is_some()
        || !url.path().starts_with("/api/")
    {
        return Err("native HTTP permits only http://127.0.0.1:817/api/".into());
    }
    // The development client uses localhost. Connect to the literal address
    // for both spellings, so hostname resolution cannot change this boundary.
    url.set_host(Some("127.0.0.1")).map_err(|_| "invalid API URL".to_string())?;
    Ok(url)
}

fn safe_header(raw: &str) -> Result<HeaderName, String> {
    let header = HeaderName::from_bytes(raw.as_bytes()).map_err(|err| err.to_string())?;
    match header.as_str() {
        "host" | "connection" | "content-length" | "transfer-encoding" | "upgrade" => {
            return Err("native HTTP does not permit transport headers".into());
        }
        name if name.starts_with("proxy-") => {
            return Err("native HTTP does not permit proxy headers".into());
        }
        _ => {}
    }
    Ok(header)
}

#[derive(Deserialize)]
pub struct HttpRequestOptions {
    method: String,
    headers: Vec<(String, String)>,
    body: Option<Vec<u8>>,
}

#[derive(Serialize)]
pub struct HttpResponse {
    status: u16,
    status_text: String,
    headers: Vec<(String, String)>,
    body: Vec<u8>,
}

async fn bounded_body(mut response: Response, limit: usize) -> Result<Vec<u8>, String> {
    if response.content_length().is_some_and(|length| length > limit as u64) {
        return Err("API response exceeds native HTTP size limit".into());
    }
    let mut body = Vec::new();
    while let Some(chunk) = response.chunk().await.map_err(|err| err.to_string())? {
        if chunk.len() > limit.saturating_sub(body.len()) {
            return Err("API response exceeds native HTTP size limit".into());
        }
        body.extend_from_slice(&chunk);
    }
    Ok(body)
}

#[tauri::command]
pub async fn send_tauri_http_request(
    client: State<'_, Client>,
    url: String,
    opts: HttpRequestOptions,
) -> Result<HttpResponse, String> {
    let url = local_api_url(&url)?;
    let method = Method::from_bytes(opts.method.as_bytes()).map_err(|err| err.to_string())?;
    if !matches!(method, Method::GET | Method::HEAD | Method::POST | Method::PUT | Method::PATCH | Method::DELETE | Method::OPTIONS) {
        return Err("unsupported API method".into());
    }
    let mut request = client.request(method, url);
    for (key, value) in opts.headers {
        request = request.header(safe_header(&key)?, &value);
    }
    if let Some(body) = opts.body {
        if body.len() > MAX_REQUEST_BYTES {
            return Err("API request exceeds native HTTP size limit".into());
        }
        request = request.body(body);
    }
    let response = request.send().await.map_err(|err| err.to_string())?;
    let status = response.status().as_u16();
    let status_text = response.status().canonical_reason().unwrap_or("").to_string();
    let headers = response.headers().iter().map(|(key, value)| {
        (key.to_string(), value.to_str().unwrap_or("").to_string())
    }).collect();
    let body = bounded_body(response, MAX_RESPONSE_BYTES).await?;
    Ok(HttpResponse { status, status_text, headers, body })
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::{io::{Read, Write}, net::TcpListener};

    #[test]
    fn rejects_non_api_origins_and_url_confusion() {
        for raw in [
            "https://127.0.0.1:817/api/v1/core/status",
            "http://127.0.0.1:818/api/v1/core/status",
            "http://127.0.0.1:817.attacker.test/api/v1/",
            "http://127.0.0.1:817@attacker.test/api/v1/",
            "http://attacker:secret@127.0.0.1:817/api/v1/",
            "http://127.0.0.1:817/api/../../private",
            "http://127.0.0.1:817/api/v1/#fragment",
            "file:///etc/passwd",
            "http://192.0.2.1:817/api/v1/",
        ] {
            assert!(local_api_url(raw).is_err(), "accepted {raw}");
        }
        assert!(local_api_url("http://127.0.0.1:817/api/v1/core/status?refresh=true").is_ok());
        assert_eq!(local_api_url("http://localhost:817/api/v1/core/status").unwrap().host_str(), Some("127.0.0.1"));
    }

    #[test]
    fn rejects_transport_header_overrides() {
        for header in ["Host", "Content-Length", "Transfer-Encoding", "Connection", "Proxy-Authorization", "Upgrade"] {
            assert!(safe_header(header).is_err(), "accepted {header}");
        }
        assert!(safe_header("Content-Type").is_ok());
        assert!(safe_header("Accept").is_ok());
    }

    fn serve(response: &'static [u8]) -> (String, std::thread::JoinHandle<()>) {
        let listener = TcpListener::bind("127.0.0.1:0").unwrap();
        let address = listener.local_addr().unwrap();
        let handle = std::thread::spawn(move || {
            let (mut connection, _) = listener.accept().unwrap();
            connection.set_read_timeout(Some(Duration::from_secs(5))).unwrap();
            let mut request = [0; 4096];
            let _ = connection.read(&mut request).unwrap();
            connection.write_all(response).unwrap();
        });
        (format!("http://{address}/"), handle)
    }

    #[tokio::test]
    async fn local_redirect_does_not_escape_allowed_origin() {
        let (url, server) = serve(b"HTTP/1.1 302 Found\r\nLocation: http://192.0.2.1/private\r\nContent-Length: 0\r\nConnection: close\r\n\r\n");
        let response = create_http_client().get(url).send().await.unwrap();
        assert_eq!(response.status(), 302);
        server.join().unwrap();
    }

    #[tokio::test]
    async fn chunked_response_cannot_bypass_body_limit() {
        let (url, server) = serve(b"HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\nConnection: close\r\n\r\n8\r\n12345678\r\n8\r\n12345678\r\n0\r\n\r\n");
        let response = create_http_client().get(url).send().await.unwrap();
        assert!(bounded_body(response, 10).await.is_err());
        server.join().unwrap();
    }
}
