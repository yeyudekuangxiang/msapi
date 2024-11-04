package search

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
)

type LocalCache struct {
	Dir  string
	once sync.Once
}

func (l *LocalCache) Set(key string, value []byte) error {
	l.once.Do(func() {
		os.MkdirAll(l.Dir, 0777)
	})
	return os.WriteFile(path.Join(l.Dir, l.md5(key)), value, 0755)
}
func (l *LocalCache) Get(key string) ([]byte, bool, error) {
	data, err := os.ReadFile(path.Join(l.Dir, l.md5(key)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, err
		}
		return nil, false, nil
	}
	return data, true, nil
}
func (l *LocalCache) md5(key string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(key)))
}

type NetEasyAPi struct {
	Domain           string
	netEasyLoginResp netEasyLoginResp
	cookieStr        string
	Cache            *LocalCache
}

func (n *NetEasyAPi) EmailLogin(email, password string) error {
	resp, err := http.Get(fmt.Sprintf("%s/login?email=%s&password=%s", n.Domain, email, password))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("Login error: " + resp.Status)
	}
	defer resp.Body.Close()

	cookieStr := ""
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "NMTID" || cookie.Name == "MUSIC_U" || cookie.Name == "__csrf" {
			cookieStr += fmt.Sprintf("; %s=%s", cookie.Name, cookie.Value)
		}
	}
	if cookieStr != "" {
		cookieStr = cookieStr[2:]
		n.cookieStr = cookieStr
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	v := netEasyLoginResp{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return err
	}
	if v.Code != 200 {
		return fmt.Errorf("Login error:%d ", v.Code)
	}
	n.netEasyLoginResp = v
	return nil
}
func (n *NetEasyAPi) SearchLrcBest(param SearchLrcParam) ([]byte, error) {
	if data, exist, _ := n.Cache.Get("SearchLrcBest" + param.Title + param.Album + param.Artist); exist {
		return data, nil
	}
	musicResult, err := n.searchMusic(fmt.Sprintf("%s %s %s", param.Title, param.Artist, param.Album), param.Offset, param.Limit)
	if err != nil {
		return nil, err
	}
	lrcContent := ""
	for _, m := range musicResult.Songs {
		if m.Name == param.Title && m.Album.Name == param.Album {
			for _, ar := range m.Artists {
				if strings.Contains(param.Artist, ar.Name) {
					lrcContent, err = n.getMusicLrc(m.Id)
					if err != nil {
						return nil, err
					}
					goto hasLrc
				}
			}

		}
		if m.Name == param.Title {
			for _, ar := range m.Artists {
				if strings.Contains(param.Artist, ar.Name) {
					lrcContent, err = n.getMusicLrc(m.Id)
					if err != nil {
						return nil, err
					}
					goto hasLrc
				}
			}
		}
		if m.Name == param.Title && m.Album.Name == param.Album {
			lrcContent, err = n.getMusicLrc(m.Id)
			if err != nil {
				return nil, err
			}
			goto hasLrc
		}
		if m.Name == param.Title {
			lrcContent, err = n.getMusicLrc(m.Id)
			if err != nil {
				return nil, err
			}
			goto hasLrc
		}

	}
hasLrc:
	if lrcContent != "" {
		n.Cache.Set("SearchLrcBest"+param.Title+param.Album+param.Artist, []byte(lrcContent))
	}
	return []byte(lrcContent), nil
}
func (n *NetEasyAPi) SearchLrc(param SearchLrcParam) ([]Lrc, error) {
	musicResult, err := n.searchMusic(fmt.Sprintf("%s %s %s", param.Title, param.Artist, param.Album), param.Offset, param.Limit)
	if err != nil {
		return nil, err
	}
	lrcList := make([]Lrc, 0)
	for _, music := range musicResult.Songs {
		lrcContent, err := n.getMusicLrc(music.Id)
		if err != nil {
			return nil, err
		}
		if lrcContent == "" {
			continue
		}
		artistName := ""
		for _, art := range music.Artists {
			artistName += "," + art.Name
		}
		if artistName != "" {
			artistName = artistName[1:]
		}
		lrcList = append(lrcList, Lrc{
			Id:     fmt.Sprintf("%d", music.Id),
			Title:  music.Name,
			Artist: artistName,
			Lyrics: lrcContent,
		})
	}
	return lrcList, nil
}

func (n *NetEasyAPi) ConfirmLrc(param ConfirmLrcParam) error {
	return nil
}

func (n *NetEasyAPi) GetMusicCover(param GetMusicCoverParam) ([]byte, error) {
	return n.GetSingerCover(GetSingerCoverParam{
		Artist: param.Artist,
	})
	/*resp, err := http.Get(fmt.Sprintf("%s/search?keywords=%s %s %s&type=1&offset=%d&limit=%d", n.Domain, param.Title, param.Artist, param.Album, param.Offset, param.Limit))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("SearchLrc %+v %s\n", param, body)
	musicResp := netEasySearchMusicResp{}
	err = json.Unmarshal(body, &musicResp)
	if err != nil {
		return nil, err
	}
	if musicResp.Code != 200 {
		return nil, fmt.Errorf("%d", musicResp.Code)
	}
	id1 := int64(0)
	id2 := int64(0)
	id3 := int64(0)
	id4 := int64(0)
	for _, music := range musicResp.Result.Songs {
		if music.Name == param.Title {
			id4 = music.Id
			if len(music.Artists) > 0 {
				for _, art := range music.Artists {
					if art.Name == param.Title {
						id2 = music.Id
					}
				}
			}
			if music.Album.Name == param.Album {
				if id2 != 0 {
					id1 = music.Id
				}
				id3 = music.Id
			}
		}
	}
	*/
}
func (n *NetEasyAPi) getMusicInfoById(id int64) {

}
func (n *NetEasyAPi) GetSingerCover(param GetSingerCoverParam) ([]byte, error) {
	info, exist, err := n.getSingerInfo(param.Artist)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	resp, err := http.Get(info.ArtistAvatarPicUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (n *NetEasyAPi) GetAlbumCover(param GetAlbumCoverParam) ([]byte, error) {
	info, exist, err := n.getAlbumInfo(param.Artist, param.Album)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	resp, err := http.Get(info.PicUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}
	return io.ReadAll(resp.Body)
}
func (n *NetEasyAPi) getSingerInfo(name string) (*NetEasySearchSingerInfo, bool, error) {

	if cData, exist, _ := n.Cache.Get("GetSingerInfo" + name); exist {
		v := NetEasySearchSingerInfo{}
		err := json.Unmarshal(cData, &v)
		if err != nil {
			log.Println("解析缓存失败", err)
		} else {
			log.Println("从缓存中获取歌手信息", name)
			return &v, true, nil
		}
	}
	infos, err := n.searchSinger(name)
	if err != nil {
		return nil, false, err
	}
	for _, info := range infos {
		if info.ArtistName == name {
			cData, err := json.Marshal(info)
			if err != nil {
				log.Println("缓存歌手信息失败", err)
			} else {
				err = n.Cache.Set("GetSingerInfo"+name, cData)
				if err != nil {
					log.Println("缓存歌手信息失败", err)
				}
			}
			return &info, true, nil
		}
	}
	return nil, false, err
}
func (n *NetEasyAPi) getAlbumInfo(artist string, album string) (*netEasySearchAlbumInfo, bool, error) {

	if cData, exist, _ := n.Cache.Get("getAlbumInfo" + artist + album); exist {
		v := netEasySearchAlbumInfo{}
		err := json.Unmarshal(cData, &v)
		if err != nil {
			log.Println("解析缓存失败", err)
		} else {
			log.Println("从缓存中获取歌手信息", artist, album)
			return &v, true, nil
		}
	}
	infos, err := n.searchAlbum(album+" "+artist, 0, 10)
	if err != nil {
		return nil, false, err
	}
	for _, info := range infos.Albums {
		if info.Name == album && info.Artist.Name == artist {
			cData, err := json.Marshal(info)
			if err != nil {
				log.Println("缓存专辑信息失败", err)
			} else {
				err = n.Cache.Set("getAlbumInfo"+artist+album, cData)
				if err != nil {
					log.Println("缓存专辑信息失败", err)
				}
			}
			return &info, true, nil
		}
	}
	return nil, false, err
}
func (n *NetEasyAPi) searchSinger(name string) ([]NetEasySearchSingerInfo, error) {
	if cData, exist, _ := n.Cache.Get("searchSinger" + name); exist {
		v := netEasySearchSingerResp{}
		err := json.Unmarshal(cData, &v)
		if err != nil {
			log.Println("解析搜索歌手缓存失败", err)
		} else {
			log.Println("从缓存中获取搜索歌手", string(cData))
			return v.Data.List, nil
		}
	}
	body, err := n.get(fmt.Sprintf("%s/ugc/artist/search?keyword=%s", n.Domain, url.QueryEscape(name)), true)
	if err != nil {
		return nil, err
	}

	v := netEasySearchSingerResp{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return nil, err
	}
	if v.Code != 200 {
		return nil, fmt.Errorf("%d", v.Code)
	}
	err = n.Cache.Set("searchSinger"+name, body)
	if err != nil {
		log.Println("缓存搜索歌手失败", err)
	}
	return v.Data.List, nil
}
func (n *NetEasyAPi) postJson(url string, data interface{}, needLogin bool) ([]byte, error) {
	reqBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if needLogin {
		req.Header.Set("Cookie", n.cookieStr)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
func (n *NetEasyAPi) get(url string, needLogin bool) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if needLogin {
		req.Header.Set("Cookie", n.cookieStr)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (n *NetEasyAPi) searchMusic(keywords string, offset, limit int) (*netEasySearchMusicResult, error) {
	cacheKey := fmt.Sprintf("searchMusic%s%d%d", keywords, offset, limit)

	if cacheData, exist, _ := n.Cache.Get(cacheKey); exist {
		v := netEasySearchMusicResp{}
		err := json.Unmarshal(cacheData, &v)
		if err != nil {
			log.Println("解析搜素缓存失败", err)
		} else {
			return &v.Result, nil
		}
	}
	respBody, err := n.get(fmt.Sprintf("%s/search?keywords=%s&type=1&offset=%d&limit=%d", n.Domain, url.QueryEscape(keywords), offset, limit), false)
	if err != nil {
		return nil, err
	}
	err = n.Cache.Set(cacheKey, respBody)
	if err != nil {
		log.Println("缓存搜索结果失败", err)
	}
	v := netEasySearchMusicResp{}
	err = json.Unmarshal(respBody, &v)
	if err != nil {
		return nil, err
	}
	if v.Code != 200 {
		return nil, fmt.Errorf("%d", v.Code)
	}
	return &v.Result, nil
}
func (n *NetEasyAPi) searchAlbum(keywords string, offset, limit int) (*netEasySearchAlbumResult, error) {
	cacheKey := fmt.Sprintf("searchAlbum%s%d%d", keywords, offset, limit)

	if cacheData, exist, _ := n.Cache.Get(cacheKey); exist {
		v := netEasySearchAlbumResp{}
		err := json.Unmarshal(cacheData, &v)
		if err != nil {
			log.Println("解析搜素缓存失败", err)
		} else {
			return &v.Result, nil
		}
	}
	respBody, err := n.get(fmt.Sprintf("%s/search?keywords=%s&type=10&offset=%d&limit=%d", n.Domain, url.QueryEscape(keywords), offset, limit), false)
	if err != nil {
		return nil, err
	}
	err = n.Cache.Set(cacheKey, respBody)
	if err != nil {
		log.Println("缓存搜索结果失败", err)
	}
	v := netEasySearchAlbumResp{}
	err = json.Unmarshal(respBody, &v)
	if err != nil {
		return nil, err
	}
	if v.Code != 200 {
		return nil, fmt.Errorf("%d", v.Code)
	}
	return &v.Result, nil
}
func (n *NetEasyAPi) getMusicLrc(id int64) (string, error) {
	if cData, exist, _ := n.Cache.Get(fmt.Sprintf("lrc%d", id)); exist {
		return string(cData), nil
	}
	lrcBody, err := n.get(fmt.Sprintf("%s/lyric?id=%d", n.Domain, id), false)
	if err != nil {
		return "", err
	}
	lrc := netEasyMusicLrcResp{}
	err = json.Unmarshal(lrcBody, &lrc)
	if err != nil {
		return "", err
	}
	if lrc.Code != 200 {
		return "", fmt.Errorf("%d", lrc.Code)
	}
	if lrc.Lrc.Lyric != "" {
		err = n.Cache.Set(fmt.Sprintf("lrc%d", id), []byte(lrc.Lrc.Lyric))
		if err != nil {
			log.Println("缓存歌词失败", err)
		}
	}

	return lrc.Lrc.Lyric, nil
}

type netEasySearchMusicResult struct {
	Songs []struct {
		Id      int64  `json:"id"`
		Name    string `json:"name"`
		Artists []struct {
			Id        int           `json:"id"`
			Name      string        `json:"name"`
			PicUrl    interface{}   `json:"picUrl"`
			Alias     []interface{} `json:"alias"`
			AlbumSize int           `json:"albumSize"`
			PicId     int           `json:"picId"`
			FansGroup interface{}   `json:"fansGroup"`
			Img1V1Url string        `json:"img1v1Url"`
			Img1V1    int           `json:"img1v1"`
			Trans     interface{}   `json:"trans"`
		} `json:"artists"`
		Album struct {
			Id     int    `json:"id"`
			Name   string `json:"name"`
			Artist struct {
				Id        int           `json:"id"`
				Name      string        `json:"name"`
				PicUrl    interface{}   `json:"picUrl"`
				Alias     []interface{} `json:"alias"`
				AlbumSize int           `json:"albumSize"`
				PicId     int           `json:"picId"`
				FansGroup interface{}   `json:"fansGroup"`
				Img1V1Url string        `json:"img1v1Url"`
				Img1V1    int           `json:"img1v1"`
				Trans     interface{}   `json:"trans"`
			} `json:"artist"`
			PublishTime int64    `json:"publishTime"`
			Size        int      `json:"size"`
			CopyrightId int      `json:"copyrightId"`
			Status      int      `json:"status"`
			PicId       int64    `json:"picId"`
			Mark        int      `json:"mark"`
			Alia        []string `json:"alia,omitempty"`
		} `json:"album"`
		Duration    int         `json:"duration"`
		CopyrightId int         `json:"copyrightId"`
		Status      int         `json:"status"`
		Alias       []string    `json:"alias"`
		Rtype       int         `json:"rtype"`
		Ftype       int         `json:"ftype"`
		Mvid        int         `json:"mvid"`
		Fee         int         `json:"fee"`
		RUrl        interface{} `json:"rUrl"`
		Mark        int64       `json:"mark"`
	} `json:"songs"`
	HasMore   bool `json:"hasMore"`
	SongCount int  `json:"songCount"`
}
type netEasySearchMusicResp struct {
	Result netEasySearchMusicResult `json:"result"`
	Code   int                      `json:"code"`
}
type netEasyMusicLrcResp struct {
	Sgc bool `json:"sgc"`
	Sfy bool `json:"sfy"`
	Qfy bool `json:"qfy"`
	Lrc struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"lrc"`
	Klyric struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"klyric"`
	Tlyric struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"tlyric"`
	Romalrc struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"romalrc"`
	Code int `json:"code"`
}

type netEasyLoginResp struct {
	LoginType  int    `json:"loginType"`
	ClientId   string `json:"clientId"`
	EffectTime int    `json:"effectTime"`
	Code       int    `json:"code"`
	Account    struct {
		Id                 int    `json:"id"`
		UserName           string `json:"userName"`
		Type               int    `json:"type"`
		Status             int    `json:"status"`
		WhitelistAuthority int    `json:"whitelistAuthority"`
		CreateTime         int64  `json:"createTime"`
		Salt               string `json:"salt"`
		TokenVersion       int    `json:"tokenVersion"`
		Ban                int    `json:"ban"`
		BaoyueVersion      int    `json:"baoyueVersion"`
		DonateVersion      int    `json:"donateVersion"`
		VipType            int    `json:"vipType"`
		ViptypeVersion     int64  `json:"viptypeVersion"`
		AnonimousUser      bool   `json:"anonimousUser"`
		Uninitialized      bool   `json:"uninitialized"`
	} `json:"account"`
	Token   string `json:"token"`
	Profile struct {
		Followed           bool   `json:"followed"`
		BackgroundUrl      string `json:"backgroundUrl"`
		AvatarImgIdStr     string `json:"avatarImgIdStr"`
		BackgroundImgIdStr string `json:"backgroundImgIdStr"`
		UserType           int    `json:"userType"`
		VipType            int    `json:"vipType"`
		AuthStatus         int    `json:"authStatus"`
		DjStatus           int    `json:"djStatus"`
		DetailDescription  string `json:"detailDescription"`
		Experts            struct {
		} `json:"experts"`
		ExpertTags                interface{} `json:"expertTags"`
		AccountStatus             int         `json:"accountStatus"`
		Nickname                  string      `json:"nickname"`
		Birthday                  int64       `json:"birthday"`
		Gender                    int         `json:"gender"`
		Province                  int         `json:"province"`
		City                      int         `json:"city"`
		AvatarImgId               int64       `json:"avatarImgId"`
		BackgroundImgId           int64       `json:"backgroundImgId"`
		AvatarUrl                 string      `json:"avatarUrl"`
		DefaultAvatar             bool        `json:"defaultAvatar"`
		Mutual                    bool        `json:"mutual"`
		RemarkName                interface{} `json:"remarkName"`
		Description               string      `json:"description"`
		UserId                    int         `json:"userId"`
		Signature                 string      `json:"signature"`
		Authority                 int         `json:"authority"`
		Followeds                 int         `json:"followeds"`
		Follows                   int         `json:"follows"`
		EventCount                int         `json:"eventCount"`
		AvatarDetail              interface{} `json:"avatarDetail"`
		PlaylistCount             int         `json:"playlistCount"`
		PlaylistBeSubscribedCount int         `json:"playlistBeSubscribedCount"`
	} `json:"profile"`
	Bindings []struct {
		BindingTime  int64  `json:"bindingTime"`
		RefreshTime  int    `json:"refreshTime"`
		TokenJsonStr string `json:"tokenJsonStr"`
		ExpiresIn    int    `json:"expiresIn"`
		Url          string `json:"url"`
		Expired      bool   `json:"expired"`
		UserId       int    `json:"userId"`
		Id           int64  `json:"id"`
		Type         int    `json:"type"`
	} `json:"bindings"`
	Cookie string `json:"cookie"`
}

type netEasySearchSingerResp struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Data struct {
		TotalCount int                       `json:"totalCount"`
		List       []NetEasySearchSingerInfo `json:"list"`
		ExtraCount struct {
		} `json:"extraCount"`
	} `json:"data"`
}
type NetEasySearchSingerInfo struct {
	ArtistId           int    `json:"artistId"`
	ArtistName         string `json:"artistName"`
	ArtistAvatarPicUrl string `json:"artistAvatarPicUrl"`
}
type netEasySearchAlbumResp struct {
	Result netEasySearchAlbumResult `json:"result"`
	Code   int                      `json:"code"`
}
type netEasySearchAlbumResult struct {
	HlWords    []string                 `json:"hlWords"`
	Albums     []netEasySearchAlbumInfo `json:"albums"`
	AlbumCount int                      `json:"albumCount"`
}
type netEasySearchAlbumInfo struct {
	Name        string `json:"name"`
	Id          int    `json:"id"`
	Type        string `json:"type"`
	Size        int    `json:"size"`
	PicId       int64  `json:"picId"`
	BlurPicUrl  string `json:"blurPicUrl"`
	CompanyId   int    `json:"companyId"`
	Pic         int64  `json:"pic"`
	PicUrl      string `json:"picUrl"`
	PublishTime int64  `json:"publishTime"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Company     string `json:"company"`
	BriefDesc   string `json:"briefDesc"`
	Artist      struct {
		Name        string   `json:"name"`
		Id          int      `json:"id"`
		PicId       int64    `json:"picId"`
		Img1V1Id    int64    `json:"img1v1Id"`
		BriefDesc   string   `json:"briefDesc"`
		PicUrl      string   `json:"picUrl"`
		Img1V1Url   string   `json:"img1v1Url"`
		AlbumSize   int      `json:"albumSize"`
		Alias       []string `json:"alias"`
		Trans       string   `json:"trans"`
		MusicSize   int      `json:"musicSize"`
		TopicPerson int      `json:"topicPerson"`
		PicIdStr    string   `json:"picId_str,omitempty"`
		Img1V1IdStr string   `json:"img1v1Id_str,omitempty"`
		Alia        []string `json:"alia"`
	} `json:"artist"`
	Songs           interface{} `json:"songs"`
	Alias           []string    `json:"alias"`
	Status          int         `json:"status"`
	CopyrightId     int         `json:"copyrightId"`
	CommentThreadId string      `json:"commentThreadId"`
	Artists         []struct {
		Name        string        `json:"name"`
		Id          int           `json:"id"`
		PicId       int           `json:"picId"`
		Img1V1Id    int64         `json:"img1v1Id"`
		BriefDesc   string        `json:"briefDesc"`
		PicUrl      string        `json:"picUrl"`
		Img1V1Url   string        `json:"img1v1Url"`
		AlbumSize   int           `json:"albumSize"`
		Alias       []interface{} `json:"alias"`
		Trans       string        `json:"trans"`
		MusicSize   int           `json:"musicSize"`
		TopicPerson int           `json:"topicPerson"`
		Img1V1IdStr string        `json:"img1v1Id_str,omitempty"`
	} `json:"artists"`
	Paid          bool   `json:"paid"`
	OnSale        bool   `json:"onSale"`
	PicIdStr      string `json:"picId_str,omitempty"`
	Alg           string `json:"alg"`
	Mark          int    `json:"mark"`
	ContainedSong string `json:"containedSong"`
}
