#!/bin/bash -ex

target_url="${1:-srt://127.0.0.1:9000}"
ffmpeg_pid=

call_ffmpeg() {
    command ffmpeg -hide_banner "$@"
}

daemon_ffmpeg() {
    call_ffmpeg "$@" &
    ffmpeg_pid=$!
}

shutdown_ffmpeg() {
    if is_ffmpeg_running
    then
        kill "$ffmpeg_pid"
        wait "$ffmpeg_pid"
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

while true
do
    found_audio_source=""

    while read -r line
    do
        declare -a "found_source=($(sed -e 's/"/\\"/g' -e "s/'/\"/g" -e 's/[][`~!@#$%^&*():;<>.,?/\|{}=+-]/\\&/g' <<< "$line"))"
        found_source[0]=$(sed -e 's/\\\([`~!@#$%^&*():;<>.,?/\|{}=+-]\)/\1/g' <<< "${found_source[0]}")
        found_source[1]=$(sed -e 's/\\\([`~!@#$%^&*():;<>.,?/\|{}=+-]\)/\1/g' <<< "${found_source[1]}")
        case "${found_source[0]}" in
        *\(IDHPC\ Main\ Audio\))
            found_audio_source="${found_source[0]}"
            ;;
        esac
    done < <(call_ffmpeg -loglevel info -extra_ips 192.168.188.21 -find_sources true -f libndi_newtek -i "dummy" 2>&1 | grep -Po "'(.+)'\s+'(.+)" | tee)

    if ! is_ffmpeg_running && [ -n "$found_audio_source" ]
    then
        echo "starting ffmpeg with audio source: $found_audio_source" >&2
        # HACK - can't use the standard mpegts here, but liquidsoap will happily accept anything ffmpeg can parse (by default)… so let's just use nut here even though it feels super duper wrong
        daemon_ffmpeg -loglevel warning -extra_ips 192.168.188.21 -f libndi_newtek -i "$found_audio_source" -c copy -f nut -write_index false "${target_url}"
    elif is_ffmpeg_running && [ -z "$found_audio_source" ]
    then
        echo "shutting down ffmpeg since no source has been found" >&2
        shutdown_ffmpeg # it won't shut down by itself unfortunately
    fi
done
