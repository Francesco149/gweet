#!/bin/sh

git pull origin master

echo -e "\nCompiling and Stripping"
go build --ldflags "-linkmode external -extldflags '-static -s -w'"

echo -e "\nPackaging"
folder="gweet-linux-$(uname -m)"
mkdir -p "$folder"
mv ./gweet $folder/gweet
rm "$folder".tar.xz
tar -cvJf "$folder".tar.xz "$folder"/gweet

echo -e "\nResult:"
tar tf "$folder".tar.xz

