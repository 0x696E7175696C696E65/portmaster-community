fn main() {
  tauri_build::try_build(tauri_build::Attributes::new().app_manifest(
    tauri_build::AppManifest::new().commands(&[
      "get_app_info", "get_service_manager_status", "start_service",
      "get_state", "set_state", "should_show", "should_handle_prompts",
      "send_tauri_http_request", "open_dir",
    ]),
  )).expect("failed to generate application command permissions");
}
