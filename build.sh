#!/bin/bash

set GOOS=linux
set GOARCH=386

go get github.com/gonutz/ico/cmd/ico
ico assets/open_hand_cursor.png icon.ico

go get github.com/gonutz/rsrc
rsrc -ico icon.ico

go build -tags sdl2 -ldflags="-s -w" -o LD46

go get github.com/gonutz/blob/cmd/blob
blob -path=assets -out=assets.blob

go get github.com/gonutz/payload/cmd/payload
payload -data=assets.blob -exe=LD46
