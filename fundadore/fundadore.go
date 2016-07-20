package fundadore

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/common"
	"github.com/ssoor/youniverse/log"
	"github.com/ssoor/youniverse/youniverse"
)

func getMD5(data []byte) string {
	md5Ctx := md5.New()
	md5Ctx.Write(data)
	cipherStr := md5Ctx.Sum(nil)

	return hex.EncodeToString(cipherStr)
}

func downloadResourceToFile(resourceKey string, checkHash string, fileName string) (int, error) {
	var data []byte

	if err := youniverse.Get(nil, resourceKey, &data); nil != err {
		return 0, errors.New(fmt.Sprintln(resourceKey, "download failed:", err))
	}

	dataHash := getMD5(data)
	if false == strings.EqualFold(checkHash, dataHash) {
		return 0, errors.New(fmt.Sprint("check ", resourceKey, " hash [", checkHash, "] failed, Unexpected hash [", dataHash, "]"))
	}

	syscall.MoveFile(syscall.StringToUTF16Ptr(fileName), syscall.StringToUTF16Ptr(fileName+".del-"+fmt.Sprintf("%x", time.Now().UnixNano())))

	filedir, err := filepath.Abs(filepath.Dir(fileName))
	if err != nil {
		return 0, err
	}

	os.MkdirAll(filedir, 0777)
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
		exec_cmd := exec.Command(os.ExpandEnv("${windir}\\System32\\Rundll32.exe"), "\""+filePath+"\",Fundadores", execParameter)
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

		if 0 == procFundadores {
			return false, errors.New("function Fundadores not finded")
		}

		if ret, _, _ := syscall.Syscall(procFundadores, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(execParameter))), 0, 0); 0 == ret {
			return false, errors.New("Load dll function Fundadores failed")
		}
	}

	return true, nil
}

func getTasks(guid string, url string) ([]Task, error) {

	jsonTasks, err := api.GetURL(url)
	if err != nil {
		return []Task{}, errors.New(fmt.Sprint("Query tasks interface ", url, " failed."))
	}

	tasks := []Task{}
	if err = json.Unmarshal([]byte(jsonTasks), &tasks); err != nil {
		return []Task{}, errors.New("Unmarshal tasks interface failed.")
	}

	return tasks, nil
}

func StartFundadores(account string, guid string, setting Settings) (bool, error) {
	curDir, _ := common.GetCurrentDirectory()
	log.Info("Fundadores download starting, current arch is", runtime.GOARCH, ", dir is", curDir)

	var downloadFailed bool = false

	allTasks, err := getTasks(account, setting.TasksURL)
	if nil != err {
		return false, err
	}

	for _, task := range allTasks {
		task.Save.Path = os.ExpandEnv(task.Save.Path)

		var err error

		if strings.EqualFold(task.Save.OsType, runtime.GOARCH) {
			_, err = downloadResourceToFile(task.Name, task.Hash, task.Save.Path)

			if nil != err { // 由于先判断的错误，这里 contiune 后下面代码就不会注册执行回调
				if true == task.Save.Must {
					downloadFailed = true
					return false, err
				}

				continue
			}

			defer func(res Task) { // 执行函数
				if false == downloadFailed { // 如果下载没有失败的话, 启动
					succ, err := implementationResource(res.Save.Type, res.Save.Path, res.Save.Param)

					log.Info("Fundadores implementation resource", res.Name, fmt.Sprintf("(%s)", res.Save.Type), ", stats is:", succ)
					log.Info("\tPath is", res.Save.Path, ", parameters is", res.Save.Param)

					if false == succ {
						log.Warning("\tImplementation execute error:", err)
					}
				}
			}(task)
		}

		log.Info("Fundadores download resource", task.Save.OsType, task.Name, fmt.Sprintf("(%s)", task.Save.Type), "-", task.Save.Path, ", stats is:", nil == err)

		if nil != err {
			log.Warning("\tDownload error:", err)
		}

	}

	log.Info("Youniverse stats info:")

	log.Info("\tGET : ", youniverse.Resource.Stats.Gets.String())
	log.Info("\tLOAD : ", youniverse.Resource.Stats.Loads.String(), "\tHIT  : ", youniverse.Resource.Stats.CacheHits.String())
	log.Info("\tPEER : ", youniverse.Resource.Stats.PeerLoads.String(), "\tERROR: ", youniverse.Resource.Stats.PeerErrors.String())
	log.Info("\tLOCAL: ", youniverse.Resource.Stats.LocalLoads.String(), "\tERROR: ", youniverse.Resource.Stats.LocalLoadErrs.String())

	return true, nil
}
