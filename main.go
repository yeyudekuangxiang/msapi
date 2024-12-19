package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/yeyudekuangxiang/common-go/db"
	"github.com/yeyudekuangxiang/msapi/pkg/neteasy"
	"github.com/yeyudekuangxiang/woozooo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	_ "image/png"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var downPath = flag.String("down", "./music", "")
var num = flag.Int("num", 2, "")
var mode = flag.String("mode", "", "")
var domain = flag.String("domain", "http://127.0.0.1:3000", "")
var runHttp = flag.Bool("http", false, "")
var host = flag.String("db", "nas.znil.cn", "")
var closeCh = make(chan struct{})

func main() {
	flag.Parse()
	// 创建一个通道来接收信号
	sigChan := make(chan os.Signal, 1)
	// 注册要监听的信号
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		Close()
	}()

	if *runHttp {
		go func() {
			log.Println("开始启动http")
			cmd := exec.Command("node", "./neteasecloudmusicapi/app.js")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			go func() {
				<-closeCh
				if cmd.Process != nil {
					log.Println("退出http")
					cmd.Process.Kill()
				}
			}()
			err := cmd.Run()
			if err != nil {
				log.Println("http运行错误", err)
			}
			log.Println("http已退出")
		}()
	}

	log.Println("5秒后开始抓取")
	time.Sleep(time.Second * 5)
	netEasy := neteasy.APi{
		Domain: *domain,
	}

	linkDb, err := db.NewMysqlDB(db.Config{
		Type:         "mysql",
		Host:         *host,
		UserName:     "jzl",
		Password:     "ZHUImeng521..",
		Database:     "freemusic",
		Port:         3306,
		TablePrefix:  "",
		MaxOpenConns: 100,
		MaxIdleConns: 50,
		MaxIdleTime:  600,
		MaxLifetime:  7200,
		Logger: logger.New(log.New(os.Stdout, "", 0), logger.Config{
			LogLevel: logger.Info,
		}),
	})
	if err != nil {
		CloseWithErr("连接数据库失败", err)
	}

	netEasy.SetCookie(os.Getenv("msapicookie"))
	//netEasy.SetCookie("nts_mail_user=zhanling_jin@163.com:-1:1; NTES_P_UTID=BunwI3FTXpsHF9qsTh7EUoIINSJ9uTOI|1732084109; NMTID=00OTl8UxJuYSkqoVEuwgtpTuEydk80AAAGTqiGvDQ; _ntes_nnid=9435e3533b3698c1f2050a4b37f60e3e,1733726154923; _ntes_nuid=9435e3533b3698c1f2050a4b37f60e3e; WM_NI=r7RMEk6NFGHSi4whdlnzNLLPL4ztbasZ0qcCr%2BE60KRl%2FANQCTRqRdciRSf6EltrM3p26B4DryDmcx8QWGdWACiVbzUCVdewkV%2B9ahJuNxWXhwprywldthgPAjW5PtAeYUc%3D; WM_NIKE=9ca17ae2e6ffcda170e2e6eed4b648a891ae83ea45b2e78fa3d55e879b8eb1d65db4b7adb8c148a29cf7acf82af0fea7c3b92ab2e99eaec861f6928aa8cb59abecb989e768aea682adcf5c90ecb695e16bb1860092e249b2f58bd7b34192b4e1a2cd5b818c9e92d170a1b79eb4e53e87b4bc93d039b4eefaa3b750b2a78c8ee274fc8f99a8e75d839da7b7e172979d9a85e87da98ae5dae4648aab8299d654aca9c093fc7db0bcbc89e65c83b699d6d38083bb97b6e237e2a3; WM_TID=TFlb%2FC9q3D5BRFFEBFbDGojJXQaaHsf%2B; sDeviceId=YD-5VAjuoAE%2FQpAAkUFVVbTWoiZWALkFOlM; __snaker__id=IaE2dtnH0IcuvbOv; ntes_kaola_ad=1; P_INFO=17624865520|1733726348|1|music|00&99|null&null&null#shh&null#10#0|&0|null|17624865520; ntes_utid=tid._.0YcyzRv8%252BVNFElQVAAfXWtncHFfkEawV._.0; JSESSIONID-WYYY=hT9v8lbIX%2FlAuWlIah0ZHd0qVvPPKlVBiKJ9BqMWpNuvJeD9hXRKWPdi%2BtWpUZb4hY3XPJ1Qnrz%2FaNd7n079Q1gOTqPtUI%2BK%5CE5Y%2FHhbtNm6AWtemoxW%5CZKBwa0NjuMI%2F6D%2BrQZfojrTGq3O5TQHqm2KxuovXtt%2BNu4OaYV3ZjsxiHqc%3A1733797640870; _iuqxldmzr_=32; gdxidpyhxdE=GMxJVCA3hUA9%2BtSlAnGUXGiRYDbERvrxdz5dKvs7ijduRDzV5f2geiUTJCQbhKALXXvRawOgGe%2B%2FN6%5CbZopIrmW7z33uETdRV%2FH8U9gZ5Xo4SqDwAugrYiKmNjd8pHPh%2B%2FiXv8YkgGY58Wg0poRwATqN5HYkXBqtWyd9CTfqg%2B9ACewn%3A1733796747025; __csrf=115b9a46ee9ab140c941456ea9150995; __remember_me=true; MUSIC_U=00922A012385F1A674B00DB8ADD996E03F74E823B19AFA47FB91DDF0A937C097F71A9B0C2945CA78BCDA77937CE9CF387C75A7E3C1705A1BAD82B744AD84DABCB312F6047F83ECCA0668FD9D189A1E9585CC8F77017AAD6DE87D71EA592727617CBC5FC11BFE786EE4CA05C6E6E99066D9A6A20A592EDEFA2B756E8FDB5FC760392DAE2BB22FF9CDBE4DAC196BE16A72549A5190D12B2938075B5DD4E8C3CD6D440E6A6519D26C36DC119C7E46971AE0AE4C3DBA6258CF8806DD9CCEAC8767CD2F31BE1A712A4E0E10FEA72C6C68E547EE3DE3AE202FDCCAC6FDA9C97E7EFD5CE52B6920920B4F3B16F464E8BF1AFEA17A34F1630BFD7FA9EB9C6BF04E0BE6BBE19D02CD23E7781A274DAF545903D7A0F49287CC68B4EB320F64A1A1C34B5DBFFFBBFF40F7E92ADACBE000BAE8BFE6862A1A1DF4BA9D66641F377F5F83114D2D5350C876396C60CDA9FAAEAFF9489D3B4A0DDBDCF5F3D2035C0A46A0664C9EDED5")
	for i := 0; i <= 20; i++ {
		if i == 10 {
			log.Println("http服务未启动,停止抓取")
			os.Exit(1)
		}
		resp, err := http.Get(*domain)
		if err != nil {
			log.Println("检测失败，5秒后重试", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				log.Println("http服务已启动")
				break
			} else {
				log.Println("http服务未启动,5秒后重试", resp.StatusCode)
			}
		}
		time.Sleep(5 * time.Second)
	}
	switch *mode {
	case "singer":
		saveAllSinger(netEasy, linkDb)
	case "musichot":
		saveAllHotMusic(netEasy, linkDb)
	case "music":
		saveAllNormalMusic(netEasy, linkDb)
	case "downhot":
		downHotArtistMusic(netEasy, linkDb)
	case "down":
		downNormalArtistMusic(netEasy, linkDb)
	}
	Close()
}

func Md5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
func down(u string) error {
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("下载状态码异常" + resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	os.WriteFile("12312.mp3", body, os.ModePerm)
	return nil
}

type Artist struct {
	ID       int64
	ArtistId string
	Name     string
	Pic      string
	IsFetch  int
	Site     string
	Sort     int64
}

func (a Artist) TableName() string {
	return "artist"
}

type Music struct {
	ID      int64
	MusicId string
	Name    string
	Pic     string
	Lyric   string
	Artist  string
	Album   string
	Time    int
	Quality string
	DownUrl string
	IsDown  int
	Path    string
	Site    string
	Sort    int64
}

func saveAllSinger(netEasy neteasy.APi, linkDb *gorm.DB) {
	// 暂时屏蔽入驻歌手
	sort := int64(0)
	okMap := make(map[int64]bool)
	for i := 1; i < 100; i++ {
		for _, tp := range []int64{1, 2, 3} {
			if okMap[tp] {
				continue
			}

			log.Printf("获取%d分类歌手列表,code:%s,page:%d", tp, "", i)
			singers, err := netEasy.GetSingerList(tp, -1, "", 100, (i-1)*100)
			if err != nil {
				log.Println("获取歌手列表失败", i, tp, err)
				continue
			}
			if len(singers) == 0 {
				log.Println("该分类歌手搜索完毕", tp)
				okMap[tp] = true
				continue
			}
			log.Println("获取到歌手", len(singers))
			artists := make([]Artist, 0)
			for _, singer := range singers {
				sort++
				artists = append(artists, Artist{
					ArtistId: strconv.Itoa(singer.Id),
					Name:     singer.Name,
					Pic:      singer.PicUrl,
					IsFetch:  0,
					Site:     "163",
					Sort:     sort,
				})
			}
			err = linkDb.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(&artists).Error
			if err != nil {
				CloseWithErr("保存歌手信息失败", err)
			}
			log.Println("5秒后继续")
			time.Sleep(5 * time.Second)
		}
	}
}
func saveAllHotMusic(netEasy neteasy.APi, linkDb *gorm.DB) {
	artists := make([]Artist, 0)
	limit := 100
	err := linkDb.Where("site = '163' and is_fetch = 0 and sort < 9999999999").Order("sort asc").Find(&artists).Error
	if err != nil {
		log.Println("查询热门歌手失败", err)
		Close()
	}
	for _, artist := range artists {
		sort := 1000 * artist.Sort
		for i := 1; i < 5; i++ {
			log.Println("搜索歌手", artist.ID, artist.Name, i)
			searchResult, err := netEasy.SearchMusic(artist.Name, (i-1)*limit, limit)
			if err != nil {
				CloseWithErr("搜索歌手歌曲失败", artist, err)
			}
			if len(searchResult.Songs) == 0 {
				log.Println("歌手搜索完毕", artist.Name)
				log.Println("30秒后继续")
				time.Sleep(30 * time.Second)
				break
			}
			log.Println("搜索成功", artist.Name, len(searchResult.Songs))
			musicList := make([]Music, 0)
			for _, item := range searchResult.Songs {
				sort++
				artistStr, _ := json.Marshal(item.Artists)
				musicList = append(musicList, Music{
					MusicId: strconv.FormatInt(item.Id, 10),
					Name:    item.Name,
					Pic:     "",
					Lyric:   "",
					Artist:  string(artistStr),
					Album:   "",
					Time:    0,
					Quality: "",
					DownUrl: "",
					IsDown:  0,
					Path:    "",
					Site:    "163",
					Sort:    sort,
				})
			}
			err = linkDb.Clauses(clause.OnConflict{DoNothing: true}).Create(&musicList).Error
			if err != nil {
				log.Println("保存歌曲信息失败", artist, i, err)
			}
			if len(searchResult.Songs) < limit {
				log.Println("歌手搜索完毕", artist.Name)
				log.Println("30秒后继续")
				time.Sleep(30 * time.Second)
				break
			}
			log.Println("30秒后继续")
			time.Sleep(30 * time.Second)
		}
		artist.IsFetch = 1
		err := linkDb.Save(&artist).Error
		if err != nil {
			log.Println("保存搜索状态失败", artist, err)
		}
	}
}
func saveAllNormalMusic(netEasy neteasy.APi, linkDb *gorm.DB) {
	artists := make([]Artist, 0)
	limit := 100
	err := linkDb.Where("site = '163' and is_fetch = 0 and sort = 9999999999").Find(&artists).Error
	if err != nil {
		log.Println("查询热门歌手失败", err)
		Close()
	}

	for _, artist := range artists {
		for i := 1; i < 5; i++ {
			log.Println("搜索歌手", artist.ID, artist.Name, i)
			searchResult, err := netEasy.SearchMusic(artist.Name, (i-1)*limit, limit)
			if err != nil {
				CloseWithErr("搜索歌手歌曲失败", artist, err)
			}
			if len(searchResult.Songs) == 0 {
				log.Println("歌手搜索完毕", artist.Name)
				log.Println("30秒后继续")
				time.Sleep(30 * time.Second)
				break
			}
			log.Println("搜索成功", artist.Name, len(searchResult.Songs))
			musicList := make([]Music, 0)
			for _, item := range searchResult.Songs {
				artistStr, _ := json.Marshal(item.Artists)
				musicList = append(musicList, Music{
					MusicId: strconv.FormatInt(item.Id, 10),
					Name:    item.Name,
					Pic:     "",
					Lyric:   "",
					Artist:  string(artistStr),
					Album:   "",
					Time:    0,
					Quality: "",
					DownUrl: "",
					IsDown:  0,
					Path:    "",
					Site:    "163",
					Sort:    9999999999,
				})
			}
			err = linkDb.Clauses(clause.OnConflict{DoNothing: true}).Create(&musicList).Error
			if err != nil {
				CloseWithErr("保存歌曲信息失败", artist, i, err)
			}
			if len(searchResult.Songs) < limit {
				log.Println("歌手搜索完毕", artist.Name)
				log.Println("30秒后继续")
				time.Sleep(30 * time.Second)
				break
			}
			log.Println("30秒后继续")
			time.Sleep(30 * time.Second)
		}
		artist.IsFetch = 1
		err := linkDb.Save(&artist).Error
		if err != nil {
			log.Println("保存搜索状态失败", artist, err)
		}
	}
}
func Close() {
	close(closeCh)
	log.Println("10秒后退出")
	time.Sleep(time.Second * 10)
	os.Exit(0)
}
func CloseWithErr(v ...any) {
	log.Println(v...)
	Close()
}

var dirLock sync.Mutex

func downHotArtistMusic(netEasy neteasy.APi, linkDb *gorm.DB) {
	dirResp, err := lanClient.ReadSubDir(woozooo.ReadSubDirReq{
		DirId: -1,
	})
	if err != nil {
		log.Println("获取网盘文件夹失败", err)
		return
	}
	if dirResp.Zt != 1 {
		log.Println("获取网盘文件夹失败", dirResp)
		return
	}
	dirMap := make(map[string]int64)
	for _, item := range dirResp.Text {
		dirId, err := strconv.ParseInt(item.FolId, 10, 64)
		if err != nil {
			log.Println("获取网盘文件夹失败", err)
			return
		}
		dirMap[item.Name] = dirId
	}

	c := make(chan int, *num)
	for {
		list := make([]Music, 0)
		err := linkDb.Where("site = '163' and is_down = 0 and sort < 9999999999").Limit(100).Order("sort asc").Find(&list).Error
		if err != nil {
			CloseWithErr("查询热门歌曲失败", err)
		}
		if len(list) == 0 {
			log.Println("下载歌曲完毕")
			return
		}
		wg := sync.WaitGroup{}
		ids := make([]int64, 0)
		musicMap := make(map[int64]Music)
		for _, m := range list {
			musicId, err := strconv.ParseInt(m.MusicId, 10, 64)
			if err != nil {
				log.Println("转换musicid失败", m.ID, m.MusicId, err)
				m.IsDown = 3
				err = linkDb.Save(&m).Error
				if err != nil {
					log.Println("保存下载状态失败", m, err)
				}
				continue
			}
			ids = append(ids, musicId)
			musicMap[musicId] = m
		}
		if len(ids) == 0 {
			continue
		}
		songUrls, err := netEasy.GetPlayUrl(ids, 999000)
		if err != nil {
			CloseWithErr("获取播放连接失败", err)
		}
		for _, song := range songUrls {
			music := musicMap[song.Id]
			if song.Url == "" {
				log.Println("未获取到下载链接", song)
				music.IsDown = 2
				err = linkDb.Save(&music).Error
				if err != nil {
					log.Println("保存下载失败状态失败", music, err)
				}
				continue
			}
			c <- 1

			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					<-c
				}()

				dir, realSingerName := getSingerName(music.Artist)
				dirLock.Lock()
				dirId, ok := dirMap[dir]
				if !ok {
					dirResp, err := lanClient.Mkdir(woozooo.MkdirReq{
						ParentId:          -1,
						FolderName:        dir,
						FolderDescription: dir,
					})
					if err != nil {
						log.Println("创建文件夹失败", err)
						dirLock.Unlock()
						return
					}
					if dirResp.Zt != 1 {
						log.Println("创建文件夹失败", dirResp)
						dirLock.Unlock()
						return
					}
					dirId, err = strconv.ParseInt(dirResp.Text, 10, 64)
					if dirResp.Zt != 1 {
						log.Println("创建文件夹失败", err)
						dirLock.Unlock()
						return
					}
					dirMap[dir] = dirId
				}
				dirLock.Unlock()
				filePath, err := down2Woo(dirId, dir, realSingerName, music.Name, song.Url)
				if err != nil {
					log.Println("下载失败", music.ID, err)
					music.IsDown = 4
					err = linkDb.Save(&music).Error
					if err != nil {
						log.Println("保存下载状态失败", music, err)
					}
				} else {
					music.IsDown = 1
					music.Path = filePath
					music.DownUrl = song.Url
					err = linkDb.Save(&music).Error
					if err != nil {
						log.Println("保存下载状态失败", music, err)
					}
				}
				return
			}()
		}
		wg.Wait()
		log.Println("30秒后继续下载")
		time.Sleep(time.Second * 30)
	}
}
func downNormalArtistMusic(netEasy neteasy.APi, linkDb *gorm.DB) {
	musicList := make([]Music, 0)
	c := make(chan int, *num)
	linkDb.Where("site = '163' and is_down = 0 and sort = 9999999999").FindInBatches(&musicList, 50, func(tx *gorm.DB, batch int) error {
		wg := sync.WaitGroup{}
		ids := make([]int64, 0)
		musicMap := make(map[int64]Music)
		for _, m := range musicList {
			musicId, err := strconv.ParseInt(m.MusicId, 10, 64)
			if err != nil {
				log.Println("转换musicid失败", m.ID, m.MusicId, err)
				continue
			}
			ids = append(ids, musicId)
			musicMap[musicId] = m
		}
		if len(ids) == 0 {
			return nil
		}
		songUrls, err := netEasy.GetPlayUrl(ids, 999000)
		if err != nil {
			CloseWithErr(err)
		}
		for _, song := range songUrls {
			music := musicMap[song.Id]
			if song.Url == "" {
				log.Println("未获取到下载链接", song)
				music.IsDown = 2
				err = linkDb.Save(&music).Error
				if err != nil {
					log.Println("保存下载失败状态失败", music, err)
				}
				continue
			}
			c <- 1
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					<-c
				}()

				dir, realSingerName := getSingerName(music.Artist)
				filePath, err := autoDown(dir, realSingerName, music.Name, song.Url)
				if err != nil {
					log.Println("下载失败", music.ID, err)
				} else {
					music.IsDown = 1
					music.Path = filePath
					music.DownUrl = song.Url
					err = tx.Save(&music).Error
					if err != nil {
						log.Println("保存下载状态失败", music, err)
					}
				}
				return
			}()
		}
		wg.Wait()
		log.Println("30秒后继续下载")
		time.Sleep(time.Second * 30)
		return nil
	})
}
func autoDown(subDirName string, singerName, musicName, u string) (string, error) {
	log.Println("downuuuuu", u)
	// 发送GET请求
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	//log.Println(resp, err)
	if err != nil {
		return "", err
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
		return "", err
	}

	if bytes.Contains(body, []byte("验证")) {
		return "", errors.New("安全验证")
	}
	pathName := path.Join(*downPath, subDirName)
	fileName = path.Join(pathName, fileName)
	err = os.MkdirAll(pathName, os.ModePerm)
	if err != nil {
		return "", err
	}
	// 创建本地文件
	out, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// 将响应Body复制到文件中
	_, err = out.Write(body)
	if err != nil {
		return "", err
	}

	return fileName, nil
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

var lanClient *woozooo.Client

func init() {
	var err error
	lanClient, err = woozooo.NewClient("https://pc.woozooo.com", os.Getenv("woozooocookie"))
	if err != nil {
		log.Panic(err)
	}
}
func down2Woo(dirId int64, subDirName string, singerName, musicName, u string) (string, error) {
	log.Println("downuuuuu", u)
	// 发送GET请求
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	//log.Println(resp, err)
	if err != nil {
		return "", err
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
		return "", err
	}

	if bytes.Contains(body, []byte("验证")) {
		return "", errors.New("安全验证")
	}

	uploadResp, err := lanClient.UploadFileFromReader(bytes.NewReader(body), fileName, dirId)
	if err != nil {
		return "", err
	}
	if uploadResp.Zt != 1 {
		return "", fmt.Errorf("上传失败%v", uploadResp)
	}
	info := uploadResp.Text[0]
	data, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
