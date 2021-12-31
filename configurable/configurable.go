package configurable

import (
	"net/http"

	"github.com/gocolly/colly/v2"
)

type (
	Storage interface {
		GetConfig(name string) (IConfig, error)
	}

	IConfig interface {
		GetBaseURL() string             // 获取开始URI
		GetBaseRequest() *colly.Request // 获取开始的request
		GetProxy() string               // 获取代理信息
		GetName() string                // 获取配置名称
		GetStep(name string) []Step     // 获取一个抓取步骤
		GetFinal() map[string]Element   // 获取结果
	}

	Step struct {
		HttpMethod string             `mapstructure:"http_method" json:"http_method"` // Http请求方式 默认：GET
		CSSPath    string             `mapstructure:"css_path" json:"css_path"`       // CSSPath提取
		Attr       string             `mapstructure:"attr" json:"attr"`               // 节点内容提取
		Next       string             `mapstructure:"next" json:"next"`               // 下一步步骤 其中start-end 分别表示开始和抽取结果
		Ext        map[string]Element `mapstructure:"ext" json:"ext"`                 // 额外节点信息
	}

	Element struct {
		CSSPath string `mapstructure:"css_path" json:"css_path"` // CSSPath提取
		Attr    string `mapstructure:"attr" json:"attr"`         // 节点内容提取
		ExtName string `mapstructure:"ext_name" json:"ext_name"` // 额外节点信息
		List    bool   `mapstructure:"list" json:"list"`         // 是否是一个列表
	}
)

func (s Step) GetHttpMethod() string {
	if len(s.HttpMethod) == 0 {
		return http.MethodGet
	}
	return s.HttpMethod
}

func (s Step) GetAttr() string {
	if len(s.Attr) == 0 {
		return "href"
	}
	return s.Attr
}
