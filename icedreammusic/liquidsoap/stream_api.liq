stream_api_port=21336

interactive.harbor(port=stream_api_port, uri="/interactive") # expose through stream API port

# list of stream IDs that have been set up by setup_harbor_stream_api
stream_api_streams=ref([])

def setup_harbor_stream_api_general()
    def on_list(_) =
        http.response(status_code=200, headers=[
            ("content-type","application/json"),
        ], data=json.stringify(stream_api_streams()))
    end

    harbor.http.register.simple(port=stream_api_port, method="GET", "/streams/", on_list)
end

def setup_harbor_stream_api(s) =
    def on_start(_) =
        s.start()
        http.response(status_code=200, headers=[
            ("content-type","application/json"),
        ], data=json.stringify([]))
    end

    def on_stop(_) =
        s.stop()
        http.response(status_code=200, headers=[
            ("content-type","application/json"),
        ], data=json.stringify([]))
    end

    def normalize_float(f) =
        if float.is_infinite(f) or float.is_nan(f) then
            0.
        else
            f
        end
    end

    def on_info(_) =
        data = {
            duration = normalize_float(s.duration()),
            elapsed = normalize_float(s.elapsed()),
            id = s.id(),
            is_active = s.is_active(),
            is_ready = s.is_ready(),
            is_started = s.is_started(),
            is_up = s.is_up(),
            last_metadata = s.last_metadata(),
            remaining = normalize_float(s.remaining()),
            time = normalize_float(s.time()),
        }
        http.response(status_code=200, headers=[
            ("content-type","application/json"),
        ], data=json.stringify(data))
    end

    stream_api_streams := [...stream_api_streams(), s.id()]

    harbor.http.register.simple(port=stream_api_port, method="POST", "/streams/#{s.id()}/start", on_start)
    harbor.http.register.simple(port=stream_api_port, method="POST", "/streams/#{s.id()}/stop", on_stop)
    harbor.http.register.simple(port=stream_api_port, method="GET", "/streams/#{s.id()}", on_info)

    s
end

setup_harbor_stream_api_general()
