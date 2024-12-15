package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/deepch/RTSPtoWeb/middleware"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

// Message resp struct
type Message struct {
	Status  int         `json:"status"`
	Payload interface{} `json:"payload"`
}

// HTTPAPIServer start http server routes
func HTTPAPIServer() {
	//Set HTTP API mode
	log.WithFields(logrus.Fields{
		"module": "http_server",
		"func":   "RTSPServer",
		"call":   "Start",
	}).Infoln("Server HTTP start")
	var public *gin.Engine
	if !Storage.ServerHTTPDebug() {
		gin.SetMode(gin.ReleaseMode)
		public = gin.New()
	} else {
		gin.SetMode(gin.DebugMode)
		public = gin.Default()
	}

	public.Use(gin.Recovery())

	public.Use(CrossOrigin())
	//Add private login password protect methods
	privat := public.Group("/")
	// if Storage.ServerHTTPLogin() != "" && Storage.ServerHTTPPassword() != "" {
	// 	privat.Use(gin.BasicAuth(gin.Accounts{Storage.ServerHTTPLogin(): Storage.ServerHTTPPassword()}))
	// }

	privat.Use(middleware.AuthHandler())

	public.LoadHTMLGlob(Storage.ServerHTTPDir() + "/templates/*")
	public.Any("/login", HTTPServerLogin)
	privat.GET("/", HTTPAPIServerIndex)
	privat.Any("/test", HTTPTest)
	privat.GET("/pages/stream/list", HTTPAPIStreamList)
	privat.GET("/pages/stream/add", HTTPAPIAddStream)
	privat.GET("/pages/stream/edit/:uuid", HTTPAPIEditStream)
	privat.GET("/pages/player/hls/:uuid/:channel", HTTPAPIPlayHls)
	privat.GET("/pages/player/mse/:uuid/:channel", HTTPAPIPlayMse)
	privat.GET("/pages/player/webrtc/:uuid/:channel", HTTPAPIPlayWebrtc)
	privat.GET("/pages/multiview", HTTPAPIMultiview)
	privat.Any("/pages/multiview/full", HTTPAPIFullScreenMultiView)
	privat.GET("/pages/documentation", HTTPAPIServerDocumentation)
	privat.GET("/pages/player/all/:uuid/:channel", HTTPAPIPlayAll)
	public.StaticFS("/static", http.Dir(Storage.ServerHTTPDir()+"/static"))

	/*
		Stream Control elements
	*/

	privat.GET("/streams", HTTPAPIServerStreams)
	privat.POST("/stream/:uuid/add", HTTPAPIServerStreamAdd)
	privat.POST("/stream/:uuid/edit", HTTPAPIServerStreamEdit)
	privat.GET("/stream/:uuid/delete", HTTPAPIServerStreamDelete)
	privat.GET("/stream/:uuid/reload", HTTPAPIServerStreamReload)
	privat.GET("/stream/:uuid/info", HTTPAPIServerStreamInfo)

	/*
		Streams Multi Control elements
	*/

	privat.POST("/streams/multi/control/add", HTTPAPIServerStreamsMultiControlAdd)
	privat.POST("/streams/multi/control/delete", HTTPAPIServerStreamsMultiControlDelete)

	/*
		Stream Channel elements
	*/

	privat.POST("/stream/:uuid/channel/:channel/add", HTTPAPIServerStreamChannelAdd)
	privat.POST("/stream/:uuid/channel/:channel/edit", HTTPAPIServerStreamChannelEdit)
	privat.GET("/stream/:uuid/channel/:channel/delete", HTTPAPIServerStreamChannelDelete)
	privat.GET("/stream/:uuid/channel/:channel/codec", HTTPAPIServerStreamChannelCodec)
	privat.GET("/stream/:uuid/channel/:channel/reload", HTTPAPIServerStreamChannelReload)
	privat.GET("/stream/:uuid/channel/:channel/info", HTTPAPIServerStreamChannelInfo)

	/*
		Stream video elements
	*/
	//HLS
	public.GET("/stream/:uuid/channel/:channel/hls/live/index.m3u8", HTTPAPIServerStreamHLSM3U8)
	public.GET("/stream/:uuid/channel/:channel/hls/live/segment/:seq/file.ts", HTTPAPIServerStreamHLSTS)
	//HLS remote record
	//public.GET("/stream/:uuid/channel/:channel/hls/rr/:s/:e/index.m3u8", HTTPAPIServerStreamRRM3U8)
	//public.GET("/stream/:uuid/channel/:channel/hls/rr/:s/:e/:seq/file.ts", HTTPAPIServerStreamRRTS)
	//HLS LL
	public.GET("/stream/:uuid/channel/:channel/hlsll/live/index.m3u8", HTTPAPIServerStreamHLSLLM3U8)
	public.GET("/stream/:uuid/channel/:channel/hlsll/live/init.mp4", HTTPAPIServerStreamHLSLLInit)
	public.GET("/stream/:uuid/channel/:channel/hlsll/live/segment/:segment/:any", HTTPAPIServerStreamHLSLLM4Segment)
	public.GET("/stream/:uuid/channel/:channel/hlsll/live/fragment/:segment/:fragment/:any", HTTPAPIServerStreamHLSLLM4Fragment)
	//MSE
	public.GET("/stream/:uuid/channel/:channel/mse", HTTPAPIServerStreamMSE)
	public.POST("/stream/:uuid/channel/:channel/webrtc", HTTPAPIServerStreamWebRTC)
	//Save fragment to mp4
	public.GET("/stream/:uuid/channel/:channel/save/mp4/fragment/:duration", HTTPAPIServerStreamSaveToMP4)
	/*
		HTTPS Mode Cert
		# Key considerations for algorithm "RSA" ≥ 2048-bit
		openssl genrsa -out server.key 2048

		# Key considerations for algorithm "ECDSA" ≥ secp384r1
		# List ECDSA the supported curves (openssl ecparam -list_curves)
		#openssl ecparam -genkey -name secp384r1 -out server.key
		#Generation of self-signed(x509) public key (PEM-encodings .pem|.crt) based on the private (.key)

		openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
	*/
	if Storage.ServerHTTPS() {
		if Storage.ServerHTTPSAutoTLSEnable() {
			go func() {
				err := autotls.Run(public, Storage.ServerHTTPSAutoTLSName()+Storage.ServerHTTPSPort())
				if err != nil {
					log.Println("Start HTTPS Server Error", err)
				}
			}()
		} else {
			go func() {
				err := public.RunTLS(Storage.ServerHTTPSPort(), Storage.ServerHTTPSCert(), Storage.ServerHTTPSKey())
				if err != nil {
					log.WithFields(logrus.Fields{
						"module": "http_router",
						"func":   "HTTPSAPIServer",
						"call":   "ServerHTTPSPort",
					}).Fatalln(err.Error())
					os.Exit(1)
				}
			}()
		}
	}
	err := public.Run(Storage.ServerHTTPPort())
	if err != nil {
		log.WithFields(logrus.Fields{
			"module": "http_router",
			"func":   "HTTPAPIServer",
			"call":   "ServerHTTPPort",
		}).Fatalln(err.Error())
		os.Exit(1)
	}

}

func HTTPTest(c *gin.Context) {
	log.Info("test")
	c.JSON(http.StatusOK, gin.H{"data": "OK"})
}

func HTTPServerLogin(c *gin.Context) {

	// Handle GET
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"port":    Storage.ServerHTTPPort(),
			"streams": Storage.Streams,
			"version": time.Now().String(),
			"page":    "login",
			"error":   "",
		})
		return
	}

	// Handle POST

	// Get request body
	req, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, Message{Status: 400, Payload: "Lỗi khi xử lý yêu cầu"})
		return
	}

	reader := bytes.NewReader(req)

	//  Forward the request data to main server
	res, err := http.DefaultClient.Post(cfg.MainServerEndpoint, "application/json", reader)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, Message{Status: 500, Payload:"Máy chủ gặp lỗi, vui lòng liên hệ nhà phát triển"})
		return 
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, Message{Status: 500, Payload:"Thông tin đăng nhập lỗi, vui lòng thử lại"})
		return
	}

	if res.StatusCode != http.StatusOK {
        log.Info(fmt.Print(res))
		c.IndentedJSON(http.StatusBadRequest, Message{Status: 401, Payload: "Thông tin đăng nhập không hợp lệ"})
		return
	}

    var loginResponse struct {
        StatusCode int    `json:"statusCode"`
        Message    string `json:"message"`
        Data       struct {
            AccessToken string `json:"accessToken"`
            UserData    struct {
                ID        int    `json:"id"`
                Name      string `json:"name"`
                Username  string `json:"username"`
                Role      string `json:"role"`
                Status    bool   `json:"status"`
                CreatedAt string `json:"createdAt"`
                UpdatedAt string `json:"updatedAt"`
            } `json:"userData"`
        } `json:"data"`
    }

    if err := json.Unmarshal(body, &loginResponse); err != nil {
		log.Error(fmt.Print("error when parsing the json data"))
		log.Info(fmt.Print(err))
		c.IndentedJSON(http.StatusInternalServerError, Message{Status: 500, Payload: "Lỗi khi phân tích dữ liệu"})
        return
    }

    if loginResponse.StatusCode != http.StatusOK || loginResponse.Data.AccessToken == "" {
		log.Error(fmt.Print("auth data from server is invalid"))
		log.Info(fmt.Print(err))
		c.IndentedJSON(http.StatusBadRequest, Message{Status: 400, Payload: "Thông tin đăng nhập không hợp lệ"})
        return
    }

    if loginResponse.Data.UserData.Role != "admin"  {
		log.Error(fmt.Print("auth data from server is invalid"))
		log.Info(fmt.Print(err))
		c.IndentedJSON(http.StatusBadRequest, Message{Status: 400, Payload: "User không có quyền truy cập"})
        return
    }

	var store = middleware.GetStore()

	session, _ := store.Get(c.Request, "auth")
    session.Values["authenticated"] = true
    session.Values["accessToken"] = loginResponse.Data.AccessToken
    session.Values["userID"] = loginResponse.Data.UserData.ID
    session.Values["username"] = loginResponse.Data.UserData.Username
    session.Values["role"] = loginResponse.Data.UserData.Role
    session.Values["status"] = loginResponse.Data.UserData.Status
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   cfg.AuthConfig.TimeoutInSeconds,
		HttpOnly: false,
		Secure:   false,
	}
	session.Save(c.Request, c.Writer)
	c.IndentedJSON(http.StatusOK, Message{Status: 200, Payload: "validated"})
	return
}

// HTTPAPIServerIndex index file
func HTTPAPIServerIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "index",
	})

}

func HTTPAPIServerDocumentation(c *gin.Context) {
	c.HTML(http.StatusOK, "documentation.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "documentation",
	})
}

func HTTPAPIStreamList(c *gin.Context) {
	c.HTML(http.StatusOK, "stream_list.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "stream_list",
	})
}

func HTTPAPIPlayHls(c *gin.Context) {
	c.HTML(http.StatusOK, "play_hls.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "play_hls",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}
func HTTPAPIPlayMse(c *gin.Context) {
	c.HTML(http.StatusOK, "play_mse.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "play_mse",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}
func HTTPAPIPlayWebrtc(c *gin.Context) {
	c.HTML(http.StatusOK, "play_webrtc.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "play_webrtc",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}
func HTTPAPIAddStream(c *gin.Context) {
	c.HTML(http.StatusOK, "add_stream.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "add_stream",
	})
}
func HTTPAPIEditStream(c *gin.Context) {
	c.HTML(http.StatusOK, "edit_stream.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "edit_stream",
		"uuid":    c.Param("uuid"),
	})
}

func HTTPAPIMultiview(c *gin.Context) {
	c.HTML(http.StatusOK, "multiview.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "multiview",
	})
}

func HTTPAPIPlayAll(c *gin.Context) {
	c.HTML(http.StatusOK, "play_all.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "play_all",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}

type MultiViewOptions struct {
	Grid   int                             `json:"grid"`
	Player map[string]MultiViewOptionsGrid `json:"player"`
}
type MultiViewOptionsGrid struct {
	UUID       string `json:"uuid"`
	Channel    int    `json:"channel"`
	PlayerType string `json:"playerType"`
}

func HTTPAPIFullScreenMultiView(c *gin.Context) {
	var createParams MultiViewOptions
	err := c.ShouldBindJSON(&createParams)
	if err != nil {
		log.WithFields(logrus.Fields{
			"module": "http_page",
			"func":   "HTTPAPIFullScreenMultiView",
			"call":   "BindJSON",
		}).Errorln(err.Error())
	}
	log.WithFields(logrus.Fields{
		"module": "http_page",
		"func":   "HTTPAPIFullScreenMultiView",
		"call":   "Options",
	}).Debugln(createParams)
	c.HTML(http.StatusOK, "fullscreenmulti.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"options": createParams,
		"page":    "fullscreenmulti",
		"query":   c.Request.URL.Query(),
	})
}

// CrossOrigin Access-Control-Allow-Origin any methods
func CrossOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		//c.Next()
	}
}
