package downloader

import "strings"

type URLRewriter interface {
	rewrite(url string) (newUrl string, useMain, useAlt int)
	callback(url string, err error)
}

type BilibiliRewriter struct {
	needTxyunRewrite bool
}

func (u *BilibiliRewriter) rewrite(url string) (newUrl string, useMain, useAlt int) {
	//onlyAlt = false
	useMain = 1
	useAlt = 1
	newUrl = url
	if strings.Contains(url, "gotcha105") {
		useAlt = 0
	} else if strings.Contains(url, "gotcha104") {
		if u.needTxyunRewrite {
			newUrl = strings.Replace(url, "https://d1--cn-gotcha104.bilivideo.com", "https://3hq4yf8r2xgz9.cfc-execute.su.baidubce.com", 1)
			newUrl = strings.Replace(newUrl, "http://d1--cn-gotcha104.bilivideo.com", "https://3hq4yf8r2xgz9.cfc-execute.su.baidubce.com", 1)
			useMain = 1
			useAlt = 0
		} else {
			useMain = 0
			useAlt = 1
		}
	} else if strings.Contains(url, "baidubce") {
		useAlt = 2
		useMain = 1
	}
	return
}

func (u *BilibiliRewriter) callback(url string, err error) {
	if err != nil && strings.HasSuffix(err.Error(), "403") {
		if strings.Contains(url, "gotcha104") {
			u.needTxyunRewrite = true
		}
	}
}

type RewriterWrap struct {
	Rewriters []URLRewriter
}

func (u *RewriterWrap) rewrite(url string) (newUrl string, useMain, useAlt int) {
	for _, rewriter := range u.Rewriters {
		newUrl, useMain, useAlt = rewriter.rewrite(url)
		if newUrl != url || useMain != 1 || useAlt != 1 {
			break
		}
	}
	return
}

func (u *RewriterWrap) callback(url string, err error) {
	for _, rewriter := range u.Rewriters {
		rewriter.callback(url, err)
	}
}

func getRewriter() URLRewriter {
	return &RewriterWrap{
		Rewriters: []URLRewriter{
			&BilibiliRewriter{},
		},
	}
}
