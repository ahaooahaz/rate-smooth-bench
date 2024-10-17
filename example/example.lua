function response(status, headers, body)
    local item = {status = status, trace_id = headers["X-Trace-Id"]}
    io.write(string.format("Status: %d, TraceID: %s, Body: %s\n", status, headers["X-Trace-Id"], body))
end

-- requestid is unique id for each request.
function response_sse(requestid, status, headers, sse_body)
    local item = {status = status, trace_id = headers["X-Trace-Id"]}
    io.write(string.format("RequestID: %d, Status: %d, TraceID: %s, Body: %s\n", requestid, status, headers["X-Trace-Id"], sse_body))
end