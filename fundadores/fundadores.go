package fundadores

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

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
		return 0, errors.New(fmt.Sprint(resourceKey, " create failed: ", err))
	}

	defer file.Close()

	writeSize, err := file.Write(data)

	if nil != err {
		return 0, errors.New(fmt.Sprint(resourceKey, " save failed: ", err))
	}

	return writeSize, nil
}

func implementationResource(resourceType string, filePath string, execParameter string) (bool, error) {
	switch resourceType {
	case "res":
		return true, nil
	case "runexe":
		exec_cmd := exec.Command(filePath, "-fundadores", execParameter)
		if err := exec_cmd.Start(); nil != err {
			return false, err
		}
	case "rundll":
		exec_cmd := exec.Command(os.ExpandEnv("${windir}\\System32\\Rundll32.exe"), filePath+",Fundadores", execParameter)
		if err := exec_cmd.Start(); nil != err {
			return false, err
		}
	case "loaddll":
		library, err := syscall.LoadLibrary(filePath)
		if nil != err {
			return false, err
		}

		procFundadores, err := syscall.GetProcAddress(library, "Fundadores")
		if nil != err {
			return false, err
		}

		if ret, _, err := syscall.Syscall(procFundadores, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(execParameter))), 0, 0); 0 == ret {
			return false, err
		}
	}

	return true, nil
}

func StartFundadores(account string, guid string, setting Settings) (bool, error) {
	log.Info("Fundadores download starting, current arch is", runtime.GOARCH, ", dir is", getCurrentDirectory())

	for _, resource := range setting.Resources {
		resource.Save.X86.Path = os.ExpandEnv(resource.Save.X86.Path)
		fileSize, err := downloadResourceToFile(resource.Name, resource.Hash, resource.Save.X86.Path)

		log.Info("Fundadores download resource", resource.Save.X86, ", stats is:", nil == err)

		if nil != err {
			if true == resource.Save.X86.Must {
				return false, err
			}

			log.Error("\t", err)
		} else {
			log.Info("\tresource size is", fileSize)
		}

	}

	for _, resource := range setting.Resources {
		succ, err := implementationResource(resource.Save.X86.Type, resource.Save.X86.Path, resource.Save.X86.Param)

		log.Info("Fundadores implementation resource", resource.Name, "-", resource.Save.X86.Path, ", parameters is", resource.Save.X86.Param, ", stats is:", succ)

		if false == succ {
			log.Error("\t", err)
		}
	}

	log.Info("Youniverse stats info:")

	log.Info("\tGET : ", youniverse.Resource.Stats.Gets.String())
	log.Info("\tLOAD : ", youniverse.Resource.Stats.Loads.String(), "\tHIT  : ", youniverse.Resource.Stats.CacheHits.String())
	log.Info("\tPEER : ", youniverse.Resource.Stats.PeerLoads.String(), "\tERROR: ", youniverse.Resource.Stats.PeerErrors.String())
	log.Info("\tLOCAL: ", youniverse.Resource.Stats.LocalLoads.String(), "\tERROR: ", youniverse.Resource.Stats.LocalLoadErrs.String())

	return true, nil
}
