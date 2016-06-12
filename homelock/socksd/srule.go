package socksd

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	//	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/ssoor/socks"
	"github.com/ssoor/socks/compiler"
	"github.com/ssoor/youniverse/log"
)

const (
	Rewrite_URL = iota
	Redirect_URL
	Rewrite_HTML
)

type RuleTypeof int32

type JSONCompiler struct {
	Type  RuleTypeof `json:"type"`
	Host  string     `json:"host"`
	Match []string   `json:"match"`
}

type JSONSRule struct {
	Compiler []JSONCompiler `json:"compilers"`
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
	Rewrite_URL  *compiler.SCompiler
	Redirect_URL *compiler.SCompiler

	Rewrite_HTML *compiler.SCompiler

	tranpoort_local  *http.Transport
	tranpoort_remote *http.Transport
}

func NewSRules(forward socks.Dialer) *SRules {

	return &SRules{
		Rewrite_URL:  compiler.NewSCompiler(),
		Redirect_URL: compiler.NewSCompiler(),
		Rewrite_HTML: compiler.NewSCompiler(),

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

func (this *SRules) ResolveRewriteHTML(req *http.Request, resp *http.Response) (newresp *http.Response) {

	if resp_type := resp.Header.Get("Content-Type"); false == strings.Contains(strings.ToLower(resp_type), "text/html") {
		return nil
	}

	if resp.ContentLength == 0 || resp.ContentLength > this.limits.MaxResponseContentLen {
		return nil
	}

	log.Info("Socksd request url:", resp.Request.URL)

	bodyReader := bufio.NewReader(resp.Body)

	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {

		read, _ := gzip.NewReader(resp.Body)

		bodyReader = bufio.NewReader(read)

		resp.Header.Del("Content-Encoding")
	}

	//dumpdata, _ := httputil.DumpResponse(resp, true)

	//log.Println(string(dumpdata))

	var bodyBuf bytes.Buffer

	bodyBuf.ReadFrom(bodyReader)

	html, err := this.Rewrite_HTML.Replace(resp.Request.Host, bodyBuf.String())

	if err != nil {
		html = bodyBuf.String()
		//log.Warning("Injection html", resp.Request.URL.String(), " failed:",err,", html size is", resp.ContentLength)
	} else {
		log.Info("Injection html", resp.Request.URL.String(), " successed, old size is", resp.ContentLength)
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

func (this *SRules) ResolveRequest(req *http.Request) (tran *http.Transport, resp *http.Response) {
	tran = this.tranpoort_local

	if dsturl, err := this.replaceURL(this.Redirect_URL, req.Host, req.URL.String()); err == nil {
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

	if dsturl, err := this.replaceURL(this.Rewrite_URL, req.Host, req.URL.String()); err == nil {
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

	if newresp := this.ResolveRewriteHTML(req, resp); nil != newresp {
		return newresp
	}

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

func (this *SRules) Add(compiler JSONCompiler) (err error) {

	switch compiler.Type {
	case Rewrite_URL:
		err = this.Rewrite_URL.Add(compiler.Host, compiler.Match)
	case Redirect_URL:
		err = this.Redirect_URL.Add(compiler.Host, compiler.Match)
	case Rewrite_HTML:
		err = this.Rewrite_HTML.Add(compiler.Host, compiler.Match)
	default:
		return errors.New("Unrecognized type of routing rules")
	}

	for i := 0; i < len(compiler.Match); i++ {
		log.Info("Sign up routing:", compiler.Type, compiler.Host, compiler.Match[i], err)
	}

	return err
}

func (this *SRules) replaceURL(scomp *compiler.SCompiler, host string, src string) (dsturl *url.URL, err error) {
	var dststr string

	if dststr, err = scomp.Replace(host, src); err != nil {
		return nil, err
	}

	if dsturl, err = url.Parse(dststr); err != nil {
		return nil, err
	}

	return dsturl, nil
}
