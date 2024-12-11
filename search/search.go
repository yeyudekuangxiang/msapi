package search

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type NetEasyAPi struct {
	Domain           string
	Email            string
	Password         string
	netEasyLoginResp netEasyLoginResp
	cookieStr        string
}

// GetPlayUrl 获取播放列表
func (n *NetEasyAPi) GetPlayUrl(ids []int64, br int64) ([]NetEasyGetPlayUrlInfo, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if br <= 0 {
		br = 999000
	}
	idStr := ""
	for _, item := range ids {
		idStr += "," + strconv.FormatInt(item, 10)
	}
	if idStr != "" {
		idStr = idStr[1:]
	}
	body, err := n.get(fmt.Sprintf("%s/song/url?id=%s&br=%d", n.Domain, idStr, br), true)
	if err != nil {
		return nil, err
	}
	v := NetEasyGetPlayUrlResp{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return nil, err
	}
	if v.Code != 200 {
		return nil, fmt.Errorf("%d", v.Code)
	}
	return v.Data, nil
}

// EmailLogin 邮箱登录
func (n *NetEasyAPi) EmailLogin() error {
	return nil
	resp, err := http.Get(fmt.Sprintf("%s/login?email=%s&password=%s", n.Domain, n.Email, n.Password))
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

func (n *NetEasyAPi) SearchSinger(name string) ([]NetEasySearchSingerInfo, error) {

	body, err := n.get(fmt.Sprintf("%s/ugc/artist/search?keyword=%s", n.Domain, url.QueryEscape(name)), true)
	if err != nil {
		return nil, err
	}
	v := NetEasySearchSingerResp{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return nil, err
	}
	if v.Code != 200 {
		return nil, fmt.Errorf("%d", v.Code)
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
		if n.cookieStr == "" {
			err = n.EmailLogin()
			if err != nil {
				log.Println("登录失败", err)
				return nil, err
			}
		}
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
		if n.cookieStr == "" {
			err = n.EmailLogin()
			if err != nil {
				log.Println("登录失败", err)
				return nil, err
			}
		}
		req.Header.Set("Cookie", n.cookieStr)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		errBody, err := io.ReadAll(resp.Body)
		if err == nil {
			return errBody, fmt.Errorf("%d", resp.StatusCode)
		}
		return nil, fmt.Errorf("%d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (n *NetEasyAPi) SearchMusic(keywords string, offset, limit int) (*NetEasySearchMusicResult, error) {

	respBody, err := n.get(fmt.Sprintf("%s/search?keywords=%s&type=1&offset=%d&limit=%d", n.Domain, url.QueryEscape(keywords), offset, limit), true)
	if err != nil {
		log.Println("搜索歌曲失败", string(respBody))
		return nil, err
	}
	v := NetEasySearchMusicResp{}
	err = json.Unmarshal(respBody, &v)
	if err != nil {
		return nil, err
	}
	if v.Code != 200 {
		return nil, fmt.Errorf("%d", v.Code)
	}
	return &v.Result, nil
}
func (n *NetEasyAPi) SearchAlbum(keywords string, offset, limit int) (*NetEasySearchAlbumResult, error) {

	respBody, err := n.get(fmt.Sprintf("%s/search?keywords=%s&type=10&offset=%d&limit=%d", n.Domain, url.QueryEscape(keywords), offset, limit), true)
	if err != nil {
		return nil, err
	}
	v := NetEasySearchAlbumResp{}
	err = json.Unmarshal(respBody, &v)
	if err != nil {
		return nil, err
	}
	if v.Code != 200 {
		return nil, fmt.Errorf("%d", v.Code)
	}
	return &v.Result, nil
}
func (n *NetEasyAPi) GetMusicLrc(id int64) (string, error) {
	lrcBody, err := n.get(fmt.Sprintf("%s/lyric?id=%d", n.Domain, id), true)
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
	return lrc.Lrc.Lyric, nil
}
func (n *NetEasyAPi) SearchSuggest(keywords string) (*NetEasySearchSuggestResult, error) {
	body, err := n.get(fmt.Sprintf("%s/search/suggest?keywords=%s", n.Domain, url.QueryEscape(keywords)), true)
	if err != nil {
		return nil, err
	}
	r := NetEasySearchSuggestResp{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	if r.Code != 200 {
		return nil, fmt.Errorf("%d", r.Code)
	}
	return &r.Result, nil
}

/*
GetSingerList

	type 取值
	1:男歌手
	2:女歌手
	3:乐队

	area 取值
	-1:全部
	7华语
	96欧美
	8:日本
	16韩国
	0:其他

	initial 取值 a-z/A-Z
*/
func (n *NetEasyAPi) GetSingerList(tp int64, area int64, code string, limit int, offset int) ([]NetEasyArtistInfo, error) {
	u := fmt.Sprintf("%s/artist/list?type=%d&area=%d&initial=%s&limit=%d&offset=%d", n.Domain, tp, area, code, limit, offset)
	body, err := n.get(u, true)
	if err != nil {
		return nil, err
	}
	resp := NetEasySingerResp{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("code err:%d", resp.Code)
	}
	return resp.Artists, nil
}

func (n *NetEasyAPi) SetCookie(cookie string) {
	n.cookieStr = cookie
}

type NetEasySearchMusicResult struct {
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
type NetEasySearchMusicResp struct {
	Result NetEasySearchMusicResult `json:"result"`
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
type NetEasySearchSingerResp struct {
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
type NetEasySearchAlbumResp struct {
	Result NetEasySearchAlbumResult `json:"result"`
	Code   int                      `json:"code"`
}
type NetEasySearchAlbumResult struct {
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
type NetEasySearchSuggestResp struct {
	Result NetEasySearchSuggestResult `json:"result"`
	Code   int                        `json:"code"`
}
type NetEasySearchSuggestResult struct {
	Albums []struct {
		Id     int    `json:"id"`
		Name   string `json:"name"`
		Artist struct {
			Id        int         `json:"id"`
			Name      string      `json:"name"`
			PicUrl    *string     `json:"picUrl"`
			Alias     []string    `json:"alias"`
			AlbumSize int         `json:"albumSize"`
			PicId     int64       `json:"picId"`
			FansGroup interface{} `json:"fansGroup"`
			Img1V1Url string      `json:"img1v1Url"`
			Img1V1    int         `json:"img1v1"`
			Alia      []string    `json:"alia,omitempty"`
			Trans     interface{} `json:"trans"`
		} `json:"artist"`
		PublishTime int64    `json:"publishTime"`
		Size        int      `json:"size"`
		CopyrightId int      `json:"copyrightId"`
		Status      int      `json:"status"`
		PicId       int64    `json:"picId"`
		Mark        int      `json:"mark"`
		TransNames  []string `json:"transNames,omitempty"`
	} `json:"albums"`
	Artists []NetEasySearchSuggestArtist `json:"artists"`
	Songs   []struct {
		Id      int    `json:"id"`
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
			PublishTime int64 `json:"publishTime"`
			Size        int   `json:"size"`
			CopyrightId int   `json:"copyrightId"`
			Status      int   `json:"status"`
			PicId       int64 `json:"picId"`
			Mark        int   `json:"mark"`
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
	Playlists []struct {
		Id            int64       `json:"id"`
		Name          string      `json:"name"`
		CoverImgUrl   string      `json:"coverImgUrl"`
		Creator       interface{} `json:"creator"`
		Subscribed    bool        `json:"subscribed"`
		TrackCount    int         `json:"trackCount"`
		UserId        int         `json:"userId"`
		PlayCount     int         `json:"playCount"`
		BookCount     int         `json:"bookCount"`
		SpecialType   int         `json:"specialType"`
		OfficialTags  interface{} `json:"officialTags"`
		Action        interface{} `json:"action"`
		ActionType    interface{} `json:"actionType"`
		RecommendText interface{} `json:"recommendText"`
		Score         interface{} `json:"score"`
		Description   string      `json:"description"`
		HighQuality   bool        `json:"highQuality"`
	} `json:"playlists"`
	Order []string `json:"order"`
}
type NetEasySearchSuggestArtist struct {
	Id        int         `json:"id"`
	Name      string      `json:"name"`
	PicUrl    string      `json:"picUrl"`
	Alias     []string    `json:"alias"`
	AlbumSize int         `json:"albumSize"`
	PicId     int64       `json:"picId"`
	FansGroup interface{} `json:"fansGroup"`
	Img1V1Url string      `json:"img1v1Url"`
	Img1V1    int64       `json:"img1v1"`
	Alia      []string    `json:"alia"`
	Trans     interface{} `json:"trans"`
}
type NetEasyGetPlayUrlResp struct {
	Code int                     `json:"code"`
	Data []NetEasyGetPlayUrlInfo `json:"data"`
}
type NetEasyGetPlayUrlInfo struct {
	Id                 int64       `json:"id"`
	Url                string      `json:"url"`
	Br                 int         `json:"br"`
	Size               int         `json:"size"`
	Md5                string      `json:"md5"`
	Code               int         `json:"code"`
	Expi               int         `json:"expi"`
	Type               string      `json:"type"`
	Gain               int         `json:"gain"`
	Peak               interface{} `json:"peak"`
	ClosedGain         int         `json:"closedGain"`
	ClosedPeak         int         `json:"closedPeak"`
	Fee                int         `json:"fee"`
	Uf                 interface{} `json:"uf"`
	Payed              int         `json:"payed"`
	Flag               int         `json:"flag"`
	CanExtend          bool        `json:"canExtend"`
	FreeTrialInfo      interface{} `json:"freeTrialInfo"`
	Level              string      `json:"level"`
	EncodeType         string      `json:"encodeType"`
	ChannelLayout      interface{} `json:"channelLayout"`
	FreeTrialPrivilege struct {
		ResConsumable      bool        `json:"resConsumable"`
		UserConsumable     bool        `json:"userConsumable"`
		ListenType         interface{} `json:"listenType"`
		CannotListenReason interface{} `json:"cannotListenReason"`
		PlayReason         interface{} `json:"playReason"`
		FreeLimitTagType   interface{} `json:"freeLimitTagType"`
	} `json:"freeTrialPrivilege"`
	FreeTimeTrialPrivilege struct {
		ResConsumable  bool `json:"resConsumable"`
		UserConsumable bool `json:"userConsumable"`
		Type           int  `json:"type"`
		RemainTime     int  `json:"remainTime"`
	} `json:"freeTimeTrialPrivilege"`
	UrlSource    int         `json:"urlSource"`
	RightSource  int         `json:"rightSource"`
	PodcastCtrp  interface{} `json:"podcastCtrp"`
	EffectTypes  interface{} `json:"effectTypes"`
	Time         int         `json:"time"`
	Message      interface{} `json:"message"`
	LevelConfuse interface{} `json:"levelConfuse"`
	MusicId      string      `json:"musicId"`
}

type NetEasySingerResp struct {
	Artists []NetEasyArtistInfo `json:"artists"`
	More    bool                `json:"more"`
	Code    int                 `json:"code"`
}

type NetEasyArtistInfo struct {
	AccountId   int64    `json:"accountId,omitempty"`
	AlbumSize   int      `json:"albumSize"`
	Alias       []string `json:"alias"`
	BriefDesc   string   `json:"briefDesc"`
	FansCount   int      `json:"fansCount"`
	Followed    bool     `json:"followed"`
	Id          int      `json:"id"`
	Img1V1Id    int64    `json:"img1v1Id"`
	Img1V1IdStr string   `json:"img1v1Id_str,omitempty"`
	Img1V1Url   string   `json:"img1v1Url"`
	MusicSize   int      `json:"musicSize"`
	Name        string   `json:"name"`
	PicId       int64    `json:"picId"`
	PicIdStr    string   `json:"picId_str,omitempty"`
	PicUrl      string   `json:"picUrl"`
	TopicPerson int      `json:"topicPerson"`
	Trans       string   `json:"trans"`
	TransNames  []string `json:"transNames,omitempty"`
}
