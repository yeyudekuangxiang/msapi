package search

type SearchLrcParam struct {
	Title    string
	Artist   string
	Path     string
	Album    string
	Duration int
	Offset   int
	Limit    int
}
type ConfirmLrcParam struct {
	Path     string `json:"path"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Lyrics   string `json:"lyrics"`
	LyricsId string `json:"lyricsId"`
}
type GetMusicCoverParam struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
}
type GetSingerCoverParam struct {
	Artist string `json:"artist"`
}
type GetAlbumCoverParam struct {
	Artist string `json:"artist"`
	Album  string `json:"album"`
}
type IApi interface {
	SearchLrc(param SearchLrcParam) ([]Lrc, error)
	ConfirmLrc(param ConfirmLrcParam) error
	GetMusicCover(param GetMusicCoverParam) ([]byte, error)
	GetSingerCover(param GetSingerCoverParam) ([]byte, error)
	GetAlbumCover(param GetAlbumCoverParam) ([]byte, error)
}

type Lrc struct {
	Id     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Lyrics string `json:"lyrics"`
}
