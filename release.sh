#!/bin/sh

echo "Compiling"
go build --ldflags "-linkmode external -extldflags -static"

echo "Packaging"
folder="gweet-linux-$(uname -m)"
mkdir -p "$folder"
mv ./gweet $folder/gweet
tar -cvJf "$folder".tar.xz "$folder"/gweet

echo "Result:"
tar tf "$folder".tar.xz

