package main

import (
	"fmt"
	"log"
    "flag"
	"strings"
	"os/exec"
	"github.com/spf13/viper"
    "github.com/rjeczalik/notify"
)

var configFile=flag.String("cf","config.toml","配置文件名")

//配置文件是一个数组，每个任务是一个四元组
//该struct用于读取配置文件的一个表数组元素
type Task struct{
    IsTemp bool `default:"false"`
    Source []string
    Target string
    Para string //rsync 附加参数
    events chan notify.EventInfo
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

    var tasks []Task  //从配置文件读取的任务
    var rtasks []Task //保存需要执行的任务
    err:=viper.UnmarshalKey("task",&tasks)
    if err != nil {
        // 处理错误
        panic(fmt.Errorf("Fatal error config file: %s \n", err))
    }
//给每个用户加上事件监控器，并向监控管理器notify申请监控任务
    for _,task := range tasks{

        if task.IsTemp == true{

            continue
        }

        task.events= make(chan notify.EventInfo, 4)

        //格式化目标文件夹格式为统一以符号/结束
        if task.Target[len(task.Target)-1:]!="/"{
            task.Target=task.Target+"/"
        }
        //申请监控任务
        for _,source :=range task.Source{

            //是否递归根据配置文件源文件夹最后三个字符是否为...决定 
            if err := notify.Watch(source, task.events, notify.InCloseWrite, notify.InMovedTo); err != nil {
                panic(fmt.Errorf("Fatal error add source path : %s \n", err))
            }
        }
        rtasks= append(rtasks,task)
    }


    //启动监控
    for _,task := range rtasks{
        go startWatch(task)

    }

    // Block main goroutine forever.
    <-make(chan struct{})
}

func startWatch(task Task) {
    startSync(task)
    for {
        switch event := <-task.events;event.Event() {
        case notify.InCloseWrite,notify.InMovedTo :

            log.Println("event:", "Editing of", event.Path(), "file is done.")

            //拼接传输的目标文件名字,分两种情况，一种是源文件夹以/结束，另一种是不以/结束
            //方法是,先获得变动文件相对于被监控文件夹的相对文件名shortName
            //,再使用目录文件夹名和shortName组合成最终的目标文件名。最终得到的shorName第一次字符不是"/"
            // linux中目录中间的分隔符"/"，可以是多条如 "cd go///rsyncd-go",不会有问题
            //这里的event.Name是带相对路径的文件名，target+event.Name就是其在目标机器上的存取路径名
            var shortName string
            for _,sourTmp:= range task.Source{
                //如果是递归监控，要去掉源文件夹最后的三个...
                if sourTmp[len(sourTmp)-3:]=="..."{
                    sourTmp=sourTmp[:len(sourTmp)-3]
                }
                //------------------------------------------

                if strings.HasPrefix(event.Path(), sourTmp){
                    shortName = strings.Replace(event.Path(),sourTmp,"",1)

                    //如果源目录是目录名，则短文件名需要带最后的目录，否则短文件名就是文件名本身。
                    if sourTmp[len(sourTmp)-1:]!="/"{

                        tpath:=strings.Split(sourTmp,"/")
                        shortName= tpath[len(tpath)-1]+shortName
                    }
                    //如果源目录数组中某个元素是变动的文件名前缀，则已经获取到本次变动文件的短文件名
                    //则马上跳出循环去执行同步
                    break
                }

            }

            cmdStr := "rsync " + " "+task.Para+" "+ event.Path()+ " "+task.Target+shortName
            log.Println(cmdStr)
            cmd:=exec.Command("bash","-c",cmdStr)
            output,outputErr:=cmd.CombinedOutput()
            if outputErr!=nil{
                log.Println("run rsync error:",outputErr)
            }else{
                log.Println("cmd output:",string(output))
            }
         default:
            log.Println("some event happend")

        }
    }
}

func startSync(task Task)bool {

    log.Println("in start sync",task)
    var sourceStr string
    for _,sourceTmp := range task.Source{

        if sourceTmp[len(sourceTmp)-3:]=="..."{
            sourceTmp=sourceTmp[:len(sourceTmp)-3]
        }
        sourceStr = sourceStr+" "+ sourceTmp
    }

    cmdStr := "rsync " + " "+task.Para+" "+ sourceStr+ " "+task.Target
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
