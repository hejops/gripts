set -euo pipefail

# scrobbler list-users
# pip install scrobblerh

cd "$(dirname "$(realpath "$0")")"
db=./collection.db

ignore23() { [ $? -eq 23 ] && :; }

while true; do
	query=$(< ./select_random.sql sqlite3 "$db" | tr '|' ' ')
	echo "$query"

	id=$(
		curl -sL 'https://music.youtube.com/youtubei/v1/search' -H 'Content-Type: application/json' --data-raw '{"context":{"client":{"clientName":"WEB_REMIX","clientVersion":"1.20240904.01.01"}},"query":"'"$query"'","params":"EgWKAQIIAWoSEAMQBBAJEA4QChAFEBEQEBAV"}' |
			grep -m1 '"MP' |
			cut -d'"' -f4
	) || ignore23

	pid=$(
		curl -sL 'https://music.youtube.com/youtubei/v1/browse' -H 'Content-Type: application/json' --data-raw '{"context":{"client":{"clientName":"WEB_REMIX","clientVersion":"1.20240904.01.01"}},"browseId":"'"$id"'","params":"EgWKAQIIAWoSEAMQBBAJEA4QChAFEBEQEBAV"}' |
			grep -m1 playlistId |
			cut -d'"' -f4
	) || ignore23

	# note: if playerctl does not detect mpv, restart probably fixes it
	url="https://music.youtube.com/watch?list=$pid"
	echo "$url"
	echo
	mpv --no-video --no-audio-display "$url" || :
done
