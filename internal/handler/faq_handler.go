package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FAQHandler struct{}

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
		Question: "这个网站是干什么的？",
		Answer:   "我们的定位是一个简历分享网站，如果你的目标是寻找国内大厂工作并且在简历撰写方面遇到困难，那么就可以来我们的网站上寻找简历参考。",
	},
	{
		ID:       2,
		Question: "这些简历是真实的吗？",
		Answer:   "是的，网站初始阶段的第一批简历来源是团队里的小伙伴和身边的大厂朋友自己真实求职使用的简历。",
	},
	{
		ID:       3,
		Question: "我如何上传自己的简历?",
		Answer:   " 由于网站定位原因，我们对简历上传有一定的门槛～首先需要你将简历上传，接着需要进行经历认证（为了确保你的经历真实，个人信息会完全保密），完成之后我们会有审核，最后你的简历就会展示在我们的网站上啦～",
	},
	{
		ID:       4,
		Question: "如果我想向简历所有者进行咨询，有什么办法吗？",
		Answer:   "可以向简历所有者进行付费咨询，请点击简历详情页的「我想咨询」按钮，给简历所有者留言。",
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
