use bytes::Bytes;
use std::sync::Arc;
use tokio::task;
use webrtc::api::APIBuilder;
use webrtc::api::media_engine::MediaEngine;
use webrtc::media::Sample;
use webrtc::peer_connection::configuration::RTCConfiguration;
use webrtc::rtp_transceiver::rtp_codec::RTCRtpCodecCapability;
use webrtc::track::track_local::track_local_static_sample::TrackLocalStaticSample;

use gstreamer as gst;
use gstreamer::prelude::*;
use gstreamer_app::{AppSink, AppSinkCallbacks};

async fn start_webrtc_stream() -> Result<(), Box<dyn std::error::Error>> {
    // ✅ Initialize GStreamer
    gst::init()?;

    // ✅ Create a WebRTC MediaEngine
    let mut media_engine = MediaEngine::default();
    media_engine.register_default_codecs()?;

    // ✅ Create WebRTC API instance
    let api = APIBuilder::new().with_media_engine(media_engine).build();

    // ✅ Define WebRTC configuration
    let config = RTCConfiguration {
        ice_servers: vec![], // No ICE servers for now
        ..Default::default()
    };

    // ✅ Create a PeerConnection
    let peer_connection = Arc::new(api.new_peer_connection(config).await?);

    // ✅ Create a WebRTC video track
    let video_track = Arc::new(TrackLocalStaticSample::new(
        RTCRtpCodecCapability {
            mime_type: "video/vp8".to_owned(),
            clock_rate: 90000,
            ..Default::default()
        },
        "video".to_owned(),
        "webrtc-rs".to_owned(),
    ));

    // ✅ Add track to PeerConnection
    peer_connection.add_track(video_track.clone()).await?;

    // ✅ Manually Create GStreamer Elements (Latest API)
    let pipeline = gst::Pipeline::new();
    let source = gst::ElementFactory::make("autovideosrc").build()?;
    let convert = gst::ElementFactory::make("videoconvert").build()?;
    let scale = gst::ElementFactory::make("videoscale").build()?;
    let sink_element = gst::ElementFactory::make("appsink").build()?; // Wrap as Element
    let sink = sink_element
        .clone()
        .downcast::<AppSink>()
        .expect("Sink element is not an AppSink");

    // ✅ Add elements to pipeline
    pipeline.add_many(&[&source, &convert, &scale, &sink_element])?; // Use `sink_element`, not `sink`

    // ✅ Link elements manually
    source.link(&convert)?;
    convert.link(&scale)?;
    scale.link(&sink_element)?; // Link to `sink_element`, not `sink`

    let video_track_clone = video_track.clone();

    // ✅ Set up GStreamer app sink callbacks to process frames
    sink.set_callbacks(
        AppSinkCallbacks::builder()
            .new_sample(move |sink| {
                let sample = sink.pull_sample().map_err(|_| gst::FlowError::Eos)?;
                let buffer = sample.buffer().ok_or(gst::FlowError::Error)?;

                // ✅ Convert buffer to readable format
                let map = buffer.map_readable().map_err(|_| gst::FlowError::Error)?;

                // ✅ Convert to Bytes format for WebRTC
                let sample_data: Bytes = map.to_vec().into();

                let video_track_clone = video_track_clone.clone();
                let timestamp = std::time::SystemTime::now(); // ✅ FIXED TIMESTAMP

                task::spawn(async move {
                    let _ = video_track_clone
                        .write_sample(&Sample {
                            data: sample_data,
                            duration: std::time::Duration::from_millis(33), // ✅ FIXED DURATION
                            timestamp,                                      // ✅ FIXED TIMESTAMP
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

    pipeline.set_state(gst::State::Playing)?;

    println!("🚀 Streaming video... Press Ctrl+C to stop.");

    // ✅ Keep the app running
    tokio::signal::ctrl_c().await?;
    pipeline.set_state(gst::State::Null)?;
    peer_connection.close().await?;

    Ok(())
}

#[tokio::main]
async fn main() {
    if let Err(err) = start_webrtc_stream().await {
        eprintln!("❌ Error: {}", err);
    }
}
