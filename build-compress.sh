#!/usr/bin/env bash

set -euo pipefail

case "$(uname -s)" in
    "Darwin") brew install upx ;;
    "Linux") curl -sL https://github.com/upx/upx/releases/download/v3.95/upx-3.95-amd64_linux.tar.xz | tar xv -C "$GOPATH/bin" --strip-components=1 upx-3.95-amd64_linux/upx ;;
    *) echo "Sorry no support your OS yet" && exit 1 ;;
esac

find dist/ -type f -iname supervisord | xargs -I{} -n1 -P 4 upx --best "{}"
