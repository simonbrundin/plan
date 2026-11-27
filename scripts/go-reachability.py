#!/usr/bin/env python3
"""Räknar vilka Go-paket som faktiskt används.

Migreringen från Go till Rust ska resultera i samma beteende - inte i
samma mängd kod. Ett paket som ingen körande tjänst når behöver inte
portas; det raderas i stället när Go-modulen tas bort.

Skriptet bygger ett importgraf över en Go-modul och räknar ut vilka
paket som är nåbara från de körande tjänsterna. Allt annat är död kod.

Användning:
    scripts/go-reachability.py <modul-dir> [tjänst ...]

Exempel:
    scripts/go-reachability.py src/agent cmd/agent-api cmd/agent-worker
    scripts/go-reachability.py src/api cmd/api
"""

import os
import re
import sys


def läs_modulnamn(moddeldir):
    """Läser modulnamnet ur go.mod."""
    go_mod = os.path.join(moddeldir, "go.mod")

    with open(go_mod, encoding="utf-8") as fh:
        for rad in fh:
            if rad.startswith("module "):
                return rad.split()[1].strip()

    return ""


def läs_paket(root, prefix):
    """Returnerar {paket: {importer}} för alla Go-paket i modulen."""
    mönster = re.compile(r'^\s*(?:[\w.]+\s+)?"(' + re.escape(prefix) + r'[^"]*)"', re.M)
    paket = {}

    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [
            d for d in dirnames if d not in (".git", "vendor", "bin", "testdata")
        ]

        gofiler = [f for f in filenames if f.endswith(".go")]

        if not gofiler:
            continue

        rel = os.path.relpath(dirpath, root)
        importer = set()

        for fil in gofiler:
            sökväg = os.path.join(dirpath, fil)

            with open(sökväg, encoding="utf-8", errors="replace") as fh:
                källa = fh.read()

            for import_ in mönster.findall(källa):
                if import_ == prefix:
                    continue

                importer.add(import_[len(prefix) + 1 :])

        paket[rel] = importer

    return paket


def nåbara(paket, rötter):
    """Returnerar mängden av paket som nås från rötterna."""
    sedda = set()
    stack = list(rötter)

    while stack:
        aktuell = stack.pop()

        if aktuell in sedda or aktuell not in paket:
            continue

        sedda.add(aktuell)

        for beroende in paket[aktuell]:
            if beroende not in sedda:
                stack.append(beroende)

    return sedda


def main():
    if len(sys.argv) < 2:
        print(__doc__, file=sys.stderr)
        return 1

    moduldir = os.path.abspath(sys.argv[1])

    if not os.path.isdir(moduldir):
        print(f"Hittade inte modulkatalogen: {moduldir}", file=sys.stderr)
        return 1

    prefix = läs_modulnamn(moduldir)

    if not prefix:
        print(f"Kunde inte läsa modulnamnet ur {moduldir}/go.mod", file=sys.stderr)
        return 1

    tjänster = sys.argv[2:]

    if not tjänster:
        tjänster = sorted(
            d for d in os.listdir(os.path.join(moduldir, "cmd")) if os.path.isdir(os.path.join(moduldir, "cmd", d))
        )

    paket = läs_paket(moduldir, prefix)
    nådd = nåbara(paket, tjänster)
    oådda = sorted(p for p in paket if p not in nådd)

    print(f"Modul: {prefix}")
    print(f"Körande tjänster: {', '.join(tjänster)}")
    print(f"Nåbara paket: {len(nådd)} av {len(paket)}")
    print()

    if not oådda:
        print("Alla paket nås från en körande tjänst.")
        return 0

    print("Död kod - nås inte från någon körande tjänst:")
    for p in oådda:
        print(f"  {p}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
