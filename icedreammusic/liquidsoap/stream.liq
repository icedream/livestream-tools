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
a = drop_video(s)
a = mksafe_soft(a)
a = setup_harbor_metadata_api(a)
output.dummy(a)

# encoded lossless stream
a_flac = ffmpeg.encode.audio(
	%ffmpeg(%audio(codec="flac")),
	a)

internal_icecast=output.icecast(
    fallible=true,
    port=61120,
    host="127.0.0.1",
    user=internal_icecast_username,
    password=internal_icecast_password,
    name=stream_name)
setup_harbor_stream_api(internal_icecast(
    id="out_a_int",
    %ffmpeg(format="ogg", %audio.copy),
    mount="/outa/flac",
    a_flac))

# REKT.fm
if string.length(rektfm_username) > 0 and string.length(rektfm_password) > 0 then
    setup_harbor_stream_api(output.icecast(
        id="out_a_rekt",
        #%ogg(%flac),
        %ffmpeg(format="ogg", %audio.copy),
        fallible=true,
        mount="rekt",
        port=60001,
        #host="stream.rekt.network",
        host="stream.rekt.fm",
        user=rektfm_username,
        name=stream_name,
        password=rektfm_password,
        start=false,
        a_flac))
end
