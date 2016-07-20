package api

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"

	"github.com/ssoor/youniverse/common"

	"bytes"
	"net/http"
)

func Decrypt(base64Code []byte) (decode []byte, err error) {
	type encodeStruct struct {
		IV   string `json:"iv"`
		Code string `json:"code"`
	}

	key := []byte("890161F37139989CFA9433BAF32BDAFB")
	var jsonEninfo []byte

	for i := 0; i < len(base64Code); i++ {
		base64Code[i] = base64Code[i] - 0x90
	}

	if jsonEninfo, err = base64.StdEncoding.DecodeString(string(base64Code)); err != nil {
		return nil, err
	}

	eninfo := encodeStruct{}

	if err := json.Unmarshal(jsonEninfo, &eninfo); err != nil {
		return nil, err
	}

	var iv, code []byte

	iv, err = base64.StdEncoding.DecodeString(eninfo.IV)

	if err != nil {
		return nil, err
	}

	code, err = base64.StdEncoding.DecodeString(eninfo.Code)

	if err != nil {
		return nil, err
	}

	var block cipher.Block

	if block, err = aes.NewCipher(key); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(code, code)

	return code, nil
}

func GetURL(srcurl string) (string, error) {
	var data []byte
	resp, err := http.Get(srcurl)

	if nil != err {
		return "", err
	}

	defer resp.Body.Close()

	var bodyBuf bytes.Buffer

	bodyBuf.ReadFrom(resp.Body)

	data, err = Decrypt(bodyBuf.Bytes())

	if err != nil {
		return "", err
	}

	return common.GetValidString(data), nil
}
