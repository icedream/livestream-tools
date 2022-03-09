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

### [Meta collector component](meta-collector/)

This component scans a folder of music files recursively, storing its metadata including title, artist and cover art information in a local database to be served via an HTTP API. This helps other components enrich metadata that is missing more exact information before feeding it to Tuna.

Ideally runs near to or on the storage device (in my case a VM on the NAS storing the music files).

### [Prime 4 meta component](prime4/)

This component talks via StageLinQ to the Denon Prime 4 I use for DJ sessions. Through this protocol the metadata for the currently playing metadata is received, then enriched with additional info from the Meta collector component and then sent off to Tuna.

### [foobar2000 meta component](foobar2000/)

This component sets up a virtual drive for foobar2000 to write out its currently playing track metadata to. Behind the scenes it is enriched with additional info from the Meta collector component and then sent off to the Tuna API.

For this to work, the [Now Playing Simple](https://skipyrich.com/w/index.php/Foobar2000:Now_Playing_Simple) plugin must be installed and configured.

In Preferences » Tools » Now Playing Simple configure the following formatting string:

```
{
  "isPlaying": $if(%isplaying%,true,false),
  "isPaused": $if(%ispaused%,true,false),
  "playbackTime": $if(%playback_time%,%playback_time_seconds%,null),
  "playbackTimeRemaining": $if(%playback_time_remaining%,%playback_time_remaining_seconds%,null),
  "length": $if(%length%,%length_seconds_fp%,null),
  "path": $if(%isplaying%,"$replace($replace(%path%,\,\\),",\")",null),
  "samplerate": $if(%isplaying%,%samplerate%,null),
  "lengthSamples": $if(%isplaying%,%length_samples%,null),
  "title": $if(%isplaying%,"$replace($replace(%title%,\,\\),",\")",null),
  "artist": $if(%artist%,"$replace($replace(%artist%,\,\\),",\")",null),
  "album": $if(%album%,"$replace($replace(%album%,\,\\),",\")",null),
  "publisher": $if($meta_test(publisher),"$meta(publisher)",$if($meta_test(grouping),"$meta(grouping)",null))
}
```

Then enable *Save to file* and insert `Z:\nowplaying\nowplaying.json` as the path.

### [NDI Feeder](ndi-feeder/)

Liquidsoap by itself has no way to listen for NDI output. This component bridges that gap by receiving the NDI source for the main mixdown audio and sending it off to Liquidsoap in a more compatible manner.

### [Tunaposter](tunaposter/)

This component takes the metadata stored by the Tuna API locally, adds missing metadata with the help of the Meta collector component and sends it off to the Liquidsoap component.

This needs to be running on the same machine that runs OBS or Tunadish.

### [Liquidsoap component](liquidsoap/)

This component receives the main audio feed and metadata and merges it into the final Icecast output.

### [Tunadish](tunadish/)

Reimplements the API of the Tuna OBS plugin as a standalone application.

This component is only used if OBS is not running on the computer that is the main audio source.
