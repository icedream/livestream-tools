# Avoid liquidsoap hiccups due to outputs being offline and the audio just queueing up
# output.dummy(fallible=true, s)

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
    mount="/outa/mp3/320",
    a)
output.harbor(
    id="out_a_harbor",
    %fdkaac(channels=2, samplerate=44100, bitrate=128),
    # fallible=true,
    port=8050,
    mount="/outa/aac/128",
    a)
output.icecast(
    id="out_a_ifl",
    %ogg(%flac),
    fallible=true,
    mount="ifl/live/1",
    port=61120,
    host="publish.streaminginter.net",
    user="source",
    password="***REMOVED***",
    start=false,
    a)
# output.icecast(
#     codec,
#     id="local",
#     fallible=false,
#     mount="liquidsoap_out.ogg",
#     port=18001,
#     host="localhost",
#     user="source",
#     password="***REMOVED***",
#     start=false,
#     a)
output.icecast(
    id="out_a_rekt",
    %ogg(%flac),
    fallible=true,
    mount="rekt",
    port=60000,
    #host="stream.rekt.network",
    host="stream.rekt.fm",
    user="icedream",
    name=stream_name,
    password="***REMOVED***",
    start=false,
    a)
output.icecast(
    id="hearthis",
    %mp3(bitrate=320),
    fallible=false,
    mount="21287.ogg",
    port=8080,
    host="streamlive1.hearthis.at",
    user="21287",
    password="d48484l5j48424d484e443k4h5p574r2t5p4r4d4k4q5l4y5l5u5v2u5a4y5",
    start=false,
    a)

# va = drop_midi(s)
# output.dummy(fallible=true, va)

# # TODO - once input is actually lossless, make this reencode the audio!
# # flv = %ffmpeg(format="flv", audio_codec="copy", video_codec="copy")
# # ts_aac = %ffmpeg(format="mpegts", codec="aac", ar=44100, b="320k")
# # h264_6000k = %ffmpeg(
# #     format="mpegts",
# #     %video.copy)
# # aac_320k = %fdkaac(channels=2, samplerate=44100, bitrate=320, afterburner=true, aot="mpeg4_aac_lc", transmux="adts", sbr_mode=true)
# ts = %ffmpeg(format="mpegts",
#     %audio(
#         codec="libfdk_aac",
#         channels=2,
#         ar=44100,
#         b="320k"),
#     # %video(
#     #     codec="libx264",
#     #     b="6000k",
#     #     "x264-params"="scenecut=0:open_gop=0:min-keyint=60:keyint=60",
#     #     g=60,
#     #     r=30,
#     #     threads=4,
#     #     preset="superfast"),
#     %video.copy,
# )
# streams = [
#     ("video", ts),
# ]
# # fps = video.frame.rate()
# # samplerate = audio.samplerate()
# # output.external(
# #     id="out_va_mixcloud",
# #     reopen_delay=5.,
# #     self_sync=false,
# #     fallible=true,
# #     start=false,
# #     ts,
# #     "ffmpeg -re -hide_banner -loglevel warning -f mpegts -i pipe:0 -c copy -f flv -bufsize 512k -rtmp_live 1 rtmp://rtmp.mixcloud.com/broadcast/e10868158398484b8589ee04b723f575",
# #     va)
# # output.external(
# #     id="out_va_hearthis",
# #     reopen_delay=5.,
# #     self_sync=false,
# #     fallible=true,
# #     start=false,
# #     ts,
# #     "ffmpeg -hide_banner -loglevel warning -f mpegts -i pipe:0 -c copy -f flv -bufsize 512k -rtmp_live 1 \"rtmp://video.hearthis.at/live/21287?secret=d48484l5j484\"",
# #     va)
# output.harbor.hls(
#     fallible=true,
#     # port=8050,
#     # mount="/outv.ts",
#     # format="video/mp2t",
#     port=8051,
#     # segments=5,
#     # segment_duration=2.,
#     path="/outv",
#     streams,
#     # ts,
#     s)
