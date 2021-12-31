package configurable

import "github.com/PuerkitoBio/goquery"

// analyzeDocumentByElements 根据 Element 解析 goquery.Document
func analyzeDocumentByElements(
	document *goquery.Document,
	elements map[string]Element,
	extInfo map[string]interface{},
) map[string]interface{} {
	result := make(map[string]interface{}, len(elements))
	for name, element := range elements {
		if len(element.ExtName) != 0 {
			result[name] = extInfo[element.ExtName]
			continue
		}
		selection := document.Find(element.CSSPath)
		if element.List {
			list := make([]string, 0, len(selection.Nodes))
			selection.Each(func(_ int, node *goquery.Selection) {
				if len(element.Attr) == 0 {
					list = append(list, node.Text())
					return
				}
				if v, ok := node.Attr(element.Attr); ok {
					list = append(list, v)
				}
			})
			result[name] = list
			continue
		}
		if len(element.Attr) == 0 {
			result[name] = selection.Text()
			continue
		}
		if v, ok := selection.Attr(element.Attr); ok {
			result[name] = v
		} else {
			result[name] = ""
		}
	}
	return result
}
