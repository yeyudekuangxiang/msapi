package inter

type Options map[string]interface{}

type Option func(options *Options)

func InitOptions(items ...Option) Options {
	options := Options{}
	for _, item := range items {
		item(&options)
	}
	return options
}

type IApi interface {
	SearchMusic(keywords string, limit int64, offset int64, options ...Option) ([]Music, error)
	GetArtist(limit int64, offset int64, options ...Option) ([]Artist, error)
	GetPlayUrl(Ids []string, options ...Option) ([]PlayInfo, error)
	GetDownUrl(Ids []string, options ...Option) ([]DownInfo, error)
}
type Artist struct {
	ID   string
	Name string
	Pic  string
}
type Music struct {
	ID       string
	Name     string
	Pic      string
	LyricUrl string
	Artist   []Artist
	Album    *Album
	Time     int
	Quality  []string
	DownUrl  string
}
type Album struct {
	ID          string
	Name        string
	Pic         string
	Artist      *Artist
	Size        int
	PublishTime int64 `json:"publishTime"`
}
type PlayInfo struct {
	ID      string
	MusicId string
	Name    string
	Url     string
}
type DownInfo struct {
	ID      string
	MusicId string
	Name    string
	Url     string
}
