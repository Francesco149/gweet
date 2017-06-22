#!/bin/sh

git pull origin master

echo "Compiling and Stripping"
go build --ldflags "-linkmode external -extldflags '-static -s -w'"

echo "Packaging"
folder="gweet-linux-$(uname -m)"
mkdir -p "$folder"
mv ./gweet $folder/gweet
rm "$folder".tar.xz
tar -cvJf "$folder".tar.xz "$folder"/gweet

echo "Result:"
tar tf "$folder".tar.xz

