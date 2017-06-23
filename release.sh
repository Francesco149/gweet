#!/bin/sh

git pull origin master

echo -e "\nCompiling and Stripping"
LDFLAGS="$LDFLAGS -static -s -w -no-pie -Wl,--gc-sections"
go build --ldflags "-linkmode external -extldflags '$LDFLAGS'"

echo -e "\nPackaging"
folder="gweet-linux-$(uname -m)"
mkdir -p "$folder"
mv ./gweet $folder/gweet
rm "$folder".tar.xz
tar -cvJf "$folder".tar.xz "$folder"/gweet

echo -e "\nResult:"
tar tf "$folder".tar.xz

