package stealth

import "strings"

type URLRewriter interface {
	Rewrite(url string) (newUrl string, useMain, useAlt int)
	Callback(url string, err error)
}

type BilibiliRewriter struct {
	needTxyunRewrite bool
}

func (u *BilibiliRewriter) Rewrite(url string) (newUrl string, useMain, useAlt int) {
	//onlyAlt = false
	useMain = 1
	useAlt = 1
	newUrl = url
	// for gotcha105 & gotcha104, never use altproxy when downloading
	if strings.Contains(url, "gotcha105") {
		useAlt = 0
	} else if strings.Contains(url, "gotcha103") {
		newUrl = strings.Replace(url, "https://d1--cn-gotcha103.bilivideo.com", "http://shnode.misty.moe:49980", 1)
		newUrl = strings.Replace(url, "http://d1--cn-gotcha103.bilivideo.com", "http://shnode.misty.moe:49980", 1)
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

func (u *BilibiliRewriter) Callback(url string, err error) {
	if err != nil && strings.HasSuffix(err.Error(), "403") {
		if strings.Contains(url, "gotcha104") {
			u.needTxyunRewrite = true
		}
	}
}

type RewriterWrap struct {
	Rewriters []URLRewriter
}

func (u *RewriterWrap) Rewrite(url string) (newUrl string, useMain, useAlt int) {
	for _, rewriter := range u.Rewriters {
		newUrl, useMain, useAlt = rewriter.Rewrite(url)
		if newUrl != url || useMain != 1 || useAlt != 1 {
			break
		}
	}
	return
}

func (u *RewriterWrap) Callback(url string, err error) {
	for _, rewriter := range u.Rewriters {
		rewriter.Callback(url, err)
	}
}

func GetRewriter() URLRewriter {
	return &RewriterWrap{
		Rewriters: []URLRewriter{
			&BilibiliRewriter{},
		},
	}
}
