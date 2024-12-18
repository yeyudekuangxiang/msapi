package freemp3

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/yeyudekuangxiang/msapi/pkg/freemp3"
	"github.com/yeyudekuangxiang/msapi/search/inter"
	"log"
	"strconv"
	"time"
)

type option inter.Options

func (o option) Quality() (string, error) {

	v, ok := o["quality"]
	if !ok {
		return "", errors.New("no quality option")
	}
	if vv, ok := v.(string); ok {
		return vv, nil
	}
	return "", errors.Errorf("error quailty option:%v", v)
}
func QualityOption(quality string) func(options *inter.Options) {
	return func(options *inter.Options) {
		(*options)["quality"] = quality
	}
}

type Client struct {
	api *freemp3.Client
}

func NewClient(domain string) *Client {
	return &Client{
		api: &freemp3.Client{
			Domain: domain,
			Header: nil,
		},
	}
}
func (c *Client) SetHeader(header map[string]string) {
	c.api.Header = header
}
func (c *Client) SearchMusic(keywords string, limit int64, offset int64, options ...inter.Option) ([]inter.Music, error) {
	limit = 20
	resp, err := c.api.Search(keywords, int(offset/limit+1))
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, errors.Errorf("search fail:%d", resp.Code)
	}
	list := make([]inter.Music, 0, len(resp.Data.List))
	for _, item := range resp.Data.List {
		artists := make([]inter.Artist, 0)
		for _, ar := range item.Artist {
			artists = append(artists, inter.Artist{
				ID:   ar.Id,
				Name: ar.Name,
				Pic:  "",
			})
		}
		qualityStr, _ := json.Marshal(item.Quality)
		quailtyList, _ := c.decodeQuality(string(qualityStr))

		list = append(list, inter.Music{
			ID:       item.Id,
			Name:     item.Name,
			Pic:      item.Pic,
			LyricUrl: item.Lyric,
			Artist:   artists,
			Album: &inter.Album{
				ID:          item.Album.Id,
				Name:        item.Album.Name,
				Pic:         item.Album.Pic,
				Artist:      &inter.Artist{},
				Size:        0,
				PublishTime: 0,
			},
			Time:    item.Time,
			Quality: quailtyList,
			DownUrl: "",
		})
	}
	return list, nil
}

func (c *Client) GetArtist(limit int64, offset int64, options ...inter.Option) ([]inter.Artist, error) {
	limit = 90
	resp, err := c.api.GetArtistList(int(offset/limit + 1))
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, errors.Errorf("search fail:%d", resp.Code)
	}
	artists := make([]inter.Artist, 0, len(resp.Data.List))
	for _, item := range resp.Data.List {
		artists = append(artists, inter.Artist{
			ID:   item.Id,
			Name: item.Name,
			Pic:  item.Pic,
		})
	}
	return artists, nil
}

func (c *Client) GetPlayUrl(Ids []string, options ...inter.Option) ([]inter.PlayInfo, error) {
	opt := inter.InitOptions(options...)
	quality, err := option(opt).Quality()
	if err != nil {
		return nil, err
	}
	list := make([]inter.PlayInfo, 0)
	for _, id := range Ids {
		downUrl, err := c.api.GetRealDownLoadUrl(id, quality)
		if err != nil {
			log.Printf("get %s playurl error:%v\n", id, err)
		}
		list = append(list, inter.PlayInfo{
			ID:      id + "_" + quality,
			MusicId: id,
			Name:    "",
			Url:     downUrl,
		})
		time.Sleep(time.Millisecond * 300)
	}
	return list, nil
}

func (c *Client) GetDownUrl(Ids []string, options ...inter.Option) ([]inter.DownInfo, error) {
	opt := inter.InitOptions(options...)
	quality, err := option(opt).Quality()
	if err != nil {
		return nil, err
	}
	list := make([]inter.DownInfo, 0)
	for _, id := range Ids {
		downUrl, err := c.api.GetRealDownLoadUrl(id, quality)
		if err != nil {
			log.Printf("get %s playurl error:%v\n", id, err)
		}
		list = append(list, inter.DownInfo{
			ID:      id + "_" + quality,
			MusicId: id,
			Name:    "",
			Url:     downUrl,
		})
		time.Sleep(time.Millisecond * 300)
	}
	return list, nil
}
func (c *Client) decodeQuality(qualityStr string) ([]string, error) {
	quality := make([]interface{}, 0)
	err := json.Unmarshal([]byte(qualityStr), &quality)
	if err != nil {
		return nil, err
	}
	list := make([]string, 0)
	if len(quality) == 0 {
		return list, nil
	}
	for _, q := range quality {
		switch qq := q.(type) {
		case int64:
			list = append(list, strconv.FormatInt(qq, 10))
		case string:
			list = append(list, qq)
		case map[string]interface{}:
			list = append(list, qq["name"].(string))
		case float64:
			list = append(list, strconv.FormatInt(int64(qq), 10))
		default:
			return list, errors.New("未识别到质量" + qualityStr)
		}
	}
	return list, nil
}
