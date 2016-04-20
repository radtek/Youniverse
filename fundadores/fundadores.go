package fundadores

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/ssoor/youniverse/log"
	"github.com/ssoor/youniverse/youniverse"
)

func getMD5(data []byte) string {
	md5Ctx := md5.New()
	md5Ctx.Write(data)
	cipherStr := md5Ctx.Sum(nil)

	return hex.EncodeToString(cipherStr)
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func downloadResourceToFile(resourceKey string, checkHash string, fileName string) (int, error) {
	var data []byte

	if err := youniverse.Get(nil, resourceKey, &data); nil != err {
		return 0, errors.New(fmt.Sprintln(resourceKey, "download failed:", err))
	}

	dataHash := getMD5(data)
	if false == strings.EqualFold(checkHash, dataHash) {
		return 0, errors.New(fmt.Sprint("check", resourceKey, "hash[", checkHash, "] failed, Unexpected hash", dataHash))
	}

	syscall.MoveFile(syscall.StringToUTF16Ptr(fileName), syscall.StringToUTF16Ptr(fileName+".del-"+strconv.Itoa(rand.Intn(10086))))
	
	file, err := os.Create(fileName)

	if nil != err {
		return 0, errors.New(fmt.Sprint(resourceKey, "create failed:", err))
	}

	defer file.Close()

	writeSize, err := file.Write(data)

	if nil != err {
		return 0, errors.New(fmt.Sprint(resourceKey, "save failed:", err))
	}

	return writeSize, nil
}

func StartFundadores(guid string, setting Settings) error {
	log.Info("Fundadores download starting, current arch is", runtime.GOARCH, ", dir is", getCurrentDirectory())

	for _, resource := range setting.Resources {
		savePath := os.ExpandEnv(resource.Save.X86.Path)
		fileSize, err := downloadResourceToFile(resource.Name, resource.Hash, savePath)

		if nil != err {
			log.Warning("Fundadores download resource failed:", err)
		} else {
			log.Info("Fundadores download", resource.Name, "to", savePath, "success, resource size is", fileSize)
		}
	}

	log.Info("Youniverse stats info:")

	log.Info("\tGET : ", youniverse.Resource.Stats.Gets.String())
	log.Info("\tLOAD : ", youniverse.Resource.Stats.Loads.String(), "\tHIT  : ", youniverse.Resource.Stats.CacheHits.String())
	log.Info("\tPEER : ", youniverse.Resource.Stats.PeerLoads.String(), "\tERROR: ", youniverse.Resource.Stats.PeerErrors.String())
	log.Info("\tLOCAL: ", youniverse.Resource.Stats.LocalLoads.String(), "\tERROR: ", youniverse.Resource.Stats.LocalLoadErrs.String())

	return nil
}
