#!/bin/bash -e

#target_url="${1:-icecast://source:source@127.0.0.1:61120/main}"
: "${TARGET_IP:=127.0.0.1}"
: "${TARGET_PORT:=61120}"
: "${TARGET_MOUNT:=/main}"
: "${TARGET_USERNAME:=source}"
: "${TARGET_PASSWORD:=source}"
: "${NDI_FEEDER_EXTRA_IP:=}"

gstreamer_pids=()

call_gstreamer() {
    command gst-launch-1.0 "$@"
}

daemon_gstreamer() {
    call_gstreamer "$@" &
    gstreamer_pids+=($!)
}

shutdown_gstreamer() {
    if is_gstreamer_running; then
        kill "$gstreamer_pid" || true
        for t in $(seq 0 10); do
            if ! kill -0 "$gstreamer_pid"; then
                break
            fi
            sleep 1
        done
        if kill -0 "$gstreamer_pid"; then
            kill -9 "$gstreamer_pid" || true
        fi
    fi
    gstreamer_pid=
}

is_gstreamer_running() {
    [ -n "$gstreamer_pid" ] && kill -0 "$gstreamer_pid"
}

on_exit() {
    shutdown_gstreamer
}
trap on_exit EXIT

offline=0

url_address=()
if [ -n "$NDI_FEEDER_EXTRA_IP" ]; then
    url_address=("url-address=$NDI_FEEDER_EXTRA_IP:5961")
fi

while true; do
    found_audio_source="$(grep --line-buffered -m 1 --color=none -Po 'ndi-name = \K.+\(ID.* Main Audio.*\)$'  < <(gst-device-monitor-1.0 -f Source/Network:application/x-ndi))"

    if [ -z "$found_audio_source" ]; then
        offline=$((offline + 1))
    else
        offline=0
    fi

    if ! is_gstreamer_running && [ -n "$found_audio_source" ]; then
        echo "starting gstreamer with audio source: $found_audio_source" >&2

        call_gstreamer ndisrc ndi-name="$found_audio_source" "${url_address[@]}" ! ndisrcdemux name=demux \
            demux.audio ! queue ! audioconvert ! audio/x-raw, channels=2, rate=48000, format=S16LE ! filesink location=/dev/stdout |
            fakesilence --samplerate 48000 --channels 2 --silence-threshold 125ms |
            daemon_gstreamer filesrc location=/dev/stdin ! rawaudioparse use-sink-caps=false format=pcm pcm-format=s16le sample-rate=48000 num-channels=2 ! queue ! audioconvert ! audioresample ! flacenc ! oggmux ! shout2send mount="$TARGET_MOUNT" port="$TARGET_PORT" username="$TARGET_USERNAME" password="$TARGET_PASSWORD" ip="$TARGET_IP"
    elif is_gstreamer_running && [ -z "$found_audio_source" ] && [ "$offline" -gt 0 ]; then
        echo "shutting down gstreamer since no source has been found" >&2
        shutdown_gstreamer # it won't shut down by itself unfortunately
    fi

    sleep 1
done
