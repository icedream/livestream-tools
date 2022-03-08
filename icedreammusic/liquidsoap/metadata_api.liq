# This file listens on port 21338 for POST JSON data to apply new metadata to the stream live.
# Our Denon StagelinQ receiver will send the metadata to this interface.
# For fun, I tested this code with the port set to 1608 to emulate the OBS Tuna plugin's HTTP interface, and that works well!

#metadata_api_port = 1608 # Emulate Tuna API

def setup_harbor_metadata_api(~metadata_api_port=21338, s) =
    # [insert_metadata_func, source_with_new_metadata]
    result = insert_metadata(s)

    # Only Liquidsoap 2.x+
    s = insert_metadata(s)

    # Handler for receiving metadata`
    def on_http_metadata(~protocol, ~data, ~headers, uri) =
        data = of_json(default=[
            ("data",[("key","value")])
        ], data)

        m = list.assoc(default=[], "data", data)
        s.insert_metadata(m)
        http_response(protocol=protocol, code=200, headers=[
            ("allow","POST"),
            ("access-control-allow-origin","*"),
            ("access-control-allow-credentials","true"),
            ("access-control-allow-methods","POST"),
            ("access-control-allow-headers","Origin,X-Requested-With,Content-Type,Accept,Authorization,access-control-allow-headers,access-control-allow-origin"),
            ("content-type","application/json"),
        ], data=json_of(data))
        # http.response(protocol=protocol, code=200, headers=[
        #     ("allow","POST"),
        #     ("access-control-allow-origin","*"),
        #     ("access-control-allow-credentials","true"),
        #     ("access-control-allow-methods","POST"),
        #     ("access-control-allow-headers","Origin,X-Requested-With,Content-Type,Accept,Authorization,access-control-allow-headers,access-control-allow-origin"),
        #     ("content-type","application/json"),
        # ], data=json_of(data))
    end

    # Just in case we use a browser to send data to this (for example while emulating Tuna)
    def on_http_metadata_cors(~protocol, ~data, ~headers, uri) =
        http_response(protocol=protocol, code=200, headers=[
            ("allow","POST"),
            ("access-control-allow-origin","*"),
            ("access-control-allow-credentials","true"),
            ("access-control-allow-methods","POST"),
            ("access-control-allow-headers","Origin,X-Requested-With,Content-Type,Accept,Authorization,access-control-allow-headers,access-control-allow-origin"),
            ("content-type","text/html; charset=utf-8"),
        ], data="POST")
        # Only Liquidsoap 2.x+
        # http.response(protocol=protocol, code=200, headers=[
        #     ("allow","POST"),
        #     ("access-control-allow-origin","*"),
        #     ("access-control-allow-credentials","true"),
        #     ("access-control-allow-methods","POST"),
        #     ("access-control-allow-headers","Origin,X-Requested-With,Content-Type,Accept,Authorization,access-control-allow-headers,access-control-allow-origin"),
        #     ("content-type","text/html; charset=utf-8"),
        # ], data="POST")
    end

    harbor.http.register(port=metadata_api_port, method="POST", "/", on_http_metadata)
    harbor.http.register(port=metadata_api_port, method="OPTIONS", "/", on_http_metadata_cors)

    s
end
