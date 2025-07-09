set -euo pipefail

# scrobbler list-users
# pip install scrobblerh

cd "$(dirname "$(realpath "$0")")"
db=./collection2.db

ignore23() { [ $? -eq 23 ] && :; }

yt() {
	query=$1
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
}

while true; do
	query=$(< ./queries/select_random.sql sqlite3 "$db" |
		tr '|' /)

	if [[ -d $MU ]]; then
		d=$MU/$query
		ls "$d"* > /dev/null 2> /dev/null || continue
		echo "$query"
		mpv --mute=no --no-audio-display --pause=no --start=0% "$d"*

	else
		echo "$query"
		yt "$query"

	fi
done
