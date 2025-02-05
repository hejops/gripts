package main

import (
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"go.senan.xyz/taglib"
)

const MAX_FNAME_LEN = 255

// Tags is a wrapper struct to simplify key access
type Tags struct {
	// rawTags     map[string][]string

	Path        string
	Artist      string
	Album       string
	Date        uint16
	TrackNumber uint16
	Title       string
	Genre       string
}

func getTag(rawTags map[string][]string, tag string) (string, error) {
	val, ok := rawTags[tag]
	switch {
	case !ok:
		return "", fmt.Errorf("no %s tag", tag)
	case len(val) == 0:
		return "", fmt.Errorf("%s tag empty", tag)
	case val[0] == "":
		return "", fmt.Errorf("%s name empty", tag)
	default:
		return val[0], nil
	}
}

func NewTags(path string) (Tags, error) { // {{{
	var tags Tags
	rawTags, err := taglib.ReadTags(path)
	if err != nil {
		return tags, err
	}
	// fmt.Println(rawTags)

	artist, err := getTag(rawTags, taglib.Artist)
	if err != nil {
		return tags, err
	}

	album, err := getTag(rawTags, taglib.Album)
	if err != nil {
		return tags, err
	}

	date, err := getTag(rawTags, taglib.Date)
	if err != nil {
		return tags, err
	}
	d, err := strconv.Atoi(date)
	if err != nil {
		return tags, err
	}

	track, err := getTag(rawTags, taglib.TrackNumber)
	if err != nil {
		return tags, err
	}
	t, err := strconv.Atoi(track)
	if err != nil {
		return tags, err
	}

	title, err := getTag(rawTags, taglib.Title)
	if err != nil {
		return tags, err
	}

	genre, err := getTag(rawTags, taglib.Genre)
	if err != nil {
		return tags, err
	}

	tags.Artist = artist
	tags.Album = album
	tags.Date = uint16(d)
	tags.TrackNumber = uint16(t)
	tags.Title = title
	tags.Genre = genre

	// tags.rawTags = rawTags
	tags.Path = path
	return tags, nil
} // }}}

// Only safe for ext4
func sanitiseTag(s string) string {
	return strings.ReplaceAll(s, "/", "-")
}

func truncatePath(p string) string {
	lastSlash := strings.LastIndexByte(p, '/')
	if lastSlash+7 > MAX_FNAME_LEN {
		// album too long, entire title must be truncated to len 6
		// (assuming <= 99 tracks), then add '/'
		// i.e. `/NN.mp3`

		// for truncation, we rely on string indexing, because it
		// handles variable rune length for us; ranging over a string
		// yields runes (which may have "length" > 1).
		//
		// however, it is safe to range over bytes in the proximity of
		// the last slash the last slash, because they are guaranteed
		// to be ascii.

		// 17 = `... (1234)/01.mp3` -- valid path must always contain these elements

		lparen := lastSlash
		for i := lastSlash; p[i] != '('; i-- {
			lparen--
		}

		space := lastSlash
		for i := lastSlash; p[i] != ' '; i++ {
			space++
		}

		// trunc := p[:MAX_FNAME_LEN-17] + "... " + p[lparen:space] + filepath.Ext(p)
		trunc := fmt.Sprintf(
			"%s... %s%s",
			p[:MAX_FNAME_LEN-17], // truncated part
			p[lparen:space],      // `(1234)/01`
			filepath.Ext(p),
		)
		return trunc
	} else {
		// only need to truncate title
		return p[:MAX_FNAME_LEN-7] + "..." + filepath.Ext(p)
	}
}

// Destination returns a file path that is guaranteed to be valid on an ext4
// file system. Compatibility with NTFS is not guaranteed!
func (t *Tags) Destination() string {
	// only single-digit track numbers are zero-padded, and only to 2
	// digits, i.e. '01', ... '99', '100'. the reason for this somewhat odd
	// decision is that in the earliest days of library management, numeric
	// sorting could not be relied upon, and albums with >100 tracks are
	// rare.
	//
	// however, mpv now seems to perform numeric sorting when playing a
	// directory.
	dest := path.Join(
		DEST_BASE,
		sanitiseTag(t.Artist),
		fmt.Sprintf(
			"%s (%d)",
			sanitiseTag(t.Album),
			t.Date,
		),
		fmt.Sprintf(
			"%02d %s%s",
			t.TrackNumber,
			sanitiseTag(t.Title),
			filepath.Ext(t.Path),
		),
	)

	if len(dest) > MAX_FNAME_LEN {
		trunc := truncatePath(dest)
		if len(trunc) > 255 { // remove eventually
			panic(1)
		}
		return trunc
	}
	return dest
}
