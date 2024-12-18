package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yeyudekuangxiang/msapi/pkg/copy"
	"github.com/yeyudekuangxiang/msapi/search/freemp3"
	"github.com/yeyudekuangxiang/msapi/search/inter"
	"github.com/yeyudekuangxiang/msapi/search/neteasy"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var netEasyClient *neteasy.Client
var freemp3Client *freemp3.Client
var copyClient = copy.NewLocalCopy("./music")
var server = gin.New()

func init() {
	netEasyClient = neteasy.NewClient(os.Getenv("neteasydomain"))
	netEasyClient.SetCookie(os.Getenv("neteasycookie"))
	freemp3Client = freemp3.NewClient(os.Getenv("freemp3domain"))
	freemp3Header := os.Getenv("freemp3header")
	header := make(map[string]string)
	err := json.Unmarshal([]byte(freemp3Header), &header)
	if err != nil {
		log.Panic("解析header失败", err)
	}
	freemp3Client.SetHeader(header)
}

// CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
func Run(closeChan chan struct{}) {

	server.Use(CORSMiddleware())
	server.Use(gin.Recovery())
	server.GET("/api/music/search", func(c *gin.Context) {
		searchReq := struct {
			Keyword string `form:"keyword" json:"keyword" binding:"required"`
			Page    int64  `form:"page" json:"page" binding:"required"`
			Size    int64  `form:"size" json:"size" binding:"required"`
			Site    string `form:"site" json:"site" binding:"required"`
		}{}

		err := c.ShouldBindQuery(&searchReq)
		if err != nil {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
			return
		}
		var list interface{}
		switch searchReq.Site {
		case "163":
			list, err = netEasyClient.SearchMusic(searchReq.Keyword, searchReq.Size, (searchReq.Page-1)*searchReq.Size)
		case "freemp3":
			list, err = freemp3Client.SearchMusic(searchReq.Keyword, searchReq.Size, (searchReq.Page-1)*searchReq.Size)
		default:
			err = errors.New("未知得平台")
		}

		if err != nil {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{
			"code": 200,
			"data": gin.H{"list": list},
		})
	})
	server.GET("/api/music/play", func(c *gin.Context) {
		id := c.Query("id")
		site := c.Query("site")
		quailty := c.Query("quality")
		var playList []inter.PlayInfo
		var err error
		switch site {
		case "163":
			playList, err = netEasyClient.GetPlayUrl([]string{id})
		case "freemp3":
			playList, err = freemp3Client.GetPlayUrl([]string{id}, freemp3.QualityOption(quailty))
		default:
			err = errors.New("未知得平台")
		}
		if err != nil {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
			return
		}
		if len(playList) == 0 {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  "未获取到播放连接",
			})
			return
		}
		info := playList[0]
		t := ""
		if strings.Contains(info.Url, ".ogg") {
			t = "audio/ogg"
		} else if strings.Contains(info.Url, ".mp3") {
			t = "audio/mp3"
		} else if strings.Contains(info.Url, ".flac") {
			t = "audio/flac"
		}
		c.JSON(200, gin.H{
			"code": 400,
			"data": gin.H{
				"song": gin.H{
					"sources": []gin.H{
						{
							"src":  info.Url,
							"type": t,
						},
					},
				},
			},
		})
		return
	})
	server.GET("/api/music/sync", func(c *gin.Context) {
		id := c.Query("id")
		site := c.Query("site")
		quality := c.Query("quality")
		singerName := c.Query("singerName")
		musicName := c.Query("musicName")

		var playList []inter.PlayInfo
		var err error
		switch site {
		case "163":
			playList, err = netEasyClient.GetPlayUrl([]string{id})
		case "freemp3":
			playList, err = freemp3Client.GetPlayUrl([]string{id}, freemp3.QualityOption(quality))
		default:
			err = errors.New("未知得平台")
		}
		if err != nil {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
			return
		}
		if len(playList) == 0 {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  "未获取到播放连接",
			})
			return
		}
		info := playList[0]
		buf, name, err := autoDown(singerName, musicName, info.Url)
		if err != nil {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
			return
		}
		singerNames := strings.Split(singerName, ",")
		err = copyClient.CopyReader(buf, path.Join(singerNames[0], name))
		if err != nil {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{
			"code": 200,
			"data": gin.H{},
		})
		return
	})
	htpServer := &http.Server{Addr: ":8080", Handler: server}
	go func() {
		err := htpServer.ListenAndServe()
		if err != nil {
			log.Println("http启动失败", err)
		}
	}()
	go func() {
		for {
			select {
			case <-closeChan:
				log.Println("10秒后关闭连接")
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				if err := htpServer.Shutdown(ctx); err != nil {
					fmt.Printf("Server Shutdown Failed:%+v", err)
				}
				cancel()
				return
			}
		}
	}()
}
func autoDown(singerName, musicName, u string) (*bytes.Reader, string, error) {
	log.Println("downuuuuu", u)
	// 发送GET请求
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := http.DefaultClient.Do(req)
	//log.Println(resp, err)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var fileName string
	ct := resp.Header.Get("Content-Type")

	if strings.Contains(ct, "audio/mpeg") {
		fileName = fmt.Sprintf("%s - %s.mp3", musicName, singerName)
	} else if strings.Contains(ct, "audio/wav") {
		fileName = fmt.Sprintf("%s - %s.wav", musicName, singerName)
	} else if strings.Contains(ct, "audio/ogg") || strings.Contains(ct, "audio/x-ogg") {
		fileName = fmt.Sprintf("%s - %s.ogg", musicName, singerName)
	} else if strings.Contains(ct, "audio/acc") {
		fileName = fmt.Sprintf("%s - %s.acc", musicName, singerName)
	} else if strings.Contains(ct, "audio/flac") || strings.Contains(ct, "audio/x-flac") {
		fileName = fmt.Sprintf("%s - %s.flac", musicName, singerName)
	} else {
		//log.Println("未知的音频格式", u, musicName, ct)
	}

	// 检查Content-Disposition头以获取文件名
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" && fileName == "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil {
			fileExt := path.Ext(params["filename"])
			if fileExt != "" {
				fileName = fmt.Sprintf("%s - %s%s", musicName, singerName, fileExt)
			} else if params["filename"] != "" {
				fileName = params["filename"]
			}
		}
	}

	if fileName == "" && strings.Contains(resp.Request.URL.Path, ".") {
		fileExt := path.Ext(resp.Request.URL.Path)
		if fileExt != "" {
			fileName = fmt.Sprintf("%s - %s%s", musicName, singerName, fileExt)
		}
	}

	if fileName == "" {
		fileName = fmt.Sprintf("%s - %s", musicName, singerName)
	}
	fileName = ReplaceFileName(fileName)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	if bytes.Contains(body, []byte("验证")) {
		return nil, "", errors.New("安全验证")
	}
	return bytes.NewReader(body), fileName, nil
}
func getSingerName(singerStr string) (string, string) {
	m := make([]map[string]interface{}, 0)
	err := json.Unmarshal([]byte(singerStr), &m)
	if err != nil || len(m) == 0 {
		log.Println("解析歌手失败", singerStr, err)
		return "未知", "未知"
	}

	realSingerName := ""
	for _, ar := range m {
		realSingerName += "," + strings.TrimSpace(ar["name"].(string))
		if len(realSingerName) > 150 {
			log.Println("歌手名字太长截断部分歌手", singerStr)
			break
		}
	}
	if len(realSingerName) > 0 {
		realSingerName = realSingerName[1:]
	}
	name := m[0]["name"].(string)
	if name == "" {
		return "未知", "未知"
	}
	return strings.TrimSpace(name), realSingerName
}
func ReplaceFileName(filename string) string {
	filename = strings.ReplaceAll(filename, "/", ",")
	m := map[string]string{
		"\\": ",",
		"/":  ",",
		":":  " ",
		"*":  " ",
		"?":  " ",
		"\"": " ",
		"<":  " ",
		">":  " ",
		"|":  " ",
	}
	for oldStr, newStr := range m {
		filename = strings.ReplaceAll(filename, oldStr, newStr)
	}
	return filename
}
func ListenFile(system http.FileSystem) {
	web := server.Group("/web")
	web.StaticFS("/", system)
}
