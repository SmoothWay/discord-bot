package music

import (
	"io"
	"log"
	"sync"

	"github.com/SmoothWay/discord-bot/internal/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
	"github.com/jung-m/dca"
)

type VoiceInstance struct {
	speaking   bool
	pause      bool
	stop       bool
	skip       bool
	GuildID    string
	voice      *discordgo.VoiceConnection
	session    *discordgo.Session
	encoder    *dca.EncodeSession
	stream     *dca.StreamingSession
	queueMutex sync.Mutex
	nowPlaying Song
	queue      []Song
}

type Song struct {
	ChannelID string
	User      string
	ID        string
	VidID     string
	Tittle    string
	VideoURL  string
}

type PkgSong struct {
	data Song
	v    *VoiceInstance
}

var (
	VoiceInstances = map[string]*VoiceInstance{}
	mutex          sync.Mutex
	SongSignal     chan PkgSong
)

func globalPlay(songSig chan PkgSong) {
	for {
		select {
		case song := <-songSig:
			go song.v.PlayQueue(song.data)
		}
	}
}

func (v *VoiceInstance) Skip() bool {
	if v.speaking {
		if v.pause {
			return true
		} else {
			if v.encoder != nil {
				v.encoder.Cleanup()
			}
		}
	}

	return false
}

func (v *VoiceInstance) Stop() {
	v.stop = true
	if v.encoder != nil {
		v.encoder.Cleanup()
	}
}

func (v *VoiceInstance) QueueAdd(song Song) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	v.queue = append(v.queue, song)
}

func (v *VoiceInstance) QueueGetSong() (song Song) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()

	if len(v.queue) != 0 {
		return v.queue[0]
	}

	return
}

func (v *VoiceInstance) QueueRemoveFirst() {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()

	if len(v.queue) != 0 {
		v.queue = v.queue[1:]
	}
}

func (v *VoiceInstance) QueueRemove() {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()

	v.queue = []Song{}
}

func (v *VoiceInstance) DCA(url string) {
	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 64
	opts.Application = "lowdelay"

	encodeSession, err := dca.EncodeFile(url, opts)

	if err != nil {
		sentry.CaptureException(err)
		log.Println("ERR: Failed creating an encoding session: ", err)
		return
	}

	v.encoder = encodeSession
	done := make(chan error)

	v.stream = dca.NewStream(encodeSession, v.voice, done)

	for {
		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				sentry.CaptureException(err)
				log.Println("ERR: An error occured ", err)
				return
			}

			// Clean up in case something went wrong
			encodeSession.Cleanup()
			return
		}
	}
}

func (v *VoiceInstance) PlayQueue(song Song) {
	v.QueueAdd(song)

	if v.speaking {
		return
	}

	go func() {
		for {
			if len(v.queue) == 0 {
				log.Println("INFO: End of queue")
				utils.SendChannelMessage(v.nowPlaying.ChannelID, "[Music] End of queue")
				return
			}

			v.nowPlaying = v.QueueGetSong()
			go utils.SendChannelMessage(v.nowPlaying.ChannelID, "[Music] Now playing: "+v.nowPlaying.Tittle)

			v.stop = false
			v.skip = false
			v.speaking = true
			v.pause = false

			err := v.voice.Speaking(true)
			if err != nil {
				sentry.CaptureException(err)
				return
			}

			v.DCA(v.nowPlaying.VideoURL)

			v.QueueRemoveFirst()

			if v.stop {
				v.QueueRemove()
			}

			v.stop = false
			v.skip = false
			v.speaking = false

			err = v.voice.Speaking(false)

			if err != nil {
				sentry.CaptureException(err)
			}
		}
	}()
}
