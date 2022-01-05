package configurable

import (
	"embed"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/gocolly/colly/v2"
)

type (
	DirStorage struct {
		path   string
		pathFs *embed.FS
	}
	FileConfig struct {
		Name           string             `json:"name"`
		BaseURL        string             `json:"url"`
		BaseHttpMethod string             `json:"http_method"`
		Proxy          string             `json:"proxy"`
		Steps          map[string][]Step  `json:"steps"`
		Final          map[string]Element `json:"final"`
	}
)

func NewDirStorage(path string, fs2 *embed.FS) (*DirStorage, error) {
	if !isDir(path) && fs2 == nil {
		return nil, fs.ErrNotExist
	}
	return &DirStorage{
		path:   path,
		pathFs: fs2,
	}, nil
}

func (d *DirStorage) getConfigByPath(name string) ([]byte, error) {
	filename := filepath.Join(d.path, name+".json")
	if !isFile(filename) {
		return nil, fs.ErrNotExist
	}
	return ioutil.ReadFile(filename)
}

func (d *DirStorage) getConfigByEmbed(name string) ([]byte, error) {
	return d.pathFs.ReadFile(name + ".json")
}

func (d *DirStorage) GetConfig(name string) (IConfig, error) {
	var bytes []byte
	var err error
	if d.pathFs == nil {
		bytes, err = d.getConfigByPath(name)
	} else {
		bytes, err = d.getConfigByEmbed(name)
	}
	if err != nil {
		return nil, err
	}
	result := new(FileConfig)
	if err := json.Unmarshal(bytes, result); err != nil {
		return nil, err
	}
	result.Name = name
	if len(result.BaseHttpMethod) == 0 {
		result.BaseHttpMethod = http.MethodGet
	}
	return result, nil
}

func (f *FileConfig) GetName() string {
	return f.Name
}

func (f *FileConfig) GetProxy() string {
	return f.Proxy
}

func (f *FileConfig) GetBaseRequest(urls ...string) *colly.Request {
	ctx := colly.NewContext()
	ctx.Put(CollyConfName, f.GetName())
	if len(f.GetStep(CollyConfStepStart)) == 0 {
		ctx.Put(CollyConfStepName, CollyConfStepEnd)
	} else {
		ctx.Put(CollyConfStepName, CollyConfStepStart)
	}
	urlStr := f.GetBaseURL()
	if len(urls) > 0 {
		urlStr = urls[0]
	}
	u, _ := url.Parse(urlStr)
	return &colly.Request{
		URL:      u,
		Method:   http.MethodGet,
		Ctx:      ctx,
		ProxyURL: f.Proxy,
	}
}

func (f *FileConfig) GetBaseURL() string {
	return f.BaseURL
}

func (f *FileConfig) GetStep(name string) []Step {
	steps, _ := f.Steps[name]
	return steps
}

func (f *FileConfig) GetFinal() map[string]Element {
	return f.Final
}
