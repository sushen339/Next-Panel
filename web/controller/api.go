package controller

import (
	"x-ui/sub"
	"x-ui/web/service"

	"github.com/gin-gonic/gin"
)

type APIController struct {
	BaseController
	inboundController *InboundController
	serverController  *ServerController
	Tgbot             service.Tgbot
}

func NewAPIController(g *gin.RouterGroup) *APIController {
	a := &APIController{}
	a.initRouter(g)
	return a
}

func (a *APIController) initRouter(g *gin.RouterGroup) {
	// Main API group
	api := g.Group("/panel/api")
	api.Use(a.checkLogin)

	// Inbounds API
	inbounds := api.Group("/inbounds")
	a.inboundController = NewInboundController(inbounds)

	// Server API
	server := api.Group("/server")
	a.serverController = NewServerController(server)

	// Subscription API
	sub := api.Group("/sub")
	sub.GET("/total-id", a.getTotalSubscriptionId)

	// Extra routes
	api.GET("/backuptotgbot", a.BackuptoTgbot)
}

func (a *APIController) BackuptoTgbot(c *gin.Context) {
	a.Tgbot.SendBackupToAdmins()
	jsonObj(c, "", nil)
}

func (a *APIController) getTotalSubscriptionId(c *gin.Context) {
	subService := sub.NewSubService(false, "")
	
	// 检查是否有 generate 参数
	generateNew := c.Query("generate")
	
	var totalId string
	var err error
	
	if generateNew == "true" {
		// 生成新的总订阅ID
		totalId, err = subService.GenerateNewTotalSubscriptionId()
	} else {
		// 只获取现有的ID，不自动生成
		totalId, err = subService.GetTotalSubscriptionId()
	}
	
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.obtain", "Total Subscription ID"), err)
		return
	}
	jsonObj(c, totalId, nil)
}
