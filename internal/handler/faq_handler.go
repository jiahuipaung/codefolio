package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FAQHandler struct {}

// FAQ数据结构
type FAQItem struct {
	ID       uint   `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// 预定义的FAQ数据
var faqData = []FAQItem{
	{
		ID:       1,
		Question: "什么是Codefolio?",
		Answer:   "Codefolio是一个基于Go语言开发的个人作品集展示平台，帮助开发者展示自己的项目和技能。",
	},
	{
		ID:       2,
		Question: "如何创建账户?",
		Answer:   "点击网站右上角的"注册"按钮，填写您的电子邮件、密码和个人信息，然后点击"创建账户"即可。",
	},
	{
		ID:       3,
		Question: "如何上传作品集?",
		Answer:   "登录后，进入个人仪表板，点击"添加新项目"按钮，然后填写项目详情并上传相关资源。",
	},
	{
		ID:       4,
		Question: "支持哪些技术栈的展示?",
		Answer:   "Codefolio支持所有主流技术栈的展示，包括但不限于Web开发、移动应用、桌面应用、人工智能项目等。",
	},
	{
		ID:       5,
		Question: "如何与其他开发者交流?",
		Answer:   "您可以在项目评论区留言，也可以通过平台内置的消息系统与其他开发者私聊交流。",
	},
}

// NewFAQHandler 创建FAQ处理器实例
func NewFAQHandler() *FAQHandler {
	return &FAQHandler{}
}

// GetAllFAQs 获取所有FAQ
func (h *FAQHandler) GetAllFAQs(c *gin.Context) {
	c.JSON(http.StatusOK, faqData)
} 