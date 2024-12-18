package neteasy

import (
	"github.com/pkg/errors"
	"github.com/yeyudekuangxiang/msapi/pkg/neteasy"
	"github.com/yeyudekuangxiang/msapi/search/inter"
	"log"
	"strconv"
)

func SingerWoman() func(option *inter.Options) {
	return func(options *inter.Options) {
		(*options)["tp"] = 2
	}
}
func SingerMan() func(option *inter.Options) {
	return func(options *inter.Options) {
		(*options)["tp"] = 1
	}
}
func SingerTeam() func(option *inter.Options) {
	return func(options *inter.Options) {
		(*options)["tp"] = 3
	}
}

type option inter.Options

func (o option) ArtistTp() int64 {
	v, ok := o["tp"]
	if !ok {
		return 1
	}
	switch vv := v.(type) {
	case int:
		return int64(vv)
	case int64:
		return vv
	case int32:
		return int64(vv)
	default:
		log.Println("err format tp", v)
		return 1
	}
}
func (o option) PlayBr() int64 {
	return 999000
}

type Client struct {
	api *neteasy.APi
}

func NewClient(domain string) *Client {
	c := &Client{
		api: &neteasy.APi{
			Domain: domain,
		},
	}
	return c
}

func (c *Client) SetCookie(cookieStr string) {
	c.api.SetCookie(cookieStr)
}

func (c *Client) SearchMusic(keywords string, limit int64, offset int64, options ...inter.Option) ([]inter.Music, error) {
	result, err := c.api.SearchMusic(keywords, int(offset), int(limit))
	if err != nil {
		return nil, err
	}
	list := make([]inter.Music, 0, len(result.Songs))
	for _, song := range result.Songs {
		artists := make([]inter.Artist, 0)
		for _, ar := range song.Artists {
			artists = append(artists, inter.Artist{
				ID:   strconv.Itoa(ar.Id),
				Name: ar.Name,
				Pic:  ar.Img1V1Url,
			})
		}
		list = append(list, inter.Music{
			ID:       strconv.FormatInt(song.Id, 10),
			Name:     song.Name,
			Pic:      "",
			LyricUrl: "",
			Artist:   artists,
			Album: &inter.Album{
				ID:   strconv.Itoa(song.Album.Id),
				Name: song.Album.Name,
				Pic:  "",
				Artist: &inter.Artist{
					ID:   strconv.Itoa(song.Album.Artist.Id),
					Name: song.Album.Artist.Name,
					Pic:  song.Album.Artist.Img1V1Url,
				},
				Size:        song.Album.Size,
				PublishTime: song.Album.PublishTime,
			},
			Time:    0,
			Quality: []string{"320000", "999000"},
			DownUrl: "",
		})
	}
	return list, nil
}

func (c *Client) GetArtist(limit int64, offset int64, options ...inter.Option) ([]inter.Artist, error) {
	opt := inter.InitOptions(options...)
	singerList, err := c.api.GetSingerList(option(opt).ArtistTp(), -1, "", int(limit), int(offset))
	if err != nil {
		return nil, err
	}
	artist := make([]inter.Artist, 0, len(singerList))
	for _, singer := range singerList {
		artist = append(artist, inter.Artist{
			ID:   strconv.Itoa(singer.Id),
			Name: singer.Name,
			Pic:  singer.PicUrl,
		})
	}
	return artist, nil
}

func (c *Client) GetPlayUrl(Ids []string, options ...inter.Option) ([]inter.PlayInfo, error) {
	opt := inter.InitOptions(options...)

	realIds := make([]int64, 0, len(Ids))
	for _, idStr := range Ids {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		realIds = append(realIds, id)
	}
	if len(realIds) == 0 {
		return nil, nil
	}
	playList, err := c.api.GetPlayUrl(realIds, option(opt).PlayBr())
	if err != nil {
		return nil, err
	}
	platInfos := make([]inter.PlayInfo, 0, len(playList))
	for _, pl := range playList {
		platInfos = append(platInfos, inter.PlayInfo{
			ID:      strconv.FormatInt(pl.Id, 10),
			MusicId: pl.MusicId,
			Name:    "",
			Url:     pl.Url,
		})
	}
	return platInfos, nil
}

func (c *Client) GetDownUrl(Ids []string, options ...inter.Option) ([]inter.DownInfo, error) {
	opt := inter.InitOptions(options...)

	realIds := make([]int64, 0, len(Ids))
	for _, idStr := range Ids {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		realIds = append(realIds, id)
	}
	if len(realIds) == 0 {
		return nil, nil
	}
	playList, err := c.api.GetPlayUrl(realIds, option(opt).PlayBr())
	if err != nil {
		return nil, err
	}
	platInfos := make([]inter.DownInfo, 0, len(playList))
	for _, pl := range playList {
		platInfos = append(platInfos, inter.DownInfo{
			ID:      strconv.FormatInt(pl.Id, 10),
			MusicId: pl.MusicId,
			Name:    "",
			Url:     pl.Url,
		})
	}
	return platInfos, nil

}
