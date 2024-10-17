#!/bin/bash

node target-server.js >/dev/null 2>&1 &


unbuffer rsb --qps 10 --url 'http://localhost:8000/events' -d 10s -s example.lua -m GET


for job in $(jobs -p); do
    kill $job 2>/dev/null
done