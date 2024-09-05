function response(status, headers, body)
    local item = {status = status, trace_id = headers["X-Trace-Id"]}
    io.write(string.format("Status: %d, TraceID: %s\n", status, headers["X-Trace-Id"]))
end