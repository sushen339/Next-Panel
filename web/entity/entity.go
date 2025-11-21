package entity

import (
	"crypto/tls"
	"math"
	"net"
	"strings"
	"time"

	"x-ui/util/common"
)

type Msg struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Obj     any    `json:"obj"`
}

type AllSetting struct {
	WebListen                   string `json:"webListen" form:"webListen"`
	WebDomain                   string `json:"webDomain" form:"webDomain"`
	WebPort                     int    `json:"webPort" form:"webPort"`
	WebCertFile                 string `json:"webCertFile" form:"webCertFile"`
	WebKeyFile                  string `json:"webKeyFile" form:"webKeyFile"`
	WebBasePath                 string `json:"webBasePath" form:"webBasePath"`
	SessionMaxAge               int    `json:"sessionMaxAge" form:"sessionMaxAge"`
	PageSize                    int    `json:"pageSize" form:"pageSize"`
	ExpireDiff                  int    `json:"expireDiff" form:"expireDiff"`
	TrafficDiff                 int    `json:"trafficDiff" form:"trafficDiff"`
	RemarkModel                 string `json:"remarkModel" form:"remarkModel"`
	TgBotEnable                 bool   `json:"tgBotEnable" form:"tgBotEnable"`
	TgBotToken                  string `json:"tgBotToken" form:"tgBotToken"`
	TgBotProxy                  string `json:"tgBotProxy" form:"tgBotProxy"`
	TgBotAPIServer              string `json:"tgBotAPIServer" form:"tgBotAPIServer"`
	TgBotChatId                 string `json:"tgBotChatId" form:"tgBotChatId"`
	TgRunTime                   string `json:"tgRunTime" form:"tgRunTime"`
	TgBotBackup                 bool   `json:"tgBotBackup" form:"tgBotBackup"`
	TgBotLoginNotify            bool   `json:"tgBotLoginNotify" form:"tgBotLoginNotify"`
	TgCpu                       int    `json:"tgCpu" form:"tgCpu"`
	TgLang                      string `json:"tgLang" form:"tgLang"`
	TimeLocation                string `json:"timeLocation" form:"timeLocation"`
	TwoFactorEnable             bool   `json:"twoFactorEnable" form:"twoFactorEnable"`
	TwoFactorToken              string `json:"twoFactorToken" form:"twoFactorToken"`
	SubEnable                   bool   `json:"subEnable" form:"subEnable"`
	SubTitle                    string `json:"subTitle" form:"subTitle"`
	SubPath                     string `json:"subPath" form:"subPath"`
	SubDomain                   string `json:"subDomain" form:"subDomain"`
	SubUpdates                  int    `json:"subUpdates" form:"subUpdates"`
	ExternalTrafficInformEnable bool   `json:"externalTrafficInformEnable" form:"externalTrafficInformEnable"`
	ExternalTrafficInformURI    string `json:"externalTrafficInformURI" form:"externalTrafficInformURI"`
	SubEncrypt                  bool   `json:"subEncrypt" form:"subEncrypt"`
	SubShowInfo                 bool   `json:"subShowInfo" form:"subShowInfo"`
	SubURI                      string `json:"subURI" form:"subURI"`
	SubJsonPath                 string `json:"subJsonPath" form:"subJsonPath"`
	SubJsonURI                  string `json:"subJsonURI" form:"subJsonURI"`
	SubJsonFragment             string `json:"subJsonFragment" form:"subJsonFragment"`
	SubJsonNoises               string `json:"subJsonNoises" form:"subJsonNoises"`
	SubJsonMux                  string `json:"subJsonMux" form:"subJsonMux"`
	SubJsonRules                string `json:"subJsonRules" form:"subJsonRules"`
	SubTotalId                  string `json:"subTotalId" form:"subTotalId"`
	Datepicker                  string `json:"datepicker" form:"datepicker"`
}

func (s *AllSetting) CheckValid() error {
	if s.WebListen != "" {
		ip := net.ParseIP(s.WebListen)
		if ip == nil {
			return common.NewError("web listen is not valid ip:", s.WebListen)
		}
	}

	if s.WebPort <= 0 || s.WebPort > math.MaxUint16 {
		return common.NewError("web port is not a valid port:", s.WebPort)
	}

	if s.WebCertFile != "" || s.WebKeyFile != "" {
		_, err := tls.LoadX509KeyPair(s.WebCertFile, s.WebKeyFile)
		if err != nil {
			return common.NewErrorf("cert file <%v> or key file <%v> invalid: %v", s.WebCertFile, s.WebKeyFile, err)
		}
	}

	if !strings.HasPrefix(s.WebBasePath, "/") {
		s.WebBasePath = "/" + s.WebBasePath
	}
	if !strings.HasSuffix(s.WebBasePath, "/") {
		s.WebBasePath += "/"
	}
	if !strings.HasPrefix(s.SubPath, "/") {
		s.SubPath = "/" + s.SubPath
	}
	if !strings.HasSuffix(s.SubPath, "/") {
		s.SubPath += "/"
	}

	if !strings.HasPrefix(s.SubJsonPath, "/") {
		s.SubJsonPath = "/" + s.SubJsonPath
	}
	if !strings.HasSuffix(s.SubJsonPath, "/") {
		s.SubJsonPath += "/"
	}

	_, err := time.LoadLocation(s.TimeLocation)
	if err != nil {
		return common.NewError("time location not exist:", s.TimeLocation)
	}

	return nil
}
