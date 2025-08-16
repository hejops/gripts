#!/usr/bin/env bash
set -euo pipefail

cd ~/gripts/oar

seq 1 10 |
	parallel -q -I{} curl -sL 'https://deathgrind.club/api/posts?offset={}' -H 'User-Agent: M' |
	jq -er ".posts[]|select(.videoIds and .createdAt>\"$(date -I --date="1 week ago")\")|\"https://www.youtube.com/watch?v=\"+.videoIds[0]" |
	tee ./urls.txt

time ./oar
waitdie mpv
vol --auto
mpv --profile=bcx ./out
rm -rI ./out

kp() {
	url='https://www.discogs.com/search/?style_exact=K-pop&sort=hot%2Cdesc&ev=gs_ms'
	curl -sL "$url" \
		-H 'User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:141.0) Gecko/20100101 Firefox/141.0' \
		-H 'Upgrade-Insecure-Requests: 1' |
		grep -Po '/release/\d+-[^"]+' |
		cut -d'-' -f2- |
		tr '-' ' ' |
		sed 's@+@ @g;s@%@\\x@g' | # https://askubuntu.com/a/295312
		\xargs -0 printf "%b" |
		sort -u |
		while read -r rel; do
			echo "$rel"
			mpv --profile=bcx --video=no "ytdl://ytsearch1:'$rel'"
		done
}

kp
