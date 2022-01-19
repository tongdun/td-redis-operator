package luc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"td-redis-operator/pkg/conf"
	. "td-redis-operator/pkg/web"
	"time"
)

var SSO string

func Luc(g *gin.Context) {
	//过滤掉不需要认证的接口
	if g.Request.URL.Path == "/api/v1alpha2/redisall" {
		return
	}
	token := g.GetHeader("x-user-token")
	if token == "" || token == "null" {
		errlog := fmt.Sprintf("luc get token failed")
		g.JSON(http.StatusBadRequest, Result(errlog, false))
		g.Abort()
		return
	}
	// Use the custom HTTP client when requesting a token.
	client := &http.Client{Timeout: 2 * time.Second}
	type User struct {
		Token string `json:"token"`
	}
	user, err := json.Marshal(User{Token: token})
	if err != nil {
		errlog := fmt.Sprintf("luc switch to json failed, error:%s", err.Error())
		g.JSON(http.StatusBadRequest, Result(errlog, false))
		g.Abort()
		return
	}
	req, err := http.NewRequest("POST", SSO+"/oauth/user/", bytes.NewBuffer(user))
	if err != nil {
		errlog := fmt.Sprintf("luc request /oauth/user/ failed, error:%s", err.Error())
		g.JSON(http.StatusBadRequest, Result(errlog, false))
		g.Abort()
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		g.JSON(http.StatusBadRequest, Result(err.Error(), false))
		g.Abort()
		return
	}
	defer resp.Body.Close()
	userInfo, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		g.JSON(http.StatusBadGateway, Result(err.Error(), false))
		g.Abort()
		return
	}
	if resp.StatusCode/200 != 1 {
		errlog := fmt.Sprintf("reasonCode:%d, %s", resp.StatusCode, string(userInfo))
		g.JSON(resp.StatusCode, Result(errlog, false))
		g.Abort()
		return
	}
	var userInfoS struct {
		Msg      string   `json:"msg"`
		Roles    []string `json:"roles"`
		Realname string   `json:"name"`
		Owner    string   `json:"alias"`
	}
	err = json.Unmarshal(userInfo, &userInfoS)
	if err != nil {
		errlog := fmt.Sprintf("token invalid(token:%s), error:%s, userInfo:%s", token, err.Error(), string(userInfo))

		g.JSON(http.StatusInsufficientStorage, Result(errlog, false))
		g.Abort()
		return
	}
	if userInfoS.Msg != "" {
		errlog := fmt.Sprintf("token:%s msg:%s", token, userInfoS.Msg)

		g.JSON(http.StatusInsufficientStorage, Result(errlog, false))
		g.Abort()
		return
	}
	g.Set("realname", userInfoS.Realname)
	g.Set("owner", userInfoS.Owner)
	if !contain("dba", userInfoS.Roles) {
		g.Set("dba", false)
	} else {
		g.Set("dba", true)
	}
}

func contain(selected string, list []string) bool {
	for _, element := range list {
		if element == selected {
			return true
		}
	}
	return false
}

func init() {
	SSO = conf.Cfg.Luc
}
