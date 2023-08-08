package music

const ytubeVideoURL string = "https://youtube.com/watch?v="

func InitializeRoutine() {
	SongSignal := make(chan PkgSong)
	go globalPlay(SongSignal)
}
