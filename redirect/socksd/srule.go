package socksd

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	//	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/log"
	"github.com/ssoor/youniverse/redirect/socksd/compiler"
)

const (
	Rewrite_URL = iota
	Redirect_URL
	Rewrite_HTML
	Rewrite_JaveScript
)

type RuleTypeof int32

type internalJSONURLMatch struct {
	compiler.JSONURLMatch
	Type RuleTypeof `json:"type"`
}

type JSONSRule struct {
	Compiler []internalJSONURLMatch `json:"compilers"`
}

type JSONLimits struct {
	MaxResponseContentLen int64 `json:"max_response_content_len"`
}

type JSONRules struct {
	Local  bool        `json:"local"`
	Limits JSONLimits  `json:"limits"`
	SRules []JSONSRule `json:"srules"`
}

type SRules struct {
	local        bool
	limits       JSONLimits
	Rewrite_URL  *compiler.URLMatch
	Redirect_URL *compiler.URLMatch

	Rewrite_HTML       *compiler.URLMatch
	Rewrite_JaveScript *compiler.URLMatch

	tranpoort_local  *http.Transport
	tranpoort_remote *http.Transport
}

func NewSRules(forward socks.Dialer) *SRules {

	return &SRules{
		Rewrite_URL:        compiler.NewURLMatch(),
		Redirect_URL:       compiler.NewURLMatch(),
		Rewrite_HTML:       compiler.NewURLMatch(),
		Rewrite_JaveScript: compiler.NewURLMatch(),

		tranpoort_remote: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return forward.Dial(network, addr)
			},
		},
		tranpoort_local: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return socks.Direct.Dial(network, addr)
			},
		},
	}
}

func (this *SRules) ResolveJson(data []byte) (err error) {

	jsonRules := JSONRules{}

	if err = json.Unmarshal(data, &jsonRules); err != nil {
		return err
	}

	this.local = jsonRules.Local
	this.limits = jsonRules.Limits

	if false == this.local {
		this.tranpoort_local = this.tranpoort_remote
	}

	for i := 0; i < len(jsonRules.SRules); i++ {
		for j := 0; j < len(jsonRules.SRules[i].Compiler); j++ {
			if err := this.Add(jsonRules.SRules[i].Compiler[j]); nil != err {
				return err
			}
		}
	}

	return nil
}

func (s *SRules) Add(internalMatch internalJSONURLMatch) (err error) {
	var match compiler.JSONURLMatch

	match.Url = internalMatch.Url
	match.Host = internalMatch.Host
	match.Match = internalMatch.Match

	switch internalMatch.Type {
	case Rewrite_URL:
		err = s.Rewrite_URL.AddMatchs(match)
	case Redirect_URL:
		err = s.Redirect_URL.AddMatchs(match)
	case Rewrite_HTML:
		err = s.Rewrite_HTML.AddMatchs(match)
	case Rewrite_JaveScript:
		err = s.Rewrite_JaveScript.AddMatchs(match)
	default:
		return errors.New("Unrecognized type of routing rules")
	}

	for i := 0; i < len(match.Match); i++ {
		log.Info("Sign up routing:", err, internalMatch.Type, fmt.Sprintf("%s(%s)", match.Host, match.Url), match.Match[i])
	}

	return err
}

func (s *SRules) replaceURL(urlmatch *compiler.URLMatch, srcurl *url.URL) (dsturl *url.URL, err error) {
	var dststr string
	if dststr, err = urlmatch.Replace(srcurl, srcurl.String()); err != nil {
		return nil, err
	}

	if dsturl, err = url.Parse(dststr); err != nil {
		return nil, err
	}

	return dsturl, nil
}

func (this *SRules) GetRewriteHTML(req *http.Request, resp *http.Response) (dst string, err error) {

	if resp_type := resp.Header.Get("Content-Type"); false == strings.Contains(strings.ToLower(resp_type), "text/html") {
		return "", errors.New("Content type mismatch.")
	}

	bodyReader := bufio.NewReader(resp.Body)

	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {

		read, err := gzip.NewReader(resp.Body)
		if nil != err {
			log.Info("Create gzip reader error:", err)
			return "", err
		}

		bodyReader = bufio.NewReader(read)

		resp.Header.Del("Content-Encoding")
	}

	//dumpdata, _ := httputil.DumpResponse(resp, true)

	//log.Println(string(dumpdata))

	var bodyBuf bytes.Buffer

	bodyBuf.ReadFrom(bodyReader)

	html, err := this.Rewrite_HTML.Replace(resp.Request.URL, bodyBuf.String())

	if err != nil {
		err = nil
		html = bodyBuf.String()
		//log.Warning("Injection html", resp.Request.URL.String(), " failed:",err,", html size is", resp.ContentLength)
	} else {
		log.Info("Injection html", resp.Request.URL.String(), " successed, old size is", resp.ContentLength)
	}

	return html, err
}

func (this *SRules) GetRewriteJaveScript(req *http.Request, resp *http.Response) (dst string, err error) {

	if resp_type := resp.Header.Get("Content-Type"); false == strings.Contains(strings.ToLower(resp_type), "text/javascript") && false == strings.Contains(strings.ToLower(resp_type), "application/javascript") && false == strings.Contains(strings.ToLower(resp_type), "application/x-javascript") {
		return "", errors.New("Content type mismatch.")
	}

	bodyReader := bufio.NewReader(resp.Body)

	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {

		read, err := gzip.NewReader(resp.Body)
		if nil != err {
			log.Info("Create gzip reader error:", err)
			return "", err
		}

		bodyReader = bufio.NewReader(read)

		resp.Header.Del("Content-Encoding")
	}

	var bodyBuf bytes.Buffer

	bodyBuf.ReadFrom(bodyReader)

	html, err := this.Rewrite_JaveScript.Replace(resp.Request.URL, bodyBuf.String())

	if err != nil {
		err = nil
		html = bodyBuf.String()
	} else {
		log.Info("Injection javascript", resp.Request.URL.String(), " successed, old size is", resp.ContentLength)
	}

	return html, err
}

func (this *SRules) ResolveRequest(req *http.Request) (tran *http.Transport, resp *http.Response) {
	tran = this.tranpoort_local

	if dsturl, err := this.replaceURL(this.Redirect_URL, req.URL); err == nil {
		if false == strings.EqualFold(req.URL.String(), dsturl.String()) {
			log.Info("Socksd redirect request", req.URL, "to", dsturl)

			req.URL = dsturl

			tran = nil
			resp = this.createRedirectResponse(dsturl.String(), req)
		} else {
			log.Info("Socksd set request", req.URL, "is remote.")

			resp = nil
			tran = this.tranpoort_remote
		}
	}

	if dsturl, err := this.replaceURL(this.Rewrite_URL, req.URL); err == nil {
		if strings.EqualFold(req.URL.Host, dsturl.Host) {
			log.Info("Socksd rewrite request", req.URL, "to", dsturl)

			req.URL = dsturl

			resp = nil
			tran = this.tranpoort_remote
		} else {
			log.Error("Socksd rewrite request", req.URL, "to", dsturl, "failed: Unauthorized jump, the host does not match")
		}
	}

	return tran, resp
}

func (this *SRules) ResolveResponse(req *http.Request, resp *http.Response) *http.Response {
	var err error
	var html string

	if resp.ContentLength == 0 || resp.ContentLength > this.limits.MaxResponseContentLen {
		return resp
	}

	if html, err = this.GetRewriteHTML(req, resp); nil == err {
		log.Info("Socksd request url(html):", resp.Request.URL)
	} else if html, err = this.GetRewriteJaveScript(req, resp); nil == err {
		log.Info("Socksd request url(javascript):", resp.Request.URL)
	} else {
		return resp
	}

	if -1 != resp.ContentLength {
		resp.ContentLength = int64(len([]byte(html)))
		resp.Header.Set("Content-Length", strconv.Itoa(len([]byte(html))))
	}

	oldBody := resp.Body
	defer oldBody.Close()

	resp.Body = ioutil.NopCloser(strings.NewReader(html))

	return resp
}

func (this *SRules) createRedirectResponse(url string, req *http.Request) (resp *http.Response) {

	resp = &http.Response{
		StatusCode: http.StatusFound,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Request:    req,
		Header: http.Header{
			"Location": []string{url},
		},
		ContentLength:    0,
		TransferEncoding: nil,
		Body:             ioutil.NopCloser(strings.NewReader("")),
		Close:            true,
	}

	return
}
