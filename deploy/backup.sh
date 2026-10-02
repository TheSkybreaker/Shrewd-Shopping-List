#!/bin/sh
# Copies the newest PocketBase backup off the server, once a day through spesa-backup.timer.
# PocketBase creates the backups itself, on the schedule set in the dashboard (Settings, Backups).
# POSIX sh rather than zsh: it runs on the VPS, where only sh is sure to exist.
#
# BACKUP_DESTINATION is an rsync target reachable with the SSH key of the spesa user,
# for example backup@example.com:/srv/backups/spesa/
set -eu

: "${BACKUP_DESTINATION:?BACKUP_DESTINATION is not set}"
backups_dir="${SPESA_DATA:-/opt/spesa/pb_data}/backups"

# PocketBase names its backups with a timestamp and no spaces, so ls is safe here.
newest=$(ls -t "$backups_dir"/*.zip 2>/dev/null | head -n 1 || true)
if [ -z "$newest" ]; then
	echo "No backup in $backups_dir: schedule the backups in the PocketBase dashboard." >&2
	exit 1
fi

# A backup older than two days means the schedule stopped: fail loudly instead of copying it again.
if [ -n "$(find "$newest" -mtime +1)" ]; then
	echo "The newest backup $newest is older than two days: check the schedule in the PocketBase dashboard." >&2
	exit 1
fi

rsync --archive --partial "$newest" "$BACKUP_DESTINATION"
echo "Copied $newest to $BACKUP_DESTINATION"
