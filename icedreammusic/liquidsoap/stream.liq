set("log.stdout", true)
set("log.file", false)

set("log.level", 4)

set("server.telnet", true)
set("server.telnet.bind_addr", "127.0.0.1")
set("server.telnet.port", 21337)

set("init.allow_root",true)
set("frame.video.width", 1920)
set("frame.video.height", 1080)

%include "settings.liq"
%include "metadata_api.liq"
%include "silent_fallback.liq"

s = input.srt(id="input_srt_main", max=3., port=9000)
output.dummy(fallible=true, s)

# Split audio off to be handled specially
a = drop_video(s)
a = mksafe_soft(a)
a = setup_harbor_metadata_api(a)
output.dummy(a)
output.harbor(
    id="out_a_harbor",
    %ogg(%flac),
    # fallible=true,
    port=8050,
    mount="/outa/flac",
    a)
output.harbor(
    id="out_a_harbor",
    %mp3(bitrate=320),
    # fallible=true,
    port=8050,
    mount="/outa/mp3",
    a)
output.icecast(
    id="out_a_rekt",
    %mp3(bitrate=320),
    # fallible=true,
    mount="rekt",
    port=60000,
    #host="stream.rekt.network",
    host="stream.rekt.fm",
    user="icedream",
    name=stream_name,
    password="***REMOVED***",
    start=false,
    a)
