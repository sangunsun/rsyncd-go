# rsyncd-go
"Real-time monitoring of multiple folders and immediate rsync synchronization of changed files to address the issue of low efficiency with rsync and inotify-tools when there are many files."

实时监控多文件夹，并对其中变动的文件立即进行rsync同步，解决文件较多时rsync+inotify-tools效率低下问题

## 功能概述
+ 本程序为解决自动调用rsync同步文件而开发， 
+ 本程序解决了rsync+inotify tools，每次都要检查完整源文件夹和目标文件夹,造成计算资源和通讯资源浪费，效率低的问题  
+ 本程序实时监控多个文件夹，并在被监控的某个文件变动时立即调用rsync同步变动文件到事先设置好的备份文件夹中  
+ 本程序运行时会先把所有被监控文件夹进行一次rsync同步，之后进行监入状态，每次只同步一个发生变化的文件  
+ 使用本程序需先安装好rsync  
+ 如果要通过ssh传输给远端目标，还需要先安装好sshpass,远端机器的ssh密码请放入一个单独的文件中并做好权限管理，具体做法参见sshpass相关文档。  
+ 使用前需要确保每个源文件夹是存在的，否则程序会退出，目标文件夹如果不存在rsync会自动创建  

## 使用步骤
1. 规划好同步任务：确定要监控的源文件夹和目标文件夹，一个任务只能有一个目标文件夹，但可以有多个源文件夹，当然多个任务也可以是同一个目标文件夹。  
2. 根据同步任务规划，在配置文件中输入源文件夹和目标文件夹的信息  
3. 根据同步任务特点，在配置文件中输入各任务需要的rsync特殊参数  
4. 如果要同步到远程服务器，还需要先安装好sshpass软件，并将远程服务器对应用户的密码存入一个密码文件中，一个任务一个密码文件，文件中只存密码一个数据，然后根据示例配置文件，配置好rsync访问远程服务器的参数。  
5. 运行本程序，本程序运行时，会先调用rsync做一次数据同步，视被监控文件的多少可能会花费一些时间，之后本程序进行持续监控并自动同步状态。  
