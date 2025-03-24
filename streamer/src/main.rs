use bytes::Bytes;
use std::sync::Arc;
use tokio::task;
use webrtc::api::APIBuilder;
use webrtc::api::media_engine::MediaEngine;
use webrtc::media::Sample;
use webrtc::peer_connection::configuration::RTCConfiguration;
// Removed unused import: use webrtc::peer_connection::sdp::session_description::RTCSessionDescription;
use tokio_tungstenite::connect_async;
use tokio_tungstenite::tungstenite::Message;
use futures_util::{StreamExt, SinkExt};
// Removed unused import: use url::Url;
use webrtc::rtp_transceiver::rtp_codec::RTCRtpCodecCapability;
use webrtc::track::track_local::track_local_static_sample::TrackLocalStaticSample;

use gstreamer as gst;
use gstreamer::prelude::*;
use gstreamer_app::{AppSink, AppSinkCallbacks};

async fn start_webrtc_stream() -> Result<(), Box<dyn std::error::Error>> {
    // ✅ Initialize GStreamer
    gst::init()?;

    // ✅ Connect to Signaling Server
    let signaling_server_url = "ws://localhost:8080/ws";
    let (ws_stream, _) = connect_async(signaling_server_url).await?;
    let (mut write, _read) = ws_stream.split(); // Prefix unused variable with underscore

    // ✅ Define WebRTC configuration (ICE servers for NAT traversal can be added later)
    let config = RTCConfiguration {
        ice_servers: vec![],
        ..Default::default()
    };

    let api = APIBuilder::new().with_media_engine(MediaEngine::default()).build();
    // ✅ Create a WebRTC PeerConnection
    let peer_connection = Arc::new(api.new_peer_connection(config).await?);
    let offer = peer_connection.create_offer(None).await?;
    peer_connection.set_local_description(offer.clone()).await?;

    // ✅ Send offer to the signaling server
    let offer_json = serde_json::to_string(&offer)?;
    println!("📡 Sending WebRTC Offer: {}", offer_json);
    write.send(Message::Text(offer_json.into())).await?;

    // ✅ Create a WebRTC video track (VP8 Codec, 90kHz clock rate)
    let video_track = Arc::new(TrackLocalStaticSample::new(
        RTCRtpCodecCapability {
            mime_type: "video/vp8".to_owned(),
            clock_rate: 90000,
            ..Default::default()
        },
        "video".to_owned(),
        "webrtc-rs".to_owned(),
    ));

    // ✅ Add the video track to the PeerConnection
    peer_connection.add_track(video_track.clone()).await?;

    // ✅ Manually Create GStreamer Elements
    let pipeline = gst::Pipeline::new(); // Pipeline contains the entire flow of elements
    let source = gst::ElementFactory::make("autovideosrc").build()?; // Video source (webcam)
    let convert = gst::ElementFactory::make("videoconvert").build()?; // Converts video format
    let scale = gst::ElementFactory::make("videoscale").build()?; // Adjusts video scaling
    let sink_element = gst::ElementFactory::make("appsink").build()?; // AppSink receives frames

    // ✅ Convert `sink_element` into `AppSink`
    let sink = sink_element
        .clone()
        .downcast::<AppSink>()
        .expect("Sink element is not an AppSink");

    // ✅ Add elements to pipeline
    pipeline.add_many(&[&source, &convert, &scale, &sink_element])?;

    // ✅ Link elements manually (Data flow: source -> convert -> scale -> appsink)
    source.link(&convert)?;
    convert.link(&scale)?;
    scale.link(&sink_element)?;

    let video_track_clone = video_track.clone();

    // ✅ Set up GStreamer AppSink to handle video frames
    sink.set_callbacks(
        AppSinkCallbacks::builder()
            .new_sample(move |sink| {
                let sample = sink.pull_sample().map_err(|_| gst::FlowError::Eos)?;
                let buffer = sample.buffer().ok_or(gst::FlowError::Error)?;

                // ✅ Convert buffer into readable format
                let map = buffer.map_readable().map_err(|_| gst::FlowError::Error)?;

                // ✅ Convert buffer to Bytes format (WebRTC compatible)
                let sample_data: Bytes = map.to_vec().into();

                let video_track_clone = video_track_clone.clone();
                let timestamp = std::time::SystemTime::now(); // ✅ Set frame timestamp

                let runtime = tokio::runtime::Runtime::new().unwrap();
                runtime.block_on(async move {  // ✅ Fix: Use Tokio runtime
                    let _ = video_track_clone
                        .write_sample(&Sample {
                            data: sample_data,
                            duration: std::time::Duration::from_millis(33),
                            timestamp,
                            prev_dropped_packets: 0,
                            prev_padding_packets: 0,
                            packet_timestamp: 0,
                        })
                        .await;
                });

                Ok(gst::FlowSuccess::Ok)
            })
            .build(),
    );

    // ✅ Start the GStreamer pipeline
    pipeline.set_state(gst::State::Playing)?;

    println!("🚀 Streaming video... Press Ctrl+C to stop.");

    // ✅ Keep the app running until user stops it
    tokio::signal::ctrl_c().await?;
    pipeline.set_state(gst::State::Null)?;
    peer_connection.close().await?;

    Ok(())
}

#[cfg(target_os = "macos")]
extern crate cocoa;

#[cfg(target_os = "macos")]
use cocoa::appkit::NSApplication;
#[cfg(target_os = "macos")]
use cocoa::base::nil;

fn initialize_macos_ui() {
    #[cfg(target_os = "macos")]
    unsafe {
        let ns_app = NSApplication::sharedApplication(nil);
        ns_app.activateIgnoringOtherApps_(true);
    }
}

#[tokio::main]
async fn main() {
    initialize_macos_ui(); // 🛠️ Ensure NSApplication is running

    if let Err(err) = start_webrtc_stream().await {
        eprintln!("❌ Error: {}", err);
    }
}