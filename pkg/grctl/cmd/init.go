// RAINBOND, Application Management Platform
// Copyright (C) 2014-2017 Goodrain Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd
import (
	"github.com/urfave/cli"
	"github.com/Sirupsen/logrus"
	"os/exec"
	"net/http"
	"io/ioutil"
	"strings"
	"bytes"
	"github.com/goodrain/rainbond/pkg/grctl/clients"
	"fmt"
	//"time"

	"encoding/json"
	"github.com/bitly/go-simplejson"
)

func NewCmdInit() cli.Command {
	c:=cli.Command{
		Name:  "init",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "etcd",
				Usage: "etcd ip,127.0.0.1",
			},
			cli.StringFlag{
				Name:  "type",
				Usage: "node type:manage/compute, manage",
			},
			cli.StringFlag{
				Name:  "mip",
				Usage: "当前节点内网IP, 10.0.0.1",
			},
			cli.StringFlag{
				Name:  "repo_ver",
				Usage: "repo version,3.4",
			},
			cli.StringFlag{
				Name:  "install_type",
				Usage: "online/offline ,online",
			},
		},
		Usage: "初始化集群。grctl init cluster",
		Action: func(c *cli.Context) error {
			return initCluster(c)
		},
	}
	return c
}





// grctl exec POD_ID COMMAND
func initCluster(c *cli.Context) error {
	//logrus.Infof("start init command")
	resp, err := http.Get("http://repo.goodrain.com/gaops/jobs/install/prepare/init.sh")

	//参数
	//$1 -- ETCD_NODE  eg: 127.0.0.1 ETCD IP
	//$2 -- NODE_TYPE  eg: manage/compute 默认 manage
	//$3 -- MIP eg: 10.0.0.1 当前机器ip
	//$4 -- REPO_VER eg: 3.4 默认3.4
	//$5 -- INSTALL_TYPE eg: online 默认online
	//若不传参数则表示
	//
	//默认为管理节点 在线安装3.4版本的etcd
	if err != nil {
		logrus.Errorf("error get init script,details %s",err.Error())
		return err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	args:=[]string{c.String("etcd"),c.String("type"),c.String("mip"),c.String("repo_ver"),c.String("install_type")}
	arg:=strings.Join(args," ")
	argCheck:=strings.Join(args,"")
	if len(argCheck) > 0 {
		arg+=";"
	}else {
		arg=""
	}
	//logrus.Infof("args is %s,len is %d",arg,len(arg))
	fmt.Println("开始初始化集群")
	cmd := exec.Command("bash", "-c",arg+string(b))
	buf:=bytes.NewBuffer(nil)
	cmd.Stderr=buf
	cmd.Run()
	out:=buf.String()
	arr:=strings.SplitN(out,"{",2)
	outJ:="{"+arr[1]
	jsonStr:=strings.TrimSpace(outJ)
	jsonStr=strings.Replace(jsonStr,"\n","",-1)
	jsonStr=strings.Replace(jsonStr," ","",-1)
	logrus.Infof(jsonStr)
	fixedJ,_:=json.Marshal(jsonStr)

	//
	js,err:=simplejson.NewJson(fixedJ)
	if err != nil {
		logrus.Errorf("error decode json,details %s",err.Error())
		return nil
	}
	//
	global:=js.Get("global").Get("OS_VER")
	if err != nil {
		logrus.Errorf("error decode status json,details %s",err.Error())
		return nil
	}
	initStatusB,_:=json.Marshal(global)
	fmt.Println("========"+string(initStatusB))
	//
	//
	//
	//
	//fmt.Println("初始化结果：")
	//for _,v:=range global{
	//	b,_:=json.Marshal(v)
	//	statusJ,err:=simplejson.NewJson(b)
	//	if err != nil {
	//		logrus.Errorf("error decode status,details %s",err.Error())
	//		return nil
	//	}
	//	task,_:=statusJ.Get("name").String()
	//	condition,_:=statusJ.Get("condition_status").String()
	//	fmt.Printf("task:%s install %s",task,condition)
	//	fmt.Println()
	//}
	err=clients.NodeClient.Tasks().Get("check_manage_base_services").Exec([]string{})
	if err != nil {
		logrus.Errorf("error execute task %s","check_manage_base_services")
	}
	Status("check_manage_base_services")

	err=clients.NodeClient.Tasks().Get("check_manage_services").Exec([]string{})
	if err != nil {
		logrus.Errorf("error execute task %s","check_manage_services")
	}
	Status("check_manage_services")
	//checkFail:=0
	//for checkFail<3  {
	//	time.Sleep(3*time.Second)
	//	status,err:=clients.NodeClient.Tasks().Get("install_manage_ready").Status()
	//	if err != nil {
	//		checkFail+=1
	//		logrus.Errorf("error get task status ,details %s",err.Error())
	//		continue
	//	}
	//	checkFail=0
	//	for k,v:=range status.Status{
	//		if v.Status!="complete" {
	//			fmt.Printf(".")
	//			continue
	//		}else {
	//			fmt.Printf("task %s is %s-----%s",k,v.Status,v.CompleStatus)
	//			return nil
	//		}
	//	}
	//}
	//一般 job会在通过grctl执行时阻塞输出，这种通过 脚本执行的，需要单独查
	return nil
}

