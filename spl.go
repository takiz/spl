package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

func main() {
	volumeLevel := flag.Float64("vol", 100, "Set volume level in percents")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "The following formats are supported: flac, mp3, ogg, wav.\n\n"+
			"Usage: %s [options] file\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n\nVersion 0.1\n")
		os.Exit(1)
	}
	flag.Parse()
	req := flag.Args()
	if len(req) == 0 {
		os.Exit(1)
	}
	r := strings.Join(req, " ")
	f, err := os.Open(r)
	if err != nil {
		log.Fatal(err)
	}

	var streamer beep.StreamSeekCloser
	var format beep.Format
	var resampler *beep.Resampler
	ratio := 1.0
	s := strings.ToLower(r)
	if strings.HasSuffix(s, ".mp3") {
		streamer, format, err = mp3.Decode(f)
	} else if strings.HasSuffix(s, ".ogg") {
		streamer, format, err = vorbis.Decode(f)
		ratio = 0.5
	} else if strings.HasSuffix(s, ".wav") {
		streamer, format, err = wav.Decode(f)
	} else {
		streamer, format, err = flac.Decode(f)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	resampler = beep.ResampleRatio(4, ratio, streamer)
	volume := &effects.Volume{
		Streamer: resampler,
		Base:     2,
		Volume:   (*volumeLevel - 100) / 10,
		Silent:   false,
	}
	// Channel, which will signal the end of the playback.
	playing := make(chan struct{})

	speaker.Play(beep.Seq(volume, beep.Callback(func() {
		// Callback after the stream ends.
		close(playing)
	})))
	<-playing
}
