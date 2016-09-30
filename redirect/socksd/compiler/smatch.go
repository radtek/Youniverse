package compiler

import (
	"errors"
	"strings"

	"github.com/dlclark/regexp2"
)

type SMatch struct {
	template     string
	contextRegex *regexp2.Regexp
}

var (
	ErrUnrecognizedSMatch = errors.New("smatch resolution fails, unrecognized smatch expression")
)

func NewSMatch(rule string) (match SMatch, err error) {

	if rule[0] != 's' && rule[0] != 'S' {
		return match, ErrUnrecognizedSMatch // errors.New("invalid rule head: " + rule)
	}

	if rule[1] != '@' && rule[0] != '|' {
		return match, ErrUnrecognizedSMatch // errors.New("invalid character segmentation: " + rule)
	}

	split := strings.Split(rule, rule[1:2])

	if len(split) != 4 {
		return match, ErrUnrecognizedSMatch // errors.New("rule string incomplete or invalid: " + rule)
	}

	match.contextRegex, err = regexp2.Compile("(?"+split[3]+")"+split[1], 0)

	if err != nil {
		return match, ErrUnrecognizedSMatch // err
	}

	match.template = split[2]

	return match, nil
}

func (s *SMatch) Replace(src string) (string, error) {
	if isMatch, _ := s.contextRegex.MatchString(src); false == isMatch { // 当出错时，返回 false
		return src, errors.New("regular expression does not match")
	}

	return s.contextRegex.Replace(src, s.template, 0, 99999)
	//return s.contextRegex.ReplaceAllString(src, s.template), nil // 此函数无法获得本次正则是否匹配
	//submatch := s.contextRegex.FindStringSubmatchIndex(src)

	//if len(submatch) == 0 {
	//	return src, errors.New("regular expression does not match")
	//}

	//var dst []byte
	//return string(s.contextRegex.ExpandString(dst, s.template, src, submatch)), nil
}
