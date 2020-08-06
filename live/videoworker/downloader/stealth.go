package downloader

import "strings"

type URLRewriter interface {
	rewrite(url string) (newUrl string, onlyAlt bool)
	callback(url string, err error)
}

type BilibiliRewriter struct {
	needTxyunRewrite bool
}

func (u *BilibiliRewriter) rewrite(url string) (newUrl string, onlyAlt bool) {
	onlyAlt = false
	newUrl = url
	if u.needTxyunRewrite {
		if strings.Contains(url, "gotcha104") {
			newUrl = strings.Replace(url, "d1--cn-gotcha104.bilivideo.com", "3hq4yf8r2xgz9.cfc-execute.su.baidubce.com", 1)
		}
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

func (u *RewriterWrap) rewrite(url string) (newUrl string, onlyAlt bool) {
	for _, rewriter := range u.Rewriters {
		newUrl, onlyAlt = rewriter.rewrite(url)
		if newUrl != url || onlyAlt {
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
