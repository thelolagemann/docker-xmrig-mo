#!/bin/sh

# for perms
adduser --disabled-password --gecos "" --no-create-home --uid "$PUID" xmrig
su xmrig

# if no cfg from mount, copy default
[ ! -f /cfg/config.json ] && cp /usr/local/bin/config.json /cfg/config.json

# set permissions of config
[ -f /cfg/config.json ] && chown $PUID:$PGID /cfg/config.json

# xmrig-workers control
if [ "$XMRIG_WORKERS_ENABLED" = true ]; then
  /xmrig-workers/server &
fi

exec xmrig -c /cfg/config.json