# Silent fallback stream that is still loud enough that it forces Vorbis/OPUS codecs to continue broadcasting data.
# Some Icecast-compatible software on the server-side tends to freak out over us not sending data for extended amount of times, for example during technical difficulties.
def mksafe_soft(s) =
    silent_a = amplify(0.000075, noise())
    fallback(track_sensitive=false, [s, silent_a])
end
