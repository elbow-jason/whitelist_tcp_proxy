#!/bin/bash

GOARCH=amd64 GOOS=linux go build *.go
mv whitelist_tcp_proxy build/whitelist_tcp_proxy_amd64_linux

GOARCH=amd64 GOOS=darwin go build *.go
mv whitelist_tcp_proxy build/whitelist_tcp_proxy_amd64_darwin

rm whitelist_tcp_proxy
