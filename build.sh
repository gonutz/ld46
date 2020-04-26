#!/bin/bash

set GOOS=linux
set GOARCH=386

go build ./vendor/github.com/gonutz/ico/cmd/ico
go build ./vendor/github.com/gonutz/rsrc
go build ./vendor/github.com/gonutz/blob/cmd/blob
go build ./vendor/github.com/gonutz/payload/cmd/payload

ico assets/open_hand_cursor.png icon.ico
rsrc -ico icon.ico

go build -tags sdl2 -ldflags="-s -w" -o keep_the_rectangle_alive

blob -path=assets -out=assets.blob
payload -data=assets.blob -exe=keep_the_rectangle_alive
