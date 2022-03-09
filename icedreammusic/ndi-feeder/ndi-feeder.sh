#!/bin/bash -e

target_url="${1:-icecast://source:source@127.0.0.1:61120/main}"
ffmpeg_pids=()

call_ffmpeg() {
    command ffmpeg -hide_banner "$@"
}

daemon_ffmpeg() {
    call_ffmpeg "$@" &
    ffmpeg_pids+=($!)
}

shutdown_ffmpeg() {
    if is_ffmpeg_running; then
        kill "$ffmpeg_pid" || true
        for t in $(seq 0 10); do
            if ! kill -0 "$ffmpeg_pid"; then
                break
            fi
            sleep 1
        done
        if kill -0 "$ffmpeg_pid"; then
            kill -9 "$ffmpeg_pid" || true
        fi
    fi
    ffmpeg_pid=
}

is_ffmpeg_running() {
    [ -n "$ffmpeg_pid" ] && kill -0 "$ffmpeg_pid"
}

on_exit() {
    shutdown_ffmpeg
}
trap on_exit EXIT

offline=0

while true; do
    found_audio_source=""

    while read -r line; do
        declare -a "found_source=($(sed -e 's/"/\\"/g' -e "s/'/\"/g" -e 's/[][`~!@#$%^&*():;<>.,?/\|{}=+-]/\\&/g' <<<"$line"))"
        found_source[0]=$(sed -e 's/\\\([`~!@#$%^&*():;<>.,?/\|{}=+-]\)/\1/g' <<<"${found_source[0]}")
        found_source[1]=$(sed -e 's/\\\([`~!@#$%^&*():;<>.,?/\|{}=+-]\)/\1/g' <<<"${found_source[1]}")
        case "${found_source[0]}" in
        *\(IDHPC\ Main\ Audio\))
            found_audio_source="${found_source[0]}"
            ;;
        esac
    done < <(call_ffmpeg -loglevel info -extra_ips 192.168.188.21 -find_sources true -f libndi_newtek -i "dummy" 2>&1 | grep -Po "'(.+)'\s+'(.+)" | tee)

    if [ -z "$found_audio_source" ]; then
        offline=$((offline + 1))
    else
        offline=0
    fi

    if ! is_ffmpeg_running && [ -n "$found_audio_source" ]; then
        echo "starting ffmpeg with audio source: $found_audio_source" >&2

        call_ffmpeg -loglevel warning \
            -analyzeduration 1 -f libndi_newtek -extra_ips 192.168.188.21 -i "$found_audio_source" \
            -map a -c:a pcm_s16le -ar 48000 -ac 2 -f s16le - |
            fakesilence --samplerate 48000 --channels 2 --silence-threshold 125ms |
            call_ffmpeg -loglevel warning \
                -ar 48000 -channels 2 -f s16le -i - \
                -map a -c:a flac -f ogg -content_type application/ogg "${target_url}" || true

        # HACK - can't use the standard mpegts here, but liquidsoap will happily accept anything ffmpeg can parse (by default)… so let's just use nut here even though it feels super duper wrong
    elif is_ffmpeg_running && [ -z "$found_audio_source" ] && [ "$offline" -gt 0 ]; then
        echo "shutting down ffmpeg since no source has been found" >&2
        shutdown_ffmpeg # it won't shut down by itself unfortunately
    fi

    sleep 1
done
