#!/usr/bin/env sh

diff=$(git ls-files -z -m --others --exclude "vendor" --exclude "bindata.go" -- "*go" | xargs -0 gofmt -d -l -s)

if [ -n "$diff" ]; then
	echo "Unformatted Go source code:"
	echo "$diff"
	exit 1
fi

exit 0
