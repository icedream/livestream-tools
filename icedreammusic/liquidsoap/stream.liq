set("log.stdout", true)
set("log.file", false)

set("log.level", 4)

set("server.telnet", true)
set("server.telnet.bind_addr", "127.0.0.1")
set("server.telnet.port", 21337)

set("init.allow_root",true)
set("frame.video.width", 1920)
set("frame.video.height", 1080)
set("audio.converter.samplerate.libsamplerate.quality", "best")
set("audio.converter.samplerate.native.quality","linear")
set("sandbox", "disabled")

%include "settings.liq"
%include "metadata_api.liq"
%include "stream_api.liq"
%include "silent_fallback.liq"

s = input.http(id="input_ice_main", max_buffer=4., "http://127.0.0.1:61120/main")

# Split audio off to be handled specially
a = drop_video(id="main", s)
a = setup_harbor_metadata_api(a)
a = mksafe_soft(a)
output.dummy(a)

def append_encoder_meta(m) =
    [("encoder", "Liquidsoap #{liquidsoap.version}")]
end
a = map_metadata(append_encoder_meta, a)

internal_icecast=output.icecast(
    fallible=true,
    port=61120,
    host="127.0.0.1",
    user=internal_icecast_username,
    password=internal_icecast_password,
    name=stream_name,
    description=stream_description)
setup_harbor_stream_api(internal_icecast(
    id="out_a_int_vorbis",
    # %ogg(%flac),
    mount="/outa/vorbis",
    start=false,
    %ffmpeg(
        format="ogg",
        %audio(
            codec="libvorbis",
            flags=2, # CODEC_FLAG_QSCALE (enables VBR mode based on qscale aka global_quality)
            global_quality=1180., # gets divided by FF_QP2LAMBDA=118
        ),
    ),
    a,
))
setup_harbor_stream_api(internal_icecast(
    id="out_a_int_flac",
    mount="/outa/flac",
    start=true,
    %ffmpeg(
        format="ogg",
        %audio(
            codec="flac",
        ),
    ),
    a,
))
setup_harbor_stream_api(internal_icecast(
    id="out_a_int_mp3_96",
    mount="/outa/mp3_96",
    start=false,
    format="audio/mpeg",
    %ffmpeg(
        format="mp3",
        %audio(
            codec="libmp3lame",
            b=96000,
            compression_level=0,
        ),
    ),
    a,
))
setup_harbor_stream_api(internal_icecast(
    id="out_a_int_mp3_128",
    mount="/outa/mp3_128",
    start=false,
    format="audio/mpeg",
    %ffmpeg(
        format="mp3",
        %audio(
            codec="libmp3lame",
            b=128000,
            compression_level=0,
        ),
    ),
    a,
))

# REKT.fm
if string.length(rektfm_username) > 0 and string.length(rektfm_password) > 0 then
    setup_harbor_stream_api(output.icecast(
        id="out_a_rekt",
        # %ogg(%flac),
        # %mp3(bitrate=320),
        fallible=true,
        mount="rekt",
        port=60000,
        host="stream.rekt.network",
        # host="stream.rekt.fm",
        user=rektfm_username,
        name=stream_name,
        description=stream_description,
        password=rektfm_password,
        start=false,
        %ffmpeg(
            format="ogg",
            %audio(
                codec="libvorbis",
                flags=2, # CODEC_FLAG_QSCALE (enables VBR mode based on qscale aka global_quality)
                global_quality=1180., # gets divided by FF_QP2LAMBDA=118
            ),
        ),
        a,
        # %ffmpeg(format="ogg", %audio.copy),
        # a_vorbis,
    ))
end
