package webalayzer

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var (
	ErrNoURL  = errors.New("no URL passed")
	ErrBadURL = errors.New("bad URL")
)

func RunWebServer(addr string) error {
	engine := gin.Default()

	engine.LoadHTMLGlob("templates/*.gohtml")

	engine.GET("", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.gohtml", &AnalyzerFormParams{})
	})
	engine.GET("test", func(c *gin.Context) {
		c.HTML(http.StatusOK, "test.gohtml", nil)
	})

	engine.GET("webalayzer", func(c *gin.Context) {
		pageParams := AnalyzerPageParams{}
		pageParams.URL = strings.TrimSpace(c.Query("url"))
		if len(pageParams.URL) == 0 {
			pageParams.Error = ErrNoURL
		}

		analyzer := new(Analyzer)
		analyzer.SkipInaccessibleCheck = c.Query("checkInaccessibleLinks") != "on"
		analyzer.SkipEmptyLinks = c.Query("countEmptyLinks") != "on"

		pageParams.SkipInaccessibleCheck = analyzer.SkipInaccessibleCheck
		pageParams.SkipEmptyLinks = analyzer.SkipEmptyLinks

		webPageStats, err := analyzer.AnalyzeURL(c, pageParams.URL)
		if err != nil {
			pageParams.Error = err
		} else {
			pageParams.Stats = webPageStats
		}

		var status int
		if pageParams.Error == nil {
			status = http.StatusOK
		} else {
			status = http.StatusBadRequest
		}
		c.HTML(status, "webalayzer.gohtml", &pageParams)
	})

	return engine.Run(addr)
}

type AnalyzerPageParams struct {
	AnalyzerFormParams

	Stats WebPageStats

	Error error
}

type AnalyzerFormParams struct {
	URL                                   string
	SkipInaccessibleCheck, SkipEmptyLinks bool
}
