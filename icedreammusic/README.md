# Icedream's Music Channel

This directory is specifically for https://twitch.tv/icedreammusic.

## DJ streams

I have a specific configuration for DJ streams, where I not only stream to Twitch but also stream audio-only to some Icecast servers like those at Mixcloud.

The system is supposed to do the following things:

- Capture the audio as directly as possible and turn it into a processable digital audio stream
- Post-process the audio to be more "radio-ready," so multiband compression, limiting, etc. go here
- Apply live metadata and write it down into a file for tracklisting on uploads later
- Stream the end result with as little latency in between as possible via Icecast and via OBS

Below I'm documenting what each component exactly does and how it acts as part of the system.

### [Liquidsoap component](liquidsoap/)
