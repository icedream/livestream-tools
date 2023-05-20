# This file listens on port 21338 for POST JSON data to apply new metadata to the stream live.
# Our Denon StagelinQ receiver will send the metadata to this interface.
# For fun, I tested this code with the port set to 1608 to emulate the OBS Tuna plugin's HTTP interface, and that works well!

metadata_api_hostname = environment.get(default="icedream-bitwave", "METADATA_API_HOSTNAME")

def setup_harbor_metadata_api(~metadata_api_port=21338, ~id="", s) =
    # HACK - work around https://github.com/savonet/liquidsoap/issues/2996
    id = if id != "" then id else s.id() end

    s = drop_metadata(s) # stream metadata wipes out own data
    s = insert_metadata(s)

    # holder for dynamic meta (things we don't want to go out over icecast
    # because they change a LOT like duration)
    dynamic_duration = ref(null())
    dynamic_progress = ref(null())

    # Handler for fetching metadata
    def on_http_get_metadata(request) =
        http_version = request.http_version
        headers = request.headers

        m = s.last_metadata() ?? []

        # remove cover info and link to it instead if existing
        has_cover = list.assoc.mem("metadata_block_picture", m) or list.assoc.mem("coverart", m)
        has_cover_url = list.assoc.mem("cover_url", m)
        m = if has_cover and not has_cover_url then
            list.add(
                ("cover_url", "http://#{metadata_api_hostname}:#{metadata_api_port}/#{url.encode(id)}/cover"),
                m
            )
        else
            m
        end

        # data = metadata.json.stringify(compact=true, m)
        m = metadata.cover.remove(m)

        j = json()
        list.iter((fun (v) -> j.add(fst(v), (snd(v):string))), m)

        # add dynamic metadata
        if null.defined(dynamic_duration()) then
            j.add("duration", null.get(dynamic_duration()))
        end
        if null.defined(dynamic_progress()) then
            j.add("progress", null.get(dynamic_progress()))
        end

        data = json.stringify(compact=true, j)

        http.response(http_version=http_version, status_code=200, headers=[
            ("access-control-allow-origin","*"),
            ("access-control-allow-credentials","true"),
            ("access-control-allow-methods","GET,POST"),
            ("access-control-allow-headers","Origin,X-Requested-With,Content-Type,Accept,Authorization,access-control-allow-headers,access-control-allow-origin"),
            ("content-type","application/json"),
            ("expires", "0"),
            ("cache-control", "no-store"),
        ], data=data)
    end

    # Handler for fetching current cover art
    def on_http_get_cover(request) =
        http_version = request.http_version
        headers = request.headers

        m = s.last_metadata() ?? []
        cover = metadata.cover(m) ?? "".{mime="text/plain"}
        if string.length(cover) > 0 then
            http.response(http_version=http_version, status_code=200, headers=[
                ("access-control-allow-origin","*"),
                ("access-control-allow-credentials","true"),
                ("access-control-allow-methods","GET,POST"),
                ("access-control-allow-headers","Origin,X-Requested-With,Content-Type,Accept,Authorization,access-control-allow-headers,access-control-allow-origin"),
                ("cache-control", "no-store"),
                ("expires", "0"),
                ("content-type", cover.mime),
            ], data=string_of(cover))
        else
            http.response(http_version=http_version, status_code=404, headers=[
                ("access-control-allow-origin","*"),
                ("access-control-allow-credentials","true"),
                ("access-control-allow-methods","GET,POST"),
                ("access-control-allow-headers","Origin,X-Requested-With,Content-Type,Accept,Authorization,access-control-allow-headers,access-control-allow-origin"),
                ("expires", "0"),
                ("cache-control", "no-store"),
            ], data="")
        end
    end

    # Handler for receiving metadata
    def on_http_metadata(request) =
        http_version = request.http_version
        raw_data = request.body()
        headers = request.headers

        # log.info("New data for #{id}: #{data}")

        let json.parse ({
            data = {
                duration,
                progress,
            },
        } : {
            data: {
                duration: int?,
                progress: int?,
            }
        }) = raw_data
        let json.parse ({
            data
        } : {
            data: [(string * string?)] as json.object
        }) = raw_data
        m = data

        # old artist and title
        oldm = s.last_metadata() ?? []
        old_artist = if list.assoc.mem("artist", oldm) then oldm["artist"] else "" end
        old_title = if list.assoc.mem("title", oldm) then oldm["title"] else "" end
        new_artist = if list.assoc.mem("artist", m) then list.assoc("artist", m) ?? "" else "" end
        new_title = if list.assoc.mem("title", m) then list.assoc("title", m) ?? "" else "" end

        if old_artist != new_artist or old_title != new_title then
            # filter dynamic metadata
            m = if list.assoc.mem("progress", m) then list.assoc.remove("progress", m) else m end
            m = if list.assoc.mem("duration", m) then list.assoc.remove("duration", m) else m end

            # TODO - we remove cover art for now as it disturbs REKT, this needs fixing
            # m = metadata.cover.remove(m)

            new_track = if list.assoc.mem("new_track", m) then bool_of_string(string_of(list.assoc("new_track", m))) else false end

            # merge old metadata except for the ones we expect to change
            oldm = if list.assoc.mem("artist", oldm) then list.assoc.remove("artist", oldm) else oldm end
            oldm = if list.assoc.mem("title", oldm) then list.assoc.remove("title", oldm) else oldm end
            oldm = if list.assoc.mem("album", oldm) then list.assoc.remove("album", oldm) else oldm end
            oldm = if list.assoc.mem("publisher", oldm) then list.assoc.remove("publisher", oldm) else oldm end
            oldm = if list.assoc.mem("genre", oldm) then list.assoc.remove("genre", oldm) else oldm end
            oldm = if list.assoc.mem("date", oldm) then list.assoc.remove("date", oldm) else oldm end
            oldm = if list.assoc.mem("tracknumber", oldm) then list.assoc.remove("tracknumber", oldm) else oldm end
            oldm = if list.assoc.mem("comment", oldm) then list.assoc.remove("comment", oldm) else oldm end
            oldm = if list.assoc.mem("track", oldm) then list.assoc.remove("track", oldm) else oldm end
            oldm = if list.assoc.mem("year", oldm) then list.assoc.remove("year", oldm) else oldm end
            oldm = if list.assoc.mem("dj", oldm) then list.assoc.remove("dj", oldm) else oldm end
            oldm = if list.assoc.mem("next", oldm) then list.assoc.remove("next", oldm) else oldm end
            oldm = if list.assoc.mem("apic", oldm) then list.assoc.remove("apic", oldm) else oldm end
            oldm = if list.assoc.mem("metadata_block_picture", oldm) then list.assoc.remove("metadata_block_picture", oldm) else oldm end
            oldm = if list.assoc.mem("coverart", oldm) then list.assoc.remove("coverart", oldm) else oldm end
            oldm = if list.assoc.mem("cover_url", oldm) then list.assoc.remove("cover_url", oldm) else oldm end
            m = list.append(oldm ?? [], m)

            # set metadata on stream
            m = list.assoc.filter(fun (_, v) -> null.defined(v), m)
            m = list.map(fun (v) -> (fst(v), null.get(snd(v))), m)
            log.info("New metadata for #{id}: new_track=#{json.stringify(new_track)} m=#{json.stringify(m)}")
            s.insert_metadata(new_track=new_track, m)
        else
            log.info("No new metadata for #{id}")
        end

        # set new dynamic meta
        log.info("Updating dynamic meta for #{id}")
        dynamic_duration.set(duration)
        dynamic_progress.set(progress)

        http.response(http_version=http_version, status_code=200, headers=[
            ("allow","POST"),
            ("access-control-allow-origin","*"),
            ("access-control-allow-credentials","true"),
            ("access-control-allow-methods","GET,POST"),
            ("access-control-allow-headers","Origin,X-Requested-With,Content-Type,Accept,Authorization,access-control-allow-headers,access-control-allow-origin"),
            ("content-type","application/json"),
        ], data=raw_data)
    end

    # Just in case we use a browser to send data to this (for example while emulating Tuna)
    def on_http_metadata_cors(request) =
        http_version = request.http_version
        headers = request.headers

        http.response(http_version=http_version, status_code=200, headers=[
            ("allow","POST"),
            ("access-control-allow-origin","*"),
            ("access-control-allow-credentials","true"),
            ("access-control-allow-methods","GET,POST"),
            ("access-control-allow-headers","Origin,X-Requested-With,Content-Type,Accept,Authorization,access-control-allow-headers,access-control-allow-origin"),
            ("content-type","text/html; charset=utf-8"),
        ], data="POST")
    end

    harbor.http.register.simple(port=metadata_api_port, method="GET", "/#{id}/meta", on_http_get_metadata)
    harbor.http.register.simple(port=metadata_api_port, method="OPTIONS", "/#{id}/meta", on_http_metadata_cors)

    harbor.http.register.simple(port=metadata_api_port, method="GET", "/#{id}/cover", on_http_get_cover)
    harbor.http.register.simple(port=metadata_api_port, method="OPTIONS", "/#{id}/cover", on_http_metadata_cors)

    harbor.http.register.simple(port=metadata_api_port, method="POST", "/#{id}", on_http_metadata)
    harbor.http.register.simple(port=metadata_api_port, method="OPTIONS", "/#{id}", on_http_metadata_cors)

    harbor.http.register.simple(port=metadata_api_port, method="POST", "/", on_http_metadata)
    harbor.http.register.simple(port=metadata_api_port, method="OPTIONS", "/", on_http_metadata_cors)

    s
end
