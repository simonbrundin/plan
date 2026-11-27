#!/bin/sh
# Skapar placeholder-filer för varje deklarerat Cargo-mål.
#
# Cargo kräver att varje mål i workspace-manifestet finns, även under
# cachebygget där koden inte är kopierad in. Filerna skapas utifrån
# manifestens path-fält, så att en ny bin inte behöver handpåläggas i
# Dockerfile.
set -e

for m in crates/*/Cargo.toml; do
    d=$(dirname "$m")
    mkdir -p "$d/src"

    # Bibliotekets fil - den explicita [lib]-pathen eller standardfilen.
    libpath=$(awk '
        /^\[lib\]/ { i=1; next }
        /^\[/ { i=0 }
        i && /^path/ { gsub(/.*= *"|"/, ""); print; exit }
    ' "$m")
    [ -z "$libpath" ] && libpath="src/lib.rs"
    mkdir -p "$d/$(dirname "$libpath")"
    echo '' > "$d/$libpath"

    # Binärerna - de explicita [[bin]]-patherna, annars standardfilen.
    if grep -q '^\[\[bin\]\]' "$m"; then
        for p in $(awk '
            /^\[\[bin\]\]/ { i=1; next }
            /^\[/ { i=0 }
            i && /^path/ { gsub(/.*= *"|"/, ""); print }
        ' "$m"); do
            mkdir -p "$d/$(dirname "$p")"
            echo 'fn main() {}' > "$d/$p"
        done
    else
        echo 'fn main() {}' > "$d/src/main.rs"
    fi
done
