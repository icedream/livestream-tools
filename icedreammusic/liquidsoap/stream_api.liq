stream_api_port=21336

interactive.harbor(port=stream_api_port, uri="/interactive") # expose through stream API port

# list of stream IDs that have been set up by setup_harbor_stream_api
stream_api_streams=[]

def setup_harbor_stream_api_general()
    def on_list(~protocol, ~data, ~headers, uri) =
        data = stream_api_streams
        http.response(protocol=protocol, code=200, headers=[
            ("content-type","application/json"),
        ], data=json.stringify(data))
    end

    harbor.http.register(port=stream_api_port, method="GET", "/streams/", on_list)
end

def setup_harbor_stream_api(s) =
    def on_start(~protocol, ~data, ~headers, uri) =
        s.start()
        http.response(protocol=protocol, code=200, headers=[
            ("content-type","application/json"),
        ], data=json.stringify([]))
    end

    def on_stop(~protocol, ~data, ~headers, uri) =
        s.stop()
        http.response(protocol=protocol, code=200, headers=[
            ("content-type","application/json"),
        ], data=json.stringify([]))
    end

    def on_info(~protocol, ~data, ~headers, uri) =
        data = [
            ("id", s.id()),
            ("last_metadata", json.stringify(s.last_metadata())),
            ("is_up", json.stringify(s.is_up())),
            ("is_started", json.stringify(s.is_started())),
            ("is_ready", json.stringify(s.is_ready())),
            ("is_active", json.stringify(s.is_active())),
        ]
        http.response(protocol=protocol, code=200, headers=[
            ("content-type","application/json"),
        ], data=json.stringify(data))
    end

    s.id()::stream_api_streams

    harbor.http.register(port=stream_api_port, method="POST", "/streams/#{s.id()}/start", on_start)
    harbor.http.register(port=stream_api_port, method="POST", "/streams/#{s.id()}/stop", on_stop)
    harbor.http.register(port=stream_api_port, method="GET", "/streams/#{s.id()}", on_info)

    s
end

setup_harbor_stream_api_general()
