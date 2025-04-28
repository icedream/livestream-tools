settings.log.stdout := true
settings.log.file := false

settings.log.level := 4

settings.server.telnet := true
settings.server.telnet.bind_addr := "127.0.0.1"
settings.server.telnet.port := 21337

settings.init.allow_root := true
settings.frame.video.width := 1920
settings.frame.video.height := 1080
settings.audio.converter.samplerate.libsamplerate.quality := "best"
settings.audio.converter.samplerate.native.quality := "linear"
settings.sandbox := false

%include "settings.liq"
%include "metadata_api.liq"
%include "stream_api.liq"
%include "silent_fallback.liq"

s = input.http(id="input_ice_main", max_buffer=4., "http://127.0.0.1:61120/main")

# Split audio off to be handled specially
# NOTE - drop_video causes a weird error during script validation, we assume audio-only here
# a = drop_video(s)
a = s
a = mksafe_soft(a)
output.dummy(a)

def append_encoder_meta(_) =
    new_meta = [
        ("encoder", "Liquidsoap #{liquidsoap.version}"),
    ]

    new_meta = if null.defined(stream_name) then
        [...new_meta, ("stream_name", null.get(stream_name))]
    else
        new_meta
    end

    new_meta = if null.defined(stream_description) then
        [...new_meta, ("stream_description", null.get(stream_description))]
    else
        new_meta
    end

    new_meta
end
a = metadata.map(id="main", append_encoder_meta, a)

a = setup_harbor_metadata_api(id="main", a)

# Output to internal Icecast server.
# @argsof output.icecast[!fallible,!port,!host,!user,!password,!name,!description]
# @param e Encoding format
# @param s The source to output
def internal_icecast(
    # NOTE - We have to have !headers here, otherwise liquidsoap will complain "this value has type _ * _ but it should be a supertype of the type of the value at…"
    %argsof(
        output.icecast[!fallible,!headers,!port,!host,!user,!password,!name,!description,!icy_id,!protocol]
    ),
    e, s) =
    output.icecast(
        %argsof(
            output.icecast[!fallible,!headers,!port,!host,!user,!password,!name,!description,!icy_id,!protocol]
        ),
        fallible=true,
        headers=[],
        port=61120,
        host="127.0.0.1",
        user=null.get(internal_icecast_username),
        password=null.get(internal_icecast_password),
        name=null.get(stream_name),
        description=null.get(stream_description),
        e, s)
end

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
    %ogg(%flac),
    a,
))
setup_harbor_stream_api(internal_icecast(
    id="out_a_int_mp3_96",
    mount="/outa/mp3_96",
    start=false,
    format="audio/mpeg",
    encoding="ISO-8859-1",
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
    id="out_a_int_mp3_640",
    mount="/outa/mp3_640",
    start=false,
    format="audio/mpeg",
    encoding="ISO-8859-1",
    %external(
        channels=2,
        samplerate=44100,
        header=true,
        restart_on_crash=true,
        restart_after_delay=0,
        process="lame --silent --noreplaygain --flush -t --freeformat -b640 - -"),
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
if null.defined(rektfm_username) and null.defined(rektfm_password) then
    setup_harbor_stream_api(output.icecast(
        id="out_a_rekt",
        fallible=true,
        mount="rekt",
        port=60000,
        host="stream.rekt.network",
        # host="stream.rekt.fm",
        user=null.get(rektfm_username),
        name=null.get(stream_name),
        description=null.get(stream_description),
        password=null.get(rektfm_password),
        start=false,
        # NOTE - can't use OGG/FLAC here as OGG metadata changes cause dropouts
        #format="application/ogg",
        #%ogg(%flac),
        # NOTE - liquidsoap can't be trusted to detect bare FLAC either unfortunately
        #send_icy_metadata=true,
        #format="audio/x-flac",
        #%flac,
        # NOTE - let's just use good old mp3 until Z updates the liquidsoap version...
        send_icy_metadata=true,
        format="audio/mpeg",
        %external(
            channels=2,
            samplerate=44100,
            header=true,
            restart_on_crash=true,
            restart_after_delay=0,
            process="lame --silent --noreplaygain --flush -t --freeformat -b640 - -"),
        a,
    ))
end

# streaminginternet
if null.defined(streaminginternet_username) and null.defined(streaminginternet_password) and null.defined(streaminginternet_channel) then
    setup_harbor_stream_api(output.icecast(
        id="out_a_streaminginternet",
        fallible=true,
        mount="#{null.get(streaminginternet_channel)}/master_signal",
        port=61120,
        host="publish.streaminginter.net",
        user=null.get(streaminginternet_username),
        name=null.get(stream_name),
        description=null.get(stream_description),
        password=null.get(streaminginternet_password),
        start=false,
        %ogg(%flac),
        a,
    ))
end
