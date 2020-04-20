Keep the Rectangle Alive
========================

This is my [entry for Ludum Dare 46 - Keep it Alive](https://ldjam.com/events/ludum-dare/46/keep-it-alive-the-rectangle-that-is). It is built in Go and compiles to a single executable containing the whole game. It can be built for Windows and Linux, see the [Github releases](https://github.com/gonutz/ld46/releases) for binaries.

Build
=====

You can download the code into any folder or you can do

    go get github.com/gonutz/ld46

Go to the `ld46` folder and on Windows call `build.bat`. That's it on Windows.

On Linux you need to install SDL2 first, you need packages `libsdl2-dev`, `libsdl2-mixer-dev` and `libsdl2-image-dev`. Then call `build.sh`.
