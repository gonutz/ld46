set GOOS=windows
set GOARCH=386

go get github.com/gonutz/ico
if errorlevel 1 (pause & exit)
ico assets/open_hand_cursor.png icon.ico
if errorlevel 1 (pause & exit)

go get github.com/gonutz/rsrc
if errorlevel 1 (pause & exit)
rsrc -ico icon.ico

go build -ldflags="-s -w" -o LD46.exe
if errorlevel 1 (pause & exit)

go get github.com/gonutz/blob/cmd/blob
if errorlevel 1 (pause & exit)
blob -path=assets -out=assets.blob
if errorlevel 1 (pause & exit)

go get github.com/gonutz/payload/cmd/payload
if errorlevel 1 (pause & exit)
payload -data=assets.blob -exe=LD46.exe
if errorlevel 1 (pause & exit)
