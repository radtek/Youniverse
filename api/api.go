package api

import (
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"io"

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

	var zipReader io.ReadCloser
	if zipReader, err = zlib.NewReader(bytes.NewBuffer(base64Code)); nil != err {
		return nil, err
	}

	codeBuff := bytes.NewBuffer(nil)
	if _, err := io.Copy(codeBuff, zipReader); nil != err {
		zipReader.Close()
		return nil, err
	}

	zipReader.Close()
	if jsonEninfo, err = base64.StdEncoding.DecodeString(codeBuff.String()); err != nil {
		return nil, err
	}

	eninfo := encodeStruct{}

	if err := json.Unmarshal(jsonEninfo, &eninfo); err != nil {
		return nil, err
	}

	var iv, tempBuff []byte

	iv, err = base64.StdEncoding.DecodeString(eninfo.IV)

	if err != nil {
		return nil, err
	}

	tempBuff, err = base64.StdEncoding.DecodeString(eninfo.Code)

	if err != nil {
		return nil, err
	}

	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(tempBuff, tempBuff)

	if zipReader, err = zlib.NewReader(bytes.NewBuffer(tempBuff)); nil != err {
		return nil, err
	}

	defer zipReader.Close()

	codeBuff.Reset()
	if _, err := io.Copy(codeBuff, zipReader); nil != err {
		return nil, err
	}

	return codeBuff.Bytes(), nil
}

func GetURL(srcurl string) (decodeData string, err error) {
	var data []byte

	var resp *http.Response
	for i := 0; i < 3; i++ {
		if resp, err = http.Get(srcurl); nil != err {
			continue
		}

		break
	}
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var bodyBuf bytes.Buffer

	bodyBuf.ReadFrom(resp.Body)

	data, err = Decrypt(bodyBuf.Bytes())
	if err != nil {
		return "", err
	}

	//log.Info("API <", srcurl, ">", common.GetValidString(data))
	return common.GetValidString(data), nil
}
