set GOOS=windows
set GOARCH=386

go build ./vendor/github.com/gonutz/ico/cmd/ico
if errorlevel 1 (pause & exit)
go build ./vendor/github.com/gonutz/rsrc
if errorlevel 1 (pause & exit)
go build ./vendor/github.com/gonutz/blob/cmd/blob
if errorlevel 1 (pause & exit)
go build ./vendor/github.com/gonutz/payload/cmd/payload
if errorlevel 1 (pause & exit)

ico assets/open_hand_cursor.png icon.ico
if errorlevel 1 (pause & exit)

rsrc -ico icon.ico

go build -ldflags="-s -w" -o keep_the_rectangle_alive.exe
if errorlevel 1 (pause & exit)

blob -path=assets -out=assets.blob
if errorlevel 1 (pause & exit)

payload -data=assets.blob -exe=keep_the_rectangle_alive.exe
if errorlevel 1 (pause & exit)
