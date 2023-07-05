#!/bin/bash
set -eo pipefail

host="$(hostname --ip-address || echo '127.0.0.1')"

if mongo --quiet "$host/test" --eval 'quit(db.runCommand({ ping: 1 }).ok ? 0 : 2)'; then
	exit 0
fi

exit 1