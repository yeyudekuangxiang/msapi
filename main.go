package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yeyudekuangxiang/msapi/search"
	"log"
	"net/http"
)

var port = flag.String("port", "3000", "http server port")
var cacheDir = flag.String("cacheDir", "./cache/", "http server cache dir")
var email = flag.String("email", "", "email address")
var password = flag.String("password", "", "password")
var netEasyDomain = flag.String("domain", "http://nas.example:3000", "http server netEasyDomain")
var auth = flag.String("auth", "", "http server auth")

func main() {
	flag.Parse()
	r := gin.New()
	netEasy := search.NetEasyAPi{
		Domain: *netEasyDomain,
		Cache: &search.LocalCache{
			Dir: *cacheDir,
		},
	}
	err := netEasy.EmailLogin(*email, *password)
	if err != nil {
		log.Panicln(err)
	}
	r.GET("/lyrics", func(c *gin.Context) {
		if c.GetHeader("Authorization") != *auth {
			c.Status(401)
			return
		}
		type param struct {
			Title    string `form:"title"`
			Artist   string `form:"artist"`
			Path     string `form:"path"`
			Album    string `form:"album"`
			Duration int    `form:"duration"`
			Offset   int    `form:"offset"`
			Limit    int    `form:"limit"`
		}
		p := param{}
		err := c.ShouldBind(&p)
		if err != nil {
			c.Status(404)
			return
		}
		lrcData, err := netEasy.SearchLrcBest(search.SearchLrcParam{
			Title:    p.Title,
			Artist:   p.Artist,
			Path:     p.Path,
			Album:    p.Album,
			Duration: p.Duration,
			Offset:   p.Offset,
			Limit:    p.Limit,
		})
		if err != nil {
			c.Status(404)
			return
		}

		cacheKey := Md5(fmt.Sprintf("lrc%s%s", p.Title, p.Artist))
		c.Header("Cache-Control", "public, max-age=3600")
		c.Header("ETag", cacheKey)
		c.Data(200, "text/plain", lrcData)
	})
	r.POST("/lyrics/confirm", func(c *gin.Context) {
		if c.GetHeader("Authorization") != *auth {
			c.Status(401)
			return
		}
		c.Status(200)
		return
	})
	r.GET("/covers", func(c *gin.Context) {
		if c.GetHeader("Authorization") != *auth {
			c.Status(401)
			return
		}
		type param struct {
			Title  string `form:"title"`
			Artist string `form:"artist"`
			Album  string `form:"album"`
		}
		p := param{}
		err := c.ShouldBind(&p)
		if err != nil {
			c.Status(404)
			return
		}
		cacheKey := Md5(fmt.Sprintf("cover%s%s%s", p.Title, p.Album, p.Artist))
		c.Header("Cache-Control", "public, max-age=3600")
		c.Header("ETag", cacheKey)
		if p.Title != "" && p.Album != "" && p.Artist != "" {
			data, err := netEasy.GetMusicCover(search.GetMusicCoverParam{
				Title:  p.Title,
				Artist: p.Artist,
				Album:  p.Album,
			})
			if err != nil {
				c.Status(404)
				return
			}
			c.Data(200, "application/jpeg", data)
			return
		}
		if p.Album != "" && p.Artist != "" {
			data, err := netEasy.GetAlbumCover(search.GetAlbumCoverParam{
				Artist: p.Artist,
				Album:  p.Album,
			})
			if err != nil {
				c.Status(404)
				return
			}
			c.Data(200, "application/jpeg", data)
			return
		}
		if p.Artist != "" {
			data, err := netEasy.GetSingerCover(search.GetSingerCoverParam{
				Artist: p.Artist,
			})
			if err != nil {
				c.Status(404)
				return
			}
			c.Data(200, "application/jpeg", data)
			return
		}
		c.Status(404)
		return
	})
	http.ListenAndServe(":"+*port, r)
}

func Md5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
