package main

import (
	"fmt"
	"log"
    "flag"
	"strings"
	"os/exec"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var configFile=flag.String("cf","config.toml","配置文件名")

type task struct{
    IsTemp bool `default:"false"`
    Source []string
    Target string
    Para string //rsync 附加参数
}
type config struct{
    RsyncPath string
    Para string  //该任务中rsync额外的附加参数
    Task []task 
}
type rsyncdTask struct{
    Task  task
    Watcher *fsnotify.Watcher
}

func init(){

    flag.Parse()

}
func getConfig(fileName string)bool{
    viper.SetConfigName(fileName)
    viper.SetConfigType("toml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("/")
    err:=viper.ReadInConfig()
    if err != nil {
        // 处理错误
        panic(fmt.Errorf("Fatal error config file: %s \n", err))
    }
    return true
}

func main() {

    getConfig(*configFile)

    var CC config
    err:=viper.Unmarshal(&CC)
    if err != nil {
        // 处理错误
        panic(fmt.Errorf("Fatal error config file: %s \n", err))
    }

    var rTask []*rsyncdTask 
    for _,task := range CC.Task{

        if task.IsTemp == true{
            continue
        }

        watcherTmp,_:=fsnotify.NewWatcher()
        var rTaskTmp= new( rsyncdTask)
        rTaskTmp.Task=task
        rTaskTmp.Watcher=watcherTmp

        for _,source:=range task.Source{
            err = watcherTmp.Add(source)
            if err != nil {
                log.Fatal("add source err",err,source)
            }

        }
        rTask=append(rTask,rTaskTmp)
        go startWatch(rTaskTmp)
    }
    for _,rt := range rTask{
        defer rt.Watcher.Close()
    }

    // Block main goroutine forever.
    <-make(chan struct{})
}

func startWatch(rtask *rsyncdTask) {
    watcher:=rtask.Watcher 
    startSync(rtask)
    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                return
            }

            log.Println("event:", event,event.Name)

            //同步一个文件
            if event.Has(fsnotify.Write|fsnotify.Create|fsnotify.Chmod) {

                //拼接传输的目标文件名字,分两种情况，一种是源文件夹以/结束，另一种是不以/结束
                //这里的event.Name是带相对路径的文件名，target+event.Name就是其在目标机器上的存取路径名
                var shortName string
                for _,sourTmp:= range rtask.Task.Source{
                    if strings.HasPrefix(event.Name,sourTmp){
                        shortName = strings.Replace(event.Name,sourTmp,"",1)

                        //如果源目录是目录名，则短文件名需要带最后的目录，否则短文件名就是文件名本身。
                        if sourTmp[len(sourTmp)-1:]!="/"{

                            tpath:=strings.Split(sourTmp,"/")
                            shortName= tpath[len(tpath)-1]+shortName
                        }
                        continue
                    }

                }

                cmdStr := "rsync " + " "+rtask.Task.Para+" "+ event.Name+ " "+rtask.Task.Target+shortName
                log.Println(cmdStr)
                cmd:=exec.Command("bash","-c",cmdStr)
                output,outputErr:=cmd.CombinedOutput()
                if outputErr!=nil{
                    log.Println("run rsync error:",outputErr)
                }else{
                    log.Println("cmd output:",string(output))
                }
            } 
        case err, ok := <-watcher.Errors:
            if !ok {
                return
            }
            log.Println("error:", err)
        }
    }
}

func startSync(rtask *rsyncdTask)bool {

    var sourceStr string
    for _,sourceTmp := range rtask.Task.Source{
        sourceStr = sourceStr+" "+ sourceTmp
    }

    cmdStr := "rsync " + " "+rtask.Task.Para+" "+ sourceStr+ " "+rtask.Task.Target
    log.Println(cmdStr)
    cmd:=exec.Command("bash","-c",cmdStr)

    output,outputErr:=cmd.CombinedOutput()
    if outputErr!=nil{
        log.Println("run rsync error:",outputErr)
    }else{
        log.Println("cmd output:",string(output))
    }
    return false
}
