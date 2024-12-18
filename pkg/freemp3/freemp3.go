package freemp3

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Domain string
	Header map[string]string
}

func (c *Client) encode(d interface{}) (string, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	basedata := base64.StdEncoding.EncodeToString(data)
	resp, err := http.Get(c.Domain + "/encode?" + basedata)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
func (c *Client) GetArtistList(page int) (*ArtistListResponse, error) {
	reqBodyObj := ArtistListReq{
		ArtistListReqToken: ArtistListReqToken{
			Initial: 0,
			Page:    page,
			T:       time.Now().UnixMilli(),
		},
	}
	token, err := c.encode(reqBodyObj.ArtistListReqToken)
	if err != nil {
		return nil, err
	}
	reqBodyObj.Token = token
	reqBody, err := json.Marshal(reqBodyObj)
	if err != nil {
		return nil, err
	}

	log.Println("GetArtistList", string(reqBody))
	req, err := http.NewRequest("POST", "https://api.liumingye.cn/m/api/artist/list", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	c.fillReq(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	v := ArtistListResponse{}
	return &v, json.Unmarshal(respBody, &v)
}
func (c *Client) Search(title string, page int) (*SearchResp, error) {
	reqBodyObj := SearchReq{
		SearchReqToken: SearchReqToken{
			Type: "YQM",
			Text: title,
			Page: page,
			V:    "beta",
			T:    time.Now().UnixMilli(),
		},
	}
	token, err := c.encode(reqBodyObj.SearchReqToken)
	if err != nil {
		return nil, err
	}
	reqBodyObj.Token = token
	reqBody, err := json.Marshal(reqBodyObj)
	if err != nil {
		return nil, err
	}
	log.Println("Search", string(reqBody))

	req, err := http.NewRequest("POST", "https://api.liumingye.cn/m/api/search", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	c.fillReq(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	v := SearchResp{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		log.Println(string(body))
		return nil, err
	}
	return &v, nil
}
func (c *Client) GetRealDownLoadUrl(id string, quality string) (string, error) {
	t := strconv.FormatInt(time.Now().UnixMilli(), 10)
	token, err := c.encode(DownLoadUrlReq{
		Id:      id,
		Quality: quality,
		T:       t,
	})
	if err != nil {
		return "", err
	}

	uuu := fmt.Sprintf("https://api.liumingye.cn/m/api/link?id=%s&quality=%s&_t=%s&token=%s", id, quality, t, token)
	log.Println("网盘跳转页", uuu)
	req, err := http.NewRequest("GET", uuu, nil)
	if err != nil {
		return "", err
	}
	c.fillReq(req)
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Del("Content-Type")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if strings.Contains(resp.Header.Get("Content-Type"), "audio") {
		if resp.StatusCode == http.StatusFound {
			location := resp.Header.Get("Location")
			if location != "" {
				return location, nil
			}
		}
		return resp.Request.URL.String(), nil
		//return uuu, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if len(respBody) == 0 {
		return "", fmt.Errorf("body长度为0 %d", resp.StatusCode)
	}
	return getLanRealDownFromBody(respBody)
}

type SearchReq struct {
	SearchReqToken
	Token string `json:"token"`
}
type SearchReqToken struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Page int    `json:"page"`
	V    string `json:"v"`
	T    int64  `json:"_t"`
}
type SearchResp struct {
	Code int `json:"code"`
	Data struct {
		List []struct {
			Id      string        `json:"id"`
			Lyric   string        `json:"lyric"`
			Name    string        `json:"name"`
			Time    int           `json:"time,omitempty"`
			Quality []interface{} `json:"quality"`
			Album   struct {
				Id   string `json:"id"`
				Name string `json:"name"`
				Pic  string `json:"pic"`
			} `json:"album,omitempty"`
			Artist []struct {
				Id   string `json:"id"`
				Name string `json:"name"`
			} `json:"artist"`
			Hash string `json:"hash,omitempty"`
			Pic  string `json:"pic,omitempty"`
		} `json:"list"`
		Total interface{} `json:"total"`
		Word  []string    `json:"word"`
	} `json:"data"`
	Msg string `json:"msg"`
}
type ArtistListResponse struct {
	Code int `json:"code"`
	Data struct {
		List []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
			Pic  string `json:"pic"`
		} `json:"list"`
	} `json:"data"`
	Msg string `json:"msg"`
}
type ArtistDetailResponse struct {
	Code int `json:"code"`
	Data struct {
		Pic        string `json:"pic"`
		Name       string `json:"name"`
		Views      int    `json:"views"`
		UpdateTime string `json:"update_time"`
		List       []struct {
			Id      string        `json:"id"`
			Name    string        `json:"name"`
			Pic     *string       `json:"pic"`
			Url     *string       `json:"url"`
			Time    int           `json:"time"`
			Lyric   *string       `json:"lyric"`
			Quality []interface{} `json:"quality"`
			Album   *struct {
				Id   string `json:"id"`
				Name string `json:"name"`
				Pic  string `json:"pic"`
			} `json:"album"`
			Artist []struct {
				Id   string `json:"id"`
				Name string `json:"name"`
			} `json:"artist"`
			Pivot struct {
				ArtistId int `json:"artist_id"`
				TrackId  int `json:"track_id"`
				Sort     int `json:"sort"`
			} `json:"pivot"`
			Hash string `json:"hash"`
		} `json:"list"`
		Desc string `json:"desc"`
	} `json:"data"`
	Msg string `json:"msg"`
}
type ArtistListReqToken struct {
	Initial int   `json:"initial"`
	Page    int   `json:"page"`
	T       int64 `json:"_t"`
}
type ArtistListReq struct {
	ArtistListReqToken
	Token string `json:"token"`
}

func (c *Client) fillReq(req *http.Request) {
	for k, v := range c.Header {
		req.Header.Add(k, v)
	}
}

type DownAjaxResp struct {
	Zt  int    `json:"zt"`
	Dom string `json:"dom"`
	Url string `json:"url"`
	Inf int    `json:"inf"`
}
type DownLoadUrlReq struct {
	Id      string `json:"id"`
	Quality string `json:"quality"`
	T       string `json:"_t"`
}

func getLanRealDownFromBody(respBody []byte) (string, error) {

	if len(respBody) == 0 {
		return "", errors.New("body长度为0")
	}
	//log.Println(string(respBody))

	srcReg, err := regexp.Compile(`iframe.*src="(.*?)".*?iframe`)
	if err != nil {
		return "", err
	}
	list := srcReg.FindStringSubmatch(string(respBody))
	if len(list) != 2 {
		return "", errors.New("没有匹配地址")
	}

	log.Println("网盘下载页", fmt.Sprintf("https://m.lanzouy.com/%s", list[1]))

	downResp, err := http.Get(fmt.Sprintf("https://m.lanzouy.com/%s", list[1]))
	if err != nil {
		return "", err
	}
	defer downResp.Body.Close()
	downBody, err := io.ReadAll(downResp.Body)
	if err != nil {
		return "", err
	}
	//log.Println(string(downBody))

	uReg, err := regexp.Compile(`(/ajaxm.php.*?)'`)
	if err != nil {
		return "", err
	}
	list = uReg.FindStringSubmatch(string(downBody))
	if len(list) != 2 {
		return "", errors.New("未查到ajaxm")
	}
	ajaxUrl := list[1]
	dataReg, err := regexp.Compile(`data.*?:(.*?\})`)
	if err != nil {
		return "", err
	}
	list = dataReg.FindStringSubmatch(string(downBody))
	if len(list) != 2 {
		return "", errors.New("未查到ajaxm参数")
	}
	ajaxBody := list[1]
	ajaxBody = strings.ReplaceAll(ajaxBody, "ajaxdata", `'?ctdf'`)
	ajaxBody = strings.ReplaceAll(ajaxBody, "ciucjdsdc", `''`)
	ajaxBody = strings.ReplaceAll(ajaxBody, "aihidcms", `'7Sij'`)
	ajaxBody = strings.ReplaceAll(ajaxBody, "kdns", `1`)
	ajaxBody = strings.ReplaceAll(ajaxBody, `'`, `"`)
	//log.Println(ajaxUrl, ajaxBody)
	return downAjax(ajaxUrl, ajaxBody)

}
func downAjax(path string, data string) (string, error) {
	uuu := "https://m.lanzouy.com" + path
	method := "POST"

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &m)
	if err != nil {
		return "", err
	}
	uv := url.Values{}
	for k, v := range m {
		switch vv := v.(type) {
		case int64:
			uv.Add(k, strconv.FormatInt(vv, 10))
		case string:
			uv.Add(k, vv)
		case float64:
			uv.Add(k, strconv.FormatInt(int64(vv), 10))
		}
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, uuu, strings.NewReader(uv.Encode()))

	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", "codelen=1; pc_ad1=1; Hm_lvt_fb7e760e987871d56396999d288238a4=1731484870; Hm_lpvt_fb7e760e987871d56396999d288238a4=1731484870; HMACCOUNT=EC959E5081FF80F7; uz_distinctid=193248a695398-0d9140bb03de21-26011951-130980-193248a69542e7; STDATA82=czst_eid%3D275618800-3821-%26ntime%3D3821; codelen=1; pc_ad1=1")
	req.Header.Add("Origin", "https://m.lanzouy.com")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Referer", "https://m.lanzouy.com/fn?A2VUPg5rAm9UMQNgBmQCMFY0ATxeJ1AmCjBRZlI5U2EHNlY1Cm8EZQZkUDAKZwcgV3oEZFVoAXAAblAxATNUPgNmVHoObgJhVFEDPAY4")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("sec-ch-ua", "\"Chromium\";v=\"130\", \"Google Chrome\";v=\"130\", \"Not?A_Brand\";v=\"99\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")

	//log.Println("请求ajax", req)
	res, err := client.Do(req)
	//fmt.Println("ajax相应", res, err)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", errors.New(res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	vvv := DownAjaxResp{}
	err = json.Unmarshal(body, &vvv)
	if err != nil {
		return "", err
	}
	if vvv.Zt == 0 {
		return "", fmt.Errorf("%d", vvv.Zt)
	}
	//https://down-load.lanrar.com/file/?BWMBP1tqUmMIAQE5U2ZdMVplVGwFvQaMVuAH5Fe7UOAC5QbZW7cPvwPvUYECuQK/UoBTsFGEApYH6QafVLtWsAWGAftb4FKrCNIBuFO7XdVaLVS3BdAGkla4B55XuFC2AoEG/FvgD/YD2VEpAjACb1JpUzVRKAJmB2kGbFRtVggFbAEyWztSPghnAWRTMl1qWjNUaQViBiJWNwdyVzZQYgIxBmhbNg9vA2FRNAJnAiVSeFMmUTMCMgcwBjJUOlZ4BTQBZ1spUjcIZgF4Uz9dPlpjVDcFYgYwVmMHM1dtUGUCOwYwWzMPbAMwUTICYgI3UjFTZVFoAjAHNgZkVGlWYwVkAWVbMVI1CDsBYlMpXTpabFQwBTkGIlYkB3JXblAjAmsGNVs7D2MDY1E1AmECM1I9U3BRegJpB20GZVRuVmoFNAFhWzVSNwhqAW9TMV1kWjNUYAV0BipWdwdnV2dQJgI/BmBbMQ9oA2RRMwJuAjdSP1NuUTsCJgd1BnBUf1ZqBTQBYFswUj4IaQFnUzVdbFo3VGYFfAZxVjgHcVc2UGACMgZlWygPagNjUT8CeAI2UjFTeFE9AjUHLgYmVGxWOAVyAThbWVJlCDUBalM3
	//"<a href="+dom_down+"/file/"+ date.url + lanosso +" target=_blank rel=noreferrer//><span class=txt>电信下载</span><span class='txt txtc'>联通下载</span><span class=txt>普通下载</span></a>
	return fmt.Sprintf("https://down-load.lanrar.com/file/?%s", vvv.Url), nil
}
