package socksd

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ssoor/youniverse/log"
	//"github.com/ssoor/socks"
)

type PACUpdater struct {
	pac     PAC
	lock    sync.RWMutex
	data    []byte
	modtime time.Time
	timer   *time.Timer
}

func NewPACUpdater(pac PAC) (*PACUpdater, error) {
	p := &PACUpdater{
		pac: pac,
	}
	go p.backgroundUpdate()
	return p, nil
}

func parseRule(reader io.Reader) ([]string, error) {
	var err error
	var line []byte
	var rules []string
	r := bufio.NewReader(reader)
	for line, _, err = r.ReadLine(); err == nil; line, _, err = r.ReadLine() {
		s := string(line)
		if s != "" {
			rules = append(rules, s)
		}
	}
	if err == io.EOF {
		err = nil
	}
	return rules, err
}

func loadLocalRule(filepath string) ([]string, error) {
	f, err := os.OpenFile(filepath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseRule(f)
}

func loadRemoteRule(ruleURL string, upstream Upstream) ([]string, error) {
	//forward, err := BuildUpstream(upstream, socks.Direct)
	//if err != nil {
	//	return nil, err
	//}
	client := &http.Client{
		Transport: &http.Transport{
		//Dial: forward.Dial,
		},
	}
	resp, err := client.Get(ruleURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return parseRule(resp.Body)
}

func (p *PACUpdater) get() ([]byte, time.Time) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	d := make([]byte, len(p.data))
	copy(d, p.data)
	return d, p.modtime
}

func (p *PACUpdater) set(data []byte) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.data = make([]byte, len(data))
	copy(p.data, data)
	p.modtime = time.Now()
}

func (p *PACUpdater) backgroundUpdate() {
	pg := NewPACGenerator(p.pac.Rules)
	for {
		duration := 1 * time.Minute

		for pacindex, pac := range p.pac.Rules {

			if rules, err := loadLocalRule(pac.LocalRules); err == nil {
				if data, err := pg.Generate(pacindex, rules); err == nil {
					p.set(data)
					log.Info("update rules from", pac.LocalRules, "succeeded")
				}
			}

			if rules, err := loadRemoteRule(pac.RemoteRules, p.pac.Upstream); err == nil {
				if data, err := pg.Generate(pacindex, rules); err == nil {
					p.set(data)
					duration = 24 * time.Hour
					log.Info("update rules from", pac.RemoteRules, "succeeded")
				}
			}
		}

		time.Sleep(duration)
	}
}
