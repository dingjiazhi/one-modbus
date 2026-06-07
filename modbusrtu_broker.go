package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"image/color"
	"io"
	"math"
	"math/rand"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/go-basic/uuid"
	"github.com/go-gomail/gomail"
	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lucasb-eyer/go-colorful"
	_ "github.com/mattn/go-sqlite3" // 导入SQLite驱动
	"github.com/shirou/gopsutil/process"
	"github.com/shirou/gopsutil/v3/mem"
	modbus "github.com/thinkgos/gomodbus"
	"github.com/xuri/excelize/v2"
	"golang.org/x/image/font/opentype"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/simplifiedchinese"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/text"
	"gonum.org/v1/plot/vg"
)

const (
	编译时间        = "编译时间: 2026-06-07 12:14"
	程序文件最大限制    = 50 << 20                 //设定错误将无法上传更新软件
	允许同名程序名同时运行 = false                    //尽量禁止，因为若多个同名软件运行，如果跟本软件同样功能的，很多默认文件比如数据库信息文件都一样了，不好
	///////////////////////////////////
	检查token是否失效使用的默认openid     = "okxRK5n3XyOtpmo-3yak3JIWxG7U"
	开发者                        = "\r\n开发者: 158148415@qq.com"
	SQLiteDB命令允许               = true //直接执行查询命令
	http服务器写超时秒                = 600  //SQLiteDB查询生成曲线图，如果很复杂则耗时很长，如果写超时太短，浏览器会反复请求，设置长，使用人会等不耐烦自己关闭浏览器请求
	SQLiteDB查询最大容忍时长秒          = 5
	SQLiteDB使用源库平均查询每记录耗时默认us  = 100 //避免初始化为0，导致可以访问海量记录，由历史经验设定
	SQLiteDB使用内存库平均查询每记录耗时默认us = 100 //避免初始化为0，导致可以访问海量记录，由历史经验设定
	SQLiteDB查询曲线图时间刻度数         = 40
	SQLiteDb                   = "SQLiteDB"
	SQLite表格名                  = "SQLite"
	每个变量目标记录数与变量实际记录数的倍数最大设定   = 2                //如果不这样做，一旦有人估计查询10年1秒间隔，那将会生成海量多个开始时间结束时间，内存爆满
	每个变量目标记录数与变量实际记录数的倍数最大设定s  = "2"              //如果不这样做，一旦有人估计查询10年1秒间隔，那将会生成海量多个开始时间结束时间，内存爆满
	SQLiteDB查询曲线图长度cm          = 50.8124648254299 //目标1920分辨率宽度
	SQLiteDB运行时备份间隔秒数          = 3600             ///调试可改30秒默认3600//都没有用，之间修改函数内参数
	生成SQLite记录间隔秒数             = 600              ///调试可改10秒默认600//都没有用，之间修改函数内参数
	最大生成SQLite记录间隔秒数           = 3600
	默认最小采集频率毫秒                 = 1000
	默认最大采集频率毫秒                 = 生成SQLite记录间隔秒数 * 1000 //影响SQLiteDB写周期
	企业微信报警系统状态推送最小间隔秒数         = 15                    //同一个ip电脑发信息太频繁会被腾讯限制
	企业微信报警系统状态推送间隔秒数           = 3600
	微信报警字段最大字符数                = 20
	本软件内存占用率最大值                = 86.25   //修改时注意要改两个地方
	本软件内存占用率最大值_s              = "86.25" //修改时注意要改两个地方
	内存百分比设定                    = 99      //修改时注意要改两个地方
	内存百分比设定_s                  = "99"    //修改时注意要改两个地方
	内存百分比连续超过设定次数限制            = 5       //修改时注意要改两个地方
	内存百分比连续超过设定次数限制_s          = "5"     //修改时注意要改两个地方
	开启文件服务器                    = true
	登录保持秒数                     = 30   //修改时注意要改两个地方
	登录保持秒数_s                   = "30" //修改时注意要改两个地方
	发邮件时间间隔秒数                  = 15
	上传时间间隔秒数                   = 15
	自动发邮件时间间隔秒数                = 1
	检查上网状态时间间隔秒数               = 10
	发送人                        = "158148415@qq.com"
	发邮件token                   = "" // 请通过 Excel 配置表「设定」页填写
	默认邮件接收者                    = "158148415@qq.com"
	全部项目代号                     = "全部项目代号"
	变量ID初始值                    = 21    //组态王自建变量id从21开始
	运行频率                       = 1     //毫秒
	最大读请求变量个数                  = 128   //修改时注意要改两个地方
	最大读请求变量个数_s                = "128" //修改时注意要改两个地方
	最大写请求变量个数                  = 32    //修改时注意要改两个地方
	最大写请求变量个数_s                = "32"  //修改时注意要改两个地方
	可上传在服务器运行的批处理文件名           = "server.bat"
	项目变量信息表名1                  = "项目变量信息.xlsx"
	项目变量信息表名备份                 = "项目变量信息备份.xlsx"
	项目变量信息表文件最大限制              = 1 << 20
	项目变量信息表文件最小限制              = 1 << 1
	//程序文件最大限制              = 15 << 20
	程序文件最小限制              = 12 << 20
	可上传在服务器运行的批处理文件最大限制   = 1 << 10
	可上传在服务器运行的批处理文件最小限制   = 1 << 1
	默认项目代号                = "A09"
	项目变量信息内存表保存间隔         = 30
	紧急遍历项目变量信息表时间间隔       = 60
	工作表名                  = "基本变量页"
	设定工作表名                = "设定"
	常用设定值表名               = "常用设定值"
	开发说明表名                = "使用说明"
	最小端口号                 = 1024    //修改时注意要改两个地方
	最小端口号_s               = "1024"  //修改时注意要改两个地方
	最大端口号                 = 65535   //修改时注意要改两个地方
	最大端口号_s               = "65535" //修改时注意要改两个地方
	项目变量信息表格保存多少行需要1秒时间间隔 = 300
	最小变量名长度               = 5
	最大变量名长度               = 31
	最大浮点数                 = 3.4e38    //修改时注意要改两个地方
	最大浮点数_s               = "3.4e38"  //修改时注意要改两个地方
	最小浮点数                 = -3.4e38   //修改时注意要改两个地方
	最小浮点数_s               = "-3.4e38" //修改时注意要改两个地方
	默认通讯异常值_FLOAT_s       = "-9E38"
	内存最大浮点数               = 1.7976931348623157e308    //修改时注意要改两个地方
	内存最大浮点数_s             = "1.7976931348623157e308"  //修改时注意要改两个地方
	内存最小浮点数               = -1.7976931348623157e308   //修改时注意要改两个地方
	内存最小浮点数_s             = "-1.7976931348623157e308" //修改时注意要改两个地方
	内存默认通讯异常值_FLOAT_s     = "-9E308"
	内存整型最小值               = -9223372036854775808
	内存整型最大值               = 9223372036854775807
	内存整型最小值_s             = "-9223372036854775808"
	内存整型最大值_s             = "9223372036854775807"
	内存默认通讯异常值_LONG_s      = "-9999999999999999999"
	默认端口号_s               = "53046" //修改时注意要改两个地方
	默认端口号                 = 53046   //修改时注意要改两个地方
	默认串口号                 = "246"
	默认设备地址_s              = "247"    //修改时注意要改两个地方
	默认设备地址                = 247      //修改时注意要改两个地方
	默认波特率_s               = "115200" //修改时注意要改两个地方
	默认波特率                 = 115200   //修改时注意要改两个地方
	默认数据位_s               = "8"      //修改时注意要改两个地方
	默认数据位                 = 8        //修改时注意要改两个地方
	默认奇偶校验                = "N"
	默认停止位_s               = "1" //修改时注意要改两个地方
	默认停止位                 = 1   //修改时注意要改两个地方
	默认字打包长度_s             = "29"
	默认位打包长度_s             = "1"
	默认打包长度                = 2
	默认读功能码_s              = "3"
	默认读写字节序               = "ABCD"
	默认采集前等待_毫秒_s          = "0"     //修改时注意要改两个地方
	默认采集前等待_毫秒            = 0       //修改时注意要改两个地方
	默认采集频率_毫秒_s           = "5000"  //修改时注意要改两个地方
	默认采集频率_毫秒             = 5000    //修改时注意要改两个地方
	默认通讯超时_毫秒_s           = "1000"  //修改时注意要改两个地方
	默认通讯超时_毫秒             = 1000    //修改时注意要改两个地方
	默认记录到数据库间隔_秒_s        = "3600"  //修改时注意要改两个地方
	默认记录到数据库间隔_秒          = 3600    //修改时注意要改两个地方//2万个变量每小时记录一次一年将近2亿条记录左右
	默认连续通讯失败多少次则认为通讯异常_s  = "3"     //修改时注意要改两个地方
	默认连续通讯失败多少次则认为通讯异常    = 3       //修改时注意要改两个地方
	默认寄存器地址_s             = "65506" //修改时注意要改两个地方
	默认寄存器地址               = 65506   //修改时注意要改两个地方
	默认批量读开始寄存器地址          = 默认寄存器地址
	默认多少变化率才记录到数据库        = 0.25
	默认通讯异常值_LONGBCD_s     = "-1"
	默认通讯异常值_LONG_s        = "-9999999999"
	int32最小值              = -2147483648   //修改时注意要改两个地方
	int32最小值_s            = "-2147483648" //修改时注意要改两个地方
	int32最大值              = 2147483647    //修改时注意要改两个地方
	int32最大值_s            = "2147483647"  //修改时注意要改两个地方
	默认通讯异常值_ULONG_s       = "-1"
	默认通讯异常值_BCD_s         = "-1"
	默认通讯异常值_SHORT_s       = "-99999"
	默认通讯异常值_USHORT_s      = "-1"
	默认通讯异常值_BIT_s         = "-1"
	默认通讯异常值_STRING        = "www.pdlei.cn"
	默认小数位_s               = "2"   //修改时注意要改两个地方
	默认小数位数                = 2     //修改时注意要改两个地方
	最大字符串字节长度             = 128   //修改时注意要改两个地方
	最大字符串字节长度_s           = "128" //修改时注意要改两个地方
	内存字符串最多占用字节数          = 384
	内存字符串最多占用字节数_s        = "384"
	最小串口号                 = 1     //修改时注意要改两个地方
	最小串口号_s               = "1"   //修改时注意要改两个地方
	最大串口号                 = 255   //修改时注意要改两个地方
	最大串口号_s               = "255" //修改时注意要改两个地方
	最小设备地址                = 1     //修改时注意要改两个地方
	最小设备地址_s              = "1"   //修改时注意要改两个地方
	最大设备地址                = 247   //修改时注意要改两个地方
	最大设备地址_s              = "247" //修改时注意要改两个地方
	最大小数位                 = 4     //修改时注意要改两个地方
	最大小数位_s               = "4"   //修改时注意要改两个地方
	波特率01                 = 110
	波特率02                 = 300
	波特率03                 = 600
	波特率04                 = 1200
	波特率05                 = 2400
	波特率06                 = 4800
	波特率07                 = 9600
	波特率08                 = 14400
	波特率09                 = 19200
	波特率10                 = 28800
	波特率11                 = 33600
	波特率12                 = 38400
	波特率13                 = 56000
	波特率14                 = 57600
	波特率15                 = 115200
	最大字打包长度               = 125     //修改时注意要改两个地方
	最大字打包长度_s             = "125"   //修改时注意要改两个地方
	最大位打包长度               = 2000    //修改时注意要改两个地方//125个字（0xFA）250个字节2000位
	最大位打包长度_s             = "2000"  //修改时注意要改两个地方
	最小时间                  = 0       //修改时注意要改两个地方
	最小时间_s                = "0"     //修改时注意要改两个地方
	最大时间                  = 60000   //修改时注意要改两个地方
	最大时间_s                = "60000" //修改时注意要改两个地方
	最小寄存器地址               = 0       //修改时注意要改两个地方
	最小寄存器地址_s             = "0"     //修改时注意要改两个地方
	最大寄存器地址               = 65534   //修改时注意要改两个地方
	最大寄存器地址_s             = "65534" //modbus地址最大表示465535或者365535，所以寄存器地址最大只能65534
	最大读写字符个数_s            = "128"
	最小计算值除原始值             = 0.000001   //修改时注意要改两个地方
	最小计算值除原始值_s           = "0.000001" //修改时注意要改两个地方
	最大计算值除原始值             = 1000000    //修改时注意要改两个地方
	最大计算值除原始值_s           = "1000000"  //修改时注意要改两个地方
	标题列数                  = 31         //注意修改！！！！！！！！
	变量ID_下标               = 0
	变量名_下标                = 1
	变量类型_下标               = 2
	初始值_下标                = 3
	通讯异常值_下标              = 4
	是否保存值_下标              = 5
	计算值除原始值_下标            = 6
	小数位数_下标               = 7
	串口号_下标                = 8
	波特率_下标                = 9
	奇偶校验_下标               = 10
	数据位_下标                = 11
	停止位_下标                = 12
	通讯超时_下标               = 13
	设备地址_下标               = 14
	读功能码_下标               = 15
	寄存器地址_下标              = 16
	多少变化率才记录到数据库_下标       = 17
	打包长度_下标               = 18
	允许通讯异常后只读变量_下标        = 19
	数据类型_下标               = 20
	读写属性_下标               = 21
	采集频率_下标               = 22
	采集前等待_下标              = 23
	连续通讯失败多少次则认为通讯异常_下标   = 24
	读字节序_下标               = 25
	写字节序_下标               = 26
	错误信息_下标               = 27
	警告信息_下标               = 28
	记录到数据库间隔_秒_下标         = 29
	单位_下标                 = 30
)

var 默认使用者, 使用者 string
var 自定义邮件接收者 = ""
var 自定义发送人 = ""
var 自定义发邮件token = ""
var webhook_url_默认 = ""
var 登录名 = ""
var 登录密码 = ""
var 允许网关启动发邮件 = true
var 允许关闭网关发邮件 = true
var 允许重启网关发邮件 = true
var 允许重启服务器发邮件 = true
var 允许更新项目变量表发邮件 = true
var 允许更新网关软件发邮件 = true
var 允许执行自定义批处理发邮件 = true
var 允许访问IP更新发邮件 = true
var 标题列号 [标题列数]int
var 遍历项目变量信息表中 bool = true
var 完成第一次遍历项目变量信息表 bool
var 遍历项目变量信息表结果 string
var 启动软件碰到的问题 字符串不重复累加互斥锁访问结构体
var 本网关UUID string

type Row struct {
	X序号               int           `json:"序号"`
	B变量名称             string        `json:"变量名称"`
	B变量类型             string        `json:"变量类型"`
	S数据类型             string        `json:"数据类型"`
	D当前值              string        `json:"当前值"`
	D当前值锁             sync.RWMutex  `json:"当前值锁"`
	C采集前等待毫秒          time.Duration `json:"采集前等待毫秒"`
	L连续通讯失败次数         uint32        `json:"连续通讯失败次数"`
	L连续通讯失败多少次则认为通讯异常 uint32        `json:"连续通讯失败多少次则认为通讯异常"`
	D读字节序             string        `json:"读字节序"`
	X写字节序             string        `json:"写字节序"`
	X项目代号             string        `json:"项目代号"`
	D读写属性             string        `json:"读写属性"`
	C初始值              string        `json:"初始值"`
	C采集频率毫秒           int64         `json:"采集频率毫秒"`
	C采集时刻毫秒           int64         `json:"采集时刻毫秒"`
	C采集成功时刻           int64         `json:"采集成功时刻"`
	C错误信息             string        `json:"错误信息"`
	T通讯异常值            string        `json:"通讯异常值"`
	S是否保存值            string        `json:"是否保存值"`
	J计算值除原始值          float64       `json:"计算值除原始值"`
	X小数位数             int           `json:"小数位数"`
	X小数点后值            int           `json:"小数点后值"`
	C串口号              string        `json:"串口号"`
	B波特率              int           `json:"波特率"`
	J奇偶校验             string        `json:"奇偶校验"`
	S数据位              int           `json:"数据位"`
	T停止位              int           `json:"停止位"`
	T通讯超时毫秒           time.Duration `json:"通讯超时毫秒"`
	S设备地址             byte          `json:"设备地址"`
	D读功能码             string        `json:"读功能码"`
	J寄存器地址            uint16        `json:"寄存器地址"`
	J寄存器地址_s          string        `json:"寄存器地址_s"`
	P批量读开始寄存器地址       uint16        `json:"批量读开始寄存器地址"`
	D多少变化率才记录到数据库     float64       `json:"多少变化率才记录到数据库"`
	D单位               string        `json:"单位"`
	D打包长度             uint16        `json:"打包长度"`
	Y允许通讯异常后只读变量      string        `json:"允许通讯异常后只读变量"`
	T记录到数据库间隔_秒       int64         `json:"记录到数据库间隔_秒"`
	S上次生成记录时刻秒        int64         `json:"上次生成记录时刻秒"`
	S上次生成记录时刻毫秒       int64         `json:"上次生成记录时刻毫秒"`
	S上次生成记录的变量值       string        `json:"上次生成记录的变量值"`
	X需要保存数据           int           `json:"需要保存数据"`
	P批量读伙伴            []int         `json:"批量读伙伴"`
}
type 项目变量信息表组2 struct {
	Rows []Row
}

var 项目变量信息表组 项目变量信息表组2
var 串口组 map[string]bool

type KVTag struct {
	//VarValue   interface{} `json:"VarValue"`
	VarValue   string       `json:"VarValue"`
	VarValue锁  sync.RWMutex `json:"VarValue锁"`
	NVarID     int          `json:"nVarID"`
	NVarType   string       `json:"nVarType"`
	StrVarName string       `json:"strVarName"`
}
type 变量组1 struct {
	KVTags []KVTag
}

var 变量组 变量组1

func 杀掉同名程序() {
	if 允许同名程序名同时运行 {
		return
	}
	var 上个进程创建时间 int64 = 0
	var 上个进程结构指针 *process.Process
	str := ""
	进程指针组, _ := process.Processes()
	for _, 进程指针 := range 进程指针组 {
		str, _ = 进程指针.Name()
		if str == 程序名 {
			i, _ := 进程指针.CreateTime()
			if 上个进程创建时间 <= 0 {
				上个进程结构指针 = 进程指针
				上个进程创建时间 = i
			} else {
				if i > 上个进程创建时间 {
					_ = 上个进程结构指针.Kill()
					上个进程结构指针 = 进程指针
					上个进程创建时间 = i
				} else {
					if i < 上个进程创建时间 {
						_ = 进程指针.Kill()
					}
				}
			}
		}
	}
} //func 杀掉同名程序() {
func 外网ip访问判断(访问连接 string) (bool, string) {
	if strings.Contains(访问连接, "[::1]") {
		return false, ""
	}
	if strings.Contains(访问连接, "localhost") {
		return false, ""
	}
	if strings.Contains(访问连接, "127.0.0.1") {
		return false, ""
	}
	dnsip := strings.Split(访问连接, ":")[0]
	ip1 := strings.Split(dnsip, ".")[0]
	if ip1 == "10" || ip1 == "172" || ip1 == "192" {
		return false, ""
	}
	_, err := strconv.Atoi(ip1)
	if err == nil {
		return true, dnsip
	}
	return false, dnsip
}
func incrementField(obj interface{}, fieldName string) error {
	// 获取传入对象的反射值
	val := reflect.ValueOf(obj).Elem()
	// 获取指定的字段
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("no such field: %s in obj", fieldName)
	}
	// 检查字段是否可设置
	if !field.CanSet() {
		return fmt.Errorf("cannot set field: %s", fieldName)
	}
	// 检查字段类型是否为整数类型
	if field.Kind() != reflect.Int {
		return fmt.Errorf("field %s is not of type int", fieldName)
	}
	// 将字段值加 1
	newValue := field.Int() + 1
	field.SetInt(newValue)
	fmt.Printf(fieldName+"：%d \n", newValue)
	return nil
}
func ips属地输入(v string) {
	ips属地锁.Lock()
	defer ips属地锁.Unlock()
	if _, ok := ips属地[v]; !ok {
		p := make([]ip属地数据结构, 1)
		ips属地[v] = &p[0]
		ips属地[v].Ip = v
	}
} //func ips属地输入(v string){
func 连接信息处理(RemoteAddr string, 被访问次数 *uint64, fieldName string) {
	atomic.AddUint64(被访问次数, 1)
	ip := 获取访问连接中的ip(RemoteAddr)
	连接信息锁.Lock()
	defer 连接信息锁.Unlock()
	if 连接信息[ip] == nil {
		p := make([]连接信息2, 1)
		连接信息[ip] = &p[0]
		if ok, v := 外网ip访问判断(RemoteAddr); ok {
			go ips属地输入(v)
		}
	}
	s := 连接信息[ip]
	fmt.Println(ip)
	err := incrementField(s, fieldName)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	连接信息[ip].L连接时间 = time.Now().Format("2006-01-02 15:04:05")
} //func 连接信息处理(RemoteAddr string, 被访问次数 *uint32) {
var 连接信息锁 sync.Mutex
var token被访问次数 uint64

func token(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &token被访问次数, "Token")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	query := r.URL.Query()
	user := query.Get("user")
	password := query.Get("password")
	if user != 本网关登录名之MD5 || password != 本网关登录密码之MD5 {
		w.Write([]byte("用户名或密码错误！"))
		return
	}
	token已编的码锁.RLock()
	w.Write(token已编的码)
	token已编的码锁.RUnlock()
}

var GetTagList被访问次数 uint64
var 各项目GetTagList已编码 = make(map[string]*已编码内容数据结构体, 0)

func GetTagList(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &GetTagList被访问次数, "GetTagList")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	query := r.URL.Query()
	user := query.Get("user")
	password := query.Get("password")
	if user != 本网关登录名之MD5 || password != 本网关登录密码之MD5 {
		w.Write([]byte("用户名或密码错误！"))
		return
	}
	if !完成第一次遍历项目变量信息表 {
		w.Write([]byte("未完成遍历项目变量信息表！"))
		return
	}
	if len(变量组.KVTags) < 1 {
		w.Write([]byte("没有变量组数据"))
		return
	}
	项目代号 := query.Get("项目代号")
	if 项目代号 == "" {
		项目代号 = 全部项目代号
	} else {
		if _, ok := 项目代号的表行们[项目代号]; !ok {
			变量数据 := "系统中不存在此项目代号：" + 项目代号
			w.Write([]byte(变量数据))
			return
		}
	}
	if _, ok := 各项目GetTagList已编码[项目代号]; ok {
		if time.Now().UnixMilli()-各项目GetTagList已编码[项目代号].Load编码时刻() < 各项目最小采集频率毫秒[项目代号] {
			w.Write(各项目GetTagList已编码[项目代号].Load编码())
			return
		}
	}
	var d []byte
	var err3 error
	if 项目代号 == 全部项目代号 {
		for i := range 变量组.KVTags {
			变量组.KVTags[i].VarValue锁.RLock()
		}
		d, err3 = json.Marshal(变量组)
		for i := range 变量组.KVTags {
			变量组.KVTags[i].VarValue锁.RUnlock()
		}
		if err3 != nil {
			w.Write([]byte("\r\n变量数据服务器执行 json Marshal(变量组) error !" + err3.Error()))
			return
		}
	} else {
		if 表行们, ok := 项目代号的表行们[项目代号]; ok {
			var 变量组2 变量组1
			变量组2.KVTags = make([]KVTag, len(表行们))
			for i, 表行 := range 表行们 {
				变量组.KVTags[表行].VarValue锁.RLock()
				变量组2.KVTags[i].VarValue = 变量组.KVTags[表行].VarValue
				变量组.KVTags[表行].VarValue锁.RUnlock()
				变量组2.KVTags[i].NVarID = 变量组.KVTags[表行].NVarID
				变量组2.KVTags[i].NVarType = 变量组.KVTags[表行].NVarType
				变量组2.KVTags[i].StrVarName = 变量组.KVTags[表行].StrVarName
			}
			d, err3 = json.Marshal(变量组2)
			if err3 != nil {
				w.Write([]byte("\r\n变量数据服务器执行 json Marshal(变量组) error !" + err3.Error()))
				return
			}
		} else {
			fmt.Println("if 表行们, ok := 项目代号的表行们[项目代号]; ok {")
			return
		}
	}
	if _, ok := 各项目GetTagList已编码[项目代号]; !ok {
		各项目GetTagList已编码[项目代号] = &已编码内容数据结构体{}
	}
	各项目GetTagList已编码[项目代号].Set编码(d)
	各项目GetTagList已编码[项目代号].Set编码时刻(time.Now().UnixMilli())
	w.Write(d)
}

var GetKVTagsValue被访问次数 uint64

func GetKVTagsValue(w http.ResponseWriter, r *http.Request) {
	type Q请求的变量 struct {
		Name string `json:"name"`
	}
	type 请求的变量组1 struct {
		Data []Q请求的变量
	}
	var 请求的变量组 请求的变量组1
	type ErrCode struct {
		Code    int    `json:"Code"`
		Message string `json:"Message"`
	}
	type 变量值 struct {
		Name  string      `json:"name"`
		Type  string      `json:"type"`
		Value interface{} `json:"value"`
	}
	type ErrCode1 struct {
		Code    int    `json:"Code"`
		Message string `json:"Message"`
		Data    []变量值  `json:"data"`
	}
	go 连接信息处理(r.RemoteAddr, &GetKVTagsValue被访问次数, "GetKVTagsValue")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	if !完成第一次遍历项目变量信息表 {
		w.Write([]byte("未完成遍历项目变量信息表！"))
		return
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&请求的变量组)
	if err != nil {
		var C错误码 ErrCode
		C错误码.Code = 4
		C错误码.Message = "Json Format Error"
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	长度 := len(请求的变量组.Data)
	if 长度 < 1 {
		var C错误码 ErrCode
		C错误码.Code = 5
		C错误码.Message = "J没有请求的变量名"
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if 长度 > 没有错误信息行数 {
		w.Write([]byte("请求的变量个数超过系统变量个数"))
		return
	}
	if 长度 > 最大读请求变量个数 {
		w.Write([]byte("请求的变量个数超过最大读请求变量个数(" + 最大读请求变量个数_s + ")"))
		return
	}
	for i := 0; i < 长度; i++ {
		if 请求的变量组.Data[i].Name == "" {
			var C错误码 ErrCode
			C错误码.Code = 6
			C错误码.Message = "有空的请求变量名！"
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	for i := 0; i < 长度; i++ {
		检查结果 := 变量名检查(请求的变量组.Data[i].Name)
		if 检查结果 != "ok" {
			var C错误码 ErrCode
			C错误码.Code = 1
			C错误码.Message = 请求的变量组.Data[i].Name + " 请求的变量名中有错误，不符合变量名命名规范！"
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	var 变量名们 = make(map[string]bool, 0)
	for i := 0; i < 长度; i++ {
		if !变量名们[请求的变量组.Data[i].Name] {
			变量名们[请求的变量组.Data[i].Name] = true
			continue
		}
		var C错误码 ErrCode
		C错误码.Code = 3
		C错误码.Message = "请求的变量名中有重复！"
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	} //for i := 0; i < 长度; i++ {
	for i := 0; i < 长度; i++ {
		if _, ok := 变量名所在行[请求的变量组.Data[i].Name]; !ok {
			var C错误码 ErrCode
			C错误码.Code = 2
			C错误码.Message = "请求的变量名中有错误，系统中不存在这些变量名！"
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	var 变量值组 ErrCode1
	变量值组.Data = make([]变量值, 长度)
	变量值组.Code = 0
	变量值组.Message = "success"
	变量名 := ""
	for i := 0; i < 长度; i++ {
		变量名 = 请求的变量组.Data[i].Name
		变量值组.Data[i].Name = 变量名
		变量值组.Data[i].Type = 变量组.KVTags[变量名所在行[变量名]].NVarType
		变量组.KVTags[变量名所在行[变量名]].VarValue锁.RLock()
		变量值组.Data[i].Value = 变量组.KVTags[变量名所在行[变量名]].VarValue
		变量组.KVTags[变量名所在行[变量名]].VarValue锁.RUnlock()
	}
	jsonStr, err3 := json.Marshal(变量值组)
	if err3 != nil {
		w.Write([]byte("\r\n服务器执行 json Marshal(变量值组) error !"))
		return
	}
	w.Write([]byte(jsonStr))
} //func GetKVTagsValue(w http.ResponseWriter, r *http.Request) {
var SetKVTagsValue被访问次数 uint64

func SetKVTagsValue(w http.ResponseWriter, r *http.Request) {
	type Q请求的变量 struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	type 请求的变量组1 struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Data     []Q请求的变量
	}
	var 请求的变量组 请求的变量组1
	type ErrCode struct {
		Code    int    `json:"Code"`
		Message string `json:"Message"`
	}
	go 连接信息处理(r.RemoteAddr, &SetKVTagsValue被访问次数, "SetKVTagsValue")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	if !完成第一次遍历项目变量信息表 {
		str := "未完成遍历项目变量信息表！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(str)
		w.Write([]byte(str))
		return
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&请求的变量组)
	if err != nil {
		var C错误码 ErrCode
		C错误码.Code = 1
		C错误码.Message = "Json Format Error" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if 请求的变量组.User != 本网关登录名之MD5 {
		var C错误码 ErrCode
		C错误码.Code = 2
		C错误码.Message = "用户名或密码错误！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if 请求的变量组.Password != 本网关登录密码之MD5 {
		var C错误码 ErrCode
		C错误码.Code = 3
		C错误码.Message = "用户名或密码错误！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	长度 := len(请求的变量组.Data)
	if 长度 < 1 {
		var C错误码 ErrCode
		C错误码.Code = 4
		C错误码.Message = "J没有请求的变量名" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if 长度 > 没有错误信息行数 {
		w.Write([]byte("请求的变量个数超过系统变量个数"))
		return
	}
	if 长度 > 最大写请求变量个数 {
		w.Write([]byte("请求的变量个数超过最大写请求变量个数(" + 最大写请求变量个数_s + ")"))
		return
	}
	for i := 0; i < 长度; i++ {
		if 请求的变量组.Data[i].Name == "" {
			var C错误码 ErrCode
			C错误码.Code = 5
			C错误码.Message = "有空的请求变量名！" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	for i := 0; i < 长度; i++ {
		检查结果 := 变量名检查(请求的变量组.Data[i].Name)
		if 检查结果 != "ok" {
			var C错误码 ErrCode
			C错误码.Code = 6
			C错误码.Message = 请求的变量组.Data[i].Name + " 请求的变量名中有错误，不符合变量名命名规范！" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	var 变量名们 = make(map[string]bool, 0)
	for i := 0; i < 长度; i++ {
		if !变量名们[请求的变量组.Data[i].Name] {
			变量名们[请求的变量组.Data[i].Name] = true
			continue
		}
		var C错误码 ErrCode
		C错误码.Code = 7
		C错误码.Message = "请求的变量名中有重复！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	} //for i := 0; i < 长度; i++ {
	for i := 0; i < 长度; i++ {
		if _, ok := 变量名所在行[请求的变量组.Data[i].Name]; !ok {
			var C错误码 ErrCode
			C错误码.Code = 8
			C错误码.Message = "请求的变量名中有错误，系统中不存在这些变量名！" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	for i := 0; i < 长度; i++ {
		判断结果 := 写值是否合法(请求的变量组.Data[i].Name, 请求的变量组.Data[i].Value)
		if 判断结果 != "ok" {
			var C错误码 ErrCode
			C错误码.Code = 9
			C错误码.Message = 判断结果 + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	var 执行结果 字符串累加互斥锁访问结构体
	//	var 等待组 sync.WaitGroup
	for i := 0; i < 长度; i++ {
		// 等待组.Add(1)
		// go 将写操作记录到相关map中(&等待组, &执行结果, 请求的变量组.Data[i].Name, 请求的变量组.Data[i].Value)
		将写操作记录到相关map中(nil, &执行结果, 请求的变量组.Data[i].Name, 请求的变量组.Data[i].Value)
	} //for i := 0; i < 长度; i++ {
	//	等待组.Wait()
	var 变量值组 ErrCode
	变量值组.Code = 0
	变量值组.Message = 执行结果.Load() + "_您的连接: " + r.RemoteAddr
	fmt.Println(变量值组.Message)
	jsonStr, err3 := json.Marshal(变量值组)
	if err3 != nil {
		w.Write([]byte("\r\n服务器执行 json Marshal(变量值组) error !"))
		return
	}
	w.Write([]byte(jsonStr))
} //func SetKVTagsValue(w http.ResponseWriter, r *http.Request) {
var xlsx被访问次数 uint64
var 各项目xlsx已编码 = make(map[string]*已编码内容数据结构体, 0)

func xlsx(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &xlsx被访问次数, "Xlsx")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	query := r.URL.Query()
	user := query.Get("user")
	password := query.Get("password")
	if user != 本网关登录名之MD5 || password != 本网关登录密码之MD5 {
		w.Write([]byte("用户名或密码错误！"))
		return
	}
	if !完成第一次遍历项目变量信息表 {
		w.Write([]byte("未完成遍历项目变量信息表！"))
		return
	}
	if 遍历项目变量信息表中 {
		w.Write([]byte("\r\n项目数据采集网关ModbusRTU->HttpAPI执行遍历项目变量信息表中，请稍等重试 !"))
		return
	}
	if len(变量组.KVTags) < 1 {
		w.Write([]byte("没有变量组数据"))
		return
	}
	项目代号 := query.Get("项目代号")
	if 项目代号 == "" {
		项目代号 = 全部项目代号
	} else {
		if _, ok := 项目代号的表行们[项目代号]; !ok {
			变量数据 := "系统中不存在此项目代号：" + 项目代号
			w.Write([]byte(变量数据))
			return
		}
	}
	if _, ok := 各项目xlsx已编码[项目代号]; ok {
		if time.Now().UnixMilli()-各项目xlsx已编码[项目代号].Load编码时刻() < 各项目最小采集频率毫秒[项目代号] {
			w.Write(各项目xlsx已编码[项目代号].Load编码())
			return
		}
	}
	var d []byte
	var err3 error
	if 项目代号 == 全部项目代号 {
		for i := range 项目变量信息表组.Rows {
			项目变量信息表组.Rows[i].D当前值锁.RLock()
		}
		d, err3 = json.Marshal(项目变量信息表组)
		for i := range 项目变量信息表组.Rows {
			项目变量信息表组.Rows[i].D当前值锁.RUnlock()
		}
		if err3 != nil {
			w.Write([]byte("\r\n项目数据采集网关ModbusRTU->HttpAPI执行 json Marshal(项目变量信息表组) error !"))
			return
		}
	} else {
		if 表行们, ok := 项目代号的表行们[项目代号]; ok {
			var 项目变量信息表组2 项目变量信息表组2
			项目变量信息表组2.Rows = make([]Row, len(表行们))
			for i, 表行 := range 表行们 {
				项目变量信息表组.Rows[表行].D当前值锁.RLock()
				项目变量信息表组2.Rows[i].C错误信息 = 项目变量信息表组.Rows[表行].C错误信息
				项目变量信息表组2.Rows[i].D当前值 = 项目变量信息表组.Rows[表行].D当前值
				项目变量信息表组.Rows[表行].D当前值锁.RUnlock()
				项目变量信息表组2.Rows[i].C采集成功时刻 = atomic.LoadInt64(&项目变量信息表组.Rows[表行].C采集成功时刻)
				项目变量信息表组2.Rows[i].C采集时刻毫秒 = atomic.LoadInt64(&项目变量信息表组.Rows[表行].C采集时刻毫秒)
				项目变量信息表组2.Rows[i].L连续通讯失败次数 = atomic.LoadUint32(&项目变量信息表组.Rows[表行].L连续通讯失败次数)
				项目变量信息表组2.Rows[i].B变量名称 = 项目变量信息表组.Rows[表行].B变量名称
				项目变量信息表组2.Rows[i].B变量类型 = 项目变量信息表组.Rows[表行].B变量类型
				项目变量信息表组2.Rows[i].B波特率 = 项目变量信息表组.Rows[表行].B波特率
				项目变量信息表组2.Rows[i].C串口号 = 项目变量信息表组.Rows[表行].C串口号
				项目变量信息表组2.Rows[i].C初始值 = 项目变量信息表组.Rows[表行].C初始值
				项目变量信息表组2.Rows[i].C采集频率毫秒 = 项目变量信息表组.Rows[表行].C采集频率毫秒
				项目变量信息表组2.Rows[i].C采集前等待毫秒 = 项目变量信息表组.Rows[表行].C采集前等待毫秒
				项目变量信息表组2.Rows[i].D读写属性 = 项目变量信息表组.Rows[表行].D读写属性
				项目变量信息表组2.Rows[i].D读功能码 = 项目变量信息表组.Rows[表行].D读功能码
				项目变量信息表组2.Rows[i].D读字节序 = 项目变量信息表组.Rows[表行].D读字节序
				项目变量信息表组2.Rows[i].D打包长度 = 项目变量信息表组.Rows[表行].D打包长度
				项目变量信息表组2.Rows[i].J奇偶校验 = 项目变量信息表组.Rows[表行].J奇偶校验
				项目变量信息表组2.Rows[i].J寄存器地址_s = 项目变量信息表组.Rows[表行].J寄存器地址_s
				项目变量信息表组2.Rows[i].J寄存器地址 = 项目变量信息表组.Rows[表行].J寄存器地址
				项目变量信息表组2.Rows[i].J计算值除原始值 = 项目变量信息表组.Rows[表行].J计算值除原始值
				项目变量信息表组2.Rows[i].L连续通讯失败多少次则认为通讯异常 = 项目变量信息表组.Rows[表行].L连续通讯失败多少次则认为通讯异常
				项目变量信息表组2.Rows[i].P批量读开始寄存器地址 = 项目变量信息表组.Rows[表行].P批量读开始寄存器地址
				项目变量信息表组2.Rows[i].P批量读伙伴 = 项目变量信息表组.Rows[表行].P批量读伙伴
				项目变量信息表组2.Rows[i].S数据位 = 项目变量信息表组.Rows[表行].S数据位
				项目变量信息表组2.Rows[i].S设备地址 = 项目变量信息表组.Rows[表行].S设备地址
				项目变量信息表组2.Rows[i].S数据类型 = 项目变量信息表组.Rows[表行].S数据类型
				项目变量信息表组2.Rows[i].S是否保存值 = 项目变量信息表组.Rows[表行].S是否保存值
				项目变量信息表组2.Rows[i].T停止位 = 项目变量信息表组.Rows[表行].T停止位
				项目变量信息表组2.Rows[i].T通讯超时毫秒 = 项目变量信息表组.Rows[表行].T通讯超时毫秒
				项目变量信息表组2.Rows[i].T通讯异常值 = 项目变量信息表组.Rows[表行].T通讯异常值
				项目变量信息表组2.Rows[i].T记录到数据库间隔_秒 = 项目变量信息表组.Rows[表行].T记录到数据库间隔_秒
				项目变量信息表组2.Rows[i].X写字节序 = 项目变量信息表组.Rows[表行].X写字节序
				项目变量信息表组2.Rows[i].X小数位数 = 项目变量信息表组.Rows[表行].X小数位数
				项目变量信息表组2.Rows[i].X小数点后值 = 项目变量信息表组.Rows[表行].X小数点后值
				项目变量信息表组2.Rows[i].X序号 = 项目变量信息表组.Rows[表行].X序号
				项目变量信息表组2.Rows[i].X需要保存数据 = 项目变量信息表组.Rows[表行].X需要保存数据
				项目变量信息表组2.Rows[i].Y允许通讯异常后只读变量 = 项目变量信息表组.Rows[表行].Y允许通讯异常后只读变量
			} //for i, 表行 := range 表行们 {
			d, err3 = json.Marshal(项目变量信息表组2)
			if err3 != nil {
				w.Write([]byte("\r\n变量数据服务器执行 json Marshal(变量组) error !" + err3.Error()))
				return
			}
		} else {
			fmt.Println("if 表行们, ok := 项目代号的表行们[项目代号]; ok {")
			return
		}
	}
	if _, ok := 各项目xlsx已编码[项目代号]; !ok {
		各项目xlsx已编码[项目代号] = &已编码内容数据结构体{}
	}
	各项目xlsx已编码[项目代号].Set编码(d)
	各项目xlsx已编码[项目代号].Set编码时刻(time.Now().UnixMilli())
	w.Write(d)
}

// var ModbusRTU_pdlei_CQ_bat bool
func 根据表行获得变量数据(i int, 忽略变量详情 string) string {
	变量数据 := ""
	v := &项目变量信息表组.Rows[i]
	v.D当前值锁.RLock()
	当前值 := v.D当前值
	单位 := v.D单位
	if 单位 != "" {
		单位 = v.D单位 + " "
	} else {
		单位 = " "
	}
	采集成功时刻 := atomic.LoadInt64(&v.C采集成功时刻)
	v.D当前值锁.RUnlock()
	变量详情 := ""
	if strings.Contains(v.B变量类型, "内存") {
		if 忽略变量详情 != "是" {
			变量详情 = "(" + v.B变量类型 + "_是否保存值：" + v.S是否保存值 + ")"
		}
		变量数据 += v.B变量名称 + ": " + 当前值 + 单位 + 变量详情 + "\r\n"
		return 变量数据
	}
	if 忽略变量详情 != "是" {
		成功采集至今秒数 := "成功采集至今" + 至今多少天时分秒(采集成功时刻)
		if 采集成功时刻 == 0 {
			成功采集至今秒数 = "还未成功采集过"
		}
		num64 := uint64(v.S设备地址)
		设备地址 := strconv.FormatUint(num64, 16) + "(" + strconv.FormatUint(num64, 10) + ")"
		num64 = uint64(v.J寄存器地址)
		寄存器地址 := strconv.FormatUint(num64, 16) + "(" + v.J寄存器地址_s + ")"
		num64 = uint64(v.B波特率)
		波特率 := strconv.FormatUint(num64, 10) + v.J奇偶校验
		num64 = uint64(v.S数据位)
		数据位 := strconv.FormatUint(num64, 10)
		num64 = uint64(v.T停止位)
		停止位 := strconv.FormatUint(num64, 10)
		波特率 = 波特率 + 数据位 + "_" + 停止位
		num64 = uint64(v.C采集频率毫秒 / 1000)
		采集频率秒 := "采频" + strconv.FormatUint(num64, 10) + "秒"
		记录到数据库间隔秒 := ""
		if v.T记录到数据库间隔_秒 != 0 {
			记录到数据库间隔秒 = "_记录间隔" + strconv.FormatInt(v.T记录到数据库间隔_秒, 10) + "秒"
		}
		通讯超时秒 := ""
		if v.T通讯超时毫秒 != 0 {
			num64 = uint64(v.T通讯超时毫秒 / 1000)
			通讯超时秒 = "_通讯超时" + strconv.FormatUint(num64, 10) + "秒"
		}
		采集前等待秒 := ""
		if v.C采集前等待毫秒 != 0 {
			num64 = uint64(v.C采集前等待毫秒 / 1000)
			采集前等待秒 = "_采集前等待" + strconv.FormatUint(num64, 10) + "秒"
		}
		const epsilon = 1e-9 // 根据实际需求调整阈值
		多少变化率才记录到数据库 := ""
		if v.D多少变化率才记录到数据库 > epsilon {
			多少变化率才记录到数据库 = fmt.Sprintf("_变化记录%.2f", v.D多少变化率才记录到数据库)
		}
		计算值除原始值 := ""
		if v.J计算值除原始值 > 1 || (v.J计算值除原始值 < 1 && v.J计算值除原始值 > epsilon) {
			计算值除原始值 = fmt.Sprintf("_系数%.2f", v.J计算值除原始值)
		}
		读字节序 := v.D读字节序
		写字节序 := v.X写字节序
		读写字节序 := ""
		if 读字节序 != "" {
			读写字节序 += "读" + 读字节序
		}
		if 写字节序 != "" {
			读写字节序 += "写" + 写字节序
		}
		变量详情 = "(" + v.B变量类型 + "_" + v.S数据类型 + "_" + v.D读写属性 + "_" +
			波特率 + "_" + "COM" + v.C串口号 + "_" + 设备地址 + "_" + v.D读功能码 + "_" + 寄存器地址 + "_" + 读写字节序 + "_" +
			成功采集至今秒数 + "_" + 采集频率秒 +
			采集前等待秒 +
			通讯超时秒 +
			多少变化率才记录到数据库 +
			记录到数据库间隔秒 +
			计算值除原始值 +
			")"
	} //if 忽略变量详情 != "是" {
	变量数据 += v.B变量名称 + ": " + 当前值 + 单位 + 变量详情 + "\r\n"
	return 变量数据
} //func 根据表行获得变量数据(i int) string {
func 获取变量数据(项目代号, 忽略变量详情 string) string {
	if len(项目变量信息表组.Rows) < 1 {
		return "len(项目变量信息表组.Rows)<1"
	}
	变量数据 := ""
	if 项目代号 == "" || 项目代号 == 全部项目代号 {
		for i := range 项目变量信息表组.Rows {
			变量数据 += 根据表行获得变量数据(i, 忽略变量详情)
		} //for _, v := range 项目变量信息表组.Rows {
		return 变量数据
	}
	if _, ok := 项目代号的表行们[项目代号]; !ok {
		变量数据 = "系统中不存在此项目代号：" + 项目代号
		return 变量数据
	}
	if 表行们, ok := 项目代号的表行们[项目代号]; ok {
		for _, 表行 := range 表行们 {
			变量数据 += 根据表行获得变量数据(表行, 忽略变量详情)
		} //for _, v := range 项目变量信息表组.Rows {
	}
	return 变量数据
} //func 获取变量数据()string{
var 数据被访问次数 uint64
var 各项目数据已编码 = make(map[string]*已编码内容数据结构体, 0)

type 已编码内容数据结构体 struct {
	锁    sync.Mutex
	已编的码 []byte
	编码时刻 int64
}

func (T *已编码内容数据结构体) Load编码() []byte {
	T.锁.Lock()
	defer T.锁.Unlock()
	return T.已编的码
}
func (T *已编码内容数据结构体) Load编码时刻() int64 {
	T.锁.Lock()
	defer T.锁.Unlock()
	return T.编码时刻
}
func (T *已编码内容数据结构体) Set编码(B []byte) {
	T.锁.Lock()
	defer T.锁.Unlock()
	T.已编的码 = B
}
func (T *已编码内容数据结构体) Set编码时刻(S int64) {
	T.锁.Lock()
	defer T.锁.Unlock()
	T.编码时刻 = S
}
func 数据(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &数据被访问次数, "S数据")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	query := r.URL.Query()
	user := query.Get("user")
	password := query.Get("password")
	if user != 本网关登录名之MD5 || password != 本网关登录密码之MD5 {
		w.Write([]byte("用户名或密码错误！"))
		return
	}
	if !完成第一次遍历项目变量信息表 {
		w.Write([]byte("未完成遍历项目变量信息表！"))
		return
	}
	if 遍历项目变量信息表中 {
		w.Write([]byte("\r\n项目数据采集网关ModbusRTU->HttpAPI执行遍历项目变量信息表中，请稍等重试 !"))
		return
	}
	if len(变量组.KVTags) < 1 {
		w.Write([]byte("没有变量组数据"))
		return
	}
	忽略变量详情 := query.Get("忽略变量详情")
	项目代号 := query.Get("项目代号")
	if 项目代号 == "" {
		项目代号 = 全部项目代号
	} else {
		if _, ok := 项目代号的表行们[项目代号]; !ok {
			变量数据 := "系统中不存在此项目代号：" + 项目代号
			w.Write([]byte(变量数据))
			return
		}
	}
	if _, ok := 各项目数据已编码[项目代号]; ok {
		if time.Now().UnixMilli()-各项目数据已编码[项目代号].Load编码时刻() < 各项目最小采集频率毫秒[项目代号] {
			w.Write(各项目数据已编码[项目代号].Load编码())
			return
		}
	}
	startTime := time.Now()
	str1 := ""
	str := 编译时间 + "\r\n"
	str1 += str
	str = 至今多少天时分秒(网关启动时刻)
	str = "启动至今: " + str + "\r\n"
	str1 += str
	str = 网关启动时间
	str1 += str
	变量个数 := "0"
	if 表行们, ok := 项目代号的表行们[项目代号]; ok {
		变量个数 = fmt.Sprintf("%d", len(表行们))
	} else {
		if 项目代号 == "" || 项目代号 == 全部项目代号 {
			变量个数 = fmt.Sprintf("%d", len(项目变量信息表组.Rows))
		}
	}
	内容 := str1 + "\r\n" + time.Now().Format("2006-01-02 15:04:05") + "\r\n第几次采集： " +
		strconv.FormatUint(atomic.LoadUint64(&第几次采集), 10) + "\r\n变量个数：" + 变量个数 +
		"\r\n" + 获取变量数据(项目代号, 忽略变量详情)
	elapsedTime := time.Since(startTime)
	最大耗时 := fmt.Sprintf("编码耗时：%v", elapsedTime)
	内容 = 最大耗时 + "\r\n" + 内容
	d := []byte(内容)
	if _, ok := 各项目数据已编码[项目代号]; !ok {
		各项目数据已编码[项目代号] = &已编码内容数据结构体{}
	}
	各项目数据已编码[项目代号].Set编码(d)
	各项目数据已编码[项目代号].Set编码时刻(time.Now().UnixMilli())
	w.Write(d)
}
func 重启软件(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadUint32(&已经触发重启程序了) == 0 {
		邮件主题 := r.RemoteAddr + "要重启软件," + 发邮件主题.Load()
		邮件正文 := r.RemoteAddr + "要重启软件," + 发邮件主题.Load()
		go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
		批处理文件名非中文提示 := strings.Replace(strings.Replace(r.RemoteAddr, ".", "_", -1), ":", "_", -1)
		go 重新启动外部程序(批处理文件名非中文提示, 程序名, 程序文件名及路径)
	}
	w.Write([]byte("项目数据采集网关ModbusRTU->HttpAPI将被重启！"))
} //func 重启软件(w http.ResponseWriter, r *http.Request){
var 重启被访问次数 uint64

func 重启(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &重启被访问次数, "C重启")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	if time.Now().Unix()-连接登录时刻[r.RemoteAddr] > 登录保持秒数 {
		提供登录页面(w, r)
		return
	}
	if !判断文件是否存在(程序文件名及路径) {
		str := 程序文件名及路径
		str = "程序文件名及路径:" + str + "\r\n"
		w.Write([]byte(str))
		str = "此文件已经不存在！" + "\r\n"
		w.Write([]byte(str))
		w.Write([]byte("所以项目数据采集网关ModbusRTU->HttpAPI将无法被重启！"))
		return
	}
	写项目变量信息表并保存()
	重启软件(w, r)
	//atomic.StoreUint32(&已经触发重启程序了, 1)
}

var 发邮件被访问次数 uint64

func 发邮件(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &发邮件被访问次数, "F发邮件")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	str := ""
	if atomic.LoadUint32(&需要发邮件) == 1 {
		str = "\r\n当前服务器在不断尝试发送邮件中,您无需再触发发邮件！发邮件一次性成功的条件是在发邮件时间间隔秒数内没有发邮件事件，之后先登录成功然后发邮件，必须注意顺序！此条件设置目的是屏蔽恶意狂刷发邮件。"
		w.Write([]byte(str))
		return
	}
	if time.Now().Unix()-atomic.LoadInt64(&上次发邮件时刻) < 发邮件时间间隔秒数 {
		atomic.StoreInt64(&上次发邮件时刻, time.Now().Unix())
		str = "距离上次触发发邮件时间小于发邮件时间间隔秒数设定，触发发邮件有自动和人工，您此次触发为人工，本次触发无效！发邮件一次性成功的条件是在发邮件时间间隔秒数内没有发邮件事件，之后先登录成功然后发邮件，必须注意顺序！此条件设置目的是屏蔽恶意狂刷发邮件。"
		w.Write([]byte(str))
		return
	}
	atomic.StoreInt64(&上次发邮件时刻, time.Now().Unix())
	if time.Now().Unix()-连接登录时刻[r.RemoteAddr] > 登录保持秒数 {
		提供登录页面(w, r)
		return
	}
	atomic.StoreUint32(&需要发邮件, 1)
	触发者信息 = r.RemoteAddr + "/发邮件"
	邮件正文前附加 = ""
	邮件主题前附加 = "人工发邮件"
	w.Write([]byte("已向项目数据采集网关ModbusRTU->HttpAPI提交发邮件申请！"))
}

var 关闭被访问次数 uint64

func 关闭(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &关闭被访问次数, "G关闭")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	if time.Now().Unix()-连接登录时刻[r.RemoteAddr] > 登录保持秒数 {
		提供登录页面(w, r)
		return
	}
	时间 := strings.Replace(time.Now().Format("2006-01-02 15:04:05"), ":", "", -1)
	文件名 := strings.Replace(strings.Replace(r.RemoteAddr, ":", "_", -1), ".", "_", -1) + "_" + 程序名 + "_关闭_" + 时间 + ".bat"
	内容 := "set SLEEP=ping 127.0.0.1 /n" + "\r\n" + "%SLEEP% 4 > nul" + "\r\n" + "taskkill.exe /f /im " + 程序名 + "\r\n" + "exit\r\n"
	data1, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(内容))
	if err != nil {
		fmt.Fprint(w, "simplifiedchinese.GBK.NewEncoder().Bytes([]byte("+内容+")发生错误！")
		return
	}
	创建文件2(目录+"bat/", 文件名, data1)
	str := "项目数据采集网关ModbusRTU->HttpAPI将被关闭，若要重启只能登录服务器人工启动，或者部署看门狗程序由它自动启动！"
	邮件主题 := r.RemoteAddr + "要关闭软件," + 发邮件主题.Load()
	邮件正文 := 邮件主题 + "\r\n" + str
	go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
	time.Sleep(3 * time.Second)
	w.Write([]byte(str))
	写项目变量信息表并保存()
	atomic.StoreUint32(&已经触发重启程序了, 1)
	c := exec.Command(目录 + "bat/" + 文件名)
	SQLiteDB访问锁.Lock()
	defer SQLiteDB访问锁.Unlock()
	写项目变量信息表并保存锁.Lock()
	defer 写项目变量信息表并保存锁.Unlock()
	if SQLiteDB内存数据库连接 != nil {
		SQLiteDB内存数据库连接.Close()
	}
	if SQLiteDB磁盘查询历史数据库连接 != nil {
		SQLiteDB磁盘查询历史数据库连接.Close()
	}
	c.Start()
}

var 重启服务器被访问次数 uint64

func 重启服务器(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &重启服务器被访问次数, "C重启服务器")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	if time.Now().Unix()-连接登录时刻[r.RemoteAddr] > 登录保持秒数 {
		提供登录页面(w, r)
		return
	}
	提示 := "项目数据采集网关ModbusRTU->HttpAPI所在的主机将10秒之后被重启！"
	时间 := strings.Replace(time.Now().Format("2006-01-02 15:04:05"), ":", "", -1)
	文件名 := strings.Replace(strings.Replace(r.RemoteAddr, ":", "_", -1), ".", "_", -1) + "_重启服务器_" + 时间 + ".bat"
	内容 := "shutdown /f /r /t 10" + "\r\n"
	创建文件(目录+"bat/", 文件名, 内容)
	邮件主题 := r.RemoteAddr + "要重启服务器," + 发邮件主题.Load()
	邮件正文 := r.RemoteAddr + "要重启服务器," + 发邮件主题.Load()
	go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
	w.Write([]byte(提示))
	写项目变量信息表并保存()
	atomic.StoreUint32(&已经触发重启程序了, 1)
	c := exec.Command(目录 + "bat/" + 文件名)
	SQLiteDB访问锁.Lock()
	defer SQLiteDB访问锁.Unlock()
	写项目变量信息表并保存锁.Lock()
	defer 写项目变量信息表并保存锁.Unlock()
	if SQLiteDB内存数据库连接 != nil {
		SQLiteDB内存数据库连接.Close()
	}
	if SQLiteDB磁盘查询历史数据库连接 != nil {
		SQLiteDB磁盘查询历史数据库连接.Close()
	}
	c.Start()
}
func 判断r_host类型(r_Host string) string {
	if strings.Contains(r_Host, "[::1]") {
		return "[::1]"
	}
	if strings.Contains(r_Host, "127.0.0.1") {
		return "127.0.0.1"
	}
	if strings.Contains(r_Host, "localhost") {
		return "localhost"
	}
	dnsip := strings.Split(r_Host, ":")[0]
	if net.ParseIP(dnsip).IsPrivate() {
		return dnsip
	}
	// ip1 := strings.Split(dnsip, ".")[0]
	// if ip1 == "10" || ip1 == "172" || ip1 == "192" {
	// 	return "私有地址"
	// }
	ip1 := strings.Split(dnsip, ".")[0]
	_, err := strconv.Atoi(ip1)
	if err == nil {
		return "公网ip"
	}
	if strings.Contains(ip1, "[") {
		return "ipv6"
	}
	return "公网域名"
} //func 判断r_host类型(r *http.Request) string {
func 提供登录页面(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("login").Parse(loginhtml1 + r.Host + loginhtml2 + r.Host + loginhtml3))
	t.Execute(w, nil)
	fmt.Fprint(w, "<br>请先登录！ 请先登录！ 请先登录！ 请先登录！ 请先登录！<br>")
	fmt.Fprint(w, "<br>欢迎使用我公司开发的软件(www.pdlei.cn)<br>")
} //func 提供登录页面(w http.ResponseWriter, r *http.Request) {
var 调试被访问次数 uint64

func 调试(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &调试被访问次数, "T调试")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	if time.Now().Unix()-连接登录时刻[r.RemoteAddr] > 登录保持秒数 {
		提供登录页面(w, r)
		return
	}
	if 允许调试 {
		允许调试 = false
		w.Write([]byte("项目数据采集网关ModbusRTU->HttpAPI将退出调试模式！"))
	} else {
		允许调试 = true
		w.Write([]byte("项目数据采集网关ModbusRTU->HttpAPI将处于调试模式！"))
	}
}

var 统计上次json编码时刻 int64
var 统计已编的码 []byte
var 统计已编的码锁 sync.RWMutex
var 统计被访问次数 uint64

func 统计(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &统计被访问次数, "T统计")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	if time.Now().Unix()-连接登录时刻[r.RemoteAddr] > 登录保持秒数 {
		提供登录页面(w, r)
		return
	}
	if time.Now().UnixMilli()-atomic.LoadInt64(&统计上次json编码时刻) < 最小采集频率毫秒 {
		统计已编的码锁.RLock()
		w.Write(统计已编的码)
		统计已编的码锁.RUnlock()
		return
	}
	d := []byte(获取统计信息())
	统计已编的码锁.Lock()
	统计已编的码 = d
	统计已编的码锁.Unlock()
	atomic.StoreInt64(&统计上次json编码时刻, time.Now().UnixMilli())
	统计已编的码锁.RLock()
	w.Write(统计已编的码)
	统计已编的码锁.RUnlock()
}
func 波特率检查(波特率 string) string {
	波特率值, err := strconv.Atoi(波特率)
	if err != nil {
		return err.Error() + "\r\n"
	}
	if 波特率值 < 波特率01 {
		return "设定的波特率" + 波特率 + "小于最小波特率" + strconv.Itoa(波特率01) + "；将被置空\r\n" + "\r\n"
	}
	if 波特率值 > 波特率15 {
		return "设定的波特率" + 波特率 + "大于最大波特率" + strconv.Itoa(波特率15) + "；将被置空\r\n" + "\r\n"
	}
	if 波特率值 != 波特率01 &&
		波特率值 != 波特率02 &&
		波特率值 != 波特率03 &&
		波特率值 != 波特率04 &&
		波特率值 != 波特率05 &&
		波特率值 != 波特率06 &&
		波特率值 != 波特率07 &&
		波特率值 != 波特率08 &&
		波特率值 != 波特率09 &&
		波特率值 != 波特率10 &&
		波特率值 != 波特率11 &&
		波特率值 != 波特率12 &&
		波特率值 != 波特率13 &&
		波特率值 != 波特率14 &&
		波特率值 != 波特率15 {
		return "设定的波特率" + 波特率 + "不被软件认可！" + "将被置空\r\n" + "\r\n"
	}
	return "ok"
} //func 波特率检查(波特率 string) string {
func 变量名检查(变量名 string) string {
	if utf8.RuneCountInString(变量名) < 最小变量名长度 {
		return "变量名(" + 变量名 + ")长度小于最小变量名长度" + strconv.Itoa(最小变量名长度) + "\r\n"
	} else {
		if utf8.RuneCountInString(变量名) > 最大变量名长度 {
			return "变量名(" + 变量名 + ")长度大于最大变量名长度" + strconv.Itoa(最大变量名长度) + "\r\n"
		}
	}
	正则表达式 := "^[A-Za-z]{1}[A-Za-z0-9]{2}_[\u4e00-\u9fa5A-Za-z0-9_]{1,81}$"
	var a = regexp.MustCompile(正则表达式)
	if !a.MatchString(变量名) {
		return "不匹配正则表达式(" + 正则表达式 + ")的变量名(" + 变量名 + ")" + "\r\n"
	}
	return "ok"
}
func UUID检查(UUID字符串 string) string {
	const (
		最小UUID字符串长度   = 36
		最大UUID字符串长度   = 36
		最小UUID字符串长度_s = "36"
		最大UUID字符串长度_s = "36"
	)
	if utf8.RuneCountInString(UUID字符串) < 最小UUID字符串长度 {
		return "UUID字符串(" + UUID字符串 + ")长度小于最小UUID字符串长度" + 最小UUID字符串长度_s + "\r\n"
	} else {
		if utf8.RuneCountInString(UUID字符串) > 最大UUID字符串长度 {
			return "UUID字符串(" + UUID字符串 + ")长度大于最大UUID字符串长度" + 最大UUID字符串长度_s + "\r\n"
		}
	}
	正则表达式 := "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"
	var a = regexp.MustCompile(正则表达式)
	if !a.MatchString(UUID字符串) {
		return "不匹配正则表达式(" + 正则表达式 + ")的UUID字符串(" + UUID字符串 + ")" + "\r\n"
	}
	return "ok"
}
func 字符串有小写字母判断(字符串 string) bool {
	正则表达式 := "[a-z]+?"
	var a = regexp.MustCompile(正则表达式)
	return a.MatchString(字符串)
}
func 求单元格代号(行号 int, 列号 int) string {
	列符号 := ""
	if 行号 < 1 || 列号 < 1 {
		return 列符号
	}
	商 := 列号
	var 模余 int
	for 商 > 0 {
		模余 = (商 - 1) % 26
		列符号 = string(byte(65+模余)) + 列符号
		商 = (int)((商 - 模余) / 26)
	}
	列符号 = 列符号 + strconv.Itoa(行号)
	return 列符号
}
func 小数点前后值分离(字符串 string) (前值, 后值 uint64, 分离结果 string) {
	分离结果 = ""
	位置 := strings.Index(字符串, ".")
	小数点前值_s := 字符串[:位置]
	小数点后值_s := 字符串[位置+1:]
	小数点前值, err := strconv.ParseUint(小数点前值_s, 10, 16)
	if err != nil {
		if 分离结果 != "" {
			分离结果 = 分离结果 + "\r\n"
		}
		分离结果 = 分离结果 + "小数点前值(" + 小数点前值_s + ")不是一个16位无符号整型！"
		小数点前值 = 0
	}
	小数点后值, err := strconv.ParseUint(小数点后值_s, 10, 8)
	if err != nil {
		if 分离结果 != "" {
			分离结果 = 分离结果 + "\r\n"
		}
		分离结果 = 分离结果 + "小数点后值(" + 小数点后值_s + ")不是一个8位无符号整型！"
		小数点后值 = 0
	}
	前值 = 小数点前值
	后值 = 小数点后值
	if 分离结果 == "" {
		分离结果 = "ok"
	}
	return 前值, 后值, 分离结果
} //func 小数点前后值分离(字符串 string) (前值, 后值 int64, 分离结果 string) {
var 写项目变量信息表并保存锁 sync.Mutex
var 写项目变量信息表并保存时发生的错误 字符串互斥锁访问结构体
var 项目变量信息内存表保存次数 uint32

func 写项目变量信息表并保存() {
	//fmt.Println("写项目变量信息表并保存")
	if atomic.LoadUint32(&已经触发重启程序了) == 1 {
		return
	}
	if !完成第一次遍历项目变量信息表 {
		return
	}
	if 遍历项目变量信息表中 {
		return
	}
	写项目变量信息表并保存锁.Lock()
	defer 写项目变量信息表并保存锁.Unlock()
	要写保存的表行们, ok := 哪些表行要写保存.获取数组对象()
	if !ok {
		return
	}
	f, err1 := excelize.OpenFile(项目变量信息表名)
	if err1 != nil {
		fmt.Println("写项目变量信息表打开时" + err1.Error())
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("写项目变量信息表并保存关闭时" + err.Error())
		}
	}()
	index, _ := f.NewSheet(工作表名)
	rows, err := f.GetRows(工作表名)
	if err != nil {
		fmt.Println("写项目变量信息表打开获取行数据时" + err.Error())
		return
	}
	项目变量信息表格行数 := len(rows)
	if 项目变量信息表格行数 < 2 {
		fmt.Println("写项目变量信息表并保存时发现项目变量信息表格行数<2")
		return
	}
	for _, 表行 := range 要写保存的表行们 {
		变量名所在表行2 := 变量名所在表行[项目变量信息表组.Rows[表行].B变量名称]
		f.SetCellValue(工作表名, 求单元格代号(变量名所在表行2+1, 标题列号[初始值_下标]+1), 项目变量信息表组.Rows[表行].C初始值)
	}
	f.SetActiveSheet(index)
	if err := f.Save(); err != nil {
		str := "写项目变量信息表保存时" + err.Error()
		fmt.Println(str)
		写项目变量信息表并保存时发生的错误.Set(str)
		return
	}
	写项目变量信息表并保存时发生的错误.Set("")
	atomic.AddUint32(&项目变量信息内存表保存次数, 1)
	for _, 表行 := range 要写保存的表行们 {
		哪些表行要写保存.Delete(表行)
	}
	延时1 := 项目变量信息表格行数 / 项目变量信息表格保存多少行需要1秒时间间隔
	if 延时1 <= 0 {
		延时1 = 1
	}
	time.Sleep(time.Second * time.Duration(延时1))
} //写项目变量信息表并保存()
var 遍历项目变量信息表保存次 uint64
var 没有错误信息行数 int
var 变量名所在行 map[string]int
var 变量名所在表行 map[string]int
var 项目变量信息表格行数 int

func 遍历项目变量信息表2() bool {
	完成第一次遍历项目变量信息表 = false
	遍历项目变量信息表中 = true
	遍历项目变量信息表结果 = ""
	表格有改动 := false
	f, err1 := excelize.OpenFile(项目变量信息表名)
	if err1 != nil {
		遍历项目变量信息表结果 = 遍历项目变量信息表结果 + err1.Error()
		遍历项目变量信息表中 = false
		return false
	}
	defer func() {
		if err := f.Close(); err != nil {
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + err.Error()
		}
	}()
	index, _ := f.NewSheet(工作表名)
	rows, err := f.GetRows(工作表名)
	if err != nil {
		遍历项目变量信息表结果 = 遍历项目变量信息表结果 + err.Error()
		遍历项目变量信息表中 = false
		return false
	}
	项目变量信息表格行数 = len(rows)
	if 项目变量信息表格行数 < 2 {
		遍历项目变量信息表结果 = "遍历项目变量信息表结果:项目变量信息表格行数<2"
		遍历项目变量信息表中 = false
		return false
	}
	变量警告信息 := ""
	检查结果 := ""
	标题列数计数 := 0
	for i, row := range rows {
		if i > 0 {
			break
		}
		for j, colCell := range row {
			个数 := strings.Count(colCell, " ")
			if 个数 > 0 {
				colCell = strings.Replace(colCell, " ", "", 个数)
				f.SetCellValue(工作表名, 求单元格代号(i+1, j+1), colCell)
				row[j] = colCell
				表格有改动 = true
			}
			switch colCell {
			case "变量ID":
				标题列数计数++
				标题列号[变量ID_下标] = j
			case "变量名":
				标题列数计数++
				标题列号[变量名_下标] = j
			case "变量类型":
				标题列数计数++
				标题列号[变量类型_下标] = j
			case "读写属性":
				标题列数计数++
				标题列号[读写属性_下标] = j
			case "采集频率(毫秒)":
				标题列数计数++
				标题列号[采集频率_下标] = j
			case "初始值；字符串(<128字节)":
				标题列数计数++
				标题列号[初始值_下标] = j
			case "通讯异常值":
				标题列数计数++
				标题列号[通讯异常值_下标] = j
			case "是否保存值":
				标题列数计数++
				标题列号[是否保存值_下标] = j
			case "计算值除原始值":
				标题列数计数++
				标题列号[计算值除原始值_下标] = j
			case "小数位数":
				标题列数计数++
				标题列号[小数位数_下标] = j
			case "数据类型":
				标题列数计数++
				标题列号[数据类型_下标] = j
			case "串口号":
				标题列数计数++
				标题列号[串口号_下标] = j
			case "波特率":
				标题列数计数++
				标题列号[波特率_下标] = j
			case "奇偶校验":
				标题列数计数++
				标题列号[奇偶校验_下标] = j
			case "数据位":
				标题列数计数++
				标题列号[数据位_下标] = j
			case "停止位":
				标题列数计数++
				标题列号[停止位_下标] = j
			case "通讯超时（毫秒）":
				标题列数计数++
				标题列号[通讯超时_下标] = j
			case "设备地址":
				标题列数计数++
				标题列号[设备地址_下标] = j
			case "读功能码":
				标题列数计数++
				标题列号[读功能码_下标] = j
			case "寄存器地址":
				标题列数计数++
				标题列号[寄存器地址_下标] = j
			case "多少变化率才记录到数据库":
				标题列数计数++
				标题列号[多少变化率才记录到数据库_下标] = j
			case "打包长度":
				标题列数计数++
				标题列号[打包长度_下标] = j
			case "允许通讯异常后只读变量":
				标题列数计数++
				标题列号[允许通讯异常后只读变量_下标] = j
			case "采集前等待(毫秒)":
				标题列数计数++
				标题列号[采集前等待_下标] = j
			case "连续通讯失败多少次则认为通讯异常":
				标题列数计数++
				标题列号[连续通讯失败多少次则认为通讯异常_下标] = j
			case "读字节序":
				标题列数计数++
				标题列号[读字节序_下标] = j
			case "写字节序":
				标题列数计数++
				标题列号[写字节序_下标] = j
			case "错误信息":
				标题列数计数++
				标题列号[错误信息_下标] = j
			case "警告信息":
				标题列数计数++
				标题列号[警告信息_下标] = j
			case "记录到数据库间隔（秒）":
				标题列数计数++
				标题列号[记录到数据库间隔_秒_下标] = j
			case "单位":
				标题列数计数++
				标题列号[单位_下标] = j
			default:
				内容 := "标题列:" + colCell + " 不被程序认可！"
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + 内容 + "\r\n"
			} //switch colCell {
		} //for j, colCell := range row {
	} //for i, row := range rows {
	if 标题列数计数 != 标题列数 {
		if 遍历项目变量信息表结果 != "" {
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "\r\n"
		}
		遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "遍历项目变量信息表发现软件认可的标题列数为" + strconv.Itoa(标题列数计数) + ",而软件认可的列数应为" +
			strconv.Itoa(标题列数) + ",于是软件认为此工作表非法！"
	}
	if 遍历项目变量信息表结果 != "" {
		遍历项目变量信息表中 = false
		return true //标题行有错误就退出
	}
	for i := 0; i < 项目变量信息表格行数; i++ {
		if i == 0 {
			continue
		}
		str, err := f.GetCellValue(工作表名, 求单元格代号(i+1, 标题列号[记录到数据库间隔_秒_下标]+1))
		if err == nil && str == "" {
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[记录到数据库间隔_秒_下标]+1), 默认记录到数据库间隔_秒_s)
			表格有改动 = true
		}
		str, err = f.GetCellValue(工作表名, 求单元格代号(i+1, 标题列号[多少变化率才记录到数据库_下标]+1))
		if err == nil && str == "" {
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[多少变化率才记录到数据库_下标]+1), 默认多少变化率才记录到数据库)
			表格有改动 = true
		}
	}
	if 表格有改动 {
		表格有改动 = false
		if err := f.Save(); err != nil {
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + err.Error() + "\r\n"
			遍历项目变量信息表中 = false
			return true
		}
		延时1 := 项目变量信息表格行数 / 项目变量信息表格保存多少行需要1秒时间间隔
		if 延时1 <= 0 {
			延时1 = 1
		}
		time.Sleep(time.Second * time.Duration(延时1))
		遍历项目变量信息表保存次++
		rows, err = f.GetRows(工作表名)
		if err != nil {
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + err.Error()
			遍历项目变量信息表中 = false
			return false
		}
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		变量警告信息 = ""
		for j, colCell := range row {
			if j == 标题列号[警告信息_下标] || j == 标题列号[错误信息_下标] {
				continue
			}
			个数 := strings.Count(colCell, " ")
			if 个数 > 0 {
				检查结果 = colCell + "<-中不能有空格，空格将被删除\r\n"
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + strconv.Itoa(i+1) + ")" + 检查结果
				变量警告信息 = 变量警告信息 + 检查结果
				colCell = strings.Replace(colCell, " ", "", 个数)
				f.SetCellValue(工作表名, 求单元格代号(i+1, j+1), colCell)
				row[j] = colCell
				表格有改动 = true
			}
		}
		if 变量警告信息 != "" {
			if !strings.Contains(row[标题列号[警告信息_下标]], 变量警告信息) {
				变量警告信息 = 变量警告信息 + row[标题列号[警告信息_下标]]
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[警告信息_下标]+1), 变量警告信息)
				表格有改动 = true
			}
		}
	}
	变量名重复的行 := make(map[int]bool)
	for i3, row3 := range rows {
		if i3 == 0 {
			continue
		}
		if _, ok := 变量名重复的行[i3]; ok {
			continue
		}
		变量名1 := row3[标题列号[变量名_下标]]
		if 变量名1 == "" {
			continue
		}
		for i4, row4 := range rows {
			if i4 == 0 {
				continue
			}
			if i4 <= i3 {
				continue
			}
			if _, ok := 变量名重复的行[i4]; ok {
				continue
			}
			变量名 := row4[标题列号[变量名_下标]]
			if 变量名 == "" {
				continue
			}
			if 变量名 == 变量名1 {
				变量名重复的行[i4] = true
			}
		} //for i4, _ := range rows {
	} //for i3, _ := range rows {
	没有错误信息的行 := make(map[int]bool)
	没有错误信息行数 = 0
	for i, row := range rows {
		if i == 0 {
			continue
		}
		变量警告信息 = ""
		置空 := false
		警告 := false
		行号2 := strconv.Itoa(i + 1)
		变量ID := row[标题列号[变量ID_下标]]
		变量名 := row[标题列号[变量名_下标]]
		变量类型 := row[标题列号[变量类型_下标]]
		读写属性 := row[标题列号[读写属性_下标]]
		读功能码 := row[标题列号[读功能码_下标]]
		数据类型 := row[标题列号[数据类型_下标]]
		初始值_s := row[标题列号[初始值_下标]]
		通讯异常值_s := row[标题列号[通讯异常值_下标]]
		是否保存值_s := row[标题列号[是否保存值_下标]]
		计算值除原始值_s := row[标题列号[计算值除原始值_下标]]
		小数位数_s := row[标题列号[小数位数_下标]]
		串口号_s := row[标题列号[串口号_下标]]
		波特率_s := row[标题列号[波特率_下标]]
		奇偶校验_s := row[标题列号[奇偶校验_下标]]
		数据位_s := row[标题列号[数据位_下标]]
		停止位_s := row[标题列号[停止位_下标]]
		通讯超时_s := row[标题列号[通讯超时_下标]]
		设备地址_s := row[标题列号[设备地址_下标]]
		寄存器地址_s := row[标题列号[寄存器地址_下标]]
		打包长度_s := row[标题列号[打包长度_下标]]
		允许通讯异常后只读变量_s := row[标题列号[允许通讯异常后只读变量_下标]]
		采集频率_s := row[标题列号[采集频率_下标]]
		采集前等待_s := row[标题列号[采集前等待_下标]]
		连续通讯失败多少次则认为通讯异常_s := row[标题列号[连续通讯失败多少次则认为通讯异常_下标]]
		读字节序_s := row[标题列号[读字节序_下标]]
		写字节序_s := row[标题列号[写字节序_下标]]
		记录到数据库间隔_秒_s := row[标题列号[记录到数据库间隔_秒_下标]]
		错误信息 := ""
		if 字符串有小写字母判断(奇偶校验_s) {
			检查结果 = "奇偶校验(" + 奇偶校验_s + ")中有小写字母，应用大写字母来表达，将被替换！\r\n"
			变量警告信息 = 变量警告信息 + 检查结果
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			奇偶校验_s = strings.ToUpper(奇偶校验_s)
			row[标题列号[奇偶校验_下标]] = 奇偶校验_s
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[奇偶校验_下标]+1), 奇偶校验_s)
			表格有改动 = true
		}
		if 字符串有小写字母判断(变量类型) {
			检查结果 = "变量类型(" + 变量类型 + ")中有小写字母，应用大写字母来表达，将被替换！\r\n"
			变量警告信息 = 变量警告信息 + 检查结果
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量类型 = strings.ToUpper(变量类型)
			row[标题列号[变量类型_下标]] = 变量类型
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[变量类型_下标]+1), 变量类型)
			表格有改动 = true
		}
		if 字符串有小写字母判断(数据类型) {
			检查结果 = "数据类型(" + 数据类型 + ")中有小写字母，应用大写字母来表达，将被替换！\r\n"
			变量警告信息 = 变量警告信息 + 检查结果
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			数据类型 = strings.ToUpper(数据类型)
			row[标题列号[数据类型_下标]] = 数据类型
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[数据类型_下标]+1), 数据类型)
			表格有改动 = true
		}
		if 字符串有小写字母判断(读字节序_s) {
			检查结果 = "读字节序(" + 读字节序_s + ")中有小写字母，应用大写字母来表达，将被替换！\r\n"
			变量警告信息 = 变量警告信息 + 检查结果
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			读字节序_s = strings.ToUpper(读字节序_s)
			row[标题列号[读字节序_下标]] = 读字节序_s
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[读字节序_下标]+1), 读字节序_s)
			表格有改动 = true
		}
		if 字符串有小写字母判断(写字节序_s) {
			检查结果 = "写字节序(" + 写字节序_s + ")中有小写字母，应用大写字母来表达，将被替换！\r\n"
			变量警告信息 = 变量警告信息 + 检查结果
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			写字节序_s = strings.ToUpper(写字节序_s)
			row[标题列号[写字节序_下标]] = 写字节序_s
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[写字节序_下标]+1), 写字节序_s)
			表格有改动 = true
		}
		var 填值 interface{}
		只写有效 := false
		IO离散点 := false
		switch 变量类型 {
		case "IO整型", "IO实型", "IO字符串":
			if 读功能码 == "3" && 读写属性 == "只写" {
				只写有效 = true
			}
		} //switch 变量类型 {
		填值 = ""
		检查结果 = ""
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 读写属性 != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定读写属性，将被软件置空\r\n"
				row[标题列号[读写属性_下标]] = ""
				读写属性 = ""
			}
		} else {
			if 读写属性 != "读写" && 读写属性 != "只读" && 读写属性 != "只写" {
				检查结果 = "读写属性既不是(读写)也不是(只读和只写)，将使用默认(只读)!\r\n"
				填值 = "只读"
				row[标题列号[读写属性_下标]] = "只读"
				读写属性 = "只读"
			} else {
				switch 变量类型 {
				case "IO实型", "IO整型", "IO字符串":
					if 读功能码 == "4" {
						if 读写属性 == "读写" || 读写属性 == "只写" {
							检查结果 = "读功能码是4的寄存器，读写属性(" + 读写属性 + ")是错误的，将修改为只读\r\n"
							填值 = "只读"
							row[标题列号[读写属性_下标]] = "只读"
							读写属性 = "只读"
						}
					}
				} //switch 变量类型{
			}
		}
		if 检查结果 != "" {
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[读写属性_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		检查结果 = ""
		if 变量ID == "" {
			检查结果 = "变量ID不能置空!" + "\r\n"
		}
		if 检查结果 != "" {
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[变量ID_下标]+1), strconv.Itoa(i))
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		if 变量名 == "" {
			检查结果 = "变量名不能置空!" + "\r\n"
		} else {
			检查结果 = 变量名检查(变量名)
			if 检查结果 == "ok" {
				检查结果 = ""
			}
			if _, ok := 变量名重复的行[i]; ok {
				检查结果 = "变量名(" + 变量名 + ")有重复" + "\r\n"
			}
		}
		if 检查结果 != "" {
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")" + 检查结果
			错误信息 = 错误信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 波特率_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定波特率，将被软件置空\r\n"
				row[标题列号[波特率_下标]] = ""
				波特率_s = ""
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[波特率_下标]+1), 填值)
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				变量警告信息 = 变量警告信息 + 检查结果
			}
		} else {
			if 波特率_s == "" {
				检查结果 = "变量类型(" + 变量类型 + ")需设定波特率，不能置空!\r\n"
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				错误信息 = 错误信息 + 检查结果
			} else {
				检查结果 = 波特率检查(波特率_s)
				if 检查结果 != "ok" {
					row[标题列号[波特率_下标]] = ""
					波特率_s = ""
					f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[波特率_下标]+1), 填值)
					遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
					错误信息 = 错误信息 + 检查结果
				}
			}
		}
		填值 = ""
		检查结果 = ""
		置空 = true
		警告 = false
		个数 := strings.Count(寄存器地址_s, ".")
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 寄存器地址_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定寄存器地址，将被软件置空\r\n"
				警告 = true
			}
			goto 寄存器地址_结果
		}
		if 寄存器地址_s == "" {
			检查结果 = "寄存器地址不能置空！\r\n"
			置空 = false
			goto 寄存器地址_结果
		}
		if 个数 > 1 {
			检查结果 = "寄存器地址(" + 寄存器地址_s + ")不被软件认可！将被软件置空\r\n"
			goto 寄存器地址_结果
		}
		if 个数 == 1 {
			if 变量类型 != "IO离散" && 变量类型 != "IO字符串" {
				检查结果 = "变量类型(" + 变量类型 + ")是没有这种寄存器地址设定(" + 寄存器地址_s + ")的，将被软件置空\r\n"
				goto 寄存器地址_结果
			}
			前值, 后值, 分离结果 := 小数点前后值分离(寄存器地址_s)
			if 分离结果 != "ok" {
				检查结果 = "寄存器地址(" + 寄存器地址_s + ")" + 分离结果 + ",将被软件置空\r\n"
				goto 寄存器地址_结果
			}
			if 前值 < 最小寄存器地址 || 前值 > 最大寄存器地址 {
				检查结果 = "寄存器地址(" + 寄存器地址_s + ")小数点前值不在正常范围内(" + 最小寄存器地址_s + "~" + 最大寄存器地址_s + ")内,将被软件置空\r\n"
				goto 寄存器地址_结果
			}
			switch 变量类型 {
			case "IO离散":
				if 后值 > 15 {
					检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")小数点后值不在正常范围内(0~15)内,将被软件置空\r\n"
					goto 寄存器地址_结果
				}
				IO离散点 = true
				if 读功能码 != "" {
					if 读功能码 != "3" && 读功能码 != "4" {
						f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[读功能码_下标]+1), "")
						检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")的读功能码(" + 读功能码 + ")不是3或4,将被软件置空\r\n"
						遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
						错误信息 = 错误信息 + 检查结果
						row[标题列号[读功能码_下标]] = ""
						读功能码 = ""
					} else {
						if 读功能码 == "4" && (读写属性 == "读写" || 读写属性 == "只写") {
							f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[读写属性_下标]+1), "只读")
							检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")的读功能码(" + 读功能码 + ")的读写属性(" + 读写属性 + ")错误,将被软件修改为只读\r\n"
							遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
							变量警告信息 = 变量警告信息 + 检查结果
							row[标题列号[读写属性_下标]] = "只读"
							读写属性 = "只读"
						}
						if 读功能码 == "3" && 读写属性 == "只写" {
							只写有效 = true
						}
					}
				}
			case "IO字符串":
				if 后值 > 最大字符串字节长度 {
					检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")小数点后值不在正常范围内(1~" + 最大字符串字节长度_s + ")内,将被软件置空\r\n"
					goto 寄存器地址_结果
				}
				if 后值 <= 0 {
					检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")小数点后值不在正常范围内(1~" + 最大字符串字节长度_s + ")内,将被软件置空\r\n"
					goto 寄存器地址_结果
				}
			default:
				检查结果 = "变量类型(" + 变量类型 + ")是没有这种寄存器地址设定(" + 寄存器地址_s + ")的，将被软件置空\r\n"
				goto 寄存器地址_结果
			} //switch 变量类型 {
			if 打包长度_s == "" {
				goto 寄存器地址_结果后
			}
			打包长度, err6 := strconv.Atoi(打包长度_s)
			if err6 != nil {
				goto 寄存器地址_结果后
			}
			if 打包长度 < 1 || 打包长度 > 最大字打包长度 {
				goto 寄存器地址_结果后
			}
			goto 寄存器地址_结果后
		} //if 个数 == 1 {
		if 个数 == 0 {
			switch 变量类型 {
			case "IO字符串":
				检查结果 = "变量类型(" + 变量类型 + ")是没有这种寄存器地址设定(" + 寄存器地址_s + ")的，将被软件置空\r\n"
				goto 寄存器地址_结果
			} //switch 变量类型 {
			寄存器地址, err6 := strconv.Atoi(寄存器地址_s)
			if err6 != nil {
				检查结果 = err6.Error() + "\r\n"
				goto 寄存器地址_结果
			}
			if 寄存器地址 < 最小寄存器地址 || 寄存器地址 > 最大寄存器地址 {
				检查结果 = "寄存器地址(" + 寄存器地址_s + ")不在正常范围内(" + 最小寄存器地址_s + "~" + 最大寄存器地址_s + ")内,将被软件置空\r\n"
				goto 寄存器地址_结果
			}
			寄存器地址2 := 寄存器地址 + 1
			寄存器地址2_s := strconv.Itoa(int(寄存器地址2))
			if ((变量类型 == "IO整型" || 变量类型 == "IO实型") && (数据类型 == "LONGBCD" || 数据类型 == "LONG" || 数据类型 == "ULONG")) || (变量类型 == "IO实型" && 数据类型 == "FLOAT") {
				if 寄存器地址2 > 最大寄存器地址 {
					检查结果 = "变量类型(" + 变量类型 + ")数据类型(" + 数据类型 + ")寄存器地址+1(" + 寄存器地址2_s + ")(此数据类型占两个字)大于最大寄存器地址(" + 最大寄存器地址_s + "),将被软件置空\r\n"
					goto 寄存器地址_结果
				}
			}
			switch 变量类型 {
			case "IO离散":
				if 读功能码 != "" {
					if 读功能码 != "1" && 读功能码 != "2" {
						f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[读功能码_下标]+1), "")
						检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")的读功能码(" + 读功能码 + ")不是1或2,将被软件置空\r\n"
						遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
						错误信息 = 错误信息 + 检查结果
						row[标题列号[读功能码_下标]] = ""
						读功能码 = ""
					} else {
						if 读功能码 == "2" && (读写属性 == "读写" || 读写属性 == "只写") {
							f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[读写属性_下标]+1), "只读")
							检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")的读功能码(" + 读功能码 + ")的读写属性(" + 读写属性 + ")错误,将被软件修改为只读\r\n"
							遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
							变量警告信息 = 变量警告信息 + 检查结果
							row[标题列号[读写属性_下标]] = "只读"
							读写属性 = "只读"
						}
						if 读功能码 == "1" && 读写属性 == "只写" {
							只写有效 = true
						}
					}
				}
			} //switch 变量类型 {
			if 打包长度_s == "" {
				goto 寄存器地址_结果后
			}
			打包长度, err6 := strconv.Atoi(打包长度_s)
			if err6 != nil {
				goto 寄存器地址_结果后
			} // else {
			if 打包长度 < 1 {
				goto 寄存器地址_结果后
			}
			清空 := false
			switch 变量类型 {
			case "IO离散":
				if 打包长度 > 最大位打包长度 {
					清空 = true
					检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")的打包长度(" + 打包长度_s + ") > 最大位打包长度(" + 最大位打包长度_s + ")，将被软件置空\r\n"
				}
			default:
				if 打包长度 > 最大字打包长度 {
					清空 = true
					检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")的打包长度(" + 打包长度_s + ") > 最大字打包长度(" + 最大字打包长度_s + ")，将被软件置空\r\n"
				}
			} //switch 变量类型 {
			if 清空 {
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[打包长度_下标]+1), 填值)
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				if 只写有效 {
					变量警告信息 = 变量警告信息 + 检查结果
				} else {
					错误信息 = 错误信息 + 检查结果
				}
				row[标题列号[打包长度_下标]] = ""
				打包长度_s = ""
				goto 寄存器地址_结果后
			}
			goto 寄存器地址_结果后
		} //if 个数 == 0 {
	寄存器地址_结果:
		if 检查结果 != "" {
			if 置空 {
				填值 = ""
				row[标题列号[寄存器地址_下标]] = ""
				寄存器地址_s = ""
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[寄存器地址_下标]+1), 填值)
			}
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			if 警告 {
				变量警告信息 = 变量警告信息 + 检查结果
			} else {
				错误信息 = 错误信息 + 检查结果
			}
		}
	寄存器地址_结果后:
		填值 = ""
		检查结果 = ""
		if 通讯异常值_s != "" {
			switch 变量类型 {
			case "IO实型", "内存实型":
				默认通讯异常值_FLOAT_s2 := ""
				switch 变量类型 {
				case "IO实型":
					默认通讯异常值_FLOAT_s2 = 默认通讯异常值_FLOAT_s
				case "内存实型":
					默认通讯异常值_FLOAT_s2 = 内存默认通讯异常值_FLOAT_s
				}
				if 通讯异常值_s != 默认通讯异常值_FLOAT_s2 {
					浮点数值, err := strconv.ParseFloat(通讯异常值_s, 64)
					if err != nil {
						检查结果 = "通讯异常值(" + 通讯异常值_s + ")不是一个64位浮点数,将使用默认通讯异常值_FLOAT代替(" + 默认通讯异常值_FLOAT_s2 + ")\r\n"
					} else {
						var 最小浮点数2 float64
						var 最大浮点数2 float64
						最小浮点数2_s := ""
						最大浮点数2_s := ""
						switch 变量类型 {
						case "IO实型":
							最小浮点数2 = 最小浮点数
							最大浮点数2 = 最大浮点数
							最小浮点数2_s = 最小浮点数_s
							最大浮点数2_s = 最大浮点数_s
						case "内存实型":
							最小浮点数2 = 内存最小浮点数
							最大浮点数2 = 内存最大浮点数
							最小浮点数2_s = 内存最小浮点数_s
							最大浮点数2_s = 内存最大浮点数_s
						}
						if 浮点数值 >= 最小浮点数2 && 浮点数值 <= 最大浮点数2 {
							检查结果 = "通讯异常值(" + 通讯异常值_s + ")在变量类型(" + 变量类型 + ")的正常值范围内(" + 最小浮点数2_s + "~" + 最大浮点数2_s + "),将使用默认通讯异常值_FLOAT代替(" + 默认通讯异常值_FLOAT_s2 + ")\r\n"
						}
					}
					if 检查结果 != "" {
						填值 = 默认通讯异常值_FLOAT_s2
						row[标题列号[通讯异常值_下标]] = 默认通讯异常值_FLOAT_s2
						通讯异常值_s = 默认通讯异常值_FLOAT_s2
					}
				}
			case "内存整型":
				if 通讯异常值_s != 内存默认通讯异常值_LONG_s {
					整型值, err := strconv.ParseInt(通讯异常值_s, 10, 64)
					if err != nil {
						检查结果 = "通讯异常值(" + 通讯异常值_s + ")不是一个64位整数,将使用" + 变量类型 + "的内存默认通讯异常值_LONG代替(" + 内存默认通讯异常值_LONG_s + ")\r\n"
					} else {
						整型最小值_s := "0"
						整型最大值_s := "0"
						var 整型最小值, 整型最大值 int64
						整型最小值 = 内存整型最小值
						整型最大值 = 内存整型最大值
						整型最小值_s = 内存整型最小值_s
						整型最大值_s = 内存整型最大值_s
						if 整型值 <= 整型最大值 && 整型值 >= 整型最小值 {
							检查结果 = "通讯异常值(" + 通讯异常值_s + ")在变量类型(" + 变量类型 + ")的正常值范围内(" + 整型最小值_s + "~" + 整型最大值_s + "),将使用内存默认通讯异常值_LONG代替(" + 内存默认通讯异常值_LONG_s + ")\r\n"
						}
					}
				}
				if 检查结果 != "" {
					填值 = 内存默认通讯异常值_LONG_s
					row[标题列号[通讯异常值_下标]] = 内存默认通讯异常值_LONG_s
					通讯异常值_s = 内存默认通讯异常值_LONG_s
				}
			case "IO整型":
				switch 数据类型 {
				case "LONG":
					if 通讯异常值_s != 默认通讯异常值_LONG_s {
						整型值, err := strconv.ParseInt(通讯异常值_s, 10, 64)
						if err != nil {
							检查结果 = "通讯异常值(" + 通讯异常值_s + ")不是一个64位整数,将使用" + 数据类型 + "的默认通讯异常值_LONG代替(" + 默认通讯异常值_LONG_s + ")\r\n"
						} else {
							var 整型最小值 int64 = int32最小值
							var 整型最大值 int64 = int32最大值
							整型最小值_s := int32最小值_s
							整型最大值_s := int32最大值_s
							if 整型值 <= 整型最大值 && 整型值 >= 整型最小值 {
								检查结果 = "通讯异常值(" + 通讯异常值_s + ")在数据类型(" + 数据类型 + ")的正常值范围内(" + 整型最小值_s + "~" + 整型最大值_s + "),将使用默认通讯异常值_LONG代替(" + 默认通讯异常值_LONG_s + ")\r\n"
							}
						}
					}
					if 检查结果 != "" {
						填值 = 默认通讯异常值_LONG_s
						row[标题列号[通讯异常值_下标]] = 默认通讯异常值_LONG_s
						通讯异常值_s = 默认通讯异常值_LONG_s
					}
				case "ULONG":
					if 通讯异常值_s != 默认通讯异常值_ULONG_s {
						整型值, err := strconv.ParseInt(通讯异常值_s, 10, 64)
						if err != nil {
							检查结果 = "通讯异常值(" + 通讯异常值_s + ")不是一个64位整数,将使用" + 数据类型 + "的默认通讯异常值_ULONG代替(" + 默认通讯异常值_ULONG_s + ")\r\n"
						} else {
							var 整型最小值 int64 = 0
							var 整型最大值 int64 = 4294967295
							整型最小值_s := "0"
							整型最大值_s := "4294967295"
							if 整型值 <= 整型最大值 && 整型值 >= 整型最小值 {
								检查结果 = "通讯异常值(" + 通讯异常值_s + ")在数据类型(" + 数据类型 + ")的正常值范围内(" + 整型最小值_s + "~" + 整型最大值_s + "),将使用默认通讯异常值_ULONG代替(" + 默认通讯异常值_ULONG_s + ")\r\n"
							}
						}
					}
					if 检查结果 != "" {
						填值 = 默认通讯异常值_ULONG_s
						row[标题列号[通讯异常值_下标]] = 默认通讯异常值_ULONG_s
						通讯异常值_s = 默认通讯异常值_ULONG_s
					}
				case "LONGBCD":
					if 通讯异常值_s != 默认通讯异常值_LONGBCD_s {
						整型值, err := strconv.ParseInt(通讯异常值_s, 10, 32)
						if err != nil {
							检查结果 = "通讯异常值(" + 通讯异常值_s + ")不是一个32位整数,将使用" + 数据类型 + "的默认通讯异常值_LONGBCD代替(" + 默认通讯异常值_LONGBCD_s + ")\r\n"
						} else {
							var 整型最小值 int64 = 0
							var 整型最大值 int64 = 99999999
							整型最小值_s := "0"
							整型最大值_s := "99999999"
							if 整型值 <= 整型最大值 && 整型值 >= 整型最小值 {
								检查结果 = "通讯异常值(" + 通讯异常值_s + ")在数据类型(" + 数据类型 + ")的正常值范围内(" + 整型最小值_s + "~" + 整型最大值_s + "),将使用默认通讯异常值_LONGBCD代替(" + 默认通讯异常值_LONGBCD_s + ")\r\n"
							}
						}
					}
					if 检查结果 != "" {
						填值 = 默认通讯异常值_LONGBCD_s
						row[标题列号[通讯异常值_下标]] = 默认通讯异常值_LONGBCD_s
						通讯异常值_s = 默认通讯异常值_LONGBCD_s
					}
				case "BCD":
					if 通讯异常值_s != 默认通讯异常值_BCD_s {
						整型值, err := strconv.ParseInt(通讯异常值_s, 10, 16)
						if err != nil {
							检查结果 = "通讯异常值(" + 通讯异常值_s + ")不是一个16位整数,将使用" + 数据类型 + "的默认通讯异常值_BCD代替(" + 默认通讯异常值_BCD_s + ")\r\n"
						} else {
							var 整型最小值 int64 = 0
							var 整型最大值 int64 = 9999
							整型最小值_s := "0"
							整型最大值_s := "9999"
							if 整型值 <= 整型最大值 && 整型值 >= 整型最小值 {
								检查结果 = "通讯异常值(" + 通讯异常值_s + ")在数据类型(" + 数据类型 + ")的正常值范围内(" + 整型最小值_s + "~" + 整型最大值_s + "),将使用默认通讯异常值_BCD代替(" + 默认通讯异常值_BCD_s + ")\r\n"
							}
						}
					}
					if 检查结果 != "" {
						填值 = 默认通讯异常值_BCD_s
						row[标题列号[通讯异常值_下标]] = 默认通讯异常值_BCD_s
						通讯异常值_s = 默认通讯异常值_BCD_s
					}
				case "SHORT":
					if 通讯异常值_s != 默认通讯异常值_SHORT_s {
						整型值, err := strconv.ParseInt(通讯异常值_s, 10, 32)
						if err != nil {
							检查结果 = "通讯异常值(" + 通讯异常值_s + ")不是一个32位整数,将使用" + 数据类型 + "的默认通讯异常值_SHORT代替(" + 默认通讯异常值_SHORT_s + ")\r\n"
						} else {
							var 整型最小值 int64 = -32768
							var 整型最大值 int64 = 32767
							整型最小值_s := "-32768"
							整型最大值_s := "32767"
							if 整型值 <= 整型最大值 && 整型值 >= 整型最小值 {
								检查结果 = "通讯异常值(" + 通讯异常值_s + ")在数据类型(" + 数据类型 + ")的正常值范围内(" + 整型最小值_s + "~" + 整型最大值_s + "),将使用默认通讯异常值_SHORT代替(" + 默认通讯异常值_SHORT_s + ")\r\n"
							}
						}
					}
					if 检查结果 != "" {
						填值 = 默认通讯异常值_SHORT_s
						row[标题列号[通讯异常值_下标]] = 默认通讯异常值_SHORT_s
						通讯异常值_s = 默认通讯异常值_SHORT_s
					}
				case "USHORT":
					if 通讯异常值_s != 默认通讯异常值_USHORT_s {
						整型值, err := strconv.ParseInt(通讯异常值_s, 10, 32)
						if err != nil {
							检查结果 = "通讯异常值(" + 通讯异常值_s + ")不是一个32位整数,将使用" + 数据类型 + "的默认通讯异常值_USHORT代替(" + 默认通讯异常值_USHORT_s + ")\r\n"
						} else {
							var 整型最小值 int64 = 0
							var 整型最大值 int64 = 65535
							整型最小值_s := "0"
							整型最大值_s := "65535"
							if 整型值 <= 整型最大值 && 整型值 >= 整型最小值 {
								检查结果 = "通讯异常值(" + 通讯异常值_s + ")在数据类型(" + 数据类型 + ")的正常值范围内(" + 整型最小值_s + "~" + 整型最大值_s + "),将使用默认通讯异常值_USHORT代替(" + 默认通讯异常值_USHORT_s + ")\r\n"
							}
						}
					}
					if 检查结果 != "" {
						填值 = 默认通讯异常值_USHORT_s
						row[标题列号[通讯异常值_下标]] = 默认通讯异常值_USHORT_s
						通讯异常值_s = 默认通讯异常值_USHORT_s
					}
				default:
				} //switch 数据类型{
			case "IO字符串", "内存字符串":
				字符串长度 := len(通讯异常值_s)
				if 字符串长度 > 最大字符串字节长度 {
					字符串长度_s := strconv.Itoa(字符串长度)
					检查结果 = "通讯异常值(" + 通讯异常值_s + ")字符串长度(" + 字符串长度_s + ")大于最大字符串字节长度(" + 最大字符串字节长度_s + "),将使用默认通讯异常值_STRING代替(" + 默认通讯异常值_STRING + ")\r\n"
				}
				if 检查结果 != "" {
					填值 = 默认通讯异常值_STRING
					row[标题列号[通讯异常值_下标]] = 默认通讯异常值_STRING
					通讯异常值_s = 默认通讯异常值_STRING
				}
			case "IO离散", "内存离散":
				if 通讯异常值_s != 默认通讯异常值_BIT_s {
					if 通讯异常值_s == "0" || 通讯异常值_s == "1" {
						检查结果 = "通讯异常值(" + 通讯异常值_s + ")在变量类型(" + 变量类型 + ")的正常值范围内(0或1),将使用默认通讯异常值_BIT代替(" + 默认通讯异常值_BIT_s + ")\r\n"
					} else {
						_, err := strconv.ParseInt(通讯异常值_s, 10, 8)
						if err != nil {
							检查结果 = "通讯异常值(" + 通讯异常值_s + ")不是一个8位整数,将使用" + 变量类型 + "的默认通讯异常值_BIT代替(" + 默认通讯异常值_BIT_s + ")\r\n"
						}
					}
				}
				if 检查结果 != "" {
					填值 = 默认通讯异常值_BIT_s
					row[标题列号[通讯异常值_下标]] = 默认通讯异常值_BIT_s
					通讯异常值_s = 默认通讯异常值_BIT_s
				}
			default:
			} //switch 变量类型 {
			if 检查结果 != "" {
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[通讯异常值_下标]+1), 填值)
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				变量警告信息 = 变量警告信息 + 检查结果
			}
		} else { //if 通讯异常值_s != "" {
			switch 变量类型 {
			case "IO实型", "内存实型":
				默认通讯异常值_FLOAT_s2 := ""
				switch 变量类型 {
				case "IO实型":
					默认通讯异常值_FLOAT_s2 = 默认通讯异常值_FLOAT_s
				case "内存实型":
					默认通讯异常值_FLOAT_s2 = 内存默认通讯异常值_FLOAT_s
				}
				填值 = 默认通讯异常值_FLOAT_s2
				row[标题列号[通讯异常值_下标]] = 默认通讯异常值_FLOAT_s2
				通讯异常值_s = 默认通讯异常值_FLOAT_s2
			case "内存整型":
				填值 = 内存默认通讯异常值_LONG_s
				row[标题列号[通讯异常值_下标]] = 内存默认通讯异常值_LONG_s
				通讯异常值_s = 内存默认通讯异常值_LONG_s
			case "IO整型":
				switch 数据类型 {
				case "LONG":
					填值 = 默认通讯异常值_LONG_s
					row[标题列号[通讯异常值_下标]] = 默认通讯异常值_LONG_s
					通讯异常值_s = 默认通讯异常值_LONG_s
				case "ULONG":
					填值 = 默认通讯异常值_ULONG_s
					row[标题列号[通讯异常值_下标]] = 默认通讯异常值_ULONG_s
					通讯异常值_s = 默认通讯异常值_ULONG_s
				case "LONGBCD":
					填值 = 默认通讯异常值_LONGBCD_s
					row[标题列号[通讯异常值_下标]] = 默认通讯异常值_LONGBCD_s
					通讯异常值_s = 默认通讯异常值_LONGBCD_s
				case "BCD":
					填值 = 默认通讯异常值_BCD_s
					row[标题列号[通讯异常值_下标]] = 默认通讯异常值_BCD_s
					通讯异常值_s = 默认通讯异常值_BCD_s
				case "SHORT":
					填值 = 默认通讯异常值_SHORT_s
					row[标题列号[通讯异常值_下标]] = 默认通讯异常值_SHORT_s
					通讯异常值_s = 默认通讯异常值_SHORT_s
				case "USHORT":
					填值 = 默认通讯异常值_USHORT_s
					row[标题列号[通讯异常值_下标]] = 默认通讯异常值_USHORT_s
					通讯异常值_s = 默认通讯异常值_USHORT_s
				default:
				} //switch 数据类型{
			case "IO字符串", "内存字符串":
				填值 = 默认通讯异常值_STRING
				row[标题列号[通讯异常值_下标]] = 默认通讯异常值_STRING
				通讯异常值_s = 默认通讯异常值_STRING
			case "IO离散", "内存离散":
				填值 = 默认通讯异常值_BIT_s
				row[标题列号[通讯异常值_下标]] = 默认通讯异常值_BIT_s
				通讯异常值_s = 默认通讯异常值_BIT_s
			default:
			} //switch 变量类型 {
			if 填值 != "" {
				检查结果 = "变量类型(" + 变量类型 + ")通讯异常值不能留空！将使用默认通讯异常值(" + 通讯异常值_s + ")代替！\r\n"
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[通讯异常值_下标]+1), 填值)
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				变量警告信息 = 变量警告信息 + 检查结果
			}
		} //}else{//if 通讯异常值_s != "" {
		检查结果 = ""
		switch 变量类型 {
		case "IO离散", "内存离散":
			if 初始值_s == "" {
				检查结果 = "初始值不能置空,将使用通讯异常值代替(" + 通讯异常值_s + ")" + "\r\n"
			} else {
				var 最小整型, 最大整型 int64 = 0, 1
				if 初始值_s != 通讯异常值_s {
					整型值, err := strconv.ParseInt(初始值_s, 10, 8)
					if err != nil {
						检查结果 = "初始值(" + 初始值_s + ")不是一个8位整型,将使用通讯异常值代替(" + 通讯异常值_s + ")" + "\r\n"
					} else {
						if 整型值 < 最小整型 {
							检查结果 = "初始值(" + 初始值_s + ")小于离散最小值0,将使用通讯异常值代替(" + 通讯异常值_s + ")" + "\r\n"
						} else {
							if 整型值 > 最大整型 {
								检查结果 = "初始值(" + 初始值_s + ")大于离散最大值1,将使用通讯异常值代替(" + 通讯异常值_s + ")" + "\r\n"
							}
						}
					}
				} //if 初始值_s != 通讯异常值_s{
			}
			if 检查结果 != "" {
				row[标题列号[初始值_下标]] = 通讯异常值_s
				初始值_s = 通讯异常值_s
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[初始值_下标]+1), 通讯异常值_s)
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				变量警告信息 = 变量警告信息 + 检查结果
			}
		case "IO字符串", "内存字符串":
			if 初始值_s == "" {
				检查结果 = "初始值不能置空,将使用通讯异常值代替(" + 通讯异常值_s + ")" + "\r\n"
			} else {
				if 初始值_s != 通讯异常值_s {
					字符串长度 := len(初始值_s)
					最大字符串字节长度2 := 0
					最大字符串字节长度2_s := ""
					switch 变量类型 {
					case "IO字符串":
						最大字符串字节长度2 = 最大字符串字节长度
						最大字符串字节长度2_s = 最大字符串字节长度_s
					case "内存字符串":
						最大字符串字节长度2 = 内存字符串最多占用字节数
						最大字符串字节长度2_s = 内存字符串最多占用字节数_s
					}
					if 字符串长度 > 最大字符串字节长度2 {
						字符串长度_s := strconv.Itoa(字符串长度)
						检查结果 = "初始值(" + 初始值_s + ")字符串长度(" + 字符串长度_s + ")大于最大字符串字节长度(" + 最大字符串字节长度2_s + "),将使用通讯异常值代替(" + 通讯异常值_s + ")\r\n"
					}
				}
			}
			if 检查结果 != "" {
				row[标题列号[初始值_下标]] = 通讯异常值_s
				初始值_s = 通讯异常值_s
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[初始值_下标]+1), 通讯异常值_s)
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				变量警告信息 = 变量警告信息 + 检查结果
			}
		case "IO实型", "内存实型":
			if 初始值_s == "" {
				检查结果 = "初始值不能置空,将使用通讯异常值代替(" + 通讯异常值_s + ")\r\n"
			} else {
				if 初始值_s != 通讯异常值_s {
					浮点数值, err := strconv.ParseFloat(初始值_s, 64)
					if err != nil {
						检查结果 = "初始值(" + 初始值_s + ")不是一个64位浮点数,将使用通讯异常值代替(" + 通讯异常值_s + ")\r\n"
					} else {
						var 最小浮点数2 float64
						var 最大浮点数2 float64
						最小浮点数2_s := ""
						最大浮点数2_s := ""
						switch 变量类型 {
						case "IO实型":
							最小浮点数2 = 最小浮点数
							最大浮点数2 = 最大浮点数
							最小浮点数2_s = 最小浮点数_s
							最大浮点数2_s = 最大浮点数_s
						case "内存实型":
							最小浮点数2 = 内存最小浮点数
							最大浮点数2 = 内存最大浮点数
							最小浮点数2_s = 内存最小浮点数_s
							最大浮点数2_s = 内存最大浮点数_s
						}
						if 浮点数值 < 最小浮点数2 {
							检查结果 = "初始值(" + 初始值_s + ")小于最小浮点数设定(" + 最小浮点数2_s + "),将使用通讯异常值代替(" + 通讯异常值_s + ")\r\n"
						} else {
							if 浮点数值 > 最大浮点数2 {
								检查结果 = "初始值(" + 初始值_s + ")大于最大浮点数设定(" + 最大浮点数2_s + "),将使用通讯异常值代替(" + 通讯异常值_s + ")\r\n"
							}
						}
					}
				} //if 初始值_s != 通讯异常值_s{
			}
			if 检查结果 != "" {
				row[标题列号[初始值_下标]] = 通讯异常值_s
				初始值_s = 通讯异常值_s
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[初始值_下标]+1), 通讯异常值_s)
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				变量警告信息 = 变量警告信息 + 检查结果
			}
		case "IO整型":
			整型最小值_s := "0"
			整型最大值_s := "0"
			var 整型值, 整型最小值, 整型最大值 int64
			switch 数据类型 {
			case "LONGBCD":
				整型最小值 = 0
				整型最大值 = 99999999
				整型最小值_s = "0"
				整型最大值_s = "99999999"
			case "BCD":
				整型最小值 = 0
				整型最大值 = 9999
				整型最小值_s = "0"
				整型最大值_s = "9999"
			case "LONG":
				整型最小值 = int32最小值
				整型最大值 = int32最大值
				整型最小值_s = int32最小值_s
				整型最大值_s = int32最大值_s
			case "ULONG":
				整型最小值 = 0
				整型最大值 = 4294967295
				整型最小值_s = "0"
				整型最大值_s = "4294967295"
			case "SHORT":
				整型最小值 = -32768
				整型最大值 = 32767
				整型最小值_s = "-32768"
				整型最大值_s = "32767"
			case "USHORT":
				整型最小值 = 0
				整型最大值 = 65535
				整型最小值_s = "0"
				整型最大值_s = "65535"
			} //switch 数据类型{
			if 整型最大值_s != "0" {
				if 初始值_s == "" {
					检查结果 = "初始值不能置空,将使用通讯异常值代替(" + 通讯异常值_s + ")\r\n"
				} else {
					if 初始值_s != 通讯异常值_s {
						整型值, err = strconv.ParseInt(初始值_s, 10, 64)
						if err != nil {
							检查结果 = "初始值(" + 初始值_s + ")不是一个64位整数,将使用" + 数据类型 + "的通讯异常值代替(" + 通讯异常值_s + ")\r\n"
						} else {
							if 整型值 < 整型最小值 {
								检查结果 = "初始值(" + 初始值_s + ")小于数据类型(" + 数据类型 + ")的最小值(" + 整型最小值_s + ")将使用通讯异常值代替(" + 通讯异常值_s + ")\r\n"
							} else {
								if 整型值 > 整型最大值 {
									检查结果 = "初始值(" + 初始值_s + ")大于数据类型(" + 数据类型 + ")的最大值(" + 整型最大值_s + ")将使用通讯异常值代替(" + 通讯异常值_s + ")\r\n"
								}
							}
						}
					} //if 初始值_s != 通讯异常值_s{
				}
			} //if 默认通讯异常值_s != "ding" {
			if 检查结果 != "" {
				row[标题列号[初始值_下标]] = 通讯异常值_s
				初始值_s = 通讯异常值_s
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[初始值_下标]+1), 通讯异常值_s)
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				变量警告信息 = 变量警告信息 + 检查结果
			}
		case "内存整型":
			整型最小值_s := "0"
			整型最大值_s := "0"
			var 整型值, 整型最小值, 整型最大值 int64
			整型最小值 = 内存整型最小值
			整型最大值 = 内存整型最大值
			整型最小值_s = 内存整型最小值_s
			整型最大值_s = 内存整型最大值_s
			if 初始值_s == "" {
				检查结果 = "初始值不能置空,将使用通讯异常值代替(" + 通讯异常值_s + ")\r\n"
			} else {
				if 初始值_s != 通讯异常值_s {
					整型值, err = strconv.ParseInt(初始值_s, 10, 64)
					if err != nil {
						检查结果 = "初始值(" + 初始值_s + ")不是一个64位整数,将使用" + 变量类型 + "的通讯异常值代替(" + 通讯异常值_s + ")\r\n"
					} else {
						if 整型值 < 整型最小值 {
							检查结果 = "初始值(" + 初始值_s + ")小于变量类型(" + 变量类型 + ")的最小值(" + 整型最小值_s + ")将使用通讯异常值代替(" + 通讯异常值_s + ")\r\n"
						} else {
							if 整型值 > 整型最大值 {
								检查结果 = "初始值(" + 初始值_s + ")大于变量类型(" + 变量类型 + ")的最大值(" + 整型最大值_s + ")将使用通讯异常值代替(" + 通讯异常值_s + ")\r\n"
							}
						}
					}
				} //if 初始值_s != 通讯异常值_s {
			}
			if 检查结果 != "" {
				row[标题列号[初始值_下标]] = 通讯异常值_s
				初始值_s = 通讯异常值_s
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[初始值_下标]+1), 通讯异常值_s)
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				变量警告信息 = 变量警告信息 + 检查结果
			}
		} //switch 变量类型{
		填值 = ""
		检查结果 = ""
		if 是否保存值_s == "" {
			检查结果 = "是否保存值不可置空，将使用默认(否)!\r\n"
		} else {
			if 是否保存值_s != "是" && 是否保存值_s != "否" {
				检查结果 = "是否保存值(" + 是否保存值_s + ")既不是(是)也不是(否)，将使用默认(否)!\r\n"
			}
		}
		if 检查结果 != "" {
			填值 = "否"
			row[标题列号[是否保存值_下标]] = "否"
			是否保存值_s = "否"
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[是否保存值_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		置空 = false
		if 只写有效 {
			if 读字节序_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定读字节序，将被软件置空\r\n"
				置空 = true
			}
		} else {
			if 变量类型 == "IO离散" || 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
				if 读字节序_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")无需设定读字节序，将被软件置空\r\n"
					置空 = true
				}
			} else {
				if (变量类型 == "IO整型" || 变量类型 == "IO实型") && (数据类型 == "BCD" || 数据类型 == "SHORT" || 数据类型 == "USHORT") {
					if 读字节序_s != "" {
						检查结果 = "变量类型(" + 变量类型 + ")数据类型(" + 数据类型 + ")无需设定读字节序，将被软件置空\r\n"
						置空 = true
					}
				} else {
					使用默认设置 := false
					if 读字节序_s != "" {
						if 读字节序_s != "ABCD" && 读字节序_s != "CDAB" && 读字节序_s != "BADC" && 读字节序_s != "DCBA" {
							检查结果 = "读字节序(" + 读字节序_s + ")不是(ABCD、CDAB、BADC、DCBA)之一，将使用默认(ABCD)!\r\n"
							使用默认设置 = true
						}
					} else {
						检查结果 = "读字节序不可置空，将使用默认(ABCD)!\r\n"
						使用默认设置 = true
					}
					if 使用默认设置 {
						填值 = "ABCD"
						row[标题列号[读字节序_下标]] = "ABCD"
						读字节序_s = "ABCD"
					}
				}
			}
		}
		if 置空 {
			row[标题列号[读字节序_下标]] = ""
			读字节序_s = ""
		}
		if 检查结果 != "" {
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[读字节序_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		置空 = false
		if 变量类型 == "IO离散" || 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 写字节序_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定写字节序，将被软件置空\r\n"
				置空 = true
			}
		} else {
			if (变量类型 == "IO整型" || 变量类型 == "IO实型") && (数据类型 == "BCD" || 数据类型 == "SHORT" || 数据类型 == "USHORT") {
				if 写字节序_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")数据类型(" + 数据类型 + ")无需设定写字节序，将被软件置空\r\n"
					置空 = true
				}
			} else {
				使用默认设置 := false
				if 写字节序_s != "" {
					if 读写属性 == "只读" {
						检查结果 = "变量类型(" + 变量类型 + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定写字节序，将被软件置空\r\n"
						置空 = true
					} else {
						if 写字节序_s != "ABCD" && 写字节序_s != "CDAB" && 写字节序_s != "BADC" && 写字节序_s != "DCBA" {
							检查结果 = "写字节序(" + 写字节序_s + ")不是(ABCD、CDAB、BADC、DCBA)之一，将使用默认(ABCD)!\r\n"
							使用默认设置 = true
						}
					}
				} else {
					if 读写属性 != "只读" {
						检查结果 = "写字节序不可置空，将使用默认(ABCD)!\r\n"
						使用默认设置 = true
					}
				}
				if 使用默认设置 {
					填值 = "ABCD"
					row[标题列号[写字节序_下标]] = "ABCD"
					写字节序_s = "ABCD"
				}
			}
		}
		if 置空 {
			row[标题列号[写字节序_下标]] = ""
			写字节序_s = ""
		}
		if 检查结果 != "" {
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[写字节序_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		置空 = false
		switch 变量类型 {
		case "内存离散", "内存字符串", "内存整型", "内存实型":
			if 允许通讯异常后只读变量_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定允许通讯异常后只读变量，将被软件置空\r\n"
			}
			置空 = true
		case "IO整型", "IO实型":
			if 只写有效 {
				if 允许通讯异常后只读变量_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")数据类型(" + 数据类型 + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定允许通讯异常后只读变量，将被软件置空\r\n"
				}
			}
		case "IO字符串":
			if 只写有效 {
				if 允许通讯异常后只读变量_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定允许通讯异常后只读变量，将被软件置空\r\n"
				}
			}
		case "IO离散":
			if 只写有效 {
				if 允许通讯异常后只读变量_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定允许通讯异常后只读变量，将被软件置空\r\n"
				}
			}
		} //switch 变量类型 {
		if 只写有效 || 置空 {
			if 允许通讯异常后只读变量_s != "" {
				row[标题列号[允许通讯异常后只读变量_下标]] = ""
				允许通讯异常后只读变量_s = ""
			}
		} else {
			不允许 := false
			if 允许通讯异常后只读变量_s == "" {
				检查结果 = "允许通讯异常后只读变量不可置空，将使用默认(不允许)!\r\n"
				不允许 = true
			} else {
				if 允许通讯异常后只读变量_s != "允许" && 允许通讯异常后只读变量_s != "不允许" {
					检查结果 = "允许通讯异常后只读变量(" + 允许通讯异常后只读变量_s + ")既不是(允许)也不是(不允许)，将使用默认(不允许)!\r\n"
					不允许 = true
				}
			}
			if 不允许 {
				填值 = "不允许"
				row[标题列号[允许通讯异常后只读变量_下标]] = "不允许"
				允许通讯异常后只读变量_s = "不允许"
			}
		}
		if 检查结果 != "" {
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[允许通讯异常后只读变量_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 设备地址_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定设备地址，将被软件置空\r\n"
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[设备地址_下标]+1), 填值)
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				变量警告信息 = 变量警告信息 + 检查结果
				row[标题列号[设备地址_下标]] = ""
				设备地址_s = ""
			}
		} else {
			if 变量类型 == "IO离散" || 变量类型 == "IO字符串" || 变量类型 == "IO整型" || 变量类型 == "IO实型" {
				if 设备地址_s == "" {
					检查结果 = "设备地址不能置空!\r\n"
					遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
					错误信息 = 错误信息 + 检查结果
				} else {
					整型值, err := strconv.ParseUint(设备地址_s, 10, 8)
					if err != nil {
						检查结果 = "设备地址(" + 设备地址_s + ")不是一个8位无符号整型,将被置空\r\n"
					} else {
						if 整型值 < 最小设备地址 {
							检查结果 = "设备地址(" + 设备地址_s + ")小于最小设备地址(" + 最小设备地址_s + ")，将被置空\r\n"
						} else {
							if 整型值 > 最大设备地址 {
								检查结果 = "设备地址(" + 设备地址_s + ")大于最大设备地址(" + 最大设备地址_s + ")，将被置空\r\n"
							}
						}
					}
					if 检查结果 != "" {
						f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[设备地址_下标]+1), 填值)
						遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
						错误信息 = 错误信息 + 检查结果
						row[标题列号[设备地址_下标]] = ""
						设备地址_s = ""
					}
				}
			}
		}
		填值 = ""
		检查结果 = ""
		置空 = false
		switch 变量类型 {
		case "内存离散", "内存字符串", "内存整型", "内存实型":
			if 连续通讯失败多少次则认为通讯异常_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定连续通讯失败多少次则认为通讯异常，将被软件置空\r\n"
			}
			置空 = true
		case "IO整型", "IO实型":
			if 只写有效 {
				if 连续通讯失败多少次则认为通讯异常_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")数据类型(" + 数据类型 + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定连续通讯失败多少次则认为通讯异常，将被软件置空\r\n"
				}
			}
		case "IO字符串":
			if 只写有效 {
				if 连续通讯失败多少次则认为通讯异常_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定连续通讯失败多少次则认为通讯异常，将被软件置空\r\n"
				}
			}
		case "IO离散":
			if 只写有效 {
				if 连续通讯失败多少次则认为通讯异常_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定连续通讯失败多少次则认为通讯异常，将被软件置空\r\n"
				}
			}
		} //switch 变量类型 {
		if 只写有效 || 置空 {
			if 连续通讯失败多少次则认为通讯异常_s != "" {
				row[标题列号[连续通讯失败多少次则认为通讯异常_下标]] = ""
				连续通讯失败多少次则认为通讯异常_s = ""
			}
		} else {
			使用默认设置 := true
			if 连续通讯失败多少次则认为通讯异常_s == "" {
				检查结果 = "连续通讯失败多少次则认为通讯异常不能置空,将使用默认连续通讯失败多少次则认为通讯异常代替(" + 默认连续通讯失败多少次则认为通讯异常_s + ")\r\n"
			} else {
				整型值, err := strconv.ParseUint(连续通讯失败多少次则认为通讯异常_s, 10, 16)
				if err != nil {
					检查结果 = "连续通讯失败多少次则认为通讯异常(" + 连续通讯失败多少次则认为通讯异常_s + ")不是一个16位无符号整型,将使用默认连续通讯失败多少次则认为通讯异常代替(" + 默认连续通讯失败多少次则认为通讯异常_s + ")\r\n"
				} else {
					if 整型值 < 1 {
						检查结果 = "连续通讯失败多少次则认为通讯异常(" + 连续通讯失败多少次则认为通讯异常_s + ")小于1，将使用默认连续通讯失败多少次则认为通讯异常代替(" + 默认连续通讯失败多少次则认为通讯异常_s + ")\r\n"
					} else {
						使用默认设置 = false
					}
				}
			}
			if 使用默认设置 {
				填值 = 默认连续通讯失败多少次则认为通讯异常_s
				row[标题列号[连续通讯失败多少次则认为通讯异常_下标]] = 默认连续通讯失败多少次则认为通讯异常_s
				连续通讯失败多少次则认为通讯异常_s = 默认连续通讯失败多少次则认为通讯异常_s
			}
		}
		if 检查结果 != "" {
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[连续通讯失败多少次则认为通讯异常_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		置空 = false
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 通讯超时_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定通讯超时，将被软件置空\r\n"
				置空 = true
			}
		} else {
			if 通讯超时_s == "" {
				检查结果 = "通讯超时不能置空,将使用默认通讯超时_毫秒代替(" + 默认通讯超时_毫秒_s + ")\r\n"
			} else {
				整型值, err := strconv.ParseUint(通讯超时_s, 10, 16)
				if err != nil {
					检查结果 = "通讯超时(" + 通讯超时_s + ")不是一个16位无符号整型,将使用默认通讯超时_毫秒代替(" + 默认通讯超时_毫秒_s + ")\r\n"
				} else {
					if 整型值 < 最小时间 || 整型值 > 最大时间 {
						检查结果 = "通讯超时(" + 通讯超时_s + ")不在正常范围内(" + 最小时间_s + "~" + 最大时间_s + ")，将使用默认通讯超时_毫秒代替(" + 默认通讯超时_毫秒_s + ")\r\n"
					}
				}
			}
		}
		if 检查结果 != "" {
			if 置空 {
				填值 = ""
				row[标题列号[通讯超时_下标]] = ""
				通讯超时_s = ""
			} else {
				填值 = 默认通讯超时_毫秒_s
				row[标题列号[通讯超时_下标]] = 默认通讯超时_毫秒_s
				通讯超时_s = 默认通讯超时_毫秒_s
			}
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[通讯超时_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		置空 = false
		switch 变量类型 {
		case "内存离散", "内存字符串", "内存整型", "内存实型":
			if 采集频率_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定采集频率，将被软件置空\r\n"
			}
			置空 = true
		case "IO整型", "IO实型":
			if 只写有效 {
				if 采集频率_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")数据类型(" + 数据类型 + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定采集频率，将被软件置空\r\n"
				}
			}
		case "IO字符串":
			if 只写有效 {
				if 采集频率_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定采集频率，将被软件置空\r\n"
				}
			}
		case "IO离散":
			if 只写有效 {
				if 采集频率_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")寄存器地址(" + 寄存器地址_s + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定采集频率，将被软件置空\r\n"
				}
			}
		} //switch 变量类型 {
		if 只写有效 || 置空 {
			if 采集频率_s != "" {
				row[标题列号[采集频率_下标]] = ""
				采集频率_s = ""
			}
		} else {
			使用默认设置 := true
			if 采集频率_s == "" {
				检查结果 = "采集频率不能置空,将使用默认采集频率代替(" + 默认采集频率_毫秒_s + ")\r\n"
			} else {
				_, err := strconv.ParseUint(采集频率_s, 10, 64)
				if err != nil {
					检查结果 = "采集频率(" + 采集频率_s + ")不是一个64位无符号整型,将使用默认采集频率代替(" + 默认采集频率_毫秒_s + ")\r\n"
				} else {
					使用默认设置 = false
				}
			}
			if 使用默认设置 {
				填值 = 默认采集频率_毫秒_s
				row[标题列号[采集频率_下标]] = 默认采集频率_毫秒_s
				采集频率_s = 默认采集频率_毫秒_s
			}
		}
		if 检查结果 != "" {
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[采集频率_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		置空 = false
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 采集前等待_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定采集前等待，将被软件置空\r\n"
				置空 = true
			}
		} else {
			if 采集前等待_s == "" {
				检查结果 = "采集前等待不能置空,将使用默认采集前等待_毫秒代替(" + 默认采集前等待_毫秒_s + ")\r\n"
			} else {
				整型值, err := strconv.ParseUint(采集前等待_s, 10, 16)
				if err != nil {
					检查结果 = "采集前等待(" + 采集前等待_s + ")不是一个16位无符号整型,将使用默认采集前等待_毫秒代替(" + 默认采集前等待_毫秒_s + ")\r\n"
				} else {
					if 整型值 < 最小时间 || 整型值 > 最大时间 {
						检查结果 = "采集前等待(" + 采集前等待_s + ")不在正常范围内(" + 最小时间_s + "~" + 最大时间_s + ")，将使用默认采集前等待_毫秒代替(" + 默认采集前等待_毫秒_s + ")\r\n"
					}
				}
			}
		}
		if 检查结果 != "" {
			if 置空 {
				填值 = ""
				row[标题列号[采集前等待_下标]] = ""
				采集前等待_s = ""
			} else {
				填值 = 默认采集前等待_毫秒_s
				row[标题列号[采集前等待_下标]] = 默认采集前等待_毫秒_s
				采集前等待_s = 默认采集前等待_毫秒_s
			}
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[采集前等待_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		if 记录到数据库间隔_秒_s == "" {
			检查结果 = "记录到数据库间隔_秒不能置空,将使用默认记录到数据库间隔_秒代替(" + 默认记录到数据库间隔_秒_s + ")\r\n"
		} else {
			整型值, err := strconv.ParseUint(记录到数据库间隔_秒_s, 10, 16)
			if err != nil {
				检查结果 = "记录到数据库间隔_秒(" + 记录到数据库间隔_秒_s + ")不是一个16位无符号整型,将使用默认记录到数据库间隔_秒代替(" + 默认记录到数据库间隔_秒_s + ")\r\n"
			} else {
				if 整型值 < 最小时间 || 整型值 > 最大时间 {
					检查结果 = "记录到数据库间隔_秒(" + 记录到数据库间隔_秒_s + ")不在正常范围内(" + 最小时间_s + "~" + 最大时间_s + ")，将使用默认记录到数据库间隔_秒代替(" + 默认记录到数据库间隔_秒_s + ")\r\n"
				}
			}
		}
		if 检查结果 != "" {
			填值 = 默认记录到数据库间隔_秒_s
			row[标题列号[记录到数据库间隔_秒_下标]] = 默认记录到数据库间隔_秒_s
			记录到数据库间隔_秒_s = 默认记录到数据库间隔_秒_s
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[记录到数据库间隔_秒_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		置空 = false
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 数据位_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定数据位，将被软件置空\r\n"
				置空 = true
			}
		} else {
			if 数据位_s == "" {
				检查结果 = "数据位不能置空,将使用默认数据位代替(" + 默认数据位_s + ")\r\n"
			} else {
				if 数据位_s != "5" && 数据位_s != "6" && 数据位_s != "7" && 数据位_s != "8" {
					检查结果 = "数据位(" + 数据位_s + ")不是5、6、7、8之一,将使用默认数据位代替(" + 默认数据位_s + ")\r\n"
				}
			}
		}
		if 检查结果 != "" {
			if 置空 {
				填值 = ""
				row[标题列号[数据位_下标]] = ""
				数据位_s = ""
			} else {
				填值 = 默认数据位_s
				row[标题列号[数据位_下标]] = 默认数据位_s
				数据位_s = 默认数据位_s
			}
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[数据位_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		置空 = false
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 停止位_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定停止位，将被软件置空\r\n"
				置空 = true
			}
		} else {
			if 停止位_s == "" {
				检查结果 = "停止位不能置空,将使用默认停止位代替(" + 默认停止位_s + ")\r\n"
			} else {
				if 停止位_s != "1" && 停止位_s != "2" {
					检查结果 = "停止位(" + 停止位_s + ")不是1、2之一,将使用默认停止位代替(" + 默认停止位_s + ")\r\n"
				}
			}
		}
		if 检查结果 != "" {
			if 置空 {
				填值 = ""
				row[标题列号[停止位_下标]] = ""
				停止位_s = ""
			} else {
				填值 = 默认停止位_s
				row[标题列号[停止位_下标]] = 默认停止位_s
				停止位_s = 默认停止位_s
			}
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[停止位_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		置数类型 := 0
		if 变量类型 == "IO实型" && (数据类型 == "LONG" || 数据类型 == "ULONG" || 数据类型 == "SHORT" || 数据类型 == "USHORT" || 数据类型 == "LONGBCD" || 数据类型 == "BCD") {
			if 计算值除原始值_s == "" {
				检查结果 = "变量类型(" + 变量类型 + ")数据类型(" + 数据类型 + ")需设定计算值除原始值，不能置空，将使用默认计算值除原始值(1)\r\n"
				置数类型 = 2
			} else {
				浮点数值, err := strconv.ParseFloat(计算值除原始值_s, 64)
				if err != nil {
					检查结果 = "计算值除原始值(" + 计算值除原始值_s + ")不是一个64位浮点数,将使用默认计算值除原始值(1)代替\r\n"
					置数类型 = 2
				} else {
					if 浮点数值 < 最小计算值除原始值 {
						检查结果 = "计算值除原始值(" + 计算值除原始值_s + ")小于最小计算值除原始值设定(" + 最小计算值除原始值_s + "),将使用最小计算值除原始值代替\r\n"
						置数类型 = 3
					} else {
						if 浮点数值 > 最大计算值除原始值 {
							检查结果 = "计算值除原始值(" + 计算值除原始值_s + ")大于最大计算值除原始值设定(" + 最大计算值除原始值_s + "),将使用最大计算值除原始值代替\r\n"
							置数类型 = 4
						}
					}
				}
			}
		} else {
			if 计算值除原始值_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")数据类型(" + 数据类型 + ")无需设定计算值除原始值，将被软件置空\r\n"
				置数类型 = 1
			}
		}
		if 检查结果 != "" {
			switch 置数类型 {
			case 1:
				填值 = ""
				row[标题列号[计算值除原始值_下标]] = ""
				计算值除原始值_s = ""
			case 2:
				填值 = "1"
				row[标题列号[计算值除原始值_下标]] = "1"
				计算值除原始值_s = "1"
			case 3:
				填值 = 最小计算值除原始值_s
				row[标题列号[计算值除原始值_下标]] = 最小计算值除原始值_s
				计算值除原始值_s = 最小计算值除原始值_s
			case 4:
				填值 = 最大计算值除原始值_s
				row[标题列号[计算值除原始值_下标]] = 最大计算值除原始值_s
				计算值除原始值_s = 最大计算值除原始值_s
			}
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[计算值除原始值_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		置空 = false
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 奇偶校验_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定奇偶校验，将被软件置空\r\n"
				置空 = true
			}
		} else {
			if 奇偶校验_s == "" {
				检查结果 = "奇偶校验不能置空,将使用默认奇偶校验代替(" + 默认奇偶校验 + ")\r\n"
			} else {
				if 奇偶校验_s != "偶校验" &&
					奇偶校验_s != "奇校验" &&
					奇偶校验_s != "无校验" &&
					奇偶校验_s != "偶" &&
					奇偶校验_s != "奇" &&
					奇偶校验_s != "无" &&
					奇偶校验_s != "EVEN" &&
					奇偶校验_s != "ODD" &&
					奇偶校验_s != "NONE" &&
					奇偶校验_s != "E" &&
					奇偶校验_s != "O" &&
					奇偶校验_s != "N" {
					检查结果 = "奇偶校验(" + 奇偶校验_s + ")不被软件认可，将被软件置默认奇偶校验(" + 默认奇偶校验 + ")\r\n"
				}
			}
		}
		if 检查结果 != "" {
			if 置空 {
				填值 = ""
				row[标题列号[奇偶校验_下标]] = ""
				奇偶校验_s = ""
			} else {
				填值 = 默认奇偶校验
				row[标题列号[奇偶校验_下标]] = 默认奇偶校验
				奇偶校验_s = 默认奇偶校验
			}
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[奇偶校验_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		填值 = ""
		检查结果 = ""
		if 变量类型 == "" {
			检查结果 = "变量类型不可置空！\r\n"
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			错误信息 = 错误信息 + 检查结果
		} else {
			if 变量类型 != "IO离散" && 变量类型 != "IO字符串" && 变量类型 != "IO整型" && 变量类型 != "IO实型" && 变量类型 != "内存离散" && 变量类型 != "内存字符串" && 变量类型 != "内存整型" && 变量类型 != "内存实型" {
				检查结果 = "变量类型(" + 变量类型 + ")不被软件认可！将被置空\r\n"
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[变量类型_下标]+1), 填值)
				遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
				错误信息 = 错误信息 + 检查结果
				row[标题列号[变量类型_下标]] = ""
				变量类型 = ""
			}
		}
		填值 = ""
		检查结果 = ""
		置空 = true
		if 变量类型 == "IO离散" || 变量类型 == "IO字符串" || 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 数据类型 != "" {
				检查结果 = "变量类型(" + 变量类型 + ")数据类型(" + 数据类型 + ")无需设定！将被置空\r\n"
				警告 = true
			}
		} else {
			switch 变量类型 {
			case "IO实型":
				if 数据类型 == "" {
					检查结果 = "变量类型(" + 变量类型 + ")数据类型不能置空\r\n"
					置空 = false
				} else {
					if 数据类型 != "FLOAT" && 数据类型 != "LONG" && 数据类型 != "ULONG" && 数据类型 != "LONGBCD" && 数据类型 != "BCD" && 数据类型 != "SHORT" && 数据类型 != "USHORT" {
						检查结果 = "变量类型(" + 变量类型 + ")数据类型(" + 数据类型 + ")不被软件认可！将被置空\r\n"
					}
				}
			case "IO整型":
				if 数据类型 == "" {
					检查结果 = "变量类型(" + 变量类型 + ")数据类型不能置空\r\n"
					置空 = false
				} else {
					if 数据类型 != "LONG" && 数据类型 != "ULONG" && 数据类型 != "LONGBCD" && 数据类型 != "BCD" && 数据类型 != "SHORT" && 数据类型 != "USHORT" {
						检查结果 = "变量类型(" + 变量类型 + ")数据类型(" + 数据类型 + ")不被软件认可！将被置空\r\n"
					}
				}
			default:
				if 数据类型 != "FLOAT" && 数据类型 != "LONG" && 数据类型 != "ULONG" && 数据类型 != "LONGBCD" && 数据类型 != "BCD" && 数据类型 != "SHORT" && 数据类型 != "USHORT" && 数据类型 != "BIT" && 数据类型 != "STRING" && 数据类型 != "" {
					检查结果 = "数据类型(" + 数据类型 + ")不被软件认可！将被置空\r\n"
				}
			} //switch 变量类型{
		}
		if 检查结果 != "" {
			if 置空 {
				row[标题列号[数据类型_下标]] = ""
				数据类型 = ""
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[数据类型_下标]+1), 填值)
			}
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			if 警告 {
				变量警告信息 = 变量警告信息 + 检查结果
			} else {
				错误信息 = 错误信息 + 检查结果
			}
		}
		填值 = ""
		检查结果 = ""
		置空 = true
		警告 = false
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 串口号_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定串口号，将被软件置空\r\n"
				警告 = true
			}
		} else {
			if 串口号_s == "" {
				检查结果 = "串口号不能置空！\r\n"
				置空 = false
			} else {
				整型值, err := strconv.ParseUint(串口号_s, 10, 8)
				if err != nil {
					检查结果 = "串口号(" + 串口号_s + ")不是一个8位无符号整型,将被置空！\r\n"
				} else {
					if 整型值 < 最小串口号 {
						检查结果 = "串口号(" + 串口号_s + ")小于最小串口号(" + 最小串口号_s + ")，将被置空！\r\n"
					} else {
						if 整型值 > 最大串口号 {
							检查结果 = "串口号(" + 串口号_s + ")大于最大串口号(" + 最大串口号_s + ")，将被置空！\r\n"
						}
					}
				}
			}
		}
		if 检查结果 != "" {
			if 置空 {
				填值 = ""
				row[标题列号[串口号_下标]] = ""
				串口号_s = ""
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[串口号_下标]+1), 填值)
			}
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			if 警告 {
				变量警告信息 = 变量警告信息 + 检查结果
			} else {
				错误信息 = 错误信息 + 检查结果
			}
		}
		填值 = ""
		检查结果 = ""
		置空 = true
		警告 = false
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 读功能码 != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定读功能码，将被软件置空\r\n"
				警告 = true
			}
		} else {
			if 读功能码 == "" {
				检查结果 = "读功能码不能置空！\r\n"
				置空 = false
			} else {
				if 读功能码 != "1" && 读功能码 != "2" && 读功能码 != "3" && 读功能码 != "4" {
					检查结果 = "读功能码(" + 读功能码 + ")不被软件认可！将被软件置空\r\n"
				} else { //if 读功能码!="1"&&读功能码!="2"&&读功能码!="3"&&读功能码!="4"{
					switch 变量类型 {
					case "IO字符串", "IO整型", "IO实型":
						if 读功能码 != "3" && 读功能码 != "4" {
							检查结果 = "读功能码(" + 读功能码 + ")不匹配变量类型(" + 变量类型 + ")！将被软件置空\r\n"
						}
					} //switch 变量类型{
				} //else{//if 读功能码!="1"&&读功能码!="2"&&读功能码!="3"&&读功能码!="4"{
			}
		}
		if 检查结果 != "" {
			if 置空 {
				填值 = ""
				row[标题列号[读功能码_下标]] = ""
				读功能码 = ""
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[读功能码_下标]+1), 填值)
			}
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			if 警告 {
				变量警告信息 = 变量警告信息 + 检查结果
			} else {
				错误信息 = 错误信息 + 检查结果
			}
		}
		填值 = ""
		检查结果 = ""
		置空 = true
		警告 = false
		if 变量类型 == "内存离散" || 变量类型 == "内存字符串" || 变量类型 == "内存整型" || 变量类型 == "内存实型" {
			if 打包长度_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定打包长度，将被软件置空\r\n"
				警告 = true
			}
		} else {
			if 只写有效 {
				if 打包长度_s != "" {
					检查结果 = "变量类型(" + 变量类型 + ")读功能码(" + 读功能码 + ")读写属性(" + 读写属性 + ")无需设定打包长度，将被软件置空\r\n"
					警告 = true
				}
			} else {
				if 打包长度_s == "" {
					检查结果 = "打包长度不能置空！\r\n"
					置空 = false
				} else {
					打包长度, err := strconv.ParseUint(打包长度_s, 10, 16)
					if err != nil {
						检查结果 = "打包长度(" + 打包长度_s + ")不是一个16位无符号整型！,将被置空！\r\n"
					} else {
						if 变量类型 != "IO离散" || IO离散点 {
							if 打包长度 < 1 || 打包长度 > 最大字打包长度 {
								检查结果 = "打包长度(" + 打包长度_s + ")不在正常范围内(" + "1" + "~" + 最大字打包长度_s + ")内,将被软件置空\r\n"
							}
						}
						if 变量类型 == "IO离散" && !IO离散点 {
							if 打包长度 < 1 || 打包长度 > 最大位打包长度 {
								检查结果 = "打包长度(" + 打包长度_s + ")不在正常范围内(" + "1" + "~" + 最大位打包长度_s + ")内,将被软件置空\r\n"
							}
						}
					}
				}
			}
		}
		if 检查结果 != "" {
			if 置空 {
				填值 = ""
				row[标题列号[打包长度_下标]] = ""
				打包长度_s = ""
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[打包长度_下标]+1), 填值)
			}
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			if 警告 {
				变量警告信息 = 变量警告信息 + 检查结果
			} else {
				错误信息 = 错误信息 + 检查结果
			}
		}
		置空 = false
		检查结果 = ""
		if 变量类型 != "内存实型" && 变量类型 != "IO实型" {
			if 小数位数_s != "" {
				检查结果 = "变量类型(" + 变量类型 + ")无需设定小数位数，将被软件置空\r\n"
				置空 = true
			}
		} else {
			if 小数位数_s == "" {
				检查结果 = "小数位数不能置空，将使用默认小数位(" + 默认小数位_s + ")\r\n"
			} else {
				小数位数, err6 := strconv.Atoi(小数位数_s)
				if err6 != nil {
					检查结果 = err6.Error() + "\r\n"
				} else {
					if 小数位数 < 0 {
						检查结果 = "小数位数< 0 ，将使用默认小数位(" + 默认小数位_s + ")\r\n"
					} else {
						if 小数位数 > 最大小数位 {
							检查结果 = "小数位数(" + 小数位数_s + ") > 最大小数位(" + 最大小数位_s + ")，将使用默认小数位(" + 默认小数位_s + ")\r\n"
						}
					}
				}
			}
		}
		if 检查结果 != "" {
			if 置空 {
				填值 = ""
				row[标题列号[小数位数_下标]] = ""
				小数位数_s = ""
			} else {
				填值 = 默认小数位_s
				row[标题列号[小数位数_下标]] = 默认小数位_s
				小数位数_s = 默认小数位_s
			}
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[小数位数_下标]+1), 填值)
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "行号(" + 行号2 + ")的变量名(" + 变量名 + ")" + 检查结果
			变量警告信息 = 变量警告信息 + 检查结果
		}
		if 变量警告信息 != "" {
			表格有改动 = true
			if !strings.Contains(row[标题列号[警告信息_下标]], 变量警告信息) {
				变量警告信息 = 变量警告信息 + row[标题列号[警告信息_下标]]
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[警告信息_下标]+1), 变量警告信息)
			}
		}
		if 错误信息 != "" {
			表格有改动 = true
			//if !strings.Contains(row[标题列号[错误信息_下标]], 错误信息) {
			f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[错误信息_下标]+1), 错误信息)
			//}
		} else {
			没有错误信息的行[i] = true
			没有错误信息行数++
			if row[标题列号[错误信息_下标]] != "" {
				f.SetCellValue(工作表名, 求单元格代号(i+1, 标题列号[错误信息_下标]+1), "")
				row[标题列号[错误信息_下标]] = ""
				表格有改动 = true
			}
		}
	} //for i, row := range rows {
	if 表格有改动 {
		f.NewSheet(工作表名)
		获取框线值(true, f)
		画框线(工作表名, 项目变量信息表格行数, 标题列数, f)
		//冻结前2列//冻结前1行//左上角单元格h2
		var 格式 excelize.Panes
		格式.Freeze = true
		格式.ActivePane = "\"bottomRight\",\"panes\":[{\"pane\":\"topLeft\"},{\"pane\":\"topRight\"},{\"pane\":\"bottomLeft\"},{\"active_cell\":\"d2\",\"sqref\":\"d2\",\"pane\":\"bottomRight\"}]"
		格式.TopLeftCell = "d2"
		格式.XSplit = 3
		格式.YSplit = 1
		f.SetPanes(工作表名, &格式)
		//	f.UpdateLinkedValue() //更新表格中公式等引用
		f.SetActiveSheet(index)
		if err := f.Save(); err != nil {
			遍历项目变量信息表结果 = 遍历项目变量信息表结果 + err.Error() + "\r\n"
		}
		延时1 := 项目变量信息表格行数 / 项目变量信息表格保存多少行需要1秒时间间隔
		if 延时1 <= 0 {
			延时1 = 1
		}
		time.Sleep(time.Second * time.Duration(延时1))
		遍历项目变量信息表保存次++
	} //if 表格有改动 {
	遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "项目建变量个数(" + strconv.Itoa(项目变量信息表格行数-1) + ")\r\n"
	遍历项目变量信息表结果 = 遍历项目变量信息表结果 + "没有错误信息行数(" + strconv.Itoa(没有错误信息行数) + ")\r\n"
	if 没有错误信息行数 <= 0 {
		遍历项目变量信息表中 = false
		return true
	}
	串口组 = make(map[string]bool)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if _, ok := 没有错误信息的行[i]; !ok {
			continue
		}
		串口号 := row[标题列号[串口号_下标]]
		if 串口号 != "" {
			串口组[串口号] = true
		}
	}
	项目变量信息表组.Rows = make([]Row, 没有错误信息行数)
	变量组.KVTags = make([]KVTag, 没有错误信息行数)
	变量名所在行 = make(map[string]int)
	变量名所在表行 = make(map[string]int)
	序号 := 变量ID初始值
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if _, ok := 没有错误信息的行[i]; !ok {
			continue
		}
		序号减变量ID初始值 := 序号 - 变量ID初始值
		项目变量信息表组.Rows[序号减变量ID初始值].X序号 = 序号
		变量组.KVTags[序号减变量ID初始值].NVarID = 序号
		for j, colCell := range row {
			switch j {
			case 标题列号[变量名_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].B变量名称 = colCell
				变量组.KVTags[序号减变量ID初始值].StrVarName = colCell
				变量名所在行[colCell] = 序号减变量ID初始值
				变量名所在表行[colCell] = i
			case 标题列号[数据类型_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].S数据类型 = colCell
			case 标题列号[变量类型_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].B变量类型 = colCell
				switch colCell {
				case "内存整型":
					变量组.KVTags[序号减变量ID初始值].NVarType = "LONG"
				case "内存实型":
					变量组.KVTags[序号减变量ID初始值].NVarType = "FLOAT"
				case "IO离散", "内存离散":
					变量组.KVTags[序号减变量ID初始值].NVarType = "BIT"
				case "IO字符串", "内存字符串":
					变量组.KVTags[序号减变量ID初始值].NVarType = "STRING"
				case "IO整型":
					switch row[标题列号[数据类型_下标]] {
					case "LONG":
						变量组.KVTags[序号减变量ID初始值].NVarType = "LONG"
					case "ULONG":
						变量组.KVTags[序号减变量ID初始值].NVarType = "ULONG"
					case "LONGBCD":
						变量组.KVTags[序号减变量ID初始值].NVarType = "LONGBCD"
					case "USHORT":
						变量组.KVTags[序号减变量ID初始值].NVarType = "USHORT"
					case "SHORT":
						变量组.KVTags[序号减变量ID初始值].NVarType = "SHORT"
					case "BCD":
						变量组.KVTags[序号减变量ID初始值].NVarType = "BCD"
					} //switch row[标题列号[数据类型_下标]] {
				case "IO实型":
					switch row[标题列号[数据类型_下标]] {
					case "FLOAT":
						变量组.KVTags[序号减变量ID初始值].NVarType = "FLOAT_FLOAT"
					case "LONG":
						变量组.KVTags[序号减变量ID初始值].NVarType = "FLOAT_LONG"
					case "ULONG":
						变量组.KVTags[序号减变量ID初始值].NVarType = "FLOAT_ULONG"
					case "LONGBCD":
						变量组.KVTags[序号减变量ID初始值].NVarType = "FLOAT_LONGBCD"
					case "USHORT":
						变量组.KVTags[序号减变量ID初始值].NVarType = "FLOAT_USHORT"
					case "SHORT":
						变量组.KVTags[序号减变量ID初始值].NVarType = "FLOAT_SHORT"
					case "BCD":
						变量组.KVTags[序号减变量ID初始值].NVarType = "FLOAT_BCD"
					} //switch row[标题列号[数据类型_下标]] {
				} //switch colCell {
			case 标题列号[读写属性_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].D读写属性 = colCell
			case 标题列号[采集前等待_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					整型值, err := strconv.ParseUint(colCell, 10, 16)
					if err == nil {
						项目变量信息表组.Rows[序号减变量ID初始值].C采集前等待毫秒 = time.Duration(整型值)
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].C采集前等待毫秒 = 默认采集前等待_毫秒
					}
				}
			case 标题列号[连续通讯失败多少次则认为通讯异常_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					整型值, err := strconv.ParseUint(colCell, 10, 16)
					if err == nil {
						项目变量信息表组.Rows[序号减变量ID初始值].L连续通讯失败多少次则认为通讯异常 = uint32(整型值)
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].L连续通讯失败多少次则认为通讯异常 = 默认连续通讯失败多少次则认为通讯异常
					}
				}
			case 标题列号[读字节序_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].D读字节序 = colCell
			case 标题列号[写字节序_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].X写字节序 = colCell
			case 标题列号[初始值_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].C初始值 = colCell
				项目变量信息表组.Rows[序号减变量ID初始值].D当前值 = colCell
				变量组.KVTags[序号减变量ID初始值].VarValue = colCell
			case 标题列号[采集频率_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					整型值, err := strconv.ParseUint(colCell, 10, 64)
					if err == nil {
						项目变量信息表组.Rows[序号减变量ID初始值].C采集频率毫秒 = int64(整型值)
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].C采集频率毫秒 = 默认采集频率_毫秒
					}
				}
			case 标题列号[通讯异常值_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].T通讯异常值 = colCell
			case 标题列号[是否保存值_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].S是否保存值 = colCell
			case 标题列号[计算值除原始值_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					整型值, err := strconv.ParseFloat(colCell, 64)
					if err == nil {
						项目变量信息表组.Rows[序号减变量ID初始值].J计算值除原始值 = 整型值
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].J计算值除原始值 = 1
					}
				}
			case 标题列号[小数位数_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					整型值, err := strconv.ParseUint(colCell, 10, 8)
					if err == nil {
						项目变量信息表组.Rows[序号减变量ID初始值].X小数位数 = int(整型值)
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].X小数位数 = 默认小数位数
					}
				}
			case 标题列号[串口号_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].C串口号 = colCell
			case 标题列号[波特率_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					整型值, err := strconv.ParseUint(colCell, 10, 32)
					if err == nil {
						项目变量信息表组.Rows[序号减变量ID初始值].B波特率 = int(整型值)
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].B波特率 = 默认波特率
					}
				}
			case 标题列号[奇偶校验_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					switch colCell {
					case "偶校验", "偶", "EVEN", "E":
						项目变量信息表组.Rows[序号减变量ID初始值].J奇偶校验 = "E"
					case "奇校验", "奇", "ODD", "O":
						项目变量信息表组.Rows[序号减变量ID初始值].J奇偶校验 = "O"
					case "无校验", "无", "NONE", "N":
						项目变量信息表组.Rows[序号减变量ID初始值].J奇偶校验 = "N"
					default:
						项目变量信息表组.Rows[序号减变量ID初始值].J奇偶校验 = "N"
					} //switch colCell{
				}
			case 标题列号[数据位_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					整型值, err := strconv.ParseUint(colCell, 10, 8)
					if err == nil {
						项目变量信息表组.Rows[序号减变量ID初始值].S数据位 = int(整型值)
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].S数据位 = 默认数据位
					}
				}
			case 标题列号[停止位_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					整型值, err := strconv.ParseUint(colCell, 10, 8)
					if err == nil {
						项目变量信息表组.Rows[序号减变量ID初始值].T停止位 = int(整型值)
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].T停止位 = 默认停止位
					}
				}
			case 标题列号[通讯超时_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					整型值, err := strconv.ParseUint(colCell, 10, 16)
					if err == nil {
						项目变量信息表组.Rows[序号减变量ID初始值].T通讯超时毫秒 = time.Duration(整型值)
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].T通讯超时毫秒 = 默认通讯超时_毫秒
					}
				}
			case 标题列号[设备地址_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					整型值, err := strconv.ParseUint(colCell, 10, 8)
					if err == nil {
						项目变量信息表组.Rows[序号减变量ID初始值].S设备地址 = byte(整型值)
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].S设备地址 = 默认设备地址
					}
				}
			case 标题列号[读功能码_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].D读功能码 = colCell
			case 标题列号[寄存器地址_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					if strings.Contains(colCell, ".") {
						前值, 后值, 分离结果 := 小数点前后值分离(colCell)
						if 分离结果 == "ok" {
							项目变量信息表组.Rows[序号减变量ID初始值].J寄存器地址_s = colCell
							项目变量信息表组.Rows[序号减变量ID初始值].J寄存器地址 = uint16(前值)
							项目变量信息表组.Rows[序号减变量ID初始值].X小数点后值 = int(后值)
						} else {
							项目变量信息表组.Rows[序号减变量ID初始值].C采集频率毫秒 = 0
						}
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].X小数点后值 = -1
						整型值, err := strconv.ParseUint(colCell, 10, 16)
						if err == nil {
							项目变量信息表组.Rows[序号减变量ID初始值].J寄存器地址_s = colCell
							项目变量信息表组.Rows[序号减变量ID初始值].J寄存器地址 = uint16(整型值)
						} else {
							项目变量信息表组.Rows[序号减变量ID初始值].C采集频率毫秒 = 0
						}
					}
				}
			case 标题列号[单位_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].D单位 = colCell
			case 标题列号[多少变化率才记录到数据库_下标]:
				整型值, err := strconv.ParseFloat(colCell, 64)
				if err == nil {
					项目变量信息表组.Rows[序号减变量ID初始值].D多少变化率才记录到数据库 = 整型值
				} else {
					项目变量信息表组.Rows[序号减变量ID初始值].D多少变化率才记录到数据库 = 默认多少变化率才记录到数据库
				}
			case 标题列号[打包长度_下标]:
				if !strings.Contains(row[标题列号[变量类型_下标]], "内存") {
					整型值, err := strconv.ParseUint(colCell, 10, 16)
					if err == nil {
						项目变量信息表组.Rows[序号减变量ID初始值].D打包长度 = uint16(整型值)
					} else {
						项目变量信息表组.Rows[序号减变量ID初始值].D打包长度 = 默认打包长度
					}
				}
			case 标题列号[允许通讯异常后只读变量_下标]:
				项目变量信息表组.Rows[序号减变量ID初始值].Y允许通讯异常后只读变量 = colCell
			case 标题列号[记录到数据库间隔_秒_下标]:
				整型值, err := strconv.ParseUint(colCell, 10, 16)
				if err == nil {
					项目变量信息表组.Rows[序号减变量ID初始值].T记录到数据库间隔_秒 = int64(整型值)
				} else {
					项目变量信息表组.Rows[序号减变量ID初始值].T记录到数据库间隔_秒 = 默认记录到数据库间隔_秒
				}
			} //switch 列号 {
		} //for j, colCell := range row {
		序号++
	} //for i, row := range rows {
	for i := range 项目变量信息表组.Rows {
		row := &项目变量信息表组.Rows[i]
		if strings.Contains(row.B变量类型, "内存") {
			continue
		}
		if row.C采集频率毫秒 < 最小采集频率毫秒 && row.C采集频率毫秒 > 0 {
			最小采集频率毫秒 = row.C采集频率毫秒
		}
		if row.C采集频率毫秒 > 最大采集频率毫秒 && row.C采集频率毫秒 > 0 {
			最大采集频率毫秒 = row.C采集频率毫秒
		}
	}
	if 最大采集频率毫秒 == 0 {
		最大采集频率毫秒 = 默认最大采集频率毫秒
	}
	if 最小采集频率毫秒 == 0 || 最小采集频率毫秒 == math.MaxInt64 {
		最小采集频率毫秒 = 默认最小采集频率毫秒
	}
	各项目最小采集频率毫秒[全部项目代号] = 最小采集频率毫秒
	//fmt.Println("最小采集频率毫秒：", 最小采集频率毫秒)
	for i := range 项目变量信息表组.Rows {
		项目代号2 := strings.Split(项目变量信息表组.Rows[i].B变量名称, "_")
		if len(项目代号2) > 0 {
			项目代号 := 项目代号2[0]
			项目变量信息表组.Rows[i].X项目代号 = 项目代号
			项目代号的表行们[项目代号] = append(项目代号的表行们[项目代号], i)
			if strings.Contains(项目变量信息表组.Rows[i].B变量类型, "内存") {
				项目代号的内存变量表行们[项目代号] = append(项目代号的表行们[项目代号], i)
			}
		}
	}
	fmt.Println("项目代号们及变量个数：")
	for 项目代号, 表行们 := range 项目代号的表行们 {
		变量个数 := len(表行们)
		fmt.Println(项目代号 + fmt.Sprintf(":%d", 变量个数))
	}
	for 项目代号, 表行们 := range 项目代号的表行们 {
		var 最小采集频率毫秒2 int64 = math.MaxInt64
		for _, 表行 := range 表行们 {
			row := &项目变量信息表组.Rows[表行]
			if strings.Contains(row.B变量类型, "内存") {
				continue
			}
			if row.C采集频率毫秒 < 最小采集频率毫秒2 && row.C采集频率毫秒 > 0 {
				最小采集频率毫秒2 = row.C采集频率毫秒
			}
		}
		if 最小采集频率毫秒2 == 0 || 最小采集频率毫秒2 == math.MaxInt64 {
			最小采集频率毫秒2 = 默认最小采集频率毫秒
		}
		各项目最小采集频率毫秒[项目代号] = 最小采集频率毫秒2
	} //for 项目代号, 表行们 := range 项目代号的表行们 {
	type 可一起打包采集的表行们信息结构体 struct {
		表行们         []int
		最大打包长度      uint16
		占用最大字数的变量字数 uint16
	} //type 可一起打包采集的表行们信息结构体 struct {
	var 可一起打包采集的表行们 = make(map[string]*可一起打包采集的表行们信息结构体, 0)
	for i := range 项目变量信息表组.Rows {
		z := &项目变量信息表组.Rows[i]
		if strings.Contains(z.B变量类型, "内存") {
			continue
		}
		if z.D读写属性 != "只读" && z.D读写属性 != "读写" {
			continue
		}
		if z.C采集频率毫秒 == 0 {
			continue
		}
		头 := 获得可一起打包采集的表行们的头(z)
		读字个数 := 获得变量读字个数(z)
		if _, ok := 可一起打包采集的表行们[头]; !ok {
			可一起打包采集的表行们[头] = &可一起打包采集的表行们信息结构体{
				表行们:         []int{},
				最大打包长度:      z.D打包长度,
				占用最大字数的变量字数: 读字个数,
			}
		} else { //if _, ok := 可一起打包采集的表行们[头]; !ok {
			表行们 := 可一起打包采集的表行们[头]
			if 表行们.最大打包长度 < z.D打包长度 {
				表行们.最大打包长度 = z.D打包长度
			}
			if 表行们.占用最大字数的变量字数 < 读字个数 {
				表行们.占用最大字数的变量字数 = 读字个数
			}
		} //} else {//if _, ok := 可一起打包采集的表行们[头]; !ok {
		可一起打包采集的表行们[头].表行们 = append(可一起打包采集的表行们[头].表行们, i)
	} //for i := range 项目变量信息表组.Rows {
	for _, 表行们 := range 可一起打包采集的表行们 {
		sort.Slice(表行们.表行们, func(i, j int) bool {
			if 项目变量信息表组.Rows[表行们.表行们[i]].J寄存器地址 < 项目变量信息表组.Rows[表行们.表行们[j]].J寄存器地址 {
				return true
			}
			if 项目变量信息表组.Rows[表行们.表行们[i]].J寄存器地址 > 项目变量信息表组.Rows[表行们.表行们[j]].J寄存器地址 {
				return false
			}
			if 项目变量信息表组.Rows[表行们.表行们[i]].X小数点后值 <
				项目变量信息表组.Rows[表行们.表行们[j]].X小数点后值 {
				return true
			}
			if 项目变量信息表组.Rows[表行们.表行们[i]].X小数点后值 >
				项目变量信息表组.Rows[表行们.表行们[j]].X小数点后值 {
				return false
			}
			return 项目变量信息表组.Rows[表行们.表行们[i]].B变量名称 < 项目变量信息表组.Rows[表行们.表行们[j]].B变量名称
		})
	} //for _, 表行们 := range 可一起打包采集的表行们 {
	fmt.Println("可一起打包采集的表行们的头及变量个数：")
	for 头, 表行们 := range 可一起打包采集的表行们 {
		变量个数 := len(表行们.表行们)
		fmt.Println(头 + fmt.Sprintf(":%d", 变量个数))
		//fmt.Println(表行们.表行们)
		if 表行们.占用最大字数的变量字数 == 0 {
			if 表行们.最大打包长度 < 1 {
				表行们.最大打包长度 = 1
			} else {
				if 表行们.最大打包长度 > 最大位打包长度 {
					表行们.最大打包长度 = 最大位打包长度
				}
			}
		} else { //if 表行们.占用最大字数的变量字数 == 0 {
			if 表行们.最大打包长度 < 表行们.占用最大字数的变量字数 {
				表行们.最大打包长度 = 表行们.占用最大字数的变量字数
			}
			if 表行们.最大打包长度 > 最大字打包长度 {
				表行们.最大打包长度 = 最大字打包长度
			}
		} //} else {//if 表行们.占用最大字数的变量字数 == 0 {
		fmt.Println("最大打包长度:" + fmt.Sprintf("%d", 表行们.最大打包长度))
		fmt.Println("占用最大字数的变量字数:" + fmt.Sprintf("%d", 表行们.占用最大字数的变量字数))
	}
	type 可一起打包采集的表行们信息结构体2 struct {
		表行们 [][]int
	} //type 可一起打包采集的表行们信息结构体 struct {
	var 可一起打包采集的表行们2 = make(map[string]*可一起打包采集的表行们信息结构体2, 0)
	for 头, 表行们 := range 可一起打包采集的表行们 {
		if _, ok := 可一起打包采集的表行们2[头]; !ok {
			可一起打包采集的表行们2[头] = &可一起打包采集的表行们信息结构体2{}
		}
		第几批次采集 := 0
		批次采集的表行们 := make([][]int, 0)
		for _, 表行 := range 表行们.表行们 {
			for len(批次采集的表行们) <= 第几批次采集 {
				批次采集的表行们 = append(批次采集的表行们, []int{})
			}
			批次采集的表行们[第几批次采集] = append(批次采集的表行们[第几批次采集], 表行)
			z := &项目变量信息表组.Rows[表行]
			变量字数 := 获得变量读字个数(z)
			var 变量最大寄存器地址 uint16
			if 变量字数 == 0 {
				变量最大寄存器地址 = z.J寄存器地址
			} else {
				变量最大寄存器地址 = z.J寄存器地址 + 获得变量读字个数(z) - 1
			}
			y := &项目变量信息表组.Rows[批次采集的表行们[第几批次采集][0]]
			if 变量最大寄存器地址-y.J寄存器地址+1 > 表行们.最大打包长度 {
				批次采集的表行们[第几批次采集] = removeElement(批次采集的表行们[第几批次采集], 表行)
				第几批次采集++
				for len(批次采集的表行们) <= 第几批次采集 {
					批次采集的表行们 = append(批次采集的表行们, []int{})
				}
				批次采集的表行们[第几批次采集] = append(批次采集的表行们[第几批次采集], 表行)
			}
		}
		可一起打包采集的表行们2[头].表行们 = 批次采集的表行们
	} //for 头, 表行们 := range 可一起打包采集的表行们 {
	for 头, 表行们 := range 可一起打包采集的表行们2 {
		采集次数 := len(表行们.表行们)
		fmt.Println(头 + fmt.Sprintf("_采集次数:%d", 采集次数))
		for 第几批次采集, 表行们 := range 表行们.表行们 {
			fmt.Printf("第几批次采集:%d,变量个数：%d\r\n", 第几批次采集+1, len(表行们))
			//fmt.Println(表行们)
			长度 := len(表行们)
			if 长度 < 1 {
				continue
			}
			首寄存器地址 := 项目变量信息表组.Rows[表行们[0]].J寄存器地址
			z := &项目变量信息表组.Rows[表行们[长度-1]]
			读字个数 := 获得变量读字个数(z)
			var 最后一个变量最大寄存器地址 uint16
			if 读字个数 == 0 {
				最后一个变量最大寄存器地址 = z.J寄存器地址
			} else {
				最后一个变量最大寄存器地址 = z.J寄存器地址 + 获得变量读字个数(z) - 1
			}
			打包长度 := 最后一个变量最大寄存器地址 - 首寄存器地址 + 1
			for _, 表行 := range 表行们 {
				z := &项目变量信息表组.Rows[表行]
				z.P批量读开始寄存器地址 = 首寄存器地址
				z.D打包长度 = 打包长度
				z.P批量读伙伴 = 表行们
			}
		} //for 第几批次采集, 表行们 := range 表行们.表行们 {
	} //for 头, 表行们 := range 可一起打包采集的表行们2 {
	遍历项目变量信息表中 = false
	完成第一次遍历项目变量信息表 = true
	return true
} //func 遍历项目变量信息表2() {
func removeElement(slice []int, value int) []int {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice // 如果没有找到元素，则返回原切片
}

var 各项目最小采集频率毫秒 = make(map[string]int64, 0)
var 项目代号的表行们 = make(map[string][]int, 0)
var 项目代号的内存变量表行们 = make(map[string][]int, 0)
var 最小采集频率毫秒 int64 = math.MaxInt64
var 最大采集频率毫秒 int64 = 0
var 第几次采集 uint64
var 允许调试 bool

func 获得变量读字个数(z *Row) uint16 {
	if strings.Contains(z.B变量类型, "内存") {
		return 0
	}
	if z.D读功能码 != "3" && z.D读功能码 != "4" {
		return 0
	}
	var 读字个数 uint16 = 1
	switch z.B变量类型 {
	case "IO整型":
		switch z.S数据类型 {
		case "LONGBCD":
			读字个数 = 2
		case "LONG":
			读字个数 = 2
		case "ULONG":
			读字个数 = 2
		} //switch z.S数据类型 {
	case "IO实型":
		switch z.S数据类型 {
		case "FLOAT":
			读字个数 = 2
		case "LONGBCD":
			读字个数 = 2
		case "LONG":
			读字个数 = 2
		case "ULONG":
			读字个数 = 2
		} //switch z.S数据类型 {
	case "IO字符串":
		if z.X小数点后值 <= 0 {
			return 0
		}
		if z.X小数点后值 > 最大字符串字节长度 {
			return 0
		}
		字符占字数 := z.X小数点后值 / 2
		if (z.X小数点后值 % 2) > 0 {
			字符占字数++
		}
		读字个数 = uint16(字符占字数)
	} //switch z.B变量类型 {
	return 读字个数
} //func 获得变量读字个数(z *Row) uint16 {
func 获得可一起打包采集的表行们的头(row *Row) string {
	头 := row.C串口号 + "_" + fmt.Sprintf("%d", row.B波特率) + "_" + row.J奇偶校验 + "_" + fmt.Sprintf("%d", row.S数据位) + "_" +
		fmt.Sprintf("%d", row.T停止位) + "_" + fmt.Sprintf("%d", row.S设备地址) + "_" + row.D读功能码
	return 头
} //func 获得可一起打包采集的表行们的头(row *Row) string {
func 设备写操作(要写的变量名 string, 要写的值 interface{}) {
	// if 要写的变量名 != "" {
	// 	return
	// }
	读功能码 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].D读功能码
	通讯超时毫秒 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].T通讯超时毫秒
	停止位 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].T停止位
	奇偶校验 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].J奇偶校验
	波特率 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].B波特率
	串口号 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].C串口号
	数据位 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].S数据位
	p := modbus.NewRTUClientProvider()
	p.Address = "\\\\.\\" + "com" + 串口号 //必须要要加"\\\\.\\"，否则com9以上会报错："无法打开设备"
	p.BaudRate = 波特率
	p.DataBits = 数据位
	p.Parity = 奇偶校验
	p.StopBits = 停止位
	p.Timeout = 通讯超时毫秒 * time.Millisecond //毫秒
	client := modbus.NewClient(p)
	if 允许调试 {
		fmt.Println(要写的变量名)
		client.LogMode(true)
	} else {
		client.LogMode(false)
	}
	err := client.Connect()
	//atomic.AddUint64(&第几次写设备操作, 1)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer client.Close()
	z := &项目变量信息表组.Rows[变量名所在行[要写的变量名]]
	switch 读功能码 {
	case "1":
		读功能码1写(client, z, 要写的值)
	case "2":
		fmt.Println("读功能码2的设备只读_" + 要写的变量名)
	case "3":
		读功能码3写(client, z, 要写的值)
	case "4":
		fmt.Println("读功能码4的设备只读_" + 要写的变量名)
	default:
		fmt.Println("无效的读功能码" + 读功能码)
		return
	}
} //func 设备写操作(要写的变量名 string,要写的值 interface{}){
var 第几次写设备操作 uint64

func 采集任务(串口号1 string) {
	var w sync.WaitGroup
	采集的行 := make(map[int]bool)
	for i := range 项目变量信息表组.Rows {
		row := &项目变量信息表组.Rows[i]
		if row.C采集频率毫秒 == 0 {
			continue
		}
		if strings.Contains(row.B变量类型, "内存") {
			continue
		}
		if row.D读写属性 != "只读" && row.D读写属性 != "读写" {
			continue
		}
		if row.C串口号 == 串口号1 {
			采集的行[i] = true
		}
	}
	ticker2 := time.NewTicker(time.Millisecond * 运行频率)
	defer ticker2.Stop()
	for range ticker2.C {
		if 遍历项目变量信息表中 {
			continue
		}
		if !完成第一次遍历项目变量信息表 {
			continue
		}
		for h := range 采集的行 {
			w.Wait()
			要写的变量们, ok := 哪个串口有哪些变量要写操作.获取数组对象(串口号1)
			if ok {
				for _, 要写的变量名及值 := range 要写的变量们 {
					for 要写的变量名, 要写的值 := range 要写的变量名及值 {
						fmt.Println("串口"+串口号1+"写>"+要写的变量名, "=", 要写的值)
						设备写操作(要写的变量名, 要写的值)
					}
				}
				for _, 要写的变量名及值 := range 要写的变量们 {
					for 要写的变量名 := range 要写的变量名及值 {
						哪个串口有哪些变量要写操作.Delete(串口号1, 要写的变量名)
					}
				}
			}
			func(z *Row) {
				当前时刻毫秒 := time.Now().UnixMilli()
				if 当前时刻毫秒-atomic.LoadInt64(&z.C采集时刻毫秒) < z.C采集频率毫秒 {
					return
				}
				w.Add(1)
				defer w.Done()
				p := modbus.NewRTUClientProvider()
				p.Address = "\\\\.\\" + "com" + z.C串口号 //必须要要加"\\\\.\\"，否则com9以上会报错："无法打开设备"
				p.BaudRate = z.B波特率
				p.DataBits = z.S数据位
				p.Parity = z.J奇偶校验
				p.StopBits = z.T停止位
				p.Timeout = z.T通讯超时毫秒 * time.Millisecond //毫秒
				client := modbus.NewClient(p)
				if 允许调试 {
					fmt.Println(z.B变量名称)
					client.LogMode(true)
				} else {
					client.LogMode(false)
				}
				err := client.Connect()
				atomic.AddUint64(&第几次采集, 1)
				if err != nil {
					fmt.Println("读" + z.B变量名称 + "发生错误_" + err.Error())
					str := err.Error()
					if 允许调试 {
						fmt.Println("com" + 串口号1)
						fmt.Println(str)
					}
					for _, i := range z.P批量读伙伴 {
						q := &项目变量信息表组.Rows[i]
						if 当前时刻毫秒-atomic.LoadInt64(&q.C采集时刻毫秒) < q.C采集频率毫秒 {
							continue
						}
						atomic.StoreInt64(&q.C采集时刻毫秒, 当前时刻毫秒)
						atomic.AddUint32(&q.L连续通讯失败次数, 1)
						if atomic.LoadUint32(&q.L连续通讯失败次数) > q.L连续通讯失败多少次则认为通讯异常 {
							q.D当前值锁.Lock()
							q.D当前值 = q.T通讯异常值
							q.D当前值锁.Unlock()
							变量组.KVTags[变量名所在行[q.B变量名称]].VarValue锁.Lock()
							变量组.KVTags[变量名所在行[q.B变量名称]].VarValue = q.T通讯异常值
							变量组.KVTags[变量名所在行[q.B变量名称]].VarValue锁.Unlock()
						}
					} //for i:= range z.P批量读伙伴 {
					return
				}
				defer client.Close()
				//time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
				switch z.D读功能码 {
				case "1":
					读功能码1(client, z)
				case "2":
					读功能码2(client, z)
				case "3":
					读功能码3(client, z)
				case "4":
					读功能码4(client, z)
				default:
					fmt.Println("无效的读功能码")
					return
				}
			}(&项目变量信息表组.Rows[h]) //func() {
		} //for h:= range 采集的行 {
	} //for {
} //func 采集任务(串口号1 string) {
func uint16SliceToByteSlice(data []uint16) []byte {
	var buf bytes.Buffer
	for _, num := range data {
		// 使用二进制编码将 uint16 转换为字节，并写入缓冲区
		err := binary.Write(&buf, binary.BigEndian, num)
		if err != nil {
			panic(err)
		}
	}
	return buf.Bytes()
}

// IsUTF8 checks if the given byte slice is valid UTF-8 encoded.
func IsUTF8(b []byte) (bool, error) {
	if !utf8.Valid(b) {
		return false, errors.New("invalid UTF-8 encoding")
	}
	return true, nil
}

// BytesToString converts a valid UTF-8 encoded byte slice to a string.
func BytesToString(b []byte) (string, error) {
	if valid, err := IsUTF8(b); !valid {
		return "", err
	}
	return string(b), nil
}
func IO字符串数据解析(vs []uint16, 批量读开始寄存器地址 uint16, z *Row) {
	atomic.StoreInt64(&z.C采集时刻毫秒, time.Now().UnixMilli())
	变量位置 := z.J寄存器地址 - 批量读开始寄存器地址
	字最大编号 := uint16(len(vs))
	if 1 > 字最大编号 {
		fmt.Println("读_" + z.B变量名称 + "_发生错误_uint16(len(vs))=0")
		return
	}
	if 变量位置 > 字最大编号-1 {
		fmt.Println("读_" + z.B变量名称 + "_发生错误_变量位置 > 字最大编号-1")
		return
	}
	if z.X小数点后值 <= 0 {
		fmt.Println("读_" + z.B变量名称 + "_发生错误_z.X小数点后值 <= 0")
		return
	}
	if z.X小数点后值 > 最大字符串字节长度 {
		fmt.Println("读_" + z.B变量名称 + "_发生错误_z.X小数点后值 > 最大字符串字节长度")
		return
	}
	字符占字数 := z.X小数点后值 / 2
	if (z.X小数点后值 % 2) > 0 {
		字符占字数++
	}
	字符串空间最大字编址 := z.J寄存器地址
	if 字符占字数 > 0 {
		字符串空间最大字编址 = z.J寄存器地址 + uint16(字符占字数) - 1
	}
	if 字符串空间最大字编址 > (批量读开始寄存器地址 + 字最大编号 - 1) {
		fmt.Println("读_" + z.B变量名称 + "_发生错误_字符串空间最大字编址 > (批量读开始寄存器地址 + 字最大编号 - 1)")
		return
	}
	字符串 := ""
	IO字符串字节数组 := uint16SliceToByteSlice(vs[变量位置 : 变量位置+uint16(字符占字数)])
	switch z.D读字节序 {
	case "ABCD":
	case "DCBA":
		reverseBytes(IO字符串字节数组)
	case "CDAB":
		swapBytes(IO字符串字节数组)
	case "BADC":
		swapPairs(IO字符串字节数组)
	default:
		fmt.Println("读_" + z.B变量名称 + "_发生错误_未知的D读字节序_" + z.D读字节序)
		字符串 = "未知的D读字节序_" + z.D读字节序
	}
	if 字符串 == "" {
		var nonZeroBytes []byte
		for _, b := range IO字符串字节数组 {
			if b == 0 {
				break // 遇到0元素时停止遍历
			}
			nonZeroBytes = append(nonZeroBytes, b)
		}
		if (z.X小数点后值 % 2) > 0 {
			if z.X小数点后值 < len(nonZeroBytes) {
				nonZeroBytes = nonZeroBytes[:z.X小数点后值]
			}
		}
		str, err := BytesToString(nonZeroBytes)
		if err != nil {
			//字符串 = err.Error()
			字符串 = "非UTF8字符串"
		} else {
			字符串 = str
		}
	}
	z.D当前值锁.Lock()
	z.D当前值 = 字符串
	z.D当前值锁.Unlock()
	atomic.StoreInt64(&z.C采集成功时刻, time.Now().Unix())
	atomic.StoreUint32(&z.L连续通讯失败次数, 0)
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = 字符串
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
} //func IO字符串数据解析(vs []uint16,批量读开始寄存器地址 uint16,z *Row){
func IO离散字数据解析(vs []uint16, 批量读开始寄存器地址 uint16, z *Row) {
	atomic.StoreInt64(&z.C采集时刻毫秒, time.Now().UnixMilli())
	变量位置 := z.J寄存器地址 - 批量读开始寄存器地址
	if int(变量位置) > len(vs)-1 {
		return
	}
	if z.X小数点后值 < 0 {
		return
	}
	if z.X小数点后值 > 15 {
		return
	}
	值 := vs[变量位置] & (1 << z.X小数点后值)
	if 值 > 0 {
		z.D当前值锁.Lock()
		z.D当前值 = "1"
		z.D当前值锁.Unlock()
		atomic.StoreInt64(&z.C采集成功时刻, time.Now().Unix())
		atomic.StoreUint32(&z.L连续通讯失败次数, 0)
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = "1"
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
	} else {
		z.D当前值锁.Lock()
		z.D当前值 = "0"
		z.D当前值锁.Unlock()
		atomic.StoreInt64(&z.C采集成功时刻, time.Now().Unix())
		atomic.StoreUint32(&z.L连续通讯失败次数, 0)
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = "0"
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
	}
} //func IO离散字数据解析(vs []uint16,批量读开始寄存器地址 uint16,z *Row){
func IO离散位数据解析(vs []byte, 批量读开始寄存器地址, 打包长度 uint16, z *Row) {
	atomic.StoreInt64(&z.C采集时刻毫秒, time.Now().UnixMilli())
	if z.J寄存器地址 < 批量读开始寄存器地址 {
		return
	}
	if z.J寄存器地址 > 批量读开始寄存器地址+打包长度-1 {
		return
	}
	if len(vs)*8 < int(打包长度) {
		return
	}
	i := z.J寄存器地址 - 批量读开始寄存器地址
	值 := vs[i/8] & (1 << (i % 8))
	if 值 > 0 {
		z.D当前值锁.Lock()
		z.D当前值 = "1"
		z.D当前值锁.Unlock()
		atomic.StoreInt64(&z.C采集成功时刻, time.Now().Unix())
		atomic.StoreUint32(&z.L连续通讯失败次数, 0)
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = "1"
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
	} else {
		z.D当前值锁.Lock()
		z.D当前值 = "0"
		z.D当前值锁.Unlock()
		atomic.StoreInt64(&z.C采集成功时刻, time.Now().Unix())
		atomic.StoreUint32(&z.L连续通讯失败次数, 0)
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = "0"
		变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
	}
} //func IO离散位数据解析(vs []byte, 批量读开始寄存器地址 uint16, z *Row) {
func IO整型数据解析(vs []uint16, 批量读开始寄存器地址 uint16, z *Row) {
	atomic.StoreInt64(&z.C采集时刻毫秒, time.Now().UnixMilli())
	i := z.J寄存器地址 - 批量读开始寄存器地址
	j := uint16(len(vs))
	if 1 > j {
		return
	}
	o := true
	switch z.S数据类型 {
	case "LONGBCD":
		o = false
	case "LONG":
		o = false
	case "ULONG":
		o = false
	} //switch z.S数据类型 {
	if o {
		if i > j-1 {
			return
		}
	} else {
		if i+1 > j-1 {
			return
		}
	}
	switch z.D读字节序 {
	case "CDAB":
		vs[i], vs[i+1] = vs[i+1], vs[i]
	case "BADC":
		vs[i] = (vs[i]&0xff00)>>8 + (vs[i]&0x00ff)<<8
		vs[i+1] = (vs[i+1]&0xff00)>>8 + (vs[i+1]&0x00ff)<<8
	case "DCBA":
		vs[i], vs[i+1] = vs[i+1], vs[i]
		vs[i] = (vs[i]&0xff00)>>8 + (vs[i]&0x00ff)<<8
		vs[i+1] = (vs[i+1]&0xff00)>>8 + (vs[i+1]&0x00ff)<<8
	} //switch 读字节序 {
	var 值 string
	switch z.S数据类型 {
	case "LONGBCD":
		值1 := HEX2LONGBCD(uint32(vs[i])<<16 + uint32(vs[i+1]))
		值 = strconv.Itoa(int(值1))
	case "LONG":
		值1 := int32(uint32(vs[i])<<16 + uint32(vs[i+1]))
		值 = strconv.Itoa(int(值1))
	case "ULONG":
		值1 := uint32(vs[i])<<16 + uint32(vs[i+1])
		值 = strconv.Itoa(int(值1))
	case "BCD":
		值1 := HEX2BCD(vs[i])
		值 = strconv.Itoa(int(值1))
	case "SHORT":
		值1 := int16(vs[i])
		值 = strconv.Itoa(int(值1))
	case "USHORT":
		值1 := vs[i]
		值 = strconv.Itoa(int(值1))
	} //switch z.S数据类型 {
	z.D当前值锁.Lock()
	z.D当前值 = 值
	z.D当前值锁.Unlock()
	atomic.StoreInt64(&z.C采集成功时刻, time.Now().Unix())
	atomic.StoreUint32(&z.L连续通讯失败次数, 0)
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = 值
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
} //func IO整型数据解析(vs []uint16,批量读开始寄存器地址 uint16,z *Row){
func IO实型数据解析(vs []uint16, 批量读开始寄存器地址 uint16, z *Row) {
	atomic.StoreInt64(&z.C采集时刻毫秒, time.Now().UnixMilli())
	if areFloatsEqual(z.J计算值除原始值, 0) {
		z.J计算值除原始值 = 1
	}
	i := z.J寄存器地址 - 批量读开始寄存器地址
	j := uint16(len(vs))
	if 1 > j {
		return
	}
	o := true
	switch z.S数据类型 {
	case "FLOAT":
		o = false
	case "LONGBCD":
		o = false
	case "LONG":
		o = false
	case "ULONG":
		o = false
	} //switch z.S数据类型 {
	if o {
		if i > j-1 {
			return
		}
	} else {
		if i+1 > j-1 {
			return
		}
	}
	switch z.D读字节序 {
	case "CDAB":
		vs[i], vs[i+1] = vs[i+1], vs[i]
	case "BADC":
		vs[i] = (vs[i]&0xff00)>>8 + (vs[i]&0x00ff)<<8
		vs[i+1] = (vs[i+1]&0xff00)>>8 + (vs[i+1]&0x00ff)<<8
	case "DCBA":
		vs[i], vs[i+1] = vs[i+1], vs[i]
		vs[i] = (vs[i]&0xff00)>>8 + (vs[i]&0x00ff)<<8
		vs[i+1] = (vs[i+1]&0xff00)>>8 + (vs[i+1]&0x00ff)<<8
	} //switch 读字节序 {
	var 值 string
	switch z.S数据类型 {
	case "FLOAT":
		值 = strconv.FormatFloat(float64(math.Float32frombits(uint32(vs[i])<<16+uint32(vs[i+1]))), 'f', z.X小数位数, 32)
	case "LONGBCD":
		值1 := HEX2LONGBCD(uint32(vs[i])<<16 + uint32(vs[i+1]))
		if !areFloatsEqual(z.J计算值除原始值, 1) {
			值2 := float64(值1) * z.J计算值除原始值
			值 = strconv.FormatFloat(值2, 'f', z.X小数位数, 32)
		} else {
			值 = strconv.Itoa(int(值1))
		}
	case "LONG":
		值1 := int32(uint32(vs[i])<<16 + uint32(vs[i+1]))
		if !areFloatsEqual(z.J计算值除原始值, 1) {
			值2 := float64(值1) * z.J计算值除原始值
			值 = strconv.FormatFloat(值2, 'f', z.X小数位数, 32)
		} else {
			值 = strconv.Itoa(int(值1))
		}
	case "ULONG":
		值1 := uint32(vs[i])<<16 + uint32(vs[i+1])
		if !areFloatsEqual(z.J计算值除原始值, 1) {
			值2 := float64(值1) * z.J计算值除原始值
			值 = strconv.FormatFloat(值2, 'f', z.X小数位数, 32)
		} else {
			值 = strconv.Itoa(int(值1))
		}
	case "BCD":
		值1 := HEX2BCD(vs[i])
		if !areFloatsEqual(z.J计算值除原始值, 1) {
			值2 := float64(值1) * z.J计算值除原始值
			值 = strconv.FormatFloat(值2, 'f', z.X小数位数, 32)
		} else {
			值 = strconv.Itoa(int(值1))
		}
	case "SHORT":
		值1 := int16(vs[i])
		if !areFloatsEqual(z.J计算值除原始值, 1) {
			值2 := float64(值1) * z.J计算值除原始值
			值 = strconv.FormatFloat(值2, 'f', z.X小数位数, 32)
		} else {
			值 = strconv.Itoa(int(值1))
		}
	case "USHORT":
		值1 := vs[i]
		if !areFloatsEqual(z.J计算值除原始值, 1) {
			值2 := float64(值1) * z.J计算值除原始值
			值 = strconv.FormatFloat(值2, 'f', z.X小数位数, 32)
		} else {
			值 = strconv.Itoa(int(值1))
		}
	} //switch z.S数据类型 {
	z.D当前值锁.Lock()
	z.D当前值 = 值
	z.D当前值锁.Unlock()
	atomic.StoreInt64(&z.C采集成功时刻, time.Now().Unix())
	atomic.StoreUint32(&z.L连续通讯失败次数, 0)
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = 值
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
} //func IO实型数据解析(vs []uint16,批量读开始寄存器地址 uint16,z *Row){
func 读功能码1写(client modbus.Client, z *Row, 要写的值 interface{}) {
	写值, ok := 要写的值.(bool)
	if !ok {
		fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非布尔值")
		return
	}
	if z.C采集前等待毫秒 != 0 {
		time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
	}
	atomic.AddUint64(&第几次写设备操作, 1)
	err := client.WriteSingleCoil(z.S设备地址, z.J寄存器地址, 写值)
	if err != nil {
		fmt.Println("写_" + z.B变量名称 + "_发生错误_" + err.Error())
		return
	}
	写值2 := "0"
	if 写值 {
		写值2 = "1"
	}
	z.D当前值锁.Lock()
	z.D当前值 = 写值2
	z.D当前值锁.Unlock()
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = 写值2
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
} //func 读功能码1(p *Row){
func 读功能码1(client modbus.Client, z *Row) {
	if z.Y允许通讯异常后只读变量 == "允许" && atomic.LoadUint32(&z.L连续通讯失败次数) > z.L连续通讯失败多少次则认为通讯异常 {
		if z.C采集前等待毫秒 != 0 {
			time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
		}
		vs, err := client.ReadCoils(z.S设备地址, z.J寄存器地址, 1)
		if err != nil {
			fmt.Println("读" + z.B变量名称 + "发生错误_" + err.Error())
			atomic.StoreInt64(&z.C采集时刻毫秒, time.Now().UnixMilli())
			atomic.AddUint32(&z.L连续通讯失败次数, 1)
			if atomic.LoadUint32(&z.L连续通讯失败次数) > z.L连续通讯失败多少次则认为通讯异常 {
				z.D当前值锁.Lock()
				z.D当前值 = z.T通讯异常值
				z.D当前值锁.Unlock()
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = z.T通讯异常值
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
			}
			return
		}
		IO离散位数据解析(vs, z.J寄存器地址, 1, z)
		return
	} //if z.Y允许通讯异常后只读变量=="允许"&&z.L连续通讯失败次数>z.L连续通讯失败多少次则认为通讯异常{
	if z.C采集前等待毫秒 != 0 {
		time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
	}
	vs, err := client.ReadCoils(z.S设备地址, z.P批量读开始寄存器地址, z.D打包长度)
	if err != nil {
		fmt.Println("读" + z.B变量名称 + "发生错误_" + err.Error())
		当前时刻毫秒 := time.Now().UnixMilli()
		for _, i := range z.P批量读伙伴 {
			q := &项目变量信息表组.Rows[i]
			if 当前时刻毫秒-atomic.LoadInt64(&q.C采集时刻毫秒) < q.C采集频率毫秒 {
				continue
			}
			atomic.StoreInt64(&q.C采集时刻毫秒, 当前时刻毫秒)
			atomic.AddUint32(&q.L连续通讯失败次数, 1)
			if atomic.LoadUint32(&q.L连续通讯失败次数) > q.L连续通讯失败多少次则认为通讯异常 {
				q.D当前值锁.Lock()
				q.D当前值 = q.T通讯异常值
				q.D当前值锁.Unlock()
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue锁.Lock()
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue = q.T通讯异常值
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue锁.Unlock()
			}
		} //for i:= range z.P批量读伙伴 {
		return
	}
	for _, i := range z.P批量读伙伴 {
		p := &项目变量信息表组.Rows[i]
		IO离散位数据解析(vs, z.P批量读开始寄存器地址, z.D打包长度, p)
	} //for i:= range z.P批量读伙伴 {
} //func 读功能码1(p *Row){
func 读功能码2(client modbus.Client, z *Row) {
	if z.Y允许通讯异常后只读变量 == "允许" && atomic.LoadUint32(&z.L连续通讯失败次数) > z.L连续通讯失败多少次则认为通讯异常 {
		if z.C采集前等待毫秒 != 0 {
			time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
		}
		vs, err := client.ReadDiscreteInputs(z.S设备地址, z.J寄存器地址, 1)
		if err != nil {
			fmt.Println("读" + z.B变量名称 + "发生错误_" + err.Error())
			atomic.StoreInt64(&z.C采集时刻毫秒, time.Now().UnixMilli())
			atomic.AddUint32(&z.L连续通讯失败次数, 1)
			if atomic.LoadUint32(&z.L连续通讯失败次数) > z.L连续通讯失败多少次则认为通讯异常 {
				z.D当前值锁.Lock()
				z.D当前值 = z.T通讯异常值
				z.D当前值锁.Unlock()
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = z.T通讯异常值
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
			}
			return
		}
		IO离散位数据解析(vs, z.J寄存器地址, 1, z)
		return
	} //if z.Y允许通讯异常后只读变量=="允许"&&z.L连续通讯失败次数>z.L连续通讯失败多少次则认为通讯异常{
	if z.C采集前等待毫秒 != 0 {
		time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
	}
	vs, err := client.ReadDiscreteInputs(z.S设备地址, z.P批量读开始寄存器地址, z.D打包长度)
	if err != nil {
		fmt.Println("读" + z.B变量名称 + "发生错误_" + err.Error())
		当前时刻毫秒 := time.Now().UnixMilli()
		for _, i := range z.P批量读伙伴 {
			q := &项目变量信息表组.Rows[i]
			if 当前时刻毫秒-atomic.LoadInt64(&q.C采集时刻毫秒) < q.C采集频率毫秒 {
				continue
			}
			atomic.StoreInt64(&q.C采集时刻毫秒, 当前时刻毫秒)
			atomic.AddUint32(&q.L连续通讯失败次数, 1)
			if atomic.LoadUint32(&q.L连续通讯失败次数) > q.L连续通讯失败多少次则认为通讯异常 {
				q.D当前值锁.Lock()
				q.D当前值 = q.T通讯异常值
				q.D当前值锁.Unlock()
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue锁.Lock()
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue = q.T通讯异常值
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue锁.Unlock()
			}
		} //for i:= range z.P批量读伙伴 {
		return
	}
	for _, i := range z.P批量读伙伴 {
		p := &项目变量信息表组.Rows[i]
		IO离散位数据解析(vs, z.P批量读开始寄存器地址, z.D打包长度, p)
	} //for i:= range z.P批量读伙伴 {
} //func 读功能码2(p *Row){
func 读功能码4(client modbus.Client, z *Row) {
	if z.Y允许通讯异常后只读变量 == "允许" && atomic.LoadUint32(&z.L连续通讯失败次数) > z.L连续通讯失败多少次则认为通讯异常 {
		var 读字个数 uint16 = 1
		读字个数 = 获得变量读字个数(z)
		if 读字个数 == 0 {
			fmt.Println("读" + z.B变量名称 + "发生错误_读字个数 == 0")
			return
		}
		if z.C采集前等待毫秒 != 0 {
			time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
		}
		vs, err := client.ReadInputRegisters(z.S设备地址, z.J寄存器地址, 读字个数)
		if err != nil {
			fmt.Println("读" + z.B变量名称 + "发生错误_" + err.Error())
			atomic.StoreInt64(&z.C采集时刻毫秒, time.Now().UnixMilli())
			atomic.AddUint32(&z.L连续通讯失败次数, 1)
			if atomic.LoadUint32(&z.L连续通讯失败次数) > z.L连续通讯失败多少次则认为通讯异常 {
				z.D当前值锁.Lock()
				z.D当前值 = z.T通讯异常值
				z.D当前值锁.Unlock()
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = z.T通讯异常值
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
			}
			return
		}
		switch z.B变量类型 {
		case "IO离散":
			IO离散字数据解析(vs, z.J寄存器地址, z)
		case "IO整型":
			IO整型数据解析(vs, z.J寄存器地址, z)
		case "IO实型":
			IO实型数据解析(vs, z.J寄存器地址, z)
		case "IO字符串":
			IO字符串数据解析(vs, z.J寄存器地址, z)
		} //switch z.B变量类型 {
		return
	} //if z.Y允许通讯异常后只读变量=="允许"&&z.L连续通讯失败次数>z.L连续通讯失败多少次则认为通讯异常{
	if z.C采集前等待毫秒 != 0 {
		time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
	}
	vs, err := client.ReadInputRegisters(z.S设备地址, z.P批量读开始寄存器地址, z.D打包长度)
	if err != nil {
		fmt.Println("读" + z.B变量名称 + "发生错误_" + err.Error())
		当前时刻毫秒 := time.Now().UnixMilli()
		for _, i := range z.P批量读伙伴 {
			q := &项目变量信息表组.Rows[i]
			if 当前时刻毫秒-atomic.LoadInt64(&q.C采集时刻毫秒) < q.C采集频率毫秒 {
				continue
			}
			atomic.StoreInt64(&q.C采集时刻毫秒, 当前时刻毫秒)
			atomic.AddUint32(&q.L连续通讯失败次数, 1)
			if atomic.LoadUint32(&q.L连续通讯失败次数) > q.L连续通讯失败多少次则认为通讯异常 {
				q.D当前值锁.Lock()
				q.D当前值 = q.T通讯异常值
				q.D当前值锁.Unlock()
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue锁.Lock()
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue = q.T通讯异常值
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue锁.Unlock()
			}
		} //for i:= range z.P批量读伙伴 {
		return
	}
	for _, i := range z.P批量读伙伴 {
		p := &项目变量信息表组.Rows[i]
		switch p.B变量类型 {
		case "IO离散":
			IO离散字数据解析(vs, z.P批量读开始寄存器地址, p)
		case "IO整型":
			IO整型数据解析(vs, z.P批量读开始寄存器地址, p)
		case "IO实型":
			IO实型数据解析(vs, z.P批量读开始寄存器地址, p)
		case "IO字符串":
			IO字符串数据解析(vs, z.P批量读开始寄存器地址, p)
		} //switch z.B变量类型 {
	} //for i:= range z.P批量读伙伴 {
} //func 读功能码4(p *Row){
func 读功能码3(client modbus.Client, z *Row) {
	//fmt.Println("读功能码3")
	if z.Y允许通讯异常后只读变量 == "允许" && atomic.LoadUint32(&z.L连续通讯失败次数) > z.L连续通讯失败多少次则认为通讯异常 {
		var 读字个数 uint16 = 1
		读字个数 = 获得变量读字个数(z)
		if 读字个数 == 0 {
			fmt.Println("读" + z.B变量名称 + "发生错误_读字个数 == 0")
			return
		}
		if z.C采集前等待毫秒 != 0 {
			time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
		}
		vs, err := client.ReadHoldingRegisters(z.S设备地址, z.J寄存器地址, 读字个数)
		if err != nil {
			fmt.Println("读" + z.B变量名称 + "发生错误_" + err.Error())
			atomic.StoreInt64(&z.C采集时刻毫秒, time.Now().UnixMilli())
			atomic.AddUint32(&z.L连续通讯失败次数, 1)
			if atomic.LoadUint32(&z.L连续通讯失败次数) > z.L连续通讯失败多少次则认为通讯异常 {
				z.D当前值锁.Lock()
				z.D当前值 = z.T通讯异常值
				z.D当前值锁.Unlock()
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = z.T通讯异常值
				变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
			}
			return
		}
		switch z.B变量类型 {
		case "IO离散":
			IO离散字数据解析(vs, z.J寄存器地址, z)
		case "IO整型":
			IO整型数据解析(vs, z.J寄存器地址, z)
		case "IO实型":
			IO实型数据解析(vs, z.J寄存器地址, z)
		case "IO字符串":
			IO字符串数据解析(vs, z.J寄存器地址, z)
		} //switch z.B变量类型 {
		return
	} //if z.Y允许通讯异常后只读变量=="允许"&&z.L连续通讯失败次数>z.L连续通讯失败多少次则认为通讯异常{
	//fmt.Println("读功能码3a")
	if z.C采集前等待毫秒 != 0 {
		time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
	}
	vs, err := client.ReadHoldingRegisters(z.S设备地址, z.P批量读开始寄存器地址, z.D打包长度)
	//fmt.Println("读功能码3b")
	if err != nil {
		fmt.Println("读" + z.B变量名称 + "发生错误_" + err.Error())
		当前时刻毫秒 := time.Now().UnixMilli()
		for _, i := range z.P批量读伙伴 {
			q := &项目变量信息表组.Rows[i]
			if 当前时刻毫秒-atomic.LoadInt64(&q.C采集时刻毫秒) < q.C采集频率毫秒 {
				continue
			}
			atomic.StoreInt64(&q.C采集时刻毫秒, 当前时刻毫秒)
			atomic.AddUint32(&q.L连续通讯失败次数, 1)
			if atomic.LoadUint32(&q.L连续通讯失败次数) > q.L连续通讯失败多少次则认为通讯异常 {
				q.D当前值锁.Lock()
				q.D当前值 = q.T通讯异常值
				q.D当前值锁.Unlock()
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue锁.Lock()
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue = q.T通讯异常值
				变量组.KVTags[变量名所在行[q.B变量名称]].VarValue锁.Unlock()
			}
		} //for i:= range z.P批量读伙伴 {
		return
	}
	//fmt.Println("读功能码3", vs)
	for _, i := range z.P批量读伙伴 {
		p := &项目变量信息表组.Rows[i]
		switch p.B变量类型 {
		case "IO离散":
			IO离散字数据解析(vs, z.P批量读开始寄存器地址, p)
		case "IO整型":
			IO整型数据解析(vs, z.P批量读开始寄存器地址, p)
		case "IO实型":
			IO实型数据解析(vs, z.P批量读开始寄存器地址, p)
		case "IO字符串":
			IO字符串数据解析(vs, z.P批量读开始寄存器地址, p)
		} //switch z.B变量类型 {
	} //for i:= range z.P批量读伙伴 {
} //func 读功能码3(p *Row){
func uint16ToBytes(num any) []byte {
	// 创建一个字节缓冲区
	buf := new(bytes.Buffer)
	// 使用 binary.Write 将 uint16 写入缓冲区
	// 注意：binary.BigEndian 表示使用大端字节序，如果是小端字节序则使用 binary.LittleEndian
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	// 返回缓冲区中的字节切片
	return buf.Bytes()
}

// B32位变量ToBytes converts a uint32 value to a byte slice with the specified byte order.
func B32位变量ToBytes(value any, z *Row) ([]byte, bool) {
	var order binary.ByteOrder
	将12字节交换并且34字节交换也交换 := false
	switch z.X写字节序 {
	case "ABCD":
		order = binary.BigEndian
	case "DCBA":
		order = binary.LittleEndian
	case "CDAB":
		order = binary.LittleEndian
		将12字节交换并且34字节交换也交换 = true
	case "BADC":
		order = binary.BigEndian
		将12字节交换并且34字节交换也交换 = true
	default:
		fmt.Println("写_" + z.B变量名称 + "_发生错误_未知的X写字节序_" + z.X写字节序)
		return nil, false
	}
	buf := new(bytes.Buffer)
	err := binary.Write(buf, order, value)
	if err != nil {
		fmt.Println("写_" + z.B变量名称 + "_发生错误_" + err.Error())
		return nil, false
	}
	if 将12字节交换并且34字节交换也交换 {
		var 字节数组 []byte = buf.Bytes()
		字节数组[0], 字节数组[1] = 字节数组[1], 字节数组[0]
		字节数组[2], 字节数组[3] = 字节数组[3], 字节数组[2]
		return 字节数组, true
	}
	return buf.Bytes(), true
} //func B32位变量ToBytes(value uint32, z *Row) ([]byte, bool) {
func float32ToString(num float32, decimals int) string {
	// 创建一个格式化字符串，其中 %.Xf 中的 X 将被替换为实际的小数位数
	format := fmt.Sprintf("%%.%df", decimals)
	// 使用 fmt.Sprintf 和格式化字符串将 float32 转换为字符串
	return fmt.Sprintf(format, num)
}
func 读功能码3写(client modbus.Client, z *Row, 要写的值 interface{}) {
	var 写字个数 uint16 = 1
	写值2 := ""
	var 要写的值2 []byte
	var 要写的值3 uint16
	switch z.B变量类型 {
	case "IO离散":
		写值, ok := 要写的值.(bool)
		if !ok {
			fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非布尔值")
			return
		}
		if z.X小数点后值 < 0 || z.X小数点后值 > 15 {
			fmt.Println("写_" + z.B变量名称 + "_IO离散_发生错误,z.X小数点后值<0||z.X小数点后值>15")
			return
		}
		if z.C采集前等待毫秒 != 0 {
			time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
		}
		原字值2, err := client.ReadHoldingRegisters(z.S设备地址, z.J寄存器地址, 1)
		if err != nil {
			fmt.Println("写_" + z.B变量名称 + "_IO离散_发生错误,无法从设备获得字数据_" + err.Error())
			return
		}
		原字值 := 原字值2[0]
		if 写值 {
			写值2 = "1"
			// 设置第n位为1
			原字值 |= uint16(1 << z.X小数点后值)
			要写的值3 = 原字值
		} else {
			写值2 = "0"
			// 清除第n位（将其设置为0）
			原字值 &^= uint16(1 << z.X小数点后值)
			要写的值3 = 原字值
		}
	case "IO整型":
		switch z.S数据类型 {
		case "LONGBCD", "ULONG":
			写字个数 = 2
			写值, ok := 要写的值.(uint32)
			if !ok {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非uint32")
				return
			}
			if 要写的值2, ok = B32位变量ToBytes(写值, z); !ok {
				return
			}
			写值2 = fmt.Sprintf("%d", 写值)
		case "LONG":
			写字个数 = 2
			写值, ok := 要写的值.(int32)
			if !ok {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非int32")
				return
			}
			if 要写的值2, ok = B32位变量ToBytes(写值, z); !ok {
				return
			}
			写值2 = fmt.Sprintf("%d", 写值)
		case "BCD", "USHORT":
			写字个数 = 1
			写值, ok := 要写的值.(uint16)
			if !ok {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非uint16")
				return
			}
			写值2 = fmt.Sprintf("%d", 写值)
			要写的值3 = 写值
		case "SHORT":
			写字个数 = 1
			写值, ok := 要写的值.(int16)
			if !ok {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非int16")
				return
			}
			写值2 = fmt.Sprintf("%d", 写值)
			要写的值3 = uint16(写值)
		default:
			fmt.Println("写_" + z.B变量名称 + "_发生错误_未知的数据类型_" + z.S数据类型)
			return
		} //switch z.S数据类型 {
	case "IO实型":
		switch z.S数据类型 {
		case "FLOAT":
			写字个数 = 2
			写值, ok := 要写的值.(float32)
			if !ok {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非float32")
				return
			}
			写值2 = float32ToString(写值, z.X小数位数)
			if 要写的值2, ok = B32位变量ToBytes(写值, z); !ok {
				return
			}
		case "LONGBCD":
			写字个数 = 2
			写值, ok := 要写的值.(float32)
			if !ok {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非float32")
				return
			}
			写值2 = float32ToString(写值, z.X小数位数)
			if areFloatsEqual(z.J计算值除原始值, 0) {
				z.J计算值除原始值 = 1
			}
			if !areFloatsEqual(z.J计算值除原始值, 1) {
				写值 /= float32(z.J计算值除原始值)
			}
			if 写值 < 0 || 写值 > 99999999 {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_写值 < 0 || 写值 > 99999999")
				return
			}
			if 要写的值2, ok = B32位变量ToBytes(uint32(int(math.Round(float64(写值)))), z); !ok {
				return
			}
		case "LONG":
			写字个数 = 2
			写值, ok := 要写的值.(float32)
			if !ok {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非float32")
				return
			}
			写值2 = float32ToString(写值, z.X小数位数)
			if areFloatsEqual(z.J计算值除原始值, 0) {
				z.J计算值除原始值 = 1
			}
			if !areFloatsEqual(z.J计算值除原始值, 1) {
				写值 /= float32(z.J计算值除原始值)
			}
			if 写值 < math.MinInt32 || 写值 > math.MaxInt32 {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_写值 < math.MinInt32 || 写值 > math.MaxInt32")
				return
			}
			if 要写的值2, ok = B32位变量ToBytes(int32(int(math.Round(float64(写值)))), z); !ok {
				return
			}
		case "ULONG":
			写字个数 = 2
			写值, ok := 要写的值.(float32)
			if !ok {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非float32")
				return
			}
			写值2 = float32ToString(写值, z.X小数位数)
			if areFloatsEqual(z.J计算值除原始值, 0) {
				z.J计算值除原始值 = 1
			}
			if !areFloatsEqual(z.J计算值除原始值, 1) {
				写值 /= float32(z.J计算值除原始值)
			}
			if 写值 < 0 || 写值 > math.MaxUint32 {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_写值 < 0 || 写值 > math.MaxUint32")
				return
			}
			if 要写的值2, ok = B32位变量ToBytes(uint32(int(math.Round(float64(写值)))), z); !ok {
				return
			}
		case "BCD":
			写字个数 = 1
			写值, ok := 要写的值.(float32)
			if !ok {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非float32")
				return
			}
			写值2 = float32ToString(写值, z.X小数位数)
			if areFloatsEqual(z.J计算值除原始值, 0) {
				z.J计算值除原始值 = 1
			}
			if !areFloatsEqual(z.J计算值除原始值, 1) {
				写值 /= float32(z.J计算值除原始值)
			}
			if 写值 < 0 || 写值 > 9999 {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_写值<0||写值>9999")
				return
			}
			要写的值3 = uint16(int(math.Round(float64(写值))))
		case "SHORT":
			写字个数 = 1
			写值, ok := 要写的值.(float32)
			if !ok {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非float32")
				return
			}
			写值2 = float32ToString(写值, z.X小数位数)
			if areFloatsEqual(z.J计算值除原始值, 0) {
				z.J计算值除原始值 = 1
			}
			if !areFloatsEqual(z.J计算值除原始值, 1) {
				写值 /= float32(z.J计算值除原始值)
			}
			if 写值 < math.MinInt16 || 写值 > math.MaxInt16 {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_写值 < math.MinInt16 || 写值 > math.MaxInt16")
				return
			}
			要写的值3 = uint16(int(math.Round(float64(写值))))
		case "USHORT":
			写字个数 = 1
			写值, ok := 要写的值.(float32)
			if !ok {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非float32")
				return
			}
			写值2 = float32ToString(写值, z.X小数位数)
			if areFloatsEqual(z.J计算值除原始值, 0) {
				z.J计算值除原始值 = 1
			}
			if !areFloatsEqual(z.J计算值除原始值, 1) {
				写值 /= float32(z.J计算值除原始值)
			}
			if 写值 < 0 || 写值 > math.MaxUint16 {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_写值 < 0 || 写值 > math.MaxUint16")
				return
			}
			要写的值3 = uint16(int(math.Round(float64(写值))))
		default:
			fmt.Println("写_" + z.B变量名称 + "_发生错误_未知的数据类型_" + z.S数据类型)
			return
		} //switch z.S数据类型 {
	case "IO字符串":
		if z.X小数点后值 <= 0 {
			return
		}
		if z.X小数点后值 > 最大字符串字节长度 {
			return
		}
		字符占字数 := z.X小数点后值 / 2
		if (z.X小数点后值 % 2) > 0 {
			字符占字数++
		}
		写字个数 = uint16(字符占字数)
		写值, ok := 要写的值.(string)
		if !ok {
			fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值非字符串")
			return
		}
		if len(写值) > z.X小数点后值 {
			fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值字符字节长度" +
				fmt.Sprintf("%d", len(写值)) + "超过设备字符字节长度" + fmt.Sprintf("%d", z.X小数点后值))
			return
		}
		写值2 = 写值
		if 要写的值2, ok = 字符串转字节数组(写值, z); !ok {
			return
		}
		if 写字个数 == 1 {
			要写的值3 = uint16(要写的值2[0])<<8 | uint16(要写的值2[1])
		}
	default:
		fmt.Println("写_" + z.B变量名称 + "_发生错误_未知的变量类型_" + z.B变量类型)
		return
	} //switch z.B变量类型 {
	if 写字个数 == 1 {
		if z.C采集前等待毫秒 != 0 {
			time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
		}
		atomic.AddUint64(&第几次写设备操作, 1)
		err := client.WriteSingleRegister(z.S设备地址, z.J寄存器地址, 要写的值3)
		if err != nil {
			fmt.Println("写_" + z.B变量名称 + "_发生错误_" + err.Error())
			fmt.Println("尝试用client.WriteMultipleRegisters写")
			if z.C采集前等待毫秒 != 0 {
				time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
			}
			atomic.AddUint64(&第几次写设备操作, 1)
			err = client.WriteMultipleRegisters(z.S设备地址, z.J寄存器地址, 1, uint16ToBytes(要写的值3))
			if err != nil {
				fmt.Println("写_" + z.B变量名称 + "_发生错误_" + err.Error())
				return
			}
		}
	} else {
		if z.C采集前等待毫秒 != 0 {
			time.Sleep(z.C采集前等待毫秒 * time.Millisecond)
		}
		atomic.AddUint64(&第几次写设备操作, 1)
		err := client.WriteMultipleRegisters(z.S设备地址, z.J寄存器地址, 写字个数, 要写的值2)
		if err != nil {
			fmt.Println("写_" + z.B变量名称 + "_发生错误_" + err.Error())
			return
		}
	}
	z.D当前值锁.Lock()
	z.D当前值 = 写值2
	z.D当前值锁.Unlock()
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Lock()
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue = 写值2
	变量组.KVTags[变量名所在行[z.B变量名称]].VarValue锁.Unlock()
} //func 读功能码3写(p *Row){
func 字符串转字节数组(value string, z *Row) ([]byte, bool) {
	写值字节数 := len(value)
	if 写值字节数 > z.X小数点后值 {
		fmt.Println("写_" + z.B变量名称 + "_发生错误_要写的值字符字节长度" +
			fmt.Sprintf("%d", 写值字节数) + "超过设备字符字节长度" + fmt.Sprintf("%d", z.X小数点后值))
		return nil, false
	}
	设备字符占字数 := z.X小数点后值 / 2
	if (z.X小数点后值 % 2) > 0 {
		设备字符占字数++
	}
	if 设备字符占字数 == 0 {
		设备字符占字数 = 1
	}
	var 字节数组 = make([]byte, 设备字符占字数*2)
	var 写值字节数组 = []byte(value)
	if 写值字节数%2 > 0 {
		写值字节数组 = append(写值字节数组, 0)
	}
	if 写值字节数 == 0 {
		写值字节数组 = []byte{0, 0}
	}
	copy(字节数组, 写值字节数组)
	switch z.X写字节序 {
	case "ABCD":
	case "DCBA":
		reverseBytes(字节数组)
	case "CDAB":
		swapBytes(字节数组)
	case "BADC":
		swapPairs(字节数组)
	default:
		fmt.Println("写_" + z.B变量名称 + "_发生错误_未知的X写字节序_" + z.X写字节序)
		return nil, false
	}
	return 字节数组, true
} //func 字符串转字节数组(写值 string, z *Row) ([]byte, bool) {
func swapPairs(b []byte) {
	if len(b)%2 != 0 {
		fmt.Println("Error: The length of the byte slice must be even.")
		return
	}
	for i := 0; i < len(b); i += 2 {
		b[i], b[i+1] = b[i+1], b[i] // Swap the current and next byte
	}
}
func reverseBytes(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}
func swapBytes(b []byte) {
	if len(b)%2 != 0 {
		fmt.Println("Error: The length of the byte slice must be even.")
		return
	}
	for i := 0; i < len(b); i += 4 {
		if i+3 >= len(b) {
			break // Remaining elements are less than 4, skip them
		}
		// Swap the first two bytes with the last two bytes
		b[i], b[i+2] = b[i+2], b[i]
		b[i+1], b[i+3] = b[i+3], b[i+1]
	}
}
func LONGBCD2HEX(bcd_data uint32) uint32 {
	if bcd_data == 0 {
		return 0
	}
	千万位 := bcd_data / 10000000 % 10
	百万位 := bcd_data / 1000000 % 10
	十万位 := bcd_data / 100000 % 10
	万位 := bcd_data / 10000 % 10
	千位 := bcd_data / 1000 % 10
	百位 := bcd_data / 100 % 10
	十位 := bcd_data / 10 % 10
	个位 := bcd_data % 10
	return uint32(千万位)<<28 + uint32(百万位)<<24 + uint32(十万位)<<20 + uint32(万位)<<16 + uint32(千位)<<12 + uint32(百位)<<8 + uint32(十位)<<4 + uint32(个位)
} //func LONGBCD2HEX(bcd_data uint32) uint32 {
func BCD2HEX(bcd_data uint16) uint16 {
	if bcd_data == 0 {
		return 0
	}
	千位 := bcd_data / 1000 % 10
	百位 := bcd_data / 100 % 10
	十位 := bcd_data / 10 % 10
	个位 := bcd_data % 10
	return uint16(千位)<<12 + uint16(百位)<<8 + uint16(十位)<<4 + uint16(个位)
} //func BCD2HEX(bcd_data uint16) uint16 {
func HEX2BCD(hex_data uint16) uint16 {
	if hex_data == 0 {
		return 0
	}
	var temp, bcd_data uint16
	temp = hex_data >> 12
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp*1000
	}
	temp = (hex_data >> 8) & 0x0f
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp*100
	}
	temp = (hex_data >> 4) & 0x0f
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp*10
	}
	temp = hex_data & 0x0f
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp
	}
	return bcd_data
} //func HEX2BCD(hex_data uint16) uint16 {
func HEX2LONGBCD(hex_data uint32) uint32 {
	if hex_data == 0 {
		return 0
	}
	var temp, bcd_data uint32
	temp = (hex_data >> 28) & 0x0f
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp*10000000
	}
	temp = (hex_data >> 24) & 0x0f
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp*1000000
	}
	temp = (hex_data >> 20) & 0x0f
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp*100000
	}
	temp = (hex_data >> 16) & 0x0f
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp*10000
	}
	temp = (hex_data >> 12) & 0x0f
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp*1000
	}
	temp = (hex_data >> 8) & 0x0f
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp*100
	}
	temp = (hex_data >> 4) & 0x0f
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp*10
	}
	temp = hex_data & 0x0f
	if temp > 9 {
		temp = 9
	}
	if temp > 0 {
		bcd_data = bcd_data + temp
	}
	return bcd_data
} //func HEX2LONGBCD(hex_data uint32) uint32 {

// 传入查询的端口号
// 返回端口号对应的进程PID，若没有找到相关进程，返回-1
func portInUse(portNumber int) int {
	res := -1
	var outBytes bytes.Buffer
	cmdStr := fmt.Sprintf("netstat -ano -p tcp | findstr %d", portNumber)
	cmd := exec.Command("cmd", "/c", cmdStr)
	cmd.Stdout = &outBytes
	cmd.Run()
	resStr := outBytes.String()
	r := regexp.MustCompile(`\s\d+\s`).FindAllString(resStr, -1)
	if len(r) > 0 {
		pid, err := strconv.Atoi(strings.TrimSpace(r[0]))
		if err != nil {
			res = -1
		} else {
			res = pid
		}
	}
	return res
} //func portInUse(portNumber int) int {
func 检查端口是否被占用() string {
	str := ""
	port, err := strconv.Atoi(端口)
	if err != nil {
		端口 = 默认端口号_s
	} else {
		if port < 最小端口号 || port > 最大端口号 {
			端口 = 默认端口号_s
		}
	}
	port, err = strconv.Atoi(端口)
	if err != nil {
		端口 = 默认端口号_s
		port = 默认端口号
	} else {
		if port < 最小端口号 || port > 最大端口号 {
			端口 = 默认端口号_s
			port = 默认端口号
		}
	}
	pid := portInUse(port)
	if pid != -1 {
		进程指针组, _ := process.Processes()
		占用端口进程名 := ""
		for _, 进程指针 := range 进程指针组 {
			pid2, _ := 进程指针.Ppid()
			pid3 := int(pid2)
			if pid3 == pid {
				str, _ = 进程指针.Name()
				占用端口进程名 = str
				str = "当前占用端口" + 端口 + "的进程名为:" + str + ",已经被杀掉，如果此进程对于您很重要，请重新设定端口号\r\n"
			}
		}
		pid1 := strconv.Itoa(pid)
		时间 := strings.Replace(time.Now().Format("2006-01-02 15:04:05"), ":", "", -1)
		文件名 := 占用端口进程名 + "_" + pid1 + "_杀掉占用端口的程序_" + 时间 + ".bat"
		内容 := "taskkill.exe /f /pid " + pid1 + "\r\n" + "exit\r\n"
		data1, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(内容))
		if err != nil {
			return err.Error()
		}
		创建文件2(目录+"bat/", 文件名, data1)
		邮件主题 := "端口被占用," + 发邮件主题.Load()
		邮件正文 := 邮件主题 + "\r\n" + str + "\r\n" + "服务器信息：\r\n" + 获取服务器信息() + "\r\n" + 获取服务器内存使用率() + "\r\n"
		go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
		c := exec.Command(目录 + "bat/" + 文件名)
		c.Start()
		time.Sleep(5 * time.Second)
	} else {
		str = "端口" + 端口 + "没有被占用\r\n"
	}
	return str
} //func 检查端口是否被占用() string {
func 创建目录(目录 string) {
	err := os.MkdirAll(目录, 0766)
	if err != nil {
		启动软件碰到的问题.Set("创建目录(" + 目录 + ")发生了错误(" + err.Error() + ")")
	}
}
func 判断目录是否存在(目录 string) (bool, error) {
	_, err := os.Stat(目录)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func 创建工作簿(工作簿名 string) {
	f := excelize.NewFile()
	index, _ := f.NewSheet(工作表名)
	f.NewSheet(设定工作表名)
	f.NewSheet(常用设定值表名)
	f.NewSheet(开发说明表名)
	写默认数据到变量工作表(f)
	写默认数据到设定工作表(f)
	写默认数据到开发说明表(f)
	写默认数据到常用设定值表(f)
	f.SetActiveSheet(index)
	if err := f.SaveAs(工作簿名); err != nil {
		启动软件碰到的问题.Set("保存工作簿(" + 工作簿名 + ")发生了错误(" + err.Error() + ")")
	}
	//time.Sleep(3 * time.Second)
}

var 细框线, 粗框线, 右粗框线, 底粗框线, 右下角粗框线 int = 0, 0, 0, 0, 0

func 获取框线值(重获值 bool, f *excelize.File) {
	if 细框线 > 0 && 粗框线 > 0 && 右粗框线 > 0 && 底粗框线 > 0 && 右下角粗框线 > 0 && !重获值 {
		return
	}
	exp := "@"
	var err error
	细框线, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		CustomNumFmt: &exp,
		Protection: &excelize.Protection{
			Hidden: false,
			Locked: false,
		},
	})
	if err != nil {
		启动软件碰到的问题.Set("细框线, err = f.NewStyle发生错误")
	}
	右粗框线, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 2},
		},
		CustomNumFmt: &exp,
		Protection: &excelize.Protection{
			Hidden: false,
			Locked: false,
		},
	})
	if err != nil {
		启动软件碰到的问题.Set("右粗框线, err = f.NewStyle发生错误")
	}
	底粗框线, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 1},
		},
		CustomNumFmt: &exp,
		Protection: &excelize.Protection{
			Hidden: false,
			Locked: false,
		},
	})
	if err != nil {
		启动软件碰到的问题.Set("底粗框线, err = f.NewStyle发生错误")
	}
	右下角粗框线, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
		},
		CustomNumFmt: &exp,
		Protection: &excelize.Protection{
			Hidden: false,
			Locked: false,
		},
	})
	if err != nil {
		启动软件碰到的问题.Set("右下角粗框线, err = f.NewStyle发生错误")
	}
	粗框线, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
		},
		CustomNumFmt: &exp,
	})
	if err != nil {
		启动软件碰到的问题.Set("粗框线, err = f.NewStyle发生错误")
	}
} //func 获取框线值(){
func 画框线(工作表名2 string, 总行数, 总列数 int, f *excelize.File) {
	总行数符号 := strconv.Itoa(总行数)
	总列数符号, _ := excelize.ColumnNumberToName(总列数)
	右下角单元格 := 总列数符号 + 总行数符号
	if 细框线 > 0 {
		if err := f.SetCellStyle(工作表名2, "A2", 右下角单元格, 细框线); err != nil {
			启动软件碰到的问题.Set("f.SetCellStyle(" + 工作表名2 + ", A2, " + 右下角单元格 + ", 细框线)发生错误")
		}
	}
	if 右粗框线 > 0 {
		if err := f.SetCellStyle(工作表名2, 总列数符号+"2", 右下角单元格, 右粗框线); err != nil {
			启动软件碰到的问题.Set("f.SetCellStyle(" + 工作表名2 + "," + 总列数符号 + "2," + 右下角单元格 +
				", 右粗框线)发生错误")
		}
	}
	if 底粗框线 > 0 {
		if err := f.SetCellStyle(工作表名2, "A"+总行数符号, 右下角单元格, 底粗框线); err != nil {
			启动软件碰到的问题.Set("f.SetCellStyle(" + 工作表名2 + ",A" + 总行数符号 + "," +
				右下角单元格 + ", 底粗框线)发生错误")
		}
	}
	if 右下角粗框线 > 0 {
		if err := f.SetCellStyle(工作表名2, 右下角单元格, 右下角单元格, 右下角粗框线); err != nil {
			启动软件碰到的问题.Set("f.SetCellStyle(" + 工作表名2 + "," + 右下角单元格 + ", " +
				右下角单元格 + ", 右下角粗框线)发生错误")
		}
	}
	if 粗框线 > 0 {
		if err := f.SetCellStyle(工作表名2, "A1", 总列数符号+"1", 粗框线); err != nil {
			启动软件碰到的问题.Set("f.SetCellStyle(工作表名2, A1, " + 总列数符号 + "1" + ", 粗框线)发生错误")
		}
	}
} //func 画框线(工作表名2,右下角单元格,总行数符号,总列数符号 string){
func 写默认数据到变量工作表(f *excelize.File) {
	总行数 := 19
	总列数 := 标题列数
	f.NewSheet(工作表名)
	获取框线值(false, f)
	画框线(工作表名, 总行数, 总列数, f)
	//冻结前2列//冻结前1行//左上角单元格h2
	var 格式 excelize.Panes
	格式.Freeze = true
	格式.ActivePane = "\"bottomRight\",\"panes\":[{\"pane\":\"topLeft\"},{\"pane\":\"topRight\"},{\"pane\":\"bottomLeft\"},{\"active_cell\":\"d2\",\"sqref\":\"d2\",\"pane\":\"bottomRight\"}]"
	格式.TopLeftCell = "d2"
	格式.XSplit = 3
	格式.YSplit = 1
	f.SetPanes(工作表名, &格式)
	for i := 0; i < 总列数; i++ {
		宽度 := 5.14
		列号, _ := excelize.ColumnNumberToName(i + 1)
		switch i {
		case 0:
			宽度 = 7.14
		case 1:
			宽度 = 9.43
		case 2:
			宽度 = 30.29
		case 3:
			宽度 = 28.29
		case 4:
			宽度 = 13.86
		case 5:
			宽度 = 11.86
		case 6:
			宽度 = 16.71
		case 7:
			宽度 = 9.43
		case 8:
			宽度 = 7.29
		case 9:
			宽度 = 7.29
		case 10:
			宽度 = 9.43
		case 11:
			宽度 = 7.29
		case 12:
			宽度 = 7.29
		case 13:
			宽度 = 17.29
		case 14:
			宽度 = 9.43
		case 15:
			宽度 = 9.43
		case 16:
			宽度 = 11.86
		case 17:
			宽度 = 23.86
		case 18:
			宽度 = 9.43
		case 19:
			宽度 = 26.29
		case 20:
			宽度 = 9.43
		case 21:
			宽度 = 9.43
		case 22:
			宽度 = 15.57
		case 23:
			宽度 = 18
		case 24:
			宽度 = 38.43
		case 25:
			宽度 = 9.43
		case 26:
			宽度 = 9.43
		case 27:
			宽度 = 9.43
		case 28:
			宽度 = 9.43
		case 29:
			宽度 = 24.45
		case 30:
			宽度 = 21.14
		}
		f.SetColWidth(工作表名, 列号, 列号, 宽度)
	} //for i := 0; i < 总列数; i++ {
	var 填值 interface{}
	填值 = ""
	for i := 0; i < 总列数; i++ {
		单元格, _ := excelize.ColumnNumberToName(i + 1)
		单元格 = 单元格 + "1"
		switch i {
		case 0:
			填值 = "变量ID"
		case 1:
			填值 = "变量类型"
		case 2:
			填值 = "变量名"
		case 3:
			填值 = "初始值；字符串(<128字节)"
		case 4:
			填值 = "通讯异常值"
		case 5:
			填值 = "是否保存值"
		case 6:
			填值 = "计算值除原始值"
		case 7:
			填值 = "小数位数"
		case 8:
			填值 = "串口号"
		case 9:
			填值 = "波特率"
		case 10:
			填值 = "奇偶校验"
		case 11:
			填值 = "数据位"
		case 12:
			填值 = "停止位"
		case 13:
			填值 = "通讯超时（毫秒）"
		case 14:
			填值 = "设备地址"
		case 15:
			填值 = "读功能码"
		case 16:
			填值 = "寄存器地址"
		case 17:
			填值 = "打包长度"
		case 18:
			填值 = "允许通讯异常后只读变量"
		case 19:
			填值 = "数据类型"
		case 20:
			填值 = "读写属性"
		case 21:
			填值 = "采集频率(毫秒)"
		case 22:
			填值 = "采集前等待(毫秒)"
		case 23:
			填值 = "连续通讯失败多少次则认为通讯异常"
		case 24:
			填值 = "读字节序"
		case 25:
			填值 = "写字节序"
		case 26:
			填值 = "错误信息"
		case 27:
			填值 = "警告信息"
		case 28:
			填值 = "单位"
		case 29:
			填值 = "多少变化率才记录到数据库"
		case 30:
			填值 = "记录到数据库间隔（秒）"
		}
		f.SetCellValue(工作表名, 单元格, 填值)
	} //for i := 0; i < 总列数; i++ {
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "a"+strconv.Itoa(i), strconv.Itoa(i-1))
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "b"+strconv.Itoa(i), "IO整型")
	}
	// for i := 2; i < 总行数+1; i++ {
	// 	f.SetCellValue(工作表名, "d"+strconv.Itoa(i), "0")
	// }
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "e"+strconv.Itoa(i), "-1")
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "f"+strconv.Itoa(i), "是")
	}
	// for i := 2; i < 总行数+1; i++ {
	// 	f.SetCellValue(工作表名, "g"+strconv.Itoa(i), "1")
	// }
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "h"+strconv.Itoa(i), "2")
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "I"+strconv.Itoa(i), 默认串口号)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "j"+strconv.Itoa(i), 默认波特率_s)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "k"+strconv.Itoa(i), 默认奇偶校验)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "l"+strconv.Itoa(i), 默认数据位_s)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "m"+strconv.Itoa(i), 默认停止位_s)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "n"+strconv.Itoa(i), 默认通讯超时_毫秒_s)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "O"+strconv.Itoa(i), 默认设备地址_s)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "P"+strconv.Itoa(i), 默认读功能码_s)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "Q"+strconv.Itoa(i), 默认寄存器地址_s)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "R"+strconv.Itoa(i), 默认字打包长度_s)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "S"+strconv.Itoa(i), "允许")
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "T"+strconv.Itoa(i), "FLOAT")
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "U"+strconv.Itoa(i), "只读")
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "V"+strconv.Itoa(i), 默认采集频率_毫秒_s)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "W"+strconv.Itoa(i), 默认采集前等待_毫秒_s)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "X"+strconv.Itoa(i), "3")
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "Y"+strconv.Itoa(i), 默认读写字节序)
	}
	// for i := 2; i < 总行数+1; i++ {
	// 	f.SetCellValue(工作表名, "AA"+strconv.Itoa(i), 默认读写字节序)
	// }
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "AD"+strconv.Itoa(i), 默认多少变化率才记录到数据库)
	}
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(工作表名, "AE"+strconv.Itoa(i), 默认记录到数据库间隔_秒_s)
	}
	for i := 2; i < 总行数+1; i++ {
		列号 := "B"
		switch i {
		case 2:
			填值 = "IO实型"
			// case 3:
		// 	填值 = "IO实型"
		// case 4:
		// 	填值 = "IO实型"
		// case 5:
		// 	填值 = "IO实型"
		// case 6:
		// 	填值 = "IO实型"
		// case 7:
		// 	填值 = "IO实型"
		// case 8:
		// 	填值 = "IO实型"
		case 9:
			填值 = "IO离散"
		case 10:
			填值 = "IO字符串"
		// case 11:
		// 	填值 = "IO实型"
		case 12:
			填值 = "IO离散"
		case 13:
			填值 = "IO离散"
		case 14:
			填值 = "IO实型"
		case 15:
			填值 = "IO实型"
		case 16:
			填值 = "内存离散"
		case 17:
			填值 = "内存字符串"
		case 18:
			填值 = "内存整型"
		case 19:
			填值 = "内存实型"
		} //switch i {
		if i == 2 || i == 9 || i == 10 || i >= 12 {
			f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
		}
	} //for i := 2; i < 总行数+1; i++ {
	for i := 2; i < 总行数+1; i++ {
		列号 := "C"
		switch i {
		case 2:
			填值 = 默认项目代号 + "_DS18B20_1_T"
		case 3:
			填值 = 默认项目代号 + "_DS18B20_1_T_LONGBCD"
		case 4:
			填值 = 默认项目代号 + "_DS18B20_1_T_LONG"
		case 5:
			填值 = 默认项目代号 + "_DS18B20_1_T_ULONG"
		case 6:
			填值 = 默认项目代号 + "_DS18B20_1_T_BCD"
		case 7:
			填值 = 默认项目代号 + "_DS18B20_1_T_SHORT"
		case 8:
			填值 = 默认项目代号 + "_DS18B20_1_T_USHORT"
		case 9:
			填值 = 默认项目代号 + "_DS18B20_1_T_BIT0"
		case 10:
			填值 = 默认项目代号 + "_服务器名"
		case 11:
			填值 = 默认项目代号 + "_服务器端口"
		case 12:
			填值 = 默认项目代号 + "_继电器1"
		case 13:
			填值 = 默认项目代号 + "_液位开关"
		case 14:
			填值 = 默认项目代号 + "_继电器1_得电温度"
		case 15:
			填值 = 默认项目代号 + "_继电器1_失电温度"
		case 16:
			填值 = 默认项目代号 + "_运行中"
		case 17:
			填值 = 默认项目代号 + "_项目编号"
		case 18:
			填值 = 默认项目代号 + "_程序周期计数"
		case 19:
			填值 = 默认项目代号 + "_报警温度设定"
		} //switch i {
		f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
	} //for i := 2; i < 总行数+1; i++ {
	for i := 2; i < 总行数+1; i++ {
		列号 := "E"
		switch i {
		case 2:
			填值 = 默认通讯异常值_FLOAT_s
			// case 3:
		// 	填值 = "IO实型"
		case 4:
			填值 = 默认通讯异常值_LONG_s
		// case 5:
		// 	填值 = "IO实型"
		// case 6:
		// 	填值 = "IO实型"
		case 7:
			填值 = 默认通讯异常值_SHORT_s
		// case 8:
		// 	填值 = "IO实型"
		// case 9:
		// 	填值 = "IO离散"
		case 10:
			填值 = 默认通讯异常值_STRING
		// case 11:
		// 	填值 = "IO实型"
		// case 12:
		// 	填值 = "IO离散"
		// case 13:
		// 	填值 = "IO离散"
		case 14:
			填值 = 默认通讯异常值_FLOAT_s
		case 15:
			填值 = 默认通讯异常值_FLOAT_s
		case 16:
			填值 = 默认通讯异常值_BIT_s
		case 17:
			填值 = 默认通讯异常值_STRING
		case 18:
			填值 = 内存默认通讯异常值_LONG_s
		case 19:
			填值 = 内存默认通讯异常值_FLOAT_s
		} //switch i {
		if i == 2 || i == 4 || i == 7 || i == 10 || i >= 14 {
			f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
		}
	} //for i := 2; i < 总行数+1; i++ {
	for i := 2; i < 总行数+1; i++ {
		列号 := "h"
		switch i {
		// case 2:
		// 	填值 = "IO实型"
		case 3:
			填值 = ""
		case 4:
			填值 = ""
		case 5:
			填值 = ""
		case 6:
			填值 = ""
		case 7:
			填值 = ""
		case 8:
			填值 = ""
		case 9:
			填值 = ""
		case 10:
			填值 = ""
		case 11:
			填值 = ""
		case 12:
			填值 = ""
		case 13:
			填值 = ""
			// case 14:
			// 	填值 = "IO实型"
			// case 15:
			// 	填值 = "IO实型"
		case 16:
			填值 = ""
		case 17:
			填值 = ""
		case 18:
			填值 = ""
			// case 19:
			// 	填值 = ""
		} //switch i {
		if i != 2 && i != 14 && i != 15 && i != 19 {
			f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
		}
	} //for i := 2; i < 总行数+1; i++ {
	for i := 2; i < 总行数+1; i++ {
		列号 := "P"
		switch i {
		case 12:
			填值 = "1"
		case 13:
			填值 = "2"
			// case 14:
			// 	填值 = "IO实型"
			// case 15:
			// 	填值 = "IO实型"
		case 16:
			填值 = ""
		case 17:
			填值 = ""
		case 18:
			填值 = ""
		case 19:
			填值 = ""
		} //switch i {
		if i == 12 || i == 13 || i >= 16 {
			f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
		}
	} //for i := 2; i < 总行数+1; i++ {
	for i := 2; i < 总行数+1; i++ {
		列号 := "Q"
		switch i {
		case 9:
			填值 = "65506.0"
		case 10:
			填值 = "65526.18"
		case 11:
			填值 = "65524"
		case 12:
			填值 = "0"
		case 13:
			填值 = "0"
		case 14:
			填值 = "65513"
		case 15:
			填值 = "65515"
		case 16:
			填值 = ""
		case 17:
			填值 = ""
		case 18:
			填值 = ""
		case 19:
			填值 = ""
		} //switch i {
		if !(i >= 2 && i <= 8) {
			f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
		}
	} //for i := 2; i < 总行数+1; i++ {
	for i := 2; i < 总行数+1; i++ {
		列号 := "R"
		switch i {
		// case 2:
		// 	填值 = "IO实型"
		// case 3:
		// 	填值 = "IO实型"
		// case 4:
		// 	填值 = "IO实型"
		// case 5:
		// 	填值 = "IO实型"
		// case 6:
		// 	填值 = "IO实型"
		// case 7:
		// 	填值 = "IO实型"
		// case 8:
		// 	填值 = "IO实型"
		// case 9:
		// 	填值 = "IO离散"
		// case 10:
		// 	填值 = "IO字符串"
		// case 11:
		// 	填值 = "IO实型"
		case 12:
			填值 = "1"
		case 13:
			填值 = "1"
			// case 14:
			// 	填值 = "IO实型"
			// case 15:
			// 	填值 = "IO实型"
		case 16:
			填值 = ""
		case 17:
			填值 = ""
		case 18:
			填值 = ""
		case 19:
			填值 = ""
		} //switch i {
		if i == 12 || i == 13 || i >= 16 {
			f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
		}
	} //for i := 2; i < 总行数+1; i++ {
	for i := 2; i < 总行数+1; i++ {
		列号 := "U"
		switch i {
		// case 2:
		// 	填值 = "IO实型"
		// case 3:
		// 	填值 = "IO实型"
		// case 4:
		// 	填值 = "IO实型"
		// case 5:
		// 	填值 = "IO实型"
		// case 6:
		// 	填值 = "IO实型"
		// case 7:
		// 	填值 = "IO实型"
		// case 8:
		// 	填值 = "IO实型"
		// case 9:
		// 	填值 = "IO离散"
		case 10:
			填值 = "读写"
		case 11:
			填值 = "读写"
		case 12:
			填值 = "读写"
		// case 13:
		// 	填值 = "IO离散"
		case 14:
			填值 = "读写"
		case 15:
			填值 = "读写"
		case 16:
			填值 = ""
		case 17:
			填值 = ""
		case 18:
			填值 = ""
		case 19:
			填值 = ""
		} //switch i {
		if i >= 10 && i != 13 {
			f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
		}
	} //for i := 2; i < 总行数+1; i++ {
	for i := 2; i < 总行数+1; i++ {
		列号 := "T"
		switch i {
		// case 2:
		// 	填值 = "IO实型"
		case 3:
			填值 = "LONGBCD"
		case 4:
			填值 = "LONG"
		case 5:
			填值 = "ULONG"
		case 6:
			填值 = "BCD"
		case 7:
			填值 = "SHORT"
		case 8:
			填值 = "USHORT"
		case 9:
			填值 = ""
		case 10:
			填值 = ""
		case 11:
			填值 = "USHORT"
		case 12:
			填值 = ""
		case 13:
			填值 = ""
		// case 14:
		// 	填值 = "IO实型"
		// case 15:
		// 	填值 = "IO实型"
		case 16:
			填值 = ""
		case 17:
			填值 = ""
		case 18:
			填值 = ""
		case 19:
			填值 = ""
		} //switch i {
		if i != 2 && i != 14 && i != 15 {
			f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
		}
	} //for i := 2; i < 总行数+1; i++ {
	for i := 2; i < 总行数+1; i++ {
		列号 := "Y"
		switch i {
		// case 2:
		// 	填值 = "IO实型"
		// case 3:
		// 	填值 = "IO实型"
		// case 4:
		// 	填值 = "IO实型"
		// case 5:
		// 	填值 = "IO实型"
		case 6:
			填值 = ""
		case 7:
			填值 = ""
		case 8:
			填值 = ""
		case 9:
			填值 = ""
		case 10:
			填值 = "BADC"
		case 11:
			填值 = ""
		case 12:
			填值 = ""
		case 13:
			填值 = ""
			// case 14:
			// 	填值 = "IO实型"
			// case 15:
			// 	填值 = "IO实型"
		case 16:
			填值 = ""
		case 17:
			填值 = ""
		case 18:
			填值 = ""
		case 19:
			填值 = ""
		} //switch i {
		if i > 5 && i != 14 && i != 15 {
			f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
		}
	} //for i := 2; i < 总行数+1; i++ {
	for i := 2; i < 总行数+1; i++ {
		列号 := "Z"
		switch i {
		// case 2:
		// 	填值 = "IO实型"
		// case 3:
		// 	填值 = "IO实型"
		// case 4:
		// 	填值 = "IO实型"
		// case 5:
		// 	填值 = "IO实型"
		// case 6:
		// 	填值 = ""
		// case 7:
		// 	填值 = ""
		// case 8:
		// 	填值 = ""
		// case 9:
		// 	填值 = ""
		case 10:
			填值 = "ABCD"
		// case 11:
		// 	填值 = ""
		// case 12:
		// 	填值 = ""
		// case 13:
		// 	填值 = ""
		case 14:
			填值 = "ABCD"
		case 15:
			填值 = "ABCD"
			// case 16:
			// 	填值 = ""
			// case 17:
			// 	填值 = ""
			// case 18:
			// 	填值 = ""
			// case 19:
			// 	填值 = ""
		} //switch i {
		if i == 10 || i == 14 || i == 15 {
			f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
		}
	} //for i := 2; i < 总行数+1; i++ {
	for i := 2; i < 总行数+1; i++ {
		列号 := "AC"
		switch i {
		case 2:
			填值 = "℃"
		case 14:
			填值 = "℃"
		case 15:
			填值 = "℃"
		case 19:
			填值 = "℃"
		} //switch i {
		if i == 2 || i == 14 || i == 15 || i == 19 {
			f.SetCellValue(工作表名, 列号+strconv.Itoa(i), 填值)
		}
	} //for i := 2; i < 总行数+1; i++ {
	填值 = ""
	for i := 16; i < 总行数+1; i++ {
		for j := 9; j < 21; j++ {
			f.SetCellValue(工作表名, 求单元格代号(i, j), 填值)
		}
	}
	for i := 16; i < 总行数+1; i++ {
		for j := 23; j < 26; j++ {
			f.SetCellValue(工作表名, 求单元格代号(i, j), 填值)
		}
	}
} //func 写默认数据到变量工作表() {
func 读设定工作表(f *excelize.File) bool {
	const (
		设定个数     = 18
		标题列数2    = 6
		序号_下标2   = 0
		设定名称_下标2 = 1
		设定值_下标2  = 2
		范围_下标2   = 3
		备注_下标2   = 4
		错误信息_下标2 = 5
	)
	var 标题列号2 [标题列数2]int
	rows, err := f.GetRows(设定工作表名)
	if err != nil {
		启动软件碰到的问题.Set("执行 rows, err := f.GetRows(设定工作表名)发生错误(" + err.Error() + ")")
		return false
	}
	表格行数 := len(rows)
	if 表格行数 < 2 {
		启动软件碰到的问题.Set("工作表(" + 设定工作表名 + ")表格行数 < 2或者此工作表不存在")
		f.DeleteSheet(设定工作表名)
		return false
	}
	应有的行数 := 设定个数 + 1
	if 表格行数 > 应有的行数 {
		启动软件碰到的问题.Set("工作表(" + 设定工作表名 + ")表格行数(" + strconv.Itoa(表格行数) +
			") > (设定个数(" + strconv.Itoa(设定个数) + ") + 1),软件将删除此表！")
		f.DeleteSheet(设定工作表名)
		return false
	}
	标题列数3 := 0
	错误信息 := ""
	for i, row := range rows {
		if i == 1 {
			if 标题列数2 != 标题列数3 {
				if 错误信息 != "" {
					错误信息 = 错误信息 + "\r\n"
				}
				错误信息 = 错误信息 + "读设定工作表发现标题列数为" + strconv.Itoa(标题列数3) + ",而软件认可的列数为" + strconv.Itoa(标题列数2) + ",于是软件认为此工作表非法！"
			}
			if 错误信息 != "" {
				启动软件碰到的问题.Set("工作表(" + 设定工作表名 + ")有错误(" + 错误信息 + "),软件将删除此表！")
				f.DeleteSheet(设定工作表名)
				return false //标题行有错误就退出
			}
		}
		错误信息 = ""
		for j, colCell := range row {
			个数 := strings.Count(colCell, " ")
			if 个数 > 0 {
				列号 := j + 1
				colCell = strings.Replace(colCell, " ", "", 个数)
				f.SetCellValue(设定工作表名, 求单元格代号(i+1, 列号), colCell)
				row[j] = colCell
			}
			if i == 0 {
				switch colCell {
				case "序号":
					标题列数3++
					标题列号2[序号_下标2] = j
				case "设定名称":
					标题列数3++
					标题列号2[设定名称_下标2] = j
				case "设定值":
					标题列数3++
					标题列号2[设定值_下标2] = j
				case "范围":
					标题列数3++
					标题列号2[范围_下标2] = j
				case "备注":
					标题列数3++
					标题列号2[备注_下标2] = j
				case "错误信息":
					标题列数3++
					标题列号2[错误信息_下标2] = j
				default:
					错误信息 = "标题列:" + colCell + " 不被程序认可！"
				} //switch colCell {
			}
		} //for j, colCell := range row {
	} //for i, row := range rows {
	设定个数1 := 0
	for i, row := range rows {
		错误信息 = ""
		设定名称 := row[标题列号2[设定名称_下标2]]
		设定名称 = strings.Replace(设定名称, " ", "", -1)
		设定名称 = strings.ToUpper(设定名称)
		设定值 := row[标题列号2[设定值_下标2]]
		设定值 = strings.Replace(设定值, " ", "", -1)
		switch 设定名称 {
		case "服务器端口":
			设定个数1++
			端口值, err1 := strconv.Atoi(设定值)
			if err1 != nil {
				错误信息 = "将服务器端口设定值字符串(" + 设定值 + ")转整数发生错误;"
			} else {
				if 端口值 < 最小端口号 || 端口值 > 最大端口号 {
					错误信息 = "服务器端口设定值(" + 设定值 + ")不在正常范围(" + 最小端口号_s + "~" + 最大端口号_s + ")内"
				} else {
					端口 = 设定值
					str2 := 检查端口是否被占用()
					启动软件碰到的问题.Set(str2)
					if !strings.Contains(row[标题列号2[错误信息_下标2]], str2) {
						f.SetCellValue(设定工作表名, 求单元格代号(i+1, 标题列号2[错误信息_下标2]+1), str2)
					}
				}
			}
			if 错误信息 != "" {
				端口 = 默认端口号_s
				错误信息 = 错误信息 + "将使用默认端口号(" + 默认端口号_s + ")\r\n"
				str1 := 检查端口是否被占用()
				错误信息 = 错误信息 + "\r\n" + str1
				启动软件碰到的问题.Set(错误信息)
				f.SetCellValue(设定工作表名, 求单元格代号(i+1, 标题列号2[设定值_下标2]+1), 默认端口号_s)
				f.SetCellValue(设定工作表名, 求单元格代号(i+1, 标题列号2[错误信息_下标2]+1), 错误信息)
			}
		case "UUID":
			设定个数1++
			检查结果 := UUID检查(设定值)
			if 检查结果 != "ok" {
				本网关UUID = uuid.New()
				错误信息 = 检查结果
			} else {
				本网关UUID = 设定值
			}
			if 错误信息 != "" {
				str2 := 检查结果 + "\r\n" + "将由软件重新生成新的UUID(" + 本网关UUID +
					")建议使用备忘的UUID，否则依赖此识别码对此网关操作的软件会工作异常"
				启动软件碰到的问题.Set(str2)
				f.SetCellValue(设定工作表名, 求单元格代号(i+1, 标题列号2[设定值_下标2]+1), 本网关UUID)
				f.SetCellValue(设定工作表名, 求单元格代号(i+1, 标题列号2[错误信息_下标2]+1), 错误信息)
			}
		case "TOKEN服务器":
			设定个数1++
			设定值 = 微信报警推送需要的token获得来源检查(设定值)
			if 设定值 != "" {
				token服务器地址 = 设定值
			}
		case "使用者":
			设定个数1++
			使用者 = 默认使用者
			if 设定值 != "" {
				使用者 = 设定值
			}
		case "邮件接收者":
			设定个数1++
			设定值 = 修正邮件接收者(设定值)
			if 设定值 != "" {
				自定义邮件接收者 = 设定值
			}
		case "邮件发送人":
			设定个数1++
			设定值 = 修正邮件接收者(设定值)
			发送人们 := strings.Split(设定值, ",")
			if len(发送人们) > 1 {
				设定值 = 发送人们[0]
			}
			if 设定值 != "" {
				自定义发送人 = 设定值
			}
		case "发邮件TOKEN":
			设定个数1++
			设定值 = 发邮件token检查(设定值)
			if 设定值 != "" {
				自定义发邮件token = 设定值
			}
		case "企业微信信息推送机器人":
			设定个数1++
			设定值 = 企业微信信息推送机器人检查(设定值)
			if 设定值 != "" {
				webhook_url_默认 = 设定值
			}
		case "登录名":
			设定个数1++
			登录名 = 设定值
		case "登录密码":
			设定个数1++
			登录密码 = 设定值
		case "网关启动发邮件":
			设定个数1++
			允许网关启动发邮件 = true
			if 设定值 != "允许" {
				允许网关启动发邮件 = false
			}
		case "关闭网关发邮件":
			设定个数1++
			允许关闭网关发邮件 = true
			if 设定值 != "允许" {
				允许关闭网关发邮件 = false
			}
		case "重启网关发邮件":
			设定个数1++
			允许重启网关发邮件 = true
			if 设定值 != "允许" {
				允许重启网关发邮件 = false
			}
		case "重启服务器发邮件":
			设定个数1++
			允许重启服务器发邮件 = true
			if 设定值 != "允许" {
				允许重启服务器发邮件 = false
			}
		case "更新项目变量表发邮件":
			设定个数1++
			允许更新项目变量表发邮件 = true
			if 设定值 != "允许" {
				允许更新项目变量表发邮件 = false
			}
		case "更新网关软件发邮件":
			设定个数1++
			允许更新网关软件发邮件 = true
			if 设定值 != "允许" {
				允许更新网关软件发邮件 = false
			}
		case "执行自定义批处理发邮件":
			设定个数1++
			允许执行自定义批处理发邮件 = true
			if 设定值 != "允许" {
				允许执行自定义批处理发邮件 = false
			}
		case "访问IP更新发邮件":
			设定个数1++
			允许访问IP更新发邮件 = true
			if 设定值 != "允许" {
				允许访问IP更新发邮件 = false
			}
		default:
			continue
		} //switch 设定名称 {
	} //for _, row := range rows {
	if 设定个数1 != 设定个数 {
		错误信息 = "读设定工作表发现设定个数为" + strconv.Itoa(设定个数1) + ",而软件认可的个数为" +
			strconv.Itoa(设定个数) + ",于是软件认为此工作表非法！"
		启动软件碰到的问题.Set(错误信息)
		f.DeleteSheet(设定工作表名)
		return false
	}
	return true
} //func 读设定工作表(f) {
func 获得设定值() string {
	str1 := ""
	str := ""
	str = "本网关端口：" + 端口 + "\r\n"
	str1 += str
	str = "本网关UUID：" + 本网关UUID + "\r\n"
	str1 += str
	str = "微信小程序报警推送所需token获得服务器地址：" + token服务器地址 + "\r\n"
	str1 += str
	str = "使用者：" + 使用者 + "\r\n"
	str1 += str
	str = "邮件接收者：" + 自定义邮件接收者 + "\r\n"
	str1 += str
	str = "邮件发送人：" + 自定义发送人 + "\r\n"
	str1 += str
	str = "邮件发送授权码：" + 自定义发邮件token + "\r\n"
	str1 += str
	str = "企业微信信息推送机器人：" + webhook_url_默认 + "\r\n"
	str1 += str
	str = "登录名：" + 登录名 + "\r\n"
	str1 += str
	str = "登录密码：" + 登录密码 + "\r\n"
	str1 += str
	str = "允许网关启动发邮件：" + fmt.Sprintf("%t", 允许网关启动发邮件) + "\r\n"
	str1 += str
	str = "允许关闭网关发邮件：" + fmt.Sprintf("%t", 允许关闭网关发邮件) + "\r\n"
	str1 += str
	str = "允许重启网关发邮件：" + fmt.Sprintf("%t", 允许重启网关发邮件) + "\r\n"
	str1 += str
	str = "允许重启服务器发邮件：" + fmt.Sprintf("%t", 允许重启服务器发邮件) + "\r\n"
	str1 += str
	str = "允许更新项目变量表发邮件：" + fmt.Sprintf("%t", 允许更新项目变量表发邮件) + "\r\n"
	str1 += str
	str = "允许更新网关软件发邮件：" + fmt.Sprintf("%t", 允许更新网关软件发邮件) + "\r\n"
	str1 += str
	str = "允许执行自定义批处理发邮件：" + fmt.Sprintf("%t", 允许执行自定义批处理发邮件) + "\r\n"
	str1 += str
	str = "允许访问IP更新发邮件：" + fmt.Sprintf("%t", 允许访问IP更新发邮件) + "\r\n"
	str1 += str
	return str1
} //func 获得设定值(){
func 写默认数据到设定工作表(f *excelize.File) {
	总行数 := 19
	总列数 := 6
	f.NewSheet(设定工作表名)
	获取框线值(false, f)
	画框线(设定工作表名, 总行数, 总列数, f)
	//冻结前4列//冻结前1行//左上角单元格h2
	var 格式 excelize.Panes
	格式.Freeze = true
	格式.ActivePane = "\"bottomRight\",\"panes\":[{\"pane\":\"topLeft\"},{\"pane\":\"topRight\"},{\"pane\":\"bottomLeft\"},{\"active_cell\":\"g2\",\"sqref\":\"g2\",\"pane\":\"bottomRight\"}]"
	格式.TopLeftCell = "g2"
	格式.XSplit = 6
	格式.YSplit = 1
	f.SetPanes(工作表名, &格式)
	f.SetColWidth(设定工作表名, "a", "a", 5.14)
	f.SetColWidth(设定工作表名, "b", "b", 11.86)
	f.SetColWidth(设定工作表名, "c", "c", 7.29)
	f.SetColWidth(设定工作表名, "d", "d", 11.00)
	f.SetColWidth(设定工作表名, "e", "e", 39.57)
	f.SetColWidth(设定工作表名, "f", "f", 39.57)
	单元格 := "A1"
	写值 := "序号"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B1"
	写值 = "设定名称"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C1"
	写值 = "设定值"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D1"
	写值 = "范围"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "E1"
	写值 = "备注"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "F1"
	写值 = "错误信息"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A2"
	写值 = "1"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B2"
	写值 = "服务器端口"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C2"
	写值 = 默认端口号_s
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D2"
	写值 = 最小端口号_s + "~" + 最大端口号_s
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "E2"
	写值 = "建议先用端口扫描软件(比如cports.exe)扫描当前系统是否已经使用了将要设定的端口；本软件如果发现您设定的端口被占用，则本软件会杀掉占用设定端口的进程\r\n默认" + 默认端口号_s + "\r\n1433：SQL Server SQL Server的TCP 端口，用于供SQL Server对外提供服务。\r\n1434：SQL Server SQL Server的UDP端口，用于返回SQL Server使用了哪个 TCP/IP 端口。\r\n1521：Oracle通信端口，服务器上部署了Oracle SQL需要放行的端口。\r\n3306：MySQL数据库对外提供服务的端口。\r\n3389：远程桌面服务端口，可以通过这个端口远程连接服务器。\r\n8080：代理端口,同80端口一样，8080 端口常用于WWW代理服务，实现网页浏览。\r\n"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "f2"
	str := 检查端口是否被占用()
	启动软件碰到的问题.Set(str)
	写值 = str + "此条信息由软件生成\r\n"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A3"
	写值 = "2"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B3"
	写值 = "UUID"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C3"
	写值 = uuid.New()
	本网关UUID = 写值
	str = "本网关UUID由软件新生成\r\n" + 本网关UUID + "\r\n"
	启动软件碰到的问题.Set(str)
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D3"
	写值 = "标准的UUID格式为：xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (8-4-4-4-12)其中每个x是0-9或a-f范围内的一个十六进制的数字。"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "E3"
	写值 = "必须由软件生成的通用唯一识别码，作为网关身份识别，一旦生成，建议备忘不再修改，否则依赖此识别码对此网关操作的软件会工作异常，样例：8b6da95d-db24-850f-7a81-119a9fb3126d"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "f3"
	写值 = "此条信息由软件生成"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A4"
	写值 = "3"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B4"
	写值 = "token服务器"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C4"
	写值 = ""
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D4"
	写值 = ""
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "E4"
	写值 = "微信报警token获取地址，样例：http://192.168.16.70:8891/token?user=登录名之MD5&password=登录密码之MD5"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "f4"
	写值 = ""
	f.SetCellValue(设定工作表名, 单元格, 写值)
	端口 = 默认端口号_s
	单元格 = "A5"
	写值 = "4"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B5"
	写值 = "使用者"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C5"
	写值 = "www.pdlei.cn"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A6"
	写值 = "5"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B6"
	写值 = "邮件接收者"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C6"
	写值 = "dingjiazh1@pdlei.cn，dingjiazh1@pdlei.cn"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A7"
	写值 = "6"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B7"
	写值 = "邮件发送人"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C7"
	写值 = "158148415@qq.com"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A8"
	写值 = "7"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B8"
	写值 = "发邮件token"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C8"
	写值 = "aibbehoicjogbobd"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A9"
	写值 = "8"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B9"
	写值 = "企业微信信息推送机器人"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C9"
	写值 = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=890xy2c8-94d6-4aa9-820b-dwh157j95626"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A10"
	写值 = "9"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B10"
	写值 = "登录名"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C10"
	写值 = "pdlei"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A11"
	写值 = "10"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B11"
	写值 = "登录密码"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C11"
	写值 = "pdlei"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A12"
	写值 = "11"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B12"
	写值 = "网关启动发邮件"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C12"
	写值 = "允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D12"
	写值 = "允许,不允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A13"
	写值 = "12"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B13"
	写值 = "关闭网关发邮件"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C13"
	写值 = "允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D13"
	写值 = "允许,不允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A14"
	写值 = "13"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B14"
	写值 = "重启网关发邮件"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C14"
	写值 = "允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D14"
	写值 = "允许,不允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A15"
	写值 = "14"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B15"
	写值 = "重启服务器发邮件"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C15"
	写值 = "允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D15"
	写值 = "允许,不允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A16"
	写值 = "15"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B16"
	写值 = "更新项目变量表发邮件"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C16"
	写值 = "允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D16"
	写值 = "允许,不允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A17"
	写值 = "16"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B17"
	写值 = "更新网关软件发邮件"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C17"
	写值 = "允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D17"
	写值 = "允许,不允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A18"
	写值 = "17"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B18"
	写值 = "执行自定义批处理发邮件"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C18"
	写值 = "允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D18"
	写值 = "允许,不允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "A19"
	写值 = "18"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "B19"
	写值 = "访问IP更新发邮件"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "C19"
	写值 = "允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
	单元格 = "D19"
	写值 = "允许,不允许"
	f.SetCellValue(设定工作表名, 单元格, 写值)
} //func 写默认数据到设定工作表() {
func 写默认数据到常用设定值表(f *excelize.File) {
	总行数 := 47
	总列数 := 4
	f.NewSheet(常用设定值表名)
	获取框线值(false, f)
	画框线(常用设定值表名, 总行数, 总列数, f)
	//冻结前4列//冻结前1行//左上角单元格h2
	var 格式 excelize.Panes
	格式.Freeze = true
	格式.ActivePane = "\"bottomRight\",\"panes\":[{\"pane\":\"topLeft\"},{\"pane\":\"topRight\"},{\"pane\":\"bottomLeft\"},{\"active_cell\":\"e2\",\"sqref\":\"e2\",\"pane\":\"bottomRight\"}]"
	格式.TopLeftCell = "e2"
	格式.XSplit = 4
	格式.YSplit = 1
	f.SetPanes(工作表名, &格式)
	f.SetColWidth(常用设定值表名, "a", "a", 5.14)
	f.SetColWidth(常用设定值表名, "b", "b", 29.86)
	f.SetColWidth(常用设定值表名, "c", "c", 181.14)
	f.SetColWidth(常用设定值表名, "d", "d", 37.71)
	f.SetCellValue(常用设定值表名, "a1", "序号")
	f.SetCellValue(常用设定值表名, "b1", "类型")
	f.SetCellValue(常用设定值表名, "c1", "常用值")
	f.SetCellValue(常用设定值表名, "d1", "备注")
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(常用设定值表名, "a"+strconv.Itoa(i), strconv.Itoa(i-1))
	}
	for i := 6; i < 9; i++ {
		f.SetCellValue(常用设定值表名, "b"+strconv.Itoa(i), "奇偶校验")
	}
	for i := 13; i < 17; i++ {
		f.SetCellValue(常用设定值表名, "b"+strconv.Itoa(i), "读功能码")
	}
	for i := 17; i < 26; i++ {
		f.SetCellValue(常用设定值表名, "b"+strconv.Itoa(i), "数据类型")
	}
	for i := 30; i < 32; i++ {
		f.SetCellValue(常用设定值表名, "b"+strconv.Itoa(i), "寄存器地址")
	}
	for i := 26; i < 30; i++ {
		f.SetCellValue(常用设定值表名, "c"+strconv.Itoa(i), 最小时间_s+"~"+最大时间_s)
	}
	f.SetCellValue(常用设定值表名, "b2", "串口号")
	f.SetCellValue(常用设定值表名, "c2", 最小串口号_s+"~"+最大串口号_s)
	f.SetCellValue(常用设定值表名, "d2", "默认"+默认串口号)
	f.SetCellValue(常用设定值表名, "b3", "设备地址")
	f.SetCellValue(常用设定值表名, "c3", 最小设备地址_s+"~"+最大设备地址_s)
	f.SetCellValue(常用设定值表名, "d3", "0和255一般用于广播；默认"+默认设备地址_s)
	f.SetCellValue(常用设定值表名, "b4", "波特率")
	f.SetCellValue(常用设定值表名, "c4", strconv.Itoa(波特率01)+"、"+strconv.Itoa(波特率02)+"、"+strconv.Itoa(波特率03)+"、"+strconv.Itoa(波特率04)+"、"+strconv.Itoa(波特率05)+"、"+strconv.Itoa(波特率06)+"、"+strconv.Itoa(波特率07)+"、"+strconv.Itoa(波特率08)+"、"+strconv.Itoa(波特率09)+"、"+strconv.Itoa(波特率10)+"、"+strconv.Itoa(波特率11)+"、"+strconv.Itoa(波特率12)+"、"+strconv.Itoa(波特率13)+"、"+strconv.Itoa(波特率14)+"、"+strconv.Itoa(波特率15))
	f.SetCellValue(常用设定值表名, "d4", "默认"+默认波特率_s)
	f.SetCellValue(常用设定值表名, "b5", "数据位")
	f.SetCellValue(常用设定值表名, "c5", "5、6、7、8")
	f.SetCellValue(常用设定值表名, "d5", "默认"+默认数据位_s)
	//f.SetCellValue(常用设定值表名, "b6", "奇偶校验")
	f.SetCellValue(常用设定值表名, "c6", "偶校验(偶、EVEN、E)")
	//f.SetCellValue(常用设定值表名, "d6", "默认N")
	//f.SetCellValue(常用设定值表名, "b7", "奇偶校验")
	f.SetCellValue(常用设定值表名, "c7", "奇校验(奇、ODD、O)")
	//f.SetCellValue(常用设定值表名, "d7", "默认N")
	//f.SetCellValue(常用设定值表名, "b8", "奇偶校验")
	f.SetCellValue(常用设定值表名, "c8", "无校验(无、NONE、N)")
	f.SetCellValue(常用设定值表名, "d8", "默认"+默认奇偶校验)
	f.SetCellValue(常用设定值表名, "b9", "停止位")
	f.SetCellValue(常用设定值表名, "c9", "1、2")
	f.SetCellValue(常用设定值表名, "d9", "虚拟串口则很多硬件参数随意，默认"+默认停止位_s)
	f.SetCellValue(常用设定值表名, "b10", "字打包长度")
	f.SetCellValue(常用设定值表名, "c10", "1~"+最大字打包长度_s)
	f.SetCellValue(常用设定值表名, "d10", "默认"+默认字打包长度_s)
	f.SetCellValue(常用设定值表名, "b11", "位打包长度")
	f.SetCellValue(常用设定值表名, "c11", "1~"+最大位打包长度_s)
	f.SetCellValue(常用设定值表名, "d11", "默认"+默认位打包长度_s)
	f.SetCellValue(常用设定值表名, "b12", "读写字节序")
	f.SetCellValue(常用设定值表名, "c12", "ABCD、BADC、CDAB、DCBA")
	f.SetCellValue(常用设定值表名, "d12", "默认"+默认读写字节序)
	//f.SetCellValue(常用设定值表名, "b13", "读功能码")
	f.SetCellValue(常用设定值表名, "c13", "1")
	f.SetCellValue(常用设定值表名, "d13", "读线圈")
	//f.SetCellValue(常用设定值表名, "b14", "读功能码")
	f.SetCellValue(常用设定值表名, "c14", "2")
	f.SetCellValue(常用设定值表名, "d14", "读离散输入")
	//f.SetCellValue(常用设定值表名, "b15", "读功能码")
	f.SetCellValue(常用设定值表名, "c15", "3")
	f.SetCellValue(常用设定值表名, "d15", "读保持寄存器")
	//f.SetCellValue(常用设定值表名, "b16", "读功能码")
	f.SetCellValue(常用设定值表名, "c16", "4")
	f.SetCellValue(常用设定值表名, "d16", "读输入寄存器")
	//f.SetCellValue(常用设定值表名, "b17", "数据类型")
	f.SetCellValue(常用设定值表名, "c17", "LONG")
	f.SetCellValue(常用设定值表名, "d17", "32位有符号数("+int32最小值_s+"~"+int32最大值_s+")；内存变量是64位")
	//f.SetCellValue(常用设定值表名, "b18", "数据类型")
	f.SetCellValue(常用设定值表名, "c18", "ULONG")
	f.SetCellValue(常用设定值表名, "d18", "32位有无符号数(0~4294967295)")
	//f.SetCellValue(常用设定值表名, "b19", "数据类型")
	f.SetCellValue(常用设定值表名, "c19", "FLOAT")
	f.SetCellValue(常用设定值表名, "d19", "浮点数(-3.4e38~3.4e38)；内存变量是64位")
	//f.SetCellValue(常用设定值表名, "b20", "数据类型")
	f.SetCellValue(常用设定值表名, "c20", "SHORT")
	f.SetCellValue(常用设定值表名, "d20", "16位有符号数(-32768~32767)")
	//f.SetCellValue(常用设定值表名, "b21", "数据类型")
	f.SetCellValue(常用设定值表名, "c21", "USHORT")
	f.SetCellValue(常用设定值表名, "d21", "16位无符号数(0~65535)")
	//f.SetCellValue(常用设定值表名, "b22", "数据类型")
	f.SetCellValue(常用设定值表名, "c22", "BCD")
	f.SetCellValue(常用设定值表名, "d22", "0~9999")
	//f.SetCellValue(常用设定值表名, "b23", "数据类型")
	f.SetCellValue(常用设定值表名, "c23", "LONGBCD")
	f.SetCellValue(常用设定值表名, "d23", "0~99999999")
	f.SetCellValue(常用设定值表名, "b24", "变量类型")
	f.SetCellValue(常用设定值表名, "c24", "IO离散、内存离散")
	f.SetCellValue(常用设定值表名, "d24", "0、1")
	f.SetCellValue(常用设定值表名, "b25", "变量类型")
	f.SetCellValue(常用设定值表名, "c25", "IO字符串、内存字符串")
	f.SetCellValue(常用设定值表名, "d25", "IO字符串：1~"+最大读写字符个数_s+"个字节；"+"内存字符串：1~"+内存字符串最多占用字节数_s+"个字节，一个中文utf8占3个字节；")
	f.SetCellValue(常用设定值表名, "b26", "采集前等待(毫秒)")
	//f.SetCellValue(常用设定值表名, "c26", "0~60000")
	f.SetCellValue(常用设定值表名, "d26", "默认"+默认采集前等待_毫秒_s+"；设定的原因是有的设备需要总线上安静一会，它要喘口气才能正常回复")
	f.SetCellValue(常用设定值表名, "b27", "采集频率(毫秒)")
	f.SetCellValue(常用设定值表名, "c27", "0~int64Max")
	f.SetCellValue(常用设定值表名, "d27", "默认"+默认采集频率_毫秒_s)
	f.SetCellValue(常用设定值表名, "b28", "通讯超时（毫秒）")
	//f.SetCellValue(常用设定值表名, "c28", "0~60000")
	f.SetCellValue(常用设定值表名, "d28", "默认"+默认通讯超时_毫秒_s+"；要调试确认最佳超时时间，不能凭感觉，一般如果打包越大，设备需要更多的时间计算后才能回复，特别是回复几千个位数据时最费时（读功能码1或2），要注意设定超时时间和采集频率，否则设备疲于奔命，无暇于正常的工作")
	f.SetCellValue(常用设定值表名, "b29", "记录到数据库间隔（秒）")
	//f.SetCellValue(常用设定值表名, "c29", "0~60000")
	f.SetCellValue(常用设定值表名, "d29", "默认"+默认记录到数据库间隔_秒_s+"；此设定不能留空也必须位于表格最右列，它是软件获取表格数据的边界标识；如果设为0则忽略间隔记录。间隔记录的目的是如果变量的值一直没有变化，夸张一点是几年都没变化，那么数据库中此变量的记录几年内都没有，那是啥意思？让人困惑")
	//f.SetCellValue(常用设定值表名, "b30", "寄存器地址")
	f.SetCellValue(常用设定值表名, "c30", 最小寄存器地址_s+"~"+最大寄存器地址_s)
	f.SetCellValue(常用设定值表名, "d30", "USHORT、SHORT、BCD")
	//f.SetCellValue(常用设定值表名, "b31", "寄存器地址")
	f.SetCellValue(常用设定值表名, "c31", 最小寄存器地址_s+"~"+strconv.Itoa(最大寄存器地址-1))
	f.SetCellValue(常用设定值表名, "d31", "FLOAT、LONG、LONGBCD、ULONG")
	f.SetCellValue(常用设定值表名, "b32", "寄存器地址m.n")
	f.SetCellValue(常用设定值表名, "c32", "m:"+最小寄存器地址_s+"~"+最大寄存器地址_s+"；n:1~"+最大读写字符个数_s+"；并且m+(n+1)/2 -1<="+最大寄存器地址_s)
	f.SetCellValue(常用设定值表名, "d32", "a) 二级通道表示  m.n\r\nm表示读/写字符串的起始地址；n表示字符串的字节长度。\r\n字节长度说明：字符串字节长度为N，由于Modbus Holding Register只能以双字节为单位写，所以实际发帧的时候共发N+1个字节（N为奇数）或N个字节（N为偶数）。\r\n即若写入N字节长度的字符串，实际在PLC中写入N+1或N个字节，即肯定是偶数字节数。\r\nn=1或2 时：\r\n    读取的字符串为地址m的当前UTF8编码。\r\n	n=N（N>2）时：\r\n    读取字节长度为N的字符串，从地址m开始。\r\n	b) 采集数据结果集支持可输入的UTF8编码。\r\nd）当写入字符串的字节长度小于n时，其他内存全部填充'\\0'。n为奇数时,实际在内存中写入n+1字节。")
	f.SetCellValue(常用设定值表名, "b33", "寄存器地址aa.bb")
	f.SetCellValue(常用设定值表名, "c33", "aa:"+最小寄存器地址_s+"~"+最大寄存器地址_s+"；bb:0~15")
	f.SetCellValue(常用设定值表名, "d33", "BIT")
	f.SetCellValue(常用设定值表名, "b34", "变量名")
	f.SetCellValue(常用设定值表名, "c34", "项目代号+“_”+中文、字母、数字、“_”的组合；项目代号为首字母+两个字母、数字组合；变量名区分大小写；总字符长度不超过31；变量名不可重复；所有的设定值中的空格会被软件去除；")
	f.SetCellValue(常用设定值表名, "d34", "比如:A01_1;aaa_a;A0a_灯；Aab_灯_1；")
	f.SetCellValue(常用设定值表名, "b35", "变量类型")
	f.SetCellValue(常用设定值表名, "c35", "IO离散、IO字符串、IO整型、IO实型；内存离散、内存字符串、内存整型、内存实型；")
	//f.SetCellValue(常用设定值表名, "d35", "")
	f.SetCellValue(常用设定值表名, "b36", "是否保存值")
	f.SetCellValue(常用设定值表名, "c36", "是、否；若设为是，那么变量的当前值被保存到变量初始值中，下次重启软件将加载此值，仅支持内存变量；设备数据在设备中，只要采集成功就有，所以不用保存到硬盘中")
	//f.SetCellValue(常用设定值表名, "d36", "")
	f.SetCellValue(常用设定值表名, "b37", "允许通讯异常后只读变量")
	f.SetCellValue(常用设定值表名, "c37", "允许、不允许；若设定为允许，那么通讯被判定异常后再次采集此变量数据不再打包，只单独采集此变量数据，采集数据量少，提高采集可靠性；有些设备需要批量采集才回应的，就设为不允许")
	//f.SetCellValue(常用设定值表名, "d37", "")
	f.SetCellValue(常用设定值表名, "b38", "读写属性")
	f.SetCellValue(常用设定值表名, "c38", "读写、只读、只写；若设为只写，软件不采集此变量数据或者打包采集后不解析此变量数据")
	//f.SetCellValue(常用设定值表名, "d38", "")
	f.SetCellValue(常用设定值表名, "b39", "初始值；字符串(<"+最大读写字符个数_s+"字节)")
	f.SetCellValue(常用设定值表名, "c39", "系统启动后采集前变量要装入的初始值")
	//f.SetCellValue(常用设定值表名, "d39", "")
	f.SetCellValue(常用设定值表名, "b40", "计算值除原始值")
	f.SetCellValue(常用设定值表名, "c40", "对于设备寄存器数据类型是LONG、ULONG、SHORT、USHORT、LONGBCD、BCD的数据，但变量类型是IO实型的变量，需要转换用到的系数值。比如采集回来的温度数据是387（原始值），需要除以10才获得实际值38.7（计算值），那么这个系数值就是0.1")
	//f.SetCellValue(常用设定值表名, "d40", "")
	f.SetCellValue(常用设定值表名, "b41", "小数位数")
	f.SetCellValue(常用设定值表名, "c41", "0~255")
	//f.SetCellValue(常用设定值表名, "d41", "")
	f.SetCellValue(常用设定值表名, "b42", "通讯异常值")
	f.SetCellValue(常用设定值表名, "c42", "使用变量非正常值作为通讯异常的表达，比如变量类型是IO离散，-1、-2等都是它的非正常值，如果选-1做通讯异常值，当依据设定条件判断通讯异常后软件将置此变量为-1，那么获取此变量值的软件就知道通讯异常，做出正确反应")
	//f.SetCellValue(常用设定值表名, "d42", "")
	f.SetCellValue(常用设定值表名, "b43", "连续通讯失败多少次则认为通讯异常")
	f.SetCellValue(常用设定值表名, "c43", "1~65535；假设设为3，那么连续通讯失败3次则认为通讯异常，软件将置变量值为设定的通讯异常值")
	//f.SetCellValue(常用设定值表名, "d43", "")
	f.SetCellValue(常用设定值表名, "b44", "多少变化率才记录到数据库")
	f.SetCellValue(常用设定值表名, "c44", "变量多少变化率才记录到数据库，如果设为0则不记录，如果设大于0，比如0.1，那么变量的值只要变化率大于等于0.1就记录到数据库，这样可以保证记录到数据库不会那么频繁")
	//f.SetCellValue(常用设定值表名, "d44", "")
	f.SetCellValue(常用设定值表名, "b45", "错误信息")
	f.SetCellValue(常用设定值表名, "c45", "软件分析变量表的错误信息，若有错误信息，软件忽略此变量数据采集")
	//f.SetCellValue(常用设定值表名, "d45", "")
	f.SetCellValue(常用设定值表名, "b46", "警告信息")
	f.SetCellValue(常用设定值表名, "c46", "软件分析变量表的警告信息，若有警告信息，表示这些信息的相关设定项被软件修改为默认值，此变量数据依然被采集")
	//f.SetCellValue(常用设定值表名, "d46", "")
	f.SetCellValue(常用设定值表名, "b47", "单位")
	f.SetCellValue(常用设定值表名, "c47", "比如温度单位℃，湿度%，电压V，电流A，功率KW,长度M,电度KWH,流量M3/H等")
	//f.SetCellValue(常用设定值表名, "d47", "")
}
func 写默认数据到开发说明表(f *excelize.File) {
	总行数 := 2
	总列数 := 2
	f.NewSheet(开发说明表名)
	获取框线值(false, f)
	画框线(开发说明表名, 总行数, 总列数, f)
	//冻结前4列//冻结前1行//左上角单元格h2
	var 格式 excelize.Panes
	格式.Freeze = true
	格式.ActivePane = "\"bottomRight\",\"panes\":[{\"pane\":\"topLeft\"},{\"pane\":\"topRight\"},{\"pane\":\"bottomLeft\"},{\"active_cell\":\"h2\",\"sqref\":\"h2\",\"pane\":\"bottomRight\"}]"
	格式.TopLeftCell = "a1"
	格式.XSplit = 1
	格式.YSplit = 1
	f.SetPanes(工作表名, &格式)
	f.SetColWidth(开发说明表名, "a", "a", 35.14)
	f.SetColWidth(开发说明表名, "b", "b", 35.14)
	f.SetCellValue(开发说明表名, "a1", "序号")
	f.SetCellValue(开发说明表名, "b1", "api样例")
	str := 生成api访问样例("127.0.0.1", 默认端口号_s, "本网关登录名之MD5", "本网关登录密码之MD5")
	str = "用上网浏览器登录127.0.0.1:" + 默认端口号_s + "/统计，输入用户名和密码登录，页面会给更加具体的使用说明，若用户名和密码还没在表格中设定，那么默认用户名和密码是空或pdlei\r\n" + str
	f.SetCellValue(开发说明表名, "b2", str)
	for i := 2; i < 总行数+1; i++ {
		f.SetCellValue(开发说明表名, "a"+strconv.Itoa(i), strconv.Itoa(i-1))
	}
} //func 写默认数据到开发说明表() {
func 创建快捷方式(文件名及路径, 快捷方式名及路径 string) {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		启动软件碰到的问题.Set("创建快捷方式(" + 快捷方式名及路径 + ")发生了错误(" + err.Error() + ")")
		return
	}
	defer oleShellObject.Release()
	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		启动软件碰到的问题.Set("oleShellObject.QueryInterface(ole.IID_IDispatch)发生了错误(" + err.Error() + ")")
		return
	}
	defer wshell.Release()
	cs, err := oleutil.CallMethod(wshell, "CreateShortcut", 快捷方式名及路径)
	if err != nil {
		启动软件碰到的问题.Set("oleutil.CallMethod(wshell, \"CreateShortcut\", 快捷方式名及路径)发生了错误(" +
			err.Error() + ")")
		return
	}
	idispatch := cs.ToIDispatch()
	oleutil.PutProperty(idispatch, "TargetPath", 文件名及路径)
	oleutil.CallMethod(idispatch, "Save")
} //func 创建快捷方式(文件名及路径, 快捷方式名及路径 string) {
var 当前用户名 string

func 获取当前用户名() string {
	currentUser, err := user.Current()
	if err != nil {
		return "无法获取当前用户名(" + err.Error() + ")"
	}
	username := currentUser.Username
	if strings.Contains(username, "\\") {
		strings.Index(username, "\\")
		username = username[strings.Index(username, "\\")+1:]
	}
	return username
}
func 判断文件是否存在(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}
func 启动软件检查表格文件() {
	if 存在, _ := 判断目录是否存在(目录); !存在 {
		启动软件碰到的问题.Set("目录(" + 目录 + ")不存在，将被创建！同时创建(" + 项目变量信息表名 + ")")
		fmt.Println("目录(" + 目录 + ")不存在，将被创建！同时创建(" + 项目变量信息表名 + ")")
		端口 = 默认端口号_s
		创建目录(目录)
		创建工作簿(项目变量信息表名)
	} else { //if 存在, _ := 判断目录是否存在(目录); 存在 != true {
		if !判断文件是否存在(项目变量信息表名) {
			err := 复制文件(项目变量信息表名备份及路径, 项目变量信息表名)
			if err != nil {
				fmt.Println("项目变量信息表不存在，软件尝试恢复备份文件失败，软件将创建默认表格样例供您参考修改使用！:", err)
				邮件主题 := "恢复项目变量信息表错误," + 发邮件主题.Load()
				邮件正文 := 邮件主题 + "\r\n" + err.Error() +
					"\r\n项目变量信息表不存在，软件尝试恢复备份文件失败，软件将创建默认表格样例供您参考修改使用！\r\n" +
					"服务器信息：\r\n" + 获取服务器信息() + "\r\n" + 获取服务器内存使用率() + "\r\n"
				go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
			} else {
				fmt.Println("项目变量信息表不存在,恢复" + 项目变量信息表名备份及路径 + "到" + 项目变量信息表名 + "成功!")
				// 批处理文件名非中文提示 := "HuiFuXiangMuBianLiangXinXiBiao"
				// go 重新启动外部程序(批处理文件名非中文提示, 程序名, 程序文件名及路径)
				return
			}
			端口 = 默认端口号_s
			启动软件碰到的问题.Set("项目变量信息表(" + 项目变量信息表名 + ")不存在，将被创建！")
			创建工作簿(项目变量信息表名)
		} else { //if !判断文件是否存在(项目变量信息表名) {
			f, err1 := excelize.OpenFile(项目变量信息表名)
			if err1 != nil {
				启动软件碰到的问题.Set(err1.Error())
				err := 复制文件(项目变量信息表名备份及路径, 项目变量信息表名)
				if err != nil {
					fmt.Println("项目变量信息表错误（" + err1.Error() + "），软件尝试恢复备份文件失败(" + err.Error() + ")，需人工介入！")
					邮件主题 := "恢复项目变量信息表错误," + 发邮件主题.Load()
					邮件正文 := 邮件主题 + "\r\n" + "\r\n项目变量信息表错误（" + err1.Error() + "），软件尝试恢复备份文件失败(" + err.Error() + ")，需人工介入！\r\n" +
						"服务器信息：\r\n" + 获取服务器信息() + "\r\n" + 获取服务器内存使用率() + "\r\n"
					go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
				} else {
					fmt.Println("项目变量信息表错误（" + err1.Error() + "），恢复" + 项目变量信息表名备份及路径 + "到" + 项目变量信息表名 + "成功!")
				}
				return
			}
			defer func() {
				if err2 := f.Close(); err2 != nil {
					启动软件碰到的问题.Set(err2.Error())
					return
				}
			}()
			index, _ := f.NewSheet(工作表名)
			rows, err := f.GetRows(工作表名)
			if err != nil {
				启动软件碰到的问题.Set(err.Error())
				return
			}
			表格行数 := len(rows)
			if 表格行数 < 2 {
				启动软件碰到的问题.Set("工作表(" + 工作表名 +
					")表格行数 < 2或者此工作表不存在，软件将建立一个表格样本供您参考使用！")
				写默认数据到变量工作表(f)
			}
			f.NewSheet(设定工作表名)
			rows, err = f.GetRows(设定工作表名)
			if err != nil {
				启动软件碰到的问题.Set(err.Error())
				端口 = 默认端口号_s
			}
			表格行数 = len(rows)
			if 表格行数 < 2 {
				启动软件碰到的问题.Set("工作表(" + 设定工作表名 +
					")表格行数 < 2或者此工作表不存在，软件将建立一个表格样本供您参考使用！")
				写默认数据到设定工作表(f)
			} else {
				if !读设定工作表(f) {
					启动软件碰到的问题.Set("读设定工作表发现错误，软件将建立一个表格样本供您参考使用！")
					写默认数据到设定工作表(f)
				}
			}
			f.DeleteSheet(常用设定值表名)
			写默认数据到常用设定值表(f)
			f.DeleteSheet(开发说明表名)
			写默认数据到开发说明表(f)
			//	f.UpdateLinkedValue() //更新表格中公式等引用
			f.SetActiveSheet(index)
			if err2 := f.Save(); err != nil {
				启动软件碰到的问题.Set(err2.Error())
			}
			time.Sleep(time.Second * 1)
		} //} else {//if !判断文件是否存在(项目变量信息表名) {
	} //} else {//if 存在, _ := 判断目录是否存在(目录); 存在 != true {
	快捷方式文件名及路径 := "C:\\Users\\" + 当前用户名 + "\\Desktop\\" + 项目变量信息表名1 + ".lnk"
	创建快捷方式(项目变量信息表名, 快捷方式文件名及路径)
	启动软件碰到的问题.Set("成功生成项目变量信息表(" + 项目变量信息表名 + ")的快捷方式(" + 快捷方式文件名及路径 +
		")放到用户桌面！")
} //func 启动软件检查表格文件(){
var 登录时刻 int64
var 登录时间 string
var 服务器域名 string
var LAN string
var login被访问次数 uint64

const 登录成功后如何操作 = "返回后刷新\r\n快捷键：按组合键Alt+<(左方向键)，然后按F5\r\n鼠标：点击浏览器左上角⬅，然后再点击🔄"

func login(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &login被访问次数, "Login")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	switch r.Method {
	case "GET":
		提供登录页面(w, r)
	default:
		r.ParseForm()
		if r.Form["username"][0] != 登录名 {
			fmt.Fprint(w, "用户名或密码错误！")
			str := "\r\n您的连接: " + r.RemoteAddr
			fmt.Fprint(w, str) //必须用这个函数，否则网页没数据
			return
		}
		if r.Form["password"][0] != 登录密码 {
			fmt.Fprint(w, "用户名或密码错误！")
			str := "\r\n您的连接: " + r.RemoteAddr
			fmt.Fprint(w, str)
			return
		}
		登录时间 = time.Now().Format("2006-01-02 15:04:05")
		if time.Now().Unix()-连接登录时刻[r.RemoteAddr] > 登录保持秒数 {
			登录时刻 = time.Now().Unix()
			连接登录时刻[r.RemoteAddr] = 登录时刻
			str := "登录成功！" + "\r\n"
			fmt.Fprint(w, str)
			str = "登录时间：" + 登录时间 + "\r\n"
			fmt.Fprint(w, str)
			str = "请在" + 登录保持秒数_s + "秒内完成所需操作！" + "\r\n"
			fmt.Fprint(w, str)
			str = 登录成功后如何操作
			fmt.Fprint(w, str)
		} else {
			登录时刻 = 0
			连接登录时刻[r.RemoteAddr] = 登录时刻
			str := "退出登录成功！" + "\r\n"
			fmt.Fprint(w, str)
			str = "退出登录时间：" + 登录时间
			fmt.Fprint(w, str)
		}
	} //switch r.Method{
	str := "\r\n您的连接: " + r.RemoteAddr
	fmt.Fprint(w, str)
} //func login(w http.ResponseWriter, r *http.Request) {
var 上传被访问次数 uint64
var 上次上传时刻 int64
var filesMax int64 = 15 << 20
var valuesMax int64 = 512

func upload(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &上传被访问次数, "Upload")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	if time.Now().Unix()-atomic.LoadInt64(&上次上传时刻) < 上传时间间隔秒数 {
		atomic.StoreInt64(&上次上传时刻, time.Now().Unix())
		w.Write([]byte("距离上次上传时间小于上传时间间隔秒数设定，本次触发无效！如果有多个上传源频繁上传，那么必须在上传时间间隔秒数内没有上传，之后先登录成功然后上传才有效！"))
		return
	}
	atomic.StoreInt64(&上次上传时刻, time.Now().Unix())
	if time.Now().Unix()-连接登录时刻[r.RemoteAddr] > 登录保持秒数 {
		提供登录页面(w, r)
		return
	}
	fmt.Fprint(w, "上传请求长度bytes:")
	fmt.Fprint(w, strconv.Itoa(int(r.ContentLength))+"\r\n")
	// 检查请求方法是否为POST
	filesMax = 程序文件最大限制 + valuesMax
	if r.ContentLength > filesMax {
		fmt.Fprint(w, "上传文件非法，上传失败！")
		return
	}
	if err := r.ParseMultipartForm(filesMax); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	file, _, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Fprint(w, "上传数据中没有文件，本次上传失败！失败描述： "+err.Error())
		return
	}
	defer file.Close()
	更新文件(w, r)
} //func upload(w http.ResponseWriter, r *http.Request) {
var filesMin int64 = 1 << 0

func 更新文件(w http.ResponseWriter, r *http.Request) {
	var 保存命名 string = ""
	var 保存目录 string
	var 文件大小 string
	上传的文件内容, handler, _ := r.FormFile("uploadfile")
	defer 上传的文件内容.Close()
	文件大小 = strconv.Itoa(int(handler.Size))
	可上传的文件名们 := []string{
		项目变量信息表名1,
		程序名,
		可上传在服务器运行的批处理文件名, //根据需要添加，添加后，在本函数内记得其他地方也要修改
	}
	for _, 可上传的文件名 := range 可上传的文件名们 {
		if strings.Contains(handler.Filename, 可上传的文件名) {
			保存命名 = 可上传的文件名
			break
		}
	}
	if 保存命名 == "" {
		fmt.Fprint(w, handler.Filename)
		fmt.Fprint(w, "\r\n此文件禁止上传！")
		return
	}
	保存目录 = 目录
	switch 保存命名 {
	case 项目变量信息表名1:
		filesMax = 项目变量信息表文件最大限制
		filesMin = 项目变量信息表文件最小限制
	case 程序名:
		filesMax = 程序文件最大限制
		filesMin = 程序文件最小限制
		保存目录 = 程序文件名及路径[:strings.LastIndex(程序文件名及路径, "\\")+1]
	case 可上传在服务器运行的批处理文件名:
		filesMax = 可上传在服务器运行的批处理文件最大限制
		filesMin = 可上传在服务器运行的批处理文件最小限制
	}
	if handler.Size > filesMax || handler.Size < filesMin {
		str := handler.Filename + " 大小为：" + 文件大小 + " byts"
		fmt.Fprint(w, str)
		fmt.Fprint(w, "上传文件非法，上传失败！")
		return
	}
	前缀 := time.Now().Format("2006-01-02 15:04:05")
	前缀 = strings.Replace(前缀, " ", "_", -1)
	前缀 = strings.Replace(前缀, ":", "_", -1)
	前缀 = "T" + 前缀 + "_"
	保存命名 = 前缀 + 保存命名
	f, err := os.OpenFile(保存目录+保存命名, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Fprint(w, handler.Filename)
		fmt.Fprint(w, "\r\n上传文件保存在服务器中文件名及路径：\r\n"+保存目录+保存命名+"\r\n估计正在打开编辑中，本次上传失败！失败描述： "+err.Error())
		return
	}
	defer f.Close()
	io.Copy(f, 上传的文件内容)
	str := handler.Filename + " 大小为：" + 文件大小 + " byts," + "上传成功"
	fmt.Fprint(w, str)
	fmt.Fprint(w, "\r\n上传文件保存在服务器中文件名及路径：\r\n"+保存目录+保存命名)
	str = "\r\n您的连接: " + r.RemoteAddr + "\r\n"
	fmt.Fprint(w, str)
	switch 保存命名 {
	case 前缀 + 项目变量信息表名1:
		时间 := strings.Replace(time.Now().Format("2006-01-02 15:04:05"), ":", "", -1)
		文件名 := strings.Replace(strings.Replace(r.RemoteAddr, ":", "_", -1), ".", "_", -1) + "_更新项目变量信息表_" + 时间 + ".bat"
		内容 := "set SLEEP=ping 127.0.0.1 /n" + "\r\n" + "%SLEEP% 4 > nul" + "\r\n" +
			"copy /y " + 保存目录 + 保存命名 + " " + 保存目录 + 项目变量信息表名1 + "\r\n" +
			"del /f /q " + 保存目录 + 保存命名 + "\r\n" + "exit\r\n"
		data1, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(内容))
		if err != nil {
			fmt.Fprint(w, "simplifiedchinese.GBK.NewEncoder().Bytes([]byte("+内容+")发生错误！")
			return
		}
		创建文件2(目录+"bat/", 文件名, data1)
		邮件主题 := r.RemoteAddr + "要更新项目变量信息表," + 发邮件主题.Load()
		邮件正文 := 邮件主题 + "\r\n" + 内容
		go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
		c := exec.Command(目录 + "bat/" + 文件名)
		err = c.Start()
		fmt.Fprint(w, 内容)
		if err != nil {
			fmt.Fprint(w, "\r\n更新项目变量信息表失败！失败描述： "+err.Error())
			return
		}
		fmt.Fprint(w, "\r\n更新项目变量信息表成功！")
		if 允许更新项目变量表发邮件 {
			atomic.StoreUint32(&需要发邮件, 1)
			触发者信息 = r.RemoteAddr
			邮件正文前附加 = "更新项目变量信息表"
			邮件主题前附加 = "更新项目变量信息表"
			i := 0
			for {
				time.Sleep(time.Second * 1)
				if atomic.LoadUint32(&需要发邮件) == 0 {
					break
				}
				if i > 自动发邮件时间间隔秒数 {
					break
				}
				i++
			}
		}
		重启软件(w, r)
	case 前缀 + 程序名:
		时间 := strings.Replace(time.Now().Format("2006-01-02 15:04:05"), ":", "", -1)
		文件名 := strings.Replace(strings.Replace(r.RemoteAddr, ":", "_", -1), ".", "_", -1) + "_" + 程序名 + "_更新程序_" + 时间 + ".bat"
		内容 := "set SLEEP=ping 127.0.0.1 /n" + "\r\n" + "%SLEEP% 4 > nul" + "\r\n" + "taskkill.exe /f /im " + 程序名 + "\r\n" +
			"%SLEEP% 4 > nul" + "\r\n" + "del /f /q " + 程序文件名及路径 + "\r\n" + "copy /y " + 保存目录 + 保存命名 + " " + 程序文件名及路径 + "\r\n" +
			"copy /y " + 保存目录 + 保存命名 + " " + 程序文件名及路径2 + "\r\n" +
			"del /f /q " + 保存目录 + 保存命名 + "\r\n" + "start " + 程序文件名及路径 + "\r\n" + "exit\r\n"
		data1, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(内容))
		if err != nil {
			fmt.Fprint(w, "simplifiedchinese.GBK.NewEncoder().Bytes([]byte("+内容+")发生错误！")
			return
		}
		创建文件2(目录+"bat/", 文件名, data1)
		邮件主题 := r.RemoteAddr + "要更新软件," + 发邮件主题.Load()
		邮件正文 := 邮件主题 + "\r\n" + 内容
		go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
		c := exec.Command(目录 + "bat/" + 文件名)
		写项目变量信息表并保存()
		atomic.StoreUint32(&已经触发重启程序了, 1)
		SQLiteDB访问锁.Lock()
		defer SQLiteDB访问锁.Unlock()
		写项目变量信息表并保存锁.Lock()
		defer 写项目变量信息表并保存锁.Unlock()
		if SQLiteDB内存数据库连接 != nil {
			SQLiteDB内存数据库连接.Close()
		}
		if SQLiteDB磁盘查询历史数据库连接 != nil {
			SQLiteDB磁盘查询历史数据库连接.Close()
		}
		err = c.Start()
		fmt.Fprint(w, 内容)
		if err != nil {
			fmt.Fprint(w, "\r\n执行升级失败！失败描述： "+err.Error())
			return
		}
		fmt.Fprint(w, "\r\n执行升级成功！")
	case 前缀 + 可上传在服务器运行的批处理文件名:
		文件名 := "Do_" + 可上传在服务器运行的批处理文件名
		内容 := "start " + 保存目录 + 保存命名 + "\r\n" + "exit\r\n"
		data1, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(内容))
		if err != nil {
			fmt.Fprint(w, "simplifiedchinese.GBK.NewEncoder().Bytes([]byte("+内容+")发生错误！")
			return
		}
		创建文件2(目录+"bat/", 文件名, data1)
		if 允许执行自定义批处理发邮件 {
			触发者信息 = r.RemoteAddr
			文件内容 := ""
			上传的文件内容.Seek(0, io.SeekStart)
			databuf, err := io.ReadAll(上传的文件内容)
			if err == nil {
				文件内容 = string(databuf)
			} else {
				文件内容 = "读取文件错误：" + err.Error()
			}
			//fmt.Println("批处理文件内容", 文件内容)
			邮件正文前附加 = "执行了自定义批处理,批处理的内容为：\r\n" + 文件内容
			邮件正文前附加 = 邮件正文前附加 + "\r\n上传文件保存在服务器中文件名及路径：" + 保存目录 + 保存命名
			邮件主题前附加 = "执行了自定义批处理"
			//fmt.Println("邮件正文前附加", 邮件正文前附加)
			atomic.StoreUint32(&需要发邮件, 1)
			i := 0
			for {
				time.Sleep(time.Second * 1)
				if atomic.LoadUint32(&需要发邮件) == 0 {
					break
				}
				if i > 自动发邮件时间间隔秒数 {
					break
				}
				i++
			}
		}
		c := exec.Command(目录 + "bat/" + 文件名)
		写项目变量信息表并保存()
		err = c.Start()
		fmt.Fprint(w, 内容)
		if err != nil {
			fmt.Fprint(w, "\r\n执行此批处理失败!失败描述： "+err.Error())
			return
		}
		fmt.Fprint(w, "\r\n执行此批处理成功!")
	}
} //func 更新文件(w http.ResponseWriter, r *http.Request) {
const (
	//loginhtml1 = "<!DOCTYPE html><html lang=\"en\"><head>	<meta charset=\"UTF-8\"><title>Title</title></head><body><form action=\"http://"
	loginhtml2 = "/login\" method=\"post\"><input type=\"password\" name=\"username\"><input type=\"password\" name=\"password\"><input type=\"submit\" value=\"登录\"></form>\r\n上传文件<form action=\"http://"
	loginhtml3 = "/upload\" method=\"post\" enctype=\"multipart/form-data\"><input type=\"file\" name=\"uploadfile\"/><input type=\"submit\" value=\"上传\"></form></body></html>"
)

var loginhtml1 = "<!DOCTYPE html><html lang=\"en\"><head>	<meta charset=\"UTF-8\"><title>" + 程序文件名及路径 + "</title></head><body><form action=\"http://"

func 创建文件2(目录, 文件名 string, 内容 []byte) {
	创建目录(目录)
	f, err := os.Create(目录 + 文件名)
	if err != nil {
		启动软件碰到的问题.Set("创建文件2(" + 目录 + 文件名 + ")发生了错误(" + err.Error() + ")")
		return
	} else {
		f.Write(内容)
	}
	defer f.Close()
}
func 创建文件(目录, 文件名, 内容 string) {
	创建目录(目录)
	f, err := os.Create(目录 + 文件名)
	if err != nil {
		启动软件碰到的问题.Set("创建文件(" + 目录 + 文件名 + ")发生了错误(" + err.Error() + ")")
		return
	} else {
		f.Write([]byte(内容))
	}
	defer f.Close()
}

func 获取本地上网ip() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return ""
	}
	defer conn.Close()
	localaddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localaddr, ":")
	return localaddr[:idx]
}

var 本机IP地址为 = ""
var 外网访问主机获得的本机公网ip及端口号 = ""
var 被访问获得的本机ipv6及端口号 = ""
var 服务器访问连接有变化 uint32 = 0
var 上次检查上网状态时刻 int64 = 0
var 自获取的外网ip 字符串互斥锁访问结构体

func 检查上网状态() {
	ticker := time.NewTicker(检查上网状态时间间隔秒数 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if atomic.LoadUint32(&已经触发重启程序了) == 1 {
			fmt.Println("已经触发重启程序了")
			continue
		}
		if time.Now().Unix()-上次检查上网状态时刻 < 检查上网状态时间间隔秒数 {
			continue
		}
		上次检查上网状态时刻 = time.Now().Unix()
		能上网吗()
		if atomic.LoadUint32(&能上网) == 0 {
			fmt.Println("不能上网 " + time.Now().Format("2006-01-02 15:04:05"))
			continue
		}
		fmt.Println("能上网 " + time.Now().Format("2006-01-02 15:04:05"))
		上网ip := 获取本地上网ip()
		if 上网ip != "" && 上网ip != 本机IP地址为 {
			fmt.Println("本机IP地址为：", 上网ip)
			本机IP地址为 = 上网ip
			if LAN != 上网ip+":"+端口 {
				LAN = 上网ip + ":" + 端口
				atomic.StoreUint32(&服务器访问连接有变化, 1)
			}
		}
		ip := Get外网IP()
		if ip != "" && ip != 自获取的外网ip.Load() {
			自获取的外网ip.Set(ip)
			atomic.StoreUint32(&服务器访问连接有变化, 1)
		}
		if atomic.LoadUint32(&服务器访问连接有变化) == 1 {
			atomic.StoreUint32(&服务器访问连接有变化, 0)
			str := LAN + "," + 服务器域名 + "," + ip + ":" + 端口 + "," +
				外网访问主机获得的本机公网ip及端口号 + "," + 被访问获得的本机ipv6及端口号 + ";"
			服务器访问连接.Set(str)
			发邮件主题2 := str + "_" + 当前用户名 + "_" + 程序名
			发邮件主题.Set(发邮件主题2)
			fmt.Println("发邮件主题：", 发邮件主题2)
			atomic.StoreInt64(&数据网关状态微信报警推送上次成功时刻秒, 0)
			邮件主题 := "服务器访问连接更新了" + 发邮件主题2
			邮件正文 := 邮件主题 + "\r\n登录名：" + 登录名 + "\r\n登录密码：" + 登录密码
			go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
			go 报警发邮件函数(自定义发送人, 自定义发邮件token, 默认邮件接收者, 邮件主题, 邮件正文)
		}
	} //for range ticker.C {
} //func 检查上网状态() {
var 获取服务器域名执行锁 sync.Mutex

func 获取服务器域名(主机名 string, 远程机器名 string) {
	获取服务器域名执行锁.Lock()
	defer 获取服务器域名执行锁.Unlock()
	switch 判断r_host类型(主机名) {
	case "公网ip":
		if 外网访问主机获得的本机公网ip及端口号 != 主机名 {
			外网访问主机获得的本机公网ip及端口号 = 主机名
			atomic.StoreUint32(&服务器访问连接有变化, 1)
			atomic.StoreInt64(&数据网关状态微信报警推送上次成功时刻秒, 0)
			邮件主题 := "获得公网ip（" + 主机名 + "）" + 发邮件主题.Load()
			邮件正文 := "由" + 远程机器名 + "的访问" + 邮件主题 + "\r\n" + "\r\n登录名：" + 登录名 + "\r\n登录密码：" + 登录密码
			go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
			go 报警发邮件函数(自定义发送人, 自定义发邮件token, 默认邮件接收者, 邮件主题, 邮件正文)
		}
	case "公网域名":
		if 服务器域名 != 主机名 {
			服务器域名 = 主机名
			atomic.StoreUint32(&服务器访问连接有变化, 1)
			atomic.StoreInt64(&数据网关状态微信报警推送上次成功时刻秒, 0)
			邮件主题 := "获得公网域名（" + 主机名 + "）" + 发邮件主题.Load()
			邮件正文 := "由" + 远程机器名 + "的访问" + 邮件主题 + "\r\n" + "\r\n登录名：" + 登录名 + "\r\n登录密码：" + 登录密码
			go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
			go 报警发邮件函数(自定义发送人, 自定义发邮件token, 默认邮件接收者, 邮件主题, 邮件正文)
		}
	case "ipv6":
		if 被访问获得的本机ipv6及端口号 != 主机名 {
			被访问获得的本机ipv6及端口号 = 主机名
			atomic.StoreUint32(&服务器访问连接有变化, 1)
			atomic.StoreInt64(&数据网关状态微信报警推送上次成功时刻秒, 0)
			邮件主题 := "获得ipv6（" + 主机名 + "）" + 发邮件主题.Load()
			邮件正文 := "由" + 远程机器名 + "的访问" + 邮件主题 + "\r\n" + "\r\n登录名：" + 登录名 + "\r\n登录密码：" + 登录密码
			go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
			go 报警发邮件函数(自定义发送人, 自定义发邮件token, 默认邮件接收者, 邮件主题, 邮件正文)
		}
	} //switch 判断r_host类型(主机名) {
} //func 获取服务器域名(r *http.Request) {
func 发送待发邮件() {
	time.Sleep(30 * time.Second) //等启动软件稳定后再发邮件
	ticker := time.NewTicker(自动发邮件时间间隔秒数 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if atomic.LoadUint32(&需要发邮件) == 0 {
			continue
		}
		发邮件函数()
	}
}
func 发送待发邮件2() {
	time.Sleep(30 * time.Second) //等启动软件稳定后再发邮件
	ticker := time.NewTicker(自动发邮件时间间隔秒数 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if atomic.LoadUint32(&需要发邮件2) == 0 {
			continue
		}
		发邮件函数2()
	}
}
func 能上网吗() bool {
	conn, err := net.DialTimeout("tcp", "8.8.8.8:53", 5*time.Second)
	if err != nil {
		if atomic.LoadUint32(&能上网) == 1 {
			atomic.StoreUint32(&能上网, 0)
		}
		return false
	}
	defer conn.Close()
	if atomic.LoadUint32(&能上网) == 0 {
		atomic.StoreUint32(&能上网, 1)
	}
	return true
}

var 需要发邮件 uint32 = 0
var 触发者信息 string = "网关启动触发"
var 上次发邮件时刻 int64
var 邮件正文前附加 string
var 邮件主题前附加 string = 触发者信息
var 服务器访问连接 字符串互斥锁访问结构体
var 发邮件错误信息 string = "\r\n上次发邮件错误信息： "
var 发邮件主题 字符串互斥锁访问结构体

func 发邮件函数() {
	if 自定义发送人 == "" {
		return
	}
	if 自定义发邮件token == "" {
		return
	}
	if 自定义邮件接收者 == "" {
		return
	}
	if 发邮件主题.Load() == "" {
		return
	}
	//fmt.Println("触发发邮件开始")
	atomic.StoreInt64(&上次发邮件时刻, time.Now().Unix())
	if atomic.LoadUint32(&能上网) == 0 {
		return
	}
	需要发送的文件们 := make(map[string]string)
	//fmt.Println("触发发邮件可以")
	if 判断文件是否存在(项目变量信息表名) {
		需要发送的文件们[项目变量信息表名1] = 项目变量信息表名 //要发送的文件
	}
	文件名 := time.Now().Format("2006-01-02 15:04:05")
	文件名 = strings.Replace(文件名, " ", "_", -1)
	文件名 = strings.Replace(文件名, ":", "_", -1)
	文件名 = "网关详情" + "_" + 文件名 + ".txt"
	if 邮件主题前附加 != "" {
		邮件主题前附加 = 邮件主题前附加 + ","
	}
	if 邮件正文前附加 != "" {
		邮件正文前附加 = 邮件正文前附加 + "\r\n"
	}
	文件名 = 邮件主题前附加 + 文件名
	主题 := 邮件主题前附加 + 发邮件主题.Load()
	str := 服务器访问连接.Load()
	邮件正文 := 主题 + "\r\n" + 邮件正文前附加 + "\r\n" + "本次邮件发送触发者信息：" + 触发者信息2 +
		"\r\n本邮件来自我们公司(www.pdlei.cn)开发的ModbusRTU->HttpAPI设备数据采集网关，让您知道软件使用情况及如何访问它\r\n" +
		"此网关可用上网浏览器访问，访问连接为: " + str + "/统计" + "\r\n登录名：" + 登录名 +
		"\r\n登录密码：" + 登录密码 + "\r\n" + 发邮件错误信息2 + 获取统计信息()
	//fmt.Println(主题 + "\r\n" + "本次邮件发送触发者信息：" + 触发者信息2)
	创建文件(目录+"txt/", 文件名, 邮件正文)
	需要发送的文件们[文件名] = 目录 + "txt/" + 文件名 //要发送的文件
	m := gomail.NewMessage()
	for 文件名, 文件名及路径 := range 需要发送的文件们 {
		m.Attach(文件名及路径,
			gomail.Rename(文件名),
			gomail.SetHeader(map[string][]string{
				"Content-Disposition": {
					fmt.Sprintf(`attachment; filename="%s"`, mime.QEncoding.Encode("UTF-8", 文件名)),
				},
			}),
		)
	}
	邮件头 := make(map[string][]string)
	邮件头["To"] = strings.Split(strings.Replace(自定义邮件接收者, " ", "", -1), ",")
	邮件头["From"] = []string{自定义发送人}
	邮件头["Subject"] = []string{主题}
	m.SetHeaders(邮件头)
	邮件正文 = strings.Replace(邮件正文, "\r\n", "<br>", -1)
	邮件正文 = strings.Replace(邮件正文, "\r", "<br>", -1)
	邮件正文 = strings.Replace(邮件正文, "\n", "<br>", -1)
	m.SetBody("text/html", 邮件正文)
	d := gomail.NewDialer("smtp.qq.com", 587, 自定义发送人, 自定义发邮件token)
	发邮件错误信息 = "\r\n上次发邮件错误信息： "
	if err := d.DialAndSend(m); err != nil {
		发邮件错误信息 = "\r\n上次发邮件错误信息： " + err.Error()
		fmt.Println(发邮件错误信息)
	}
	atomic.StoreUint32(&需要发邮件, 0)
} //func 发邮件函数(){
var 需要发邮件2 uint32 = 0
var 触发者信息2 string
var 能上网 uint32 = 0
var 发邮件错误信息2 string = "\r\n上次发邮件错误信息： "

func 发邮件函数2() {
	if 发送人 == "" {
		return
	}
	if 发邮件token == "" {
		return
	}
	if 默认邮件接收者 == "" {
		return
	}
	if 发邮件主题.Load() == "" {
		return
	}
	if atomic.LoadUint32(&能上网) == 0 {
		return
	}
	需要发送的文件们 := make(map[string]string)
	if 判断文件是否存在(项目变量信息表名) {
		需要发送的文件们[项目变量信息表名1] = 项目变量信息表名 //要发送的文件
	}
	主题 := 发邮件主题.Load()
	str := 服务器访问连接.Load()
	邮件正文 := 主题 + "\r\n" + 邮件正文前附加 + "\r\n" + "本次邮件发送触发者信息：" + 触发者信息2 +
		"\r\n本邮件来自我们公司(www.pdlei.cn)开发的ModbusRTU->HttpAPI设备数据采集网关，让您知道软件使用情况及如何访问它\r\n" +
		"此网关可用上网浏览器访问，访问连接为: " + str + "/统计" + "\r\n登录名：" + 登录名 +
		"\r\n登录密码：" + 登录密码 + "\r\n" + 发邮件错误信息2 + 获取统计信息()
	文件名 := time.Now().Format("2006-01-02 15:04:05")
	文件名 = strings.Replace(文件名, " ", "_", -1)
	文件名 = strings.Replace(文件名, ":", "_", -1)
	文件名 = "网关详情" + "_" + 文件名 + ".txt"
	创建文件(目录+"txt/", 文件名, 邮件正文)
	需要发送的文件们[文件名] = 目录 + "txt/" + 文件名 //要发送的文件
	m := gomail.NewMessage()
	for 文件名, 文件名及路径 := range 需要发送的文件们 {
		m.Attach(文件名及路径,
			gomail.Rename(文件名),
			gomail.SetHeader(map[string][]string{
				"Content-Disposition": {
					fmt.Sprintf(`attachment; filename="%s"`, mime.QEncoding.Encode("UTF-8", 文件名)),
				},
			}),
		)
	}
	邮件头 := make(map[string][]string)
	邮件头["To"] = strings.Split(strings.Replace(默认邮件接收者, " ", "", -1), ",")
	邮件头["From"] = []string{发送人}
	邮件头["Subject"] = []string{主题}
	m.SetHeaders(邮件头)
	邮件正文 = strings.Replace(邮件正文, "\r\n", "<br>", -1)
	邮件正文 = strings.Replace(邮件正文, "\r", "<br>", -1)
	邮件正文 = strings.Replace(邮件正文, "\n", "<br>", -1)
	m.SetBody("text/html", 邮件正文)
	d := gomail.NewDialer("smtp.qq.com", 587, 发送人, 发邮件token)
	发邮件错误信息2 = "\r\n上次发邮件错误信息： "
	if err := d.DialAndSend(m); err != nil {
		发邮件错误信息2 = "\r\n上次发邮件错误信息： " + err.Error()
	}
	atomic.StoreUint32(&需要发邮件2, 0)
} //func 发邮件函数2(){
type LSysInfo struct {
	MemAll         uint64
	MemFree        uint64
	MemUsed        uint64
	MemUsedPercent float64
	锁              sync.Mutex
}

func 获取进程资源占用情况(进程名 string) (float64, float32) {
	str := ""
	进程指针组, _ := process.Processes()
	for _, 进程指针 := range 进程指针组 {
		str, _ = 进程指针.Name()
		if str == 进程名 {
			f, _ := 进程指针.MemoryPercent()
			f2, _ := 进程指针.CPUPercent()
			return f2, f
		}
	}
	return -9.9, -8.8
}

var info LSysInfo
var 内存百分比连续超过设定次数 int

func GetSysInfo() {
	if atomic.LoadUint32(&已经触发重启程序了) == 1 {
		return
	}
	unit := uint64(1024 * 1024) // MB
	v, _ := mem.VirtualMemory()
	info.锁.Lock()
	info.MemAll = v.Total
	info.MemFree = v.Free
	info.MemUsed = info.MemAll - info.MemFree
	// 注:使用SwapMemory或VirtualMemory，在不同系统中使用率不一样，因此直接计算一次
	info.MemUsedPercent = float64(info.MemUsed) / float64(info.MemAll) * 100.0
	info.MemAll /= unit
	info.MemUsed /= unit
	info.MemFree /= unit
	info.锁.Unlock()
	MemUsedPercent := info.MemUsedPercent
	if MemUsedPercent > 内存百分比设定 {
		内存百分比连续超过设定次数++
		if 内存百分比连续超过设定次数 > 内存百分比连续超过设定次数限制 {
			//			str := "内存百分比连续超过设定次数 > 内存百分比连续超过设定次数限制，项目数据采集网关ModbusRTU->HttpAPI将被重启，若反复重启，请人工解决系统内存占用过高问题！"
			if atomic.LoadUint32(&已经触发重启程序了) == 0 {
				str1 := ""
				str := 至今多少天时分秒(网关启动时刻)
				str = "启动至今: " + str + "\r\n"
				str1 += str
				str = 网关启动时间 + "\r\n"
				str1 += str
				str = time.Now().Format("2006-01-02 15:04:05") + "\r\n"
				str1 += str
				//批处理文件名非中文提示 = "内存百分比连续超过设定次数软件重启，若反复重启，请人工解决系统内存占用过高问题！"
				邮件主题 := "内存百分比连续超过设定次数软件重启," + 发邮件主题.Load()
				邮件正文 := 邮件主题 + "\r\n" + "内存百分比(" +
					strconv.FormatFloat(MemUsedPercent, 'f', 2, 32) +
					")连续超过设定(" + 内存百分比设定_s + ")次数(" +
					内存百分比连续超过设定次数限制_s +
					")软件重启，若反复重启，请人工解决系统内存占用过高问题！" +
					str1 + "\r\n内存占用率高数据分析：" + 内存占用率高数据分析()
				go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
				批处理文件名非中文提示 := "NeiCunBaiFenBiLianXuChaoGuoSheDingCiShu"
				go 重新启动外部程序(批处理文件名非中文提示, 程序名, 程序文件名及路径)
			}
		} //if 内存百分比连续超过设定次数 > 内存百分比连续超过设定次数限制 {
	} else {
		内存百分比连续超过设定次数 = 0
	}
} //func GetSysInfo() {
func 获取系统信息() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		GetSysInfo()
	}
}

var 本网关登录密码之MD5 string
var 本网关登录名之MD5 string

func getMd5(待加密字符串 string) string {
	// 进行md5加密，因为Sum函数接受的是字节数组，因此需要注意类型转换
	srcCode := md5.Sum([]byte(待加密字符串))
	// md5.Sum函数加密后返回的是字节数组，需要转换成16进制形式
	code := fmt.Sprintf("%x", srcCode)
	return string(code)
}

var 端口 string
var 程序名 string = "ModbusRTU->HttpAPI网关"
var 程序所在文件夹 string
var 程序文件名及路径 string
var 程序文件名及路径2 string
var 网关启动时刻 int64
var 网关启动时间 string
var 目录 string
var 项目变量信息表名 string
var 项目变量信息表名备份及路径 string
var 目录2 string

func main() {
	网关启动时间 = time.Now().Format("2006-01-02 15:04:05")
	网关启动时刻 = time.Now().Unix()
	sigChan := make(chan os.Signal, 1)
	// 通知 signal 包使用 sigChan 通道接收信号
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	// 启动一个 goroutine 来处理信号
	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				fmt.Println("Received interrupt signal, shutting down gracefully...")
				写项目变量信息表并保存()
				atomic.StoreUint32(&已经触发重启程序了, 1)
				fmt.Println("Cleanup done, exiting now.")
				SQLiteDB访问锁.Lock()
				写项目变量信息表并保存锁.Lock()
				if SQLiteDB内存数据库连接 != nil {
					SQLiteDB内存数据库连接.Close()
				}
				if SQLiteDB磁盘查询历史数据库连接 != nil {
					SQLiteDB磁盘查询历史数据库连接.Close()
				}
				os.Exit(0)
			case syscall.SIGHUP:
				fmt.Println("Received hangup signal.")
				// 在这里处理 SIGHUP 信号，通常用于重新加载配置等
				// ...
			default:
				fmt.Printf("Received unsupported signal: %v\n", sig)
			}
		}
	}()
	go 检查上网状态()
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		程序名 = "无法获取程序名"
		程序所在文件夹 = "无法获取程序所在文件夹"
		程序文件名及路径 = "无法获取程序文件名及路径"
	} else {
		if strings.Contains(os.Args[0], ".exe") {
			if strings.Contains(os.Args[0], "\\") {
				位置 := strings.LastIndex(os.Args[0], "\\")
				程序名 = os.Args[0][位置+1:]
				程序所在文件夹 = os.Args[0][:位置]
				程序文件名及路径 = os.Args[0]
			} else {
				程序名 = os.Args[0]
				程序所在文件夹 = dir
				程序文件名及路径 = dir + "\\" + 程序名
			}
		} else {
			程序名 = os.Args[0] + ".exe"
			程序所在文件夹 = dir
			程序文件名及路径 = dir + "\\" + 程序名
		}
	}
	loginhtml1 = "<!DOCTYPE html><html lang=\"en\"><head>	<meta charset=\"UTF-8\"><title>" + 程序文件名及路径 + "</title></head><body><form action=\"http://"
	目录 = 程序文件名及路径[:strings.Index(程序文件名及路径, ":")] + ":\\" + strings.Replace(程序名, ".exe", "", -1) + "\\"
	目录2 = 程序文件名及路径[:strings.Index(程序文件名及路径, ":")] + ":\\" + strings.Replace(程序名, ".exe", "", -1) + "2\\"
	程序文件名及路径2 = 目录2 + 程序名 + "2"
	项目变量信息表名 = 目录 + 项目变量信息表名1
	项目变量信息表名备份及路径 = 目录 + 项目变量信息表名备份
	默认使用者 = "请在 " + 项目变量信息表名 + " 中填入您的信息，方便我公司为您主动提供服务！\r\n"
	使用者 = "\r\n使用者: " + 默认使用者
	当前用户名 = 获取当前用户名()
	连接信息 = make(map[string]*连接信息2, 10)
	连接登录时刻 = make(map[string]int64, 1)
	ips属地 = make(map[string]*ip属地数据结构, 10)
	启动软件碰到的问题.Set("")
	启动软件检查表格文件()
	遍历项目变量信息表2()
	fmt.Println(启动软件碰到的问题.Load())
	fmt.Println(获得设定值())
	webhook_url = webhook_url_默认
	杀掉同名程序()
	if 完成第一次遍历项目变量信息表 {
		err := 复制文件(项目变量信息表名, 项目变量信息表名备份及路径)
		if err != nil {
			fmt.Println("备份项目变量信息表错误:", err)
			邮件主题 := "备份项目变量信息表错误," + 发邮件主题.Load()
			邮件正文 := 邮件主题 + "\r\n" + err.Error() + "\r\n" + "服务器信息：\r\n" + 获取服务器信息() + "\r\n" + 获取服务器内存使用率() + "\r\n"
			go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
		} else {
			fmt.Println("备份" + 项目变量信息表名 + "到" + 项目变量信息表名备份及路径 + "成功!")
		}
		for c := range 串口组 {
			go 采集任务(c)
		}
	} else {
		err := 复制文件(项目变量信息表名备份及路径, 项目变量信息表名)
		if err != nil {
			fmt.Println("恢复项目变量信息表错误:", err)
			邮件主题 := "恢复项目变量信息表错误," + 发邮件主题.Load()
			邮件正文 := 邮件主题 + "\r\n" + err.Error() + "\r\n项目变量信息表错误，软件尝试恢复备份文件失败，需人工介入！\r\n" + "服务器信息：\r\n" +
				获取服务器信息() + "\r\n" + 获取服务器内存使用率() + "\r\n"
			go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
		} else {
			fmt.Println("恢复" + 项目变量信息表名备份及路径 + "到" + 项目变量信息表名 + "成功!")
			批处理文件名非中文提示 := "HuiFuXiangMuBianLiangXinXiBiao"
			go 重新启动外部程序(批处理文件名非中文提示, 程序名, 程序文件名及路径)
		}
		检查端口是否被占用()
	}
	if !允许网关启动发邮件 {
		atomic.StoreUint32(&需要发邮件, 0)
	} else {
		atomic.StoreUint32(&需要发邮件, 1)
	}
	本网关登录密码之MD5 = getMd5(登录密码)
	本网关登录名之MD5 = getMd5(登录名)
	SQLiteDB使用源库查询情况.使用源库平均查询每记录耗时2 = time.Duration(SQLiteDB使用源库平均查询每记录耗时默认us * time.Microsecond)
	SQLiteDB使用源库查询情况.使用源库平均查询每记录耗时2s = fmt.Sprintf("%v", time.Duration(SQLiteDB使用源库平均查询每记录耗时默认us*time.Microsecond))
	从文件中读入SQLiteDB库文件备份目录信息()
	从文件中读入SQLiteDB文件名及路径信息()
	检查修正SQLitedbPath()
	创建SQLiteDB内存数据库()
	go 周期生成SQLite记录()
	go SQLite()
	go GetDiskInfoWindows()
	go 发送待发邮件()
	go 发送待发邮件2() //当网关能外网访问时发邮件给开发公司
	go 获取系统信息()
	go 获取ips属地()
	go 紧急获取token2()
	go 企业微信报警系统状态推送()
	go 检查token是否失效1()
	go 项目变量信息表写保存()
	if 存在, _ := 判断目录是否存在(目录2); !存在 {
		启动软件碰到的问题.Set("目录2(" + 目录2 + ")不存在，将被创建用于文件服务器！")
		fmt.Println("目录2(" + 目录2 + ")不存在，将被创建用于文件服务器！")
		创建目录(目录2)
	}
	//文件服务器初始化开始
	登录验证Cookie名称 = 本网关UUID + 本网关登录密码之MD5 + 本网关登录名之MD5
	//登录验证Cookie名称 = 程序名
	username = 登录名  // 登录用户名
	password = 登录密码 // 登录密码
	//文件服务器初始化结束
	mux := http.NewServeMux()
	mux.HandleFunc("/token", token)
	mux.HandleFunc("/GetTagList", GetTagList)
	mux.HandleFunc("/api/GetKVTagsValue", GetKVTagsValue)
	mux.HandleFunc("/api/SetKVTagsValue", SetKVTagsValue)
	mux.HandleFunc("/SetTagValue", SetTagValue)
	mux.HandleFunc("/SQLiteDB查询", SQLiteDB查询)
	mux.HandleFunc("/SQLiteDB复制", SQLiteDB复制)
	mux.HandleFunc("/GetTagValue", GetTagValue)
	mux.HandleFunc("/xlsx", xlsx)
	mux.HandleFunc("/读表", 重启)
	mux.HandleFunc("/统计", 统计)
	mux.HandleFunc("/重启", 重启)
	mux.HandleFunc("/数据", 数据)
	mux.HandleFunc("/发邮件", 发邮件)
	mux.HandleFunc("/重启服务器", 重启服务器)
	mux.HandleFunc("/关闭", 关闭)
	mux.HandleFunc("/调试", 调试)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/upload", upload)
	//mux.HandleFunc("/", pdlei)
	mux.HandleFunc("/", loginHandler)
	mux.HandleFunc("/files", authMiddleware(fileHandler))
	mux.HandleFunc("/download", authMiddleware(downloadHandler))
	s := &http.Server{
		Addr:         ":" + 端口,
		WriteTimeout: time.Second * http服务器写超时秒,
		Handler:      mux,
	}
	fmt.Println("端口：" + 端口)
	s.ListenAndServe()
} //func main() {
// func pdlei(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("Hi, " + r.RemoteAddr + " !"))
// }

const (
	ANSIC       = "Mon Jan _2 15:04:05 2006"
	UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
	RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
	RFC822      = "02 Jan 06 15:04 MST"
	RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
	RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
	RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	RFC3339     = "2006-01-02T15:04:05Z07:00"
	RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
	Kitchen     = "3:04PM"
	Stamp       = "Jan _2 15:04:05"
	StampMilli  = "Jan _2 15:04:05.000"
	StampMicro  = "Jan _2 15:04:05.000000"
	StampNano   = "Jan _2 15:04:05.000000000"
)

func 至今多少天时分秒(pastUnixTimestamp int64) string {
	// 假设过去某时刻是一个固定的Unix时间戳，例如：1609459200 (2021-01-01 00:00:00 UTC)
	pastTime := time.Unix(pastUnixTimestamp, 0)
	// 获取当前时间
	now := time.Now()
	// 计算时间差
	duration := now.Sub(pastTime)
	结果 := ""
	// 将时间差转换为天、小时、分钟和秒
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	if days > 0 {
		结果 += fmt.Sprintf("%d天", days)
	}
	if hours > 0 {
		结果 += fmt.Sprintf("%d小时", hours)
	}
	if minutes > 0 {
		结果 += fmt.Sprintf("%d分钟", minutes)
	}
	if seconds > 0 {
		结果 += fmt.Sprintf("%d秒", seconds)
	} else {
		结果 += "0秒"
	}
	return 结果
	//return fmt.Sprintf("从过去某时刻至今：%d天 %d小时 %d分钟 %d秒\r\n", days, hours, minutes, seconds)
} //func 至今多少天时分秒(pastUnixTimestamp int64) string {
func 内存占用率高数据分析() string {
	str1 := ""
	str := 数据网关状态微信报警内容收集() + "\r\n"
	str1 += str
	str = 启动软件碰到的问题.Load()
	str = "启动软件碰到的问题:\r\n" + str
	str1 += str
	str = "当前开启文件服务器设定为允许" + "\r\n"
	if !开启文件服务器 {
		str = "当前开启文件服务器设定为禁止" + "\r\n"
	}
	str1 += str
	/////////////////
	str = "数据网关状态微信报警推送错误信息：\n" + 数据网关状态微信报警推送错误信息.Load() + "\r\n"
	str1 += str
	str = "发邮件错误信息： " + 发邮件错误信息 + "\r\n"
	str1 += str
	str = 获取访问连接信息()
	str1 += str
	return str1
} //func 内存占用率高数据分析() string {
func 生成api访问样例(本机IP地址为 string, 端口 string, 本网关登录名之MD5 string, 本网关登录密码之MD5 string) string {
	str := ""
	str1 := ""
	str = "软件是通过>>" + `net.Dial("udp", "8.8.8.8:53")` +
		">>获取本机局域网IP的，所以本机要能上网，否则很多依赖于本机局域网ip的自动化功能无法实现，比如以下api访问样例" + "\r\n"
	str1 += str
	api访问样例 := "http://" + 本机IP地址为 + ":" + 端口 + "/数据?user=" + 本网关登录名之MD5 +
		"&password=" + 本网关登录密码之MD5 +
		"&忽略变量详情=是" +
		"&项目代号=A09\r\n" + "要查询的项目代号，比如 A09,如果为空，则获取全部项目数据"
	str = "api访问样例:\r\n"
	str1 += str
	str = api访问样例 + "\r\n"
	str1 += str
	api访问样例 = "http://" + 本机IP地址为 + ":" + 端口 + "/GetTagList?user=" + 本网关登录名之MD5 +
		"&password=" + 本网关登录密码之MD5 +
		"&项目代号=A09"
	str = api访问样例 + "\r\n"
	str1 += str
	api访问样例 = "http://" + 本机IP地址为 + ":" + 端口 + "/xlsx?user=" + 本网关登录名之MD5 +
		"&password=" + 本网关登录密码之MD5 +
		"&项目代号=A09"
	str = api访问样例 + "\r\n"
	str1 += str
	api访问样例 = "http://" + 本机IP地址为 + ":" + 端口 + "/token?user=" + 本网关登录名之MD5 +
		"&password=" + 本网关登录密码之MD5
	str = api访问样例 + "\r\n"
	str1 += str
	api访问样例 = "http://" + 本机IP地址为 + ":" + 端口 + "/SetTagValue?user=" + 本网关登录名之MD5 +
		"&password=" + 本网关登录密码之MD5 + "&strTagName=" + "要写的变量名" + "&strSetTagValue=" + "要写的值"
	str = api访问样例 + "\r\n"
	str1 += str
	api访问样例 = "http://" + 本机IP地址为 + ":" + 端口 + "/GetTagValue?user=" + 本网关登录名之MD5 +
		"&password=" + 本网关登录密码之MD5 +
		"&忽略变量详情=是" +
		"&strTagNameS=" + "要读的变量名1,要读的变量名2"
	str = api访问样例 + "\r\n"
	str1 += str
	str = "一次最多读请求变量个数：" + 最大读请求变量个数_s + "\r\n"
	str1 += str
	api访问样例 = "http://" + 本机IP地址为 + ":" + 端口 +
		"/SQLiteDB查询?开始时间=" + time.Now().AddDate(0, 0, -1).Format("2006-01-02_15:04:05") +
		"&结束时间=" + time.Now().Format("2006-01-02_15:04:05") +
		"&间隔时间=1小时" +
		"&user=" + 本网关登录名之MD5 +
		"&password=" + 本网关登录密码之MD5 +
		"&曲线图Y值使用实际值=" + "是或空" +
		"&忽略异常值=是" +
		"&表格历史空数据自动填充=是" +
		"&忽略特定值们=0,0.00,9999.00" +
		"&忽略曲线说明=是或空" +
		"&计算差值=是或空" +
		"&回复格式=" + "TEXT或JSON或XLSX或GRAPH" +
		"&要查询的变量名们=" + "要查询的变量名1:累积，9999.00,0.00|要查询的变量名2:9999.00,0.00" +
		"\r\n" +
		"GRAPH：生成曲线图需要在工作目录(" + 目录 + ")下有msyh.TTF(微软雅黑字体文件，否则无法显示像μ这样的特殊字符，可自行下载)。另外要注意的是如果生成的曲线很多并且要求计算差值，PNG图像就非常复杂，生成此图像就非常耗时，本软件回复请求超时后客户端浏览器就会反复请求（请反复按ESC键退出请求），但是文件服务器(" + "http://" + 本机IP地址为 + ":" + 端口 + "/files" + ")中已经有此图像，可以下载浏览" + "\r\n" +
		"一次最多查询请求变量个数：" + 最大读请求变量个数_s + "\r\n" +
		"忽略异常值：忽略通信失败的异常值\r\n" +
		"忽略特定值们：非通信失败的异常值，是设备厂家自己定义的，比如温度探头接触不良，设备返回9999.00，这项作用于所有请求的变量\r\n" +
		"要查询的变量名1:累积，9999.00,0.00|(|或中文|变量间分隔符) 9999.00,0.00表示这个变量的这些历史数据是非法值，请忽略。累积表示变量值只增不减，对于一个类似电表读数到了1万度，更换电表又从0开始计量，软件在统计耗电量时不会算错\r\n" +
		"表格历史空数据自动填充：若是，则使用前一行数据填充当前行，如果当前行是第一个历史数据且为空，那么就往后寻找最先碰到的数据来填充，若计算差值，会自动填充0"
	str = api访问样例 + "\r\n"
	str1 += str
	api访问样例 = "http://" + 本机IP地址为 + ":" + 端口 + "/SQLiteDB复制?user=" + 本网关登录名之MD5 +
		"&password=" + 本网关登录密码之MD5 +
		"&被复制的库文件们=" + "被复制的库文件1,被复制的库文件2" + "\r\n" +
		"如：Y:\\SQLiteDB_20250212181216，Y:\\SQLiteDB_20250213160314" +
		"，如果复制的记录数很大比如10万条，每条插入目标库10ms,那么需要1000秒，请耐心等待，期间不可关闭浏览器，请选择合适时间操作"
	str = api访问样例 + "\r\n"
	str1 += str
	str = "一次最多写请求变量个数：" + 最大写请求变量个数_s + "\r\n"
	str1 += str
	api访问样例 = "http://" + 本机IP地址为 + ":" + 端口 + "/api/SetKVTagsValue\r\n" + "{\r\n\"user\":\"" + 本网关登录名之MD5 + "\",\r\n\"password\":\"" + 本网关登录密码之MD5 + "\",\r\n\"data\":[\r\n{\"name\":\"要写的变量名1\",\"value\":\"要写的值1\"},\r\n{\"name\":\"要写的变量名2\",\"value\":\"要写的值2\"},]}"
	str = api访问样例 + "\r\n"
	str1 += str
	str = "一次最多读请求变量个数：" + 最大读请求变量个数_s + "\r\n"
	str1 += str
	api访问样例 = "http://" + 本机IP地址为 + ":" + 端口 + "/api/GetKVTagsValue\r\n" + "{\"data\":[\r\n{\"name\":\"要读的变量名1\"},\r\n{\"name\":\"要读的变量名2\"},]}"
	str = api访问样例 + "\r\n"
	str1 += str
	str = "如果本网关没有外网可访问的域名或ip,您又想在任何地方访问它，可以使用lcx.exe帮忙（lcx.exe是一个用于端口转发的工具，在网络安全领域常被用于内网穿透。它可以将内网中的某台计算机（通常被称为“肉鸡”）上的端口转发到具有外网IP的另一台计算机上，从而使得外部用户能够访问到内网中的服务）"
	str1 += str
	return str1
} //func 生成api访问样例(本机IP地址为 string)string{
func 获取统计信息() string {
	startTime := time.Now()
	str1 := ""
	str := "\r\n欢迎使用我公司开发的软件(www.pdlei.cn)"
	str1 += str
	str = 开发者 + "\r\n"
	str1 += str
	str = 服务器访问连接.Load()
	str = "网关多种访问连接: " + str + "\r\n"
	str1 += str
	本机IP地址为2 := 本机IP地址为
	端口2 := 端口
	本网关登录名之MD5a := 本网关登录名之MD5
	本网关登录密码之MD5a := 本网关登录密码之MD5
	if 本机IP地址为 == "" {
		本机IP地址为2 = "127.0.0.1"
	}
	if 端口 == "" {
		端口2 = "端口号"
	}
	if 本网关登录名之MD5 == "" {
		本网关登录名之MD5a = "本网关登录名之MD5"
	}
	if 本网关登录密码之MD5 == "" {
		本网关登录密码之MD5a = "本网关登录密码之MD5"
	}
	str = 生成api访问样例(本机IP地址为2, 端口2, 本网关登录名之MD5a, 本网关登录密码之MD5a)
	str1 += str + "\r\n"
	str = "可上传在服务器运行的批处理文件名:" + 可上传在服务器运行的批处理文件名 +
		"(可以对服务器执行几乎所有操作，比如删除某目录下 *.txt 等,生成此批处理文件可以用代码或记事本的方式，但要特别注意保存时最好选择ANSI或GB*编码及以\\r\\n换行，才能操作中文文件名的文件)\r\n"
	str1 += str
	str = "企业微信信息推送机器人推送的信息：\r\n开始******************************************开始\r\n"
	str1 += str
	str = 数据网关状态微信报警内容收集() + "\r\n"
	str1 += str
	str = "结束******************************************结束\r\n"
	str1 += str
	str = 启动软件碰到的问题.Load()
	str = "启动软件碰到的问题:\r\n" + str
	str1 += str
	str = 程序名
	str = "程序名: " + str + "\r\n"
	str1 += str
	str = 程序所在文件夹
	str = "程序所在文件夹: " + str + "\r\n"
	str1 += str
	str = 程序文件名及路径
	str = "程序文件名及路径: " + str + "\r\n"
	str1 += str
	str = 目录
	str = "软件工作目录: " + str + "\r\n"
	str1 += str
	str = "文件服务器目录: " + 目录2 + "\r\n"
	str1 += str
	str = 获得设定值() + "\r\n"
	str1 += str
	if 完成第一次遍历项目变量信息表 {
		str = "成功遍历项目变量信息表且此表能用"
		if !strings.Contains(遍历项目变量信息表结果, str) {
			遍历项目变量信息表结果 = str + "\r\n" + 遍历项目变量信息表结果
		}
	} else {
		str = "遍历项目变量信息表未成功！"
		if !strings.Contains(遍历项目变量信息表结果, str) {
			遍历项目变量信息表结果 = str + "\r\n" + 遍历项目变量信息表结果
		}
	}
	str = 遍历项目变量信息表结果
	str = "遍历项目变量信息表结果:\r\n" + str + "\r\n"
	str1 += str
	str = "数据网关状态微信报警推送错误信息：\r\n" + 数据网关状态微信报警推送错误信息.Load() + "\r\n"
	str1 += str
	str = "发邮件错误信息：\r\n" + 发邮件错误信息 + "\r\n"
	str1 += str
	str = "SQLite错误信息：\r\n" + SQLite错误信息.Load() + "\r\n"
	str1 += str
	str = 获取访问连接信息() + "\r\n"
	str1 += str
	elapsedTime := time.Since(startTime)
	最大耗时 := fmt.Sprintf("获取统计信息字符串最大耗时：%v", elapsedTime)
	fmt.Println(最大耗时)
	str = 最大耗时
	str1 += str
	return str1
} //func 获取统计信息() string {
func 获取访问连接信息() string {
	连接信息锁.Lock()
	defer 连接信息锁.Unlock()
	间隔符 := "_"
	str1 := ""
	str := strconv.FormatInt(int64(len(连接信息)), 10)
	str = "访问连接数：" + str + "\r\n"
	str1 += str
	str = "访问连接" + 间隔符 + "连接时间" + 间隔符 + "登录时间" + 间隔符 + "退出登录时间" + 间隔符 +
		"统计" + 间隔符 + "xlsx" + 间隔符 + "token" + 间隔符 + "GetTagList" + 间隔符 + "api/GetKVTagsValue" +
		间隔符 + "api/SetKVTagsValue" + 间隔符 + "login" + 间隔符 + "upload" + 间隔符 + "调试" + 间隔符 + "关闭" + 间隔符 + "重启" +
		间隔符 + "重启服务器" + 间隔符 + "/" + 间隔符 + "文件浏览" + 间隔符 + "下载文件" + 间隔符 + "发邮件" + 间隔符 + "数据" + 间隔符 + "设定数据库信息" + 间隔符 + "GetTagValue" + 间隔符 + "SQLiteDB查询" + 间隔符 + "SQLiteDB复制" + 间隔符 + "SetTagValue" + 间隔符 + "SetTagValue请求信息" + "\r\n"
	str1 += str
	for i, v := range 连接信息 {
		str = i + 间隔符 + v.L连接时间 + 间隔符 + v.D登录时间 + 间隔符 + v.T退出登录时间 + 间隔符 +
			strconv.Itoa(v.T统计) + 间隔符 + strconv.Itoa(v.Xlsx) + 间隔符 + strconv.Itoa(v.Token) +
			间隔符 + strconv.Itoa(v.GetTagList) + 间隔符 + strconv.Itoa(v.GetKVTagsValue) + 间隔符 + strconv.Itoa(v.SetKVTagsValue) + 间隔符 +
			strconv.Itoa(v.Login) + 间隔符 + strconv.Itoa(v.Upload) + 间隔符 + strconv.Itoa(v.T调试) +
			间隔符 + strconv.Itoa(v.G关闭) + 间隔符 + strconv.Itoa(v.C重启) + 间隔符 +
			strconv.Itoa(v.C重启服务器) + 间隔符 + strconv.Itoa(v.G根目录) + 间隔符 + strconv.Itoa(v.W文件浏览) + 间隔符 + strconv.Itoa(v.X下载文件) + 间隔符 + strconv.Itoa(v.F发邮件) +
			间隔符 + strconv.Itoa(v.S数据) + 间隔符 + strconv.Itoa(v.S设定数据库信息) + 间隔符 + strconv.Itoa(v.GetTagValue) + 间隔符 + strconv.Itoa(v.SQLiteDB查询) + 间隔符 + strconv.Itoa(v.SQLiteDB复制) + 间隔符 + strconv.Itoa(v.SetTagValue) + 间隔符 + v.SetTagValue请求信息 + "\r\n"
		str1 += str
	} //for i,v:=range 连接信息{
	return str1
} //func 获取访问连接信息() string {
func GetIP属地(ip string) (bool, string) {
	网站 := "https://opendata.baidu.com/api.php?co=&resource_id=6006&oe=utf8&query="
	str := 网站 + ip
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", str, nil)
	if err != nil {
		return false, err.Error()
	}
	req.Header.Set("User-Agent", "www.pdlei.cn")
	resp, err := client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	input, err := io.ReadAll(resp.Body) //读取流数据
	if err != nil {
		return false, err.Error()
	}
	//{"status":"0","t":"","set_cache_time":"","data":[{"ExtendedLocation":"","OriginQuery":"222.216.40.41","appinfo":"","disp_type":0,"fetchkey":"222.216.40.41","location":"广西壮族自治区南宁市 电信",
	//"origip":"222.216.40.41","origipquery":"222.216.40.41","resourceid":"6006","role_id":0,"shareImage":1,"showLikeShare":1,"showlamp":"1","titlecont":"IP地址查询","tplt":"ip"}]}
	type 查询结果结构 struct {
		ExtendedLocation string `json:"ExtendedLocation"`
		OriginQuery      string `json:"OriginQuery"`
		Appinfo          string `json:"appinfo"`
		Disp_type        int    `json:"disp_type"`
		Fetchkey         string `json:"fetchkey"`
		Location         string `json:"location"`
		Origip           string `json:"origip"`
		Origipquery      string `json:"origipquery"`
		Resourceid       string `json:"resourceid"`
		Role_id          int    `json:"role_id"`
		ShareImage       int    `json:"shareImage"`
		ShowLikeShare    int    `json:"showLikeShare"`
		Showlamp         string `json:"showlamp"`
		Titlecont        string `json:"titlecont"`
		Tplt             string `json:"tplt"`
	} //type 查询结果结构 struct {
	type ip_api结构2 struct {
		Status         string   `json:"status"`
		T              string   `json:"t"`
		Set_cache_time string   `json:"set_cache_time"`
		Data           []查询结果结构 `json:"data"`
	}
	var ip_api结构 ip_api结构2
	err = json.Unmarshal(input, &ip_api结构) //解析json数据
	if err != nil {
		return false, err.Error()
	}
	str = ""
	分隔符 := "_"
	if ip_api结构.Status == "0" { //判断有无解析数据
		str = ip_api结构.Data[0].Origip + 分隔符 +
			ip_api结构.Data[0].Location
		return true, str
	} else {
		return false, ip_api结构.Status + "_查询失败"
	}
} //func GetIP属地(ip string)(bool string) {
type ip属地数据结构 struct {
	Ip  string `json:"ip"`
	S属地 string `json:"属地"`
}

var ips属地 map[string]*ip属地数据结构
var ips属地锁 sync.Mutex
var ips属地信息 字符串互斥锁访问结构体

func 获取ips属地() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if atomic.LoadUint32(&能上网) == 0 {
			continue
		}
		ips属地锁.Lock()
		长度 := len(ips属地)
		ips属地锁.Unlock()
		if 长度 < 1 {
			continue
		}
		var 需要查的ip = make(map[string]*ip属地数据结构, 0)
		ips属地锁.Lock()
		for _, v := range ips属地 {
			if v.S属地 != "" {
				continue
			}
			if _, ok := 需要查的ip[v.Ip]; !ok {
				需要查的ip[v.Ip] = &ip属地数据结构{}
				需要查的ip[v.Ip].Ip = v.Ip
			}
		} //for _, v := range ips属地 {
		ips属地锁.Unlock()
		有更改 := false
		for _, v := range 需要查的ip {
			ok, va := GetIP属地(v.Ip)
			if !ok {
				continue
			}
			v.S属地 = va
			有更改 = true
		}
		if !有更改 {
			continue
		}
		var 需要查的ip2 = make(map[string]*ip属地数据结构, 0)
		for _, v := range 需要查的ip {
			if v.S属地 == "" {
				continue
			}
			if _, ok := 需要查的ip2[v.Ip]; !ok {
				需要查的ip2[v.Ip] = &ip属地数据结构{}
				需要查的ip2[v.Ip].Ip = v.Ip
				需要查的ip2[v.Ip].S属地 = v.S属地
			}
		}
		ips属地锁.Lock()
		for _, v := range 需要查的ip2 {
			ips属地[v.Ip].S属地 = v.S属地
		}
		str := ""
		for _, v := range ips属地 {
			str += v.S属地 + "\r\n"
		}
		ips属地锁.Unlock()
		ips属地信息.Set(str)
	}
} //func 获取ips属地() {
func 获取访问连接中的ip(访问连接 string) string {
	if strings.Contains(访问连接, "[::1]") {
		return "[::1]"
	}
	if strings.Contains(访问连接, "localhost") {
		return "localhost"
	}
	if strings.Contains(访问连接, "127.0.0.1") {
		return "127.0.0.1"
	}
	return 访问连接[:strings.LastIndex(访问连接, ":")]
}

type 连接信息2 struct {
	L连接时间           string `json:"连接时间"`
	L连接时刻秒          int64  `json:"连接时刻秒"`
	D登录时间           string `json:"登录时间"`
	D登录时刻           int64  `json:"登录时刻"`
	T退出登录时间         string `json:"退出登录时间"`
	T统计             int    `json:"统计"`
	Xlsx            int    `json:"xlsx"`
	Token           int    `json:"token"`
	GetTagList      int    `json:"GetTagList"`
	GetKVTagsValue  int    `json:"GetKVTagsValue"`
	SetKVTagsValue  int    `json:"SetKVTagsValue"`
	SetTagValue     int    `json:"SetTagValue"`
	GetTagValue     int    `json:"GetTagValue"`
	SQLiteDB复制      int    `json:"SQLiteDB复制"`
	SQLiteDB查询      int    `json:"SQLiteDB查询"`
	S设定数据库信息        int    `json:"S设定数据库信息"`
	SetTagValue请求信息 string `json:"SetTagValue请求信息"`
	Login           int    `json:"login"`
	Upload          int    `json:"upload"`
	T调试             int    `json:"调试"`
	G关闭             int    `json:"关闭"`
	C重启             int    `json:"重启"`
	C重启服务器          int    `json:"重启服务器"`
	G根目录            int    `json:"根目录"`
	W文件浏览           int    `json:"文件浏览"`
	X下载文件           int    `json:"下载文件"`
	F发邮件            int    `json:"发邮件"`
	S数据             int    `json:"数据"`
} //type 连接信息 struct {
var 连接信息 map[string]*连接信息2
var 连接登录时刻 map[string]int64

func Float32ToByte(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	//binary.LittleEndian.PutUint32(bytes, bits)
	binary.BigEndian.PutUint32(bytes, bits)
	return bytes
}
func ByteToFloat32(bytes []byte) float32 {
	//bits := binary.LittleEndian.Uint32(bytes)
	bits := binary.BigEndian.Uint32(bytes)
	return math.Float32frombits(bits)
}

// 多个[]byte数组合并成一个[]byte
func BytesCombine(pBytes ...[]byte) []byte {
	len := len(pBytes)
	s := make([][]byte, len)
	for index := 0; index < len; index++ {
		s[index] = pBytes[index]
	}
	sep := []byte("")
	return bytes.Join(s, sep)
}

//	func 测试() {
//		//var a map[[]byte][]byte
//	}
//
// //多项式x16+x15+x2+1; 11000000000000101; 0x18005; 简记0x8005; 逆式0xa001;
var token服务器地址 string

type token2 struct {
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
}

var token3 token2
var errcode errcode2

type errcode2 struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

var token已编的码 []byte
var token已编的码锁 sync.RWMutex

func 将token编码(错误码 int, 错误信息 string) {
	access_token锁.Lock()
	j := token3.Expires_in
	d, err3 := json.Marshal(token3)
	if j <= 0 {
		if errcode.Errcode == 0 {
			errcode.Errcode = 错误码
			errcode.Errmsg = 错误信息
		}
		d, err3 = json.Marshal(errcode)
	}
	access_token锁.Unlock()
	if err3 != nil {
		fmt.Println("\r\ntoken服务器执行 json Marshal(token3或者errcode) error !" + err3.Error())
		return
	}
	token已编的码锁.Lock()
	token已编的码 = d
	token已编的码锁.Unlock()
} //func 将token编码(错误码 int, 错误信息 string) {
var 上次成功获取token时刻秒 int64

func 获取token() { //无需判断能否上外网，因为有时token服务器地址在局域网内
	access_token锁.Lock()
	errcode.Errcode = -9
	errcode.Errmsg = "开始获取"
	access_token锁.Unlock()
	将token编码(errcode.Errcode, errcode.Errmsg)
	if 遍历项目变量信息表中 {
		access_token锁.Lock()
		errcode.Errmsg = "遍历项目变量信息表中,等待获取本网关获取token的服务器地址！"
		token3.Expires_in = -1
		access_token锁.Unlock()
		将token编码(errcode.Errcode, errcode.Errmsg)
		return
	}
	if token服务器地址 == "" {
		access_token锁.Lock()
		errcode.Errmsg = "本网关设置表中没有token服务器地址！"
		token3.Expires_in = -2
		access_token锁.Unlock()
		将token编码(errcode.Errcode, errcode.Errmsg)
		return
	}
	if !strings.Contains(token服务器地址, "https://") && !strings.Contains(token服务器地址, "http://") {
		access_token锁.Lock()
		errcode.Errmsg = "无效的token服务器地址：" + token服务器地址
		token3.Expires_in = -3
		access_token锁.Unlock()
		将token编码(errcode.Errcode, errcode.Errmsg)
		return
	}
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", token服务器地址, nil)
	if err != nil {
		access_token锁.Lock()
		errcode.Errmsg = err.Error()
		token3.Expires_in = -4
		access_token锁.Unlock()
		将token编码(errcode.Errcode, errcode.Errmsg)
		return
	}
	req.Header.Set("User-Agent", "www.pdlei.cn")
	resp, err1 := client.Do(req)
	if err1 != nil {
		access_token锁.Lock()
		errcode.Errmsg = err1.Error()
		token3.Expires_in = -5
		access_token锁.Unlock()
		将token编码(errcode.Errcode, errcode.Errmsg)
		return
	}
	defer resp.Body.Close()
	bodyText, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		access_token锁.Lock()
		errcode.Errmsg = err2.Error()
		token3.Expires_in = -6
		access_token锁.Unlock()
		将token编码(errcode.Errcode, errcode.Errmsg)
		return
	}
	var token4 token2
	var errcode4 errcode2
	err3 := json.Unmarshal(bodyText, &token4)
	if err3 != nil {
		access_token锁.Lock()
		errcode.Errmsg = err3.Error()
		token3.Expires_in = -7
		access_token锁.Unlock()
		将token编码(errcode.Errcode, errcode.Errmsg)
		return
	}
	access_token锁.Lock()
	token3.Expires_in = token4.Expires_in
	token3.Access_token = token4.Access_token
	access_token锁.Unlock()
	err3 = json.Unmarshal(bodyText, &errcode4)
	if err3 != nil {
		access_token锁.Lock()
		errcode.Errmsg = err3.Error()
		token3.Expires_in = -8
		access_token锁.Unlock()
		将token编码(errcode.Errcode, errcode.Errmsg)
		return
	}
	access_token锁.Lock()
	errcode.Errmsg = errcode4.Errmsg
	errcode.Errcode = errcode4.Errcode
	access_token锁.Unlock()
	str := "system error"
	if strings.Contains(errcode.Errmsg, str) {
		access_token锁.Lock()
		errcode.Errmsg = "system error,微信服务器忙！"
		token3.Expires_in = -9
		access_token锁.Unlock()
		将token编码(errcode.Errcode, errcode.Errmsg)
		return
	}
	将token编码(errcode.Errcode, errcode.Errmsg)
	上次成功获取token时刻秒 = time.Now().Unix()
} //func 获取token() { //无需判断能否上外网，因为有时token服务器地址在局域网内
var access_token锁 = &sync.Mutex{}

func 紧急获取token2() {
	access_token锁.Lock()
	errcode.Errcode = -10
	errcode.Errmsg = "还没获取"
	access_token锁.Unlock()
	将token编码(errcode.Errcode, errcode.Errmsg)
	ticker := time.NewTicker(10 * time.Second) //由每天可获取token次数决定
	defer ticker.Stop()
	for range ticker.C {
		//  //无需判断能否上外网，因为有时token服务器地址在局域网内
		//
		//
		// if atomic.LoadUint32(&能上网) == 0 {
		// 	continue
		// }
		access_token锁.Lock()
		j := token3.Expires_in
		access_token锁.Unlock()
		if time.Now().Unix()-上次成功获取token时刻秒 < int64(j) {
			continue
		}
		获取token()
	}
}

var 数据网关状态微信报警推送上次成功时刻秒 int64
var 数据网关状态微信报警推送错误信息 = 字符串不重复累加互斥锁访问结构体{内容: "还未推送过"}

func 获取服务器内存使用率() string {
	type LSysInfo struct {
		MemAll         uint64
		MemFree        uint64
		MemUsed        uint64
		MemUsedPercent float64
	}
	var info LSysInfo
	unit := uint64(1024 * 1024) // MB
	v, _ := mem.VirtualMemory()
	info.MemAll = v.Total
	info.MemFree = v.Free
	info.MemUsed = info.MemAll - info.MemFree
	// 注:使用SwapMemory或VirtualMemory，在不同系统中使用率不一样，因此直接计算一次
	info.MemUsedPercent = float64(info.MemUsed) / float64(info.MemAll) * 100.0
	info.MemAll /= unit
	info.MemUsed /= unit
	info.MemFree /= unit
	// cpu, ram := 获取进程资源占用情况(获取当前进程名称())
	// cpu1 := "cpu " + strconv.FormatFloat(cpu, 'f', 2, 32) + "%;"
	// ram1 := "ram " + strconv.FormatFloat(float64(ram), 'f', 2, 32) + "%"
	// str := cpu1 + ram1
	return "总内存" + strconv.FormatUint(info.MemAll, 10) + "MB" + "\r\n" +
		// "已使用内存" + strconv.FormatUint(info.MemUsed, 10) + "MB" + "\r\n" +
		// "空闲内存" + strconv.FormatUint(info.MemFree, 10) + "MB" + "\r\n" +
		"使用率" + strconv.FormatFloat(info.MemUsedPercent, 'f', 2, 64) + "%"
} //func 获取服务器内存使用率() string {
var 服务器信息 = ""

func 获取服务器信息() string {
	if 服务器信息 != "" {
		return 服务器信息
	} else {
		cmd := exec.Command("wmic", "computerSystem", "get", "username,manufacturer,model")
		out, _ := cmd.Output()
		服务器信息 = strings.TrimSpace(string(out))
		return 服务器信息
	}
} //func 获取服务器信息() string {
var 当前进程名称及路径 = ""

func 获取当前进程名称及路径() string {
	if 当前进程名称及路径 != "" {
		return 当前进程名称及路径
	}
	// 获取进程自身的完整路径
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	// 获取进程名称（不包含路径）
	当前进程名称及路径 = path.Base(exePath)
	return 当前进程名称及路径
} //func 获取当前进程名称及路径() string {
// var 当前进程名称 = ""
//
//	func 获取当前进程名称() string {
//		if 当前进程名称 != "" {
//			return 当前进程名称
//		}
//		当前进程名称 = 获取当前进程名称及路径()
//		index := strings.LastIndex(当前进程名称, "\\")
//		当前进程名称 = 当前进程名称[index+1:]
//		return 当前进程名称
//	} //func 获取当前进程名称() string {
func 获取进程资源占用情况并判断和发邮件() string {
	str1 := ""
	cpu, ram := 获取进程资源占用情况(程序名)
	cpu1 := "cpu " + strconv.FormatFloat(cpu, 'f', 2, 32) + "%;"
	ram1 := "ram " + strconv.FormatFloat(float64(ram), 'f', 2, 32) + "%"
	str := cpu1 + ram1 + "\r\n"
	str1 += str
	if ram > 本软件内存占用率最大值 {
		str = "本软件内存占用率(" + ram1 + ")已经超过设定值(" + 本软件内存占用率最大值_s + ")" + "软件将被重启！\r\n"
		str1 += str
		if atomic.LoadUint32(&已经触发重启程序了) == 0 {
			//批处理文件名非中文提示 = "内存百分比连续超过设定次数软件重启，若反复重启，请人工解决系统内存占用过高问题！"
			邮件主题 := "本软件内存占用率已经超过设定值," + 发邮件主题.Load()
			邮件正文 := 邮件主题 + "\r\n" + str + "\r\n内存占用率高数据分析：" + 内存占用率高数据分析()
			go 报警发邮件函数(自定义发送人, 自定义发邮件token, 自定义邮件接收者, 邮件主题, 邮件正文)
			批处理文件名非中文提示 := "JinChengNeiCunZhanYongLvChaoGuoSheDingZhi"
			go 重新启动外部程序(批处理文件名非中文提示, 程序名, 程序文件名及路径)
		}
	}
	return str1
} //func 获取进程资源占用情况并判断和发邮件() string {
func 数据网关状态微信报警内容收集() string {
	str1 := ""
	str := 服务器访问连接.Load() + "\r\n"
	str1 += str
	str = 获取服务器信息() + "\r\n"
	str1 += str
	str = 获取当前进程名称及路径() + "\r\n"
	str1 += str
	str = "工作目录：" + 目录 + "\r\n"
	str1 += str
	//GetSysInfo()
	str = 获取服务器内存使用率() + "\r\n"
	str1 += str
	str = 获取进程资源占用情况并判断和发邮件()
	str1 += str
	str = GetDiskInfoWindows()
	if str != "" {
		str1 += str
	}
	str = 获取文件绝对路径及文件名(目录+SQLiteDB磁盘查询历史数据库文件名) + "(" + 获取文件大小(目录+SQLiteDB磁盘查询历史数据库文件名) + ")\r\n"
	str1 += str
	SQLitedbPath := SQLiteDB库文件备份目录信息.Load()
	str = "SQLitedb备份目录: " + "\r\n" + SQLitedbPath + "\r\n"
	str1 += str
	str = "SQLitedb备份目录记录在: " + "\r\n" + 目录 + SQLiteDB库文件备份目录信息文件名 + "\r\n"
	str1 += str
	SQLitedbPath = SQLiteDB文件名及路径信息.Load()
	str = "SQLitedbPath: " + "\r\n" + SQLitedbPath + "(" + 获取文件大小(SQLitedbPath) + ")\r\n"
	str1 += str
	str = "SQLitedbPath记录在: " + "\r\n" + 目录 + SQLiteDB文件名及路径信息文件名 + "\r\n"
	str1 += str
	SQLiteDB批量插入结果.锁.Lock()
	最小耗时记录数 := SQLiteDB批量插入结果.SQLiteDB批量插入最小耗时记录数
	最小耗时 := fmt.Sprintf("%v", SQLiteDB批量插入结果.SQLiteDB批量插入最小耗时)
	平均记录次最小耗时 := SQLiteDB批量插入结果.平均记录次最小耗时
	最小耗时发生在 := SQLiteDB批量插入结果.最小耗时发生时间
	SQLiteDB批量插入结果.锁.Unlock()
	str = "SQLiteDB批量插入情况: " + "\r\n"
	str1 += str
	str = strconv.Itoa(int((最小耗时记录数)))
	str = "最小耗时记录数: " + str + "\r\n"
	str1 += str
	str = "最小耗时: " + 最小耗时 + "\r\n"
	str1 += str
	str = "最小耗时发生在: " + 最小耗时发生在 + "\r\n"
	str1 += str
	str = "单条记录最小耗时: " + 平均记录次最小耗时 + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&SQLiteDB第几次批量插入)))
	str = "第几次批量插入: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&SQLiteDB插入记录总数)))
	str = "插入记录总数: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&SQLiteDB表格记录总数)))
	str = "库表格记录总数: " + str + "\r\n"
	str1 += str
	SQLiteDB访问锁.Lock()
	SQLiteDB表格记录总数查询耗时2 := SQLiteDB表格记录总数查询耗时
	SQLiteDB访问锁.Unlock()
	str = SQLiteDB表格记录总数查询耗时2 + "\r\n"
	str1 += str
	var 记录数 uint64 = 0
	等待插入SQLiteDB的记录们.Range(func(key, value interface{}) bool {
		记录数++
		return true // 继续遍历
	})
	str = strconv.Itoa(int(atomic.LoadUint64(&记录数)))
	str = "等待插入记录总数: " + str + "\r\n"
	str1 += str
	str = "SQLiteDB查询情况: " + "\r\n"
	str1 += str
	SQLiteDB使用源库查询情况.锁.Lock()
	str = "查询时间: " + SQLiteDB使用源库查询情况.查询时间 + "\r\n"
	str1 += str
	str = "磁盘查询历史数据库记录总数: " + SQLiteDB使用源库查询情况.SQLiteDB磁盘查询历史数据库记录总数 + "\r\n"
	str1 += str
	str = "内存数据库记录总数: " + SQLiteDB使用源库查询情况.SQLiteDB内存数据库记录总数 + "\r\n"
	str1 += str
	str = "查询变量个数: " + SQLiteDB使用源库查询情况.查询变量个数 + "\r\n"
	str1 += str
	str = "查询的url:\r\n" + SQLiteDB使用源库查询情况.查询的url + "\r\n"
	str1 += str
	str = "查询结果记录总数: " + SQLiteDB使用源库查询情况.查询结果记录总数 + "\r\n"
	str1 += str
	str = "查询耗时: " + SQLiteDB使用源库查询情况.查询耗时 + "\r\n"
	str1 += str
	str = "平均查询每记录耗时: " + SQLiteDB使用源库查询情况.平均查询每记录耗时 + "\r\n"
	str1 += str
	str = "访问总耗时: " + SQLiteDB使用源库查询情况.访问总耗时 + "\r\n"
	str1 += str
	str = "写内存库记录数: " + SQLiteDB使用源库查询情况.写内存库记录数 + "\r\n"
	str1 += str
	str = "查询人: " + SQLiteDB使用源库查询情况.查询人 + "\r\n"
	str1 += str
	SQLiteDB使用源库查询情况.锁.Unlock()
	str = "SQLiteDB复制情况: " + "\r\n" + SQLiteDB复制情况.Load()
	str1 += str
	str = "被访问次数: " + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&统计被访问次数)))
	str = "统计: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&数据被访问次数)))
	str = "数据: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&xlsx被访问次数)))
	str = "xlsx: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&token被访问次数)))
	str = "token: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&GetTagList被访问次数)))
	str = "GetTagList: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&GetKVTagsValue被访问次数)))
	str = "api/GetKVTagsValue: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&SetKVTagsValue被访问次数)))
	str = "api/SetKVTagsValue: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&SetTagValue被访问次数)))
	str = "SetTagValue: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&GetTagValue被访问次数)))
	str = "GetTagValue: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&SQLiteDB复制被访问次数)))
	str = "SQLiteDB复制: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&SQLiteDB查询被访问次数)))
	str = "SQLiteDB查询: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&login被访问次数)))
	str = "login: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&上传被访问次数)))
	str = "upload: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&重启被访问次数)))
	str = "重启: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&发邮件被访问次数)))
	str = "发邮件: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&重启服务器被访问次数)))
	str = "重启服务器: " + str + "\r\n"
	str1 += str
	str = "目录: " + "当前无法统计\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&调试被访问次数)))
	str = "调试: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&关闭被访问次数)))
	str = "关闭: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&根目录被访问次数)))
	str = "/: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&文件浏览被访问次数)))
	str = "文件浏览: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&下载文件被访问次数)))
	str = "下载文件: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(项目变量信息表格行数 - 1)
	str = "变量数: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(项目变量信息表格行数 - 1 - 没有错误信息行数)
	str = "错误变量数: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint64(&遍历项目变量信息表保存次)))
	str = "变量信息表保存次: " + str + "\r\n"
	str1 += str
	str = strconv.Itoa(int(atomic.LoadUint32(&项目变量信息内存表保存次数)))
	str = "变量信息内存表保存次: " + str + "\r\n"
	str1 += str
	str = 写项目变量信息表并保存时发生的错误.Load()
	if str != "" {
		str1 += str + "\r\n"
	}
	str = strconv.FormatUint(atomic.LoadUint64(&第几次采集), 10)
	str = "第几次采集: " + str + "\r\n"
	str1 += str
	str = strconv.FormatUint(atomic.LoadUint64(&第几次写设备操作), 10)
	str = "第几次写设备操作: " + str + "\r\n"
	str1 += str
	str = strconv.FormatInt(int64(len(连接信息)), 10)
	str = "访问连接数：" + str + "\r\n"
	str1 += str
	str = 编译时间 + "\r\n"
	str1 += str
	str = 至今多少天时分秒(网关启动时刻)
	str = "启动至今: " + str + "\r\n"
	str1 += str
	str = 网关启动时间 + "\r\n"
	str1 += str
	str = time.Now().Format("2006-01-02 15:04:05") + "\r\n"
	str1 += str
	str = "ip属地信息：" + "\r\n"
	str1 += str
	str = ips属地信息.Load()
	str1 += str
	// str1 = strings.Replace(str1, "\r\n", ";", -1)
	// str1 = strings.Replace(str1, " ", "", -1)
	return str1
} //func 数据网关状态微信报警内容收集()string{
var 已经触发重启程序了 uint32 = 0

func 重新启动外部程序(批处理文件名非中文提示, 程序名字, 程序文件名字及路径 string) {
	if atomic.LoadUint32(&已经触发重启程序了) == 1 {
		return
	}
	写项目变量信息表并保存()
	时间 := strings.Replace(time.Now().Format("2006-01-02 15:04:05"), ":", "", -1)
	文件名 := 批处理文件名非中文提示 + "_" + 程序名字 + "_重启_" + 时间 + ".bat"
	内容 := "set SLEEP=ping 127.0.0.1 /n" + "\r\n" + "%SLEEP% 4 > nul" + "\r\n" + "taskkill.exe /f /im " + 程序名字 + "\r\n" + "%SLEEP% 4 > nul" + "\r\n" + "start " + 程序文件名字及路径 + "\r\n" + "exit\r\n"
	data1, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(内容))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	创建文件2(目录+"bat/", 文件名, data1)
	c := exec.Command(目录 + "bat/" + 文件名)
	atomic.StoreUint32(&已经触发重启程序了, 1)
	SQLiteDB访问锁.Lock()
	defer SQLiteDB访问锁.Unlock()
	写项目变量信息表并保存锁.Lock()
	defer 写项目变量信息表并保存锁.Unlock()
	if SQLiteDB内存数据库连接 != nil {
		SQLiteDB内存数据库连接.Close()
	}
	if SQLiteDB磁盘查询历史数据库连接 != nil {
		SQLiteDB磁盘查询历史数据库连接.Close()
	}
	c.Start()
} //func 启动外部程序(程序名字 string) {
func 项目变量信息表写保存() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if atomic.LoadUint32(&已经触发重启程序了) == 1 {
			continue
		}
		写项目变量信息表并保存()
	}
}

var webhook_url string = webhook_url_默认

func 企业微信报警系统状态推送() {
	atomic.StoreInt64(&数据网关状态微信报警推送上次成功时刻秒, time.Now().Unix())
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if atomic.LoadUint32(&已经触发重启程序了) == 1 {
			continue
		}
		if atomic.LoadUint32(&能上网) == 0 {
			continue
		}
		if atomic.LoadInt64(&数据网关状态微信报警推送上次成功时刻秒) == 0 {
			fmt.Println("数据网关状态微信报警推送上次成功时刻秒为0")
			atomic.StoreInt64(&数据网关状态微信报警推送上次成功时刻秒, time.Now().Unix())
			continue
		}
		if time.Now().Unix()-atomic.LoadInt64(&数据网关状态微信报警推送上次成功时刻秒) < 企业微信报警系统状态推送最小间隔秒数 {
			continue
		}
		执行结果 := 企业微信报警系统状态推送1("")
		if 执行结果 == "ok" {
			atomic.StoreInt64(&数据网关状态微信报警推送上次成功时刻秒, time.Now().Unix()+企业微信报警系统状态推送间隔秒数)
			数据网关状态微信报警推送错误信息.Set("ok")
			continue
		}
		atomic.StoreInt64(&数据网关状态微信报警推送上次成功时刻秒, time.Now().Unix()+企业微信报警系统状态推送最小间隔秒数)
		数据网关状态微信报警推送错误信息.Set(执行结果)
	}
} //func 企业微信报警系统状态推送() {
func 企业微信报警系统状态推送1(推送的内容 string) string {
	type 内容2 struct {
		Content string `json:"content"`
	}
	type 内容1 struct {
		Msgtype string `json:"msgtype"`
		Text    内容2    `json:"text"`
	}
	var 内容 内容1
	type errcode2 struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	var Errcode errcode2
	内容.Msgtype = "text"
	if 推送的内容 == "" {
		内容.Text.Content = 数据网关状态微信报警内容收集()
	} else {
		内容.Text.Content = 推送的内容
	}
	url := webhook_url
	client := &http.Client{Timeout: 5 * time.Second}
	jsonStr, err3 := json.Marshal(内容)
	if err3 != nil {
		return "json.Marshal(Data)错误"
	}
	//fmt.Println(string(jsonStr))
	data := bytes.NewBuffer(jsonStr)
	req, err4 := http.NewRequest("POST", url, data)
	if err4 != nil {
		return "http.NewRequest错误"
	}
	resp, err5 := client.Do(req)
	if err5 != nil {
		fmt.Println(err5.Error())
		return "向企业微信群机器人写数据无法到达"
	}
	defer resp.Body.Close()
	input, err1 := io.ReadAll(resp.Body)
	if err1 != nil {
		return err1.Error()
	}
	err2 := json.Unmarshal(input, &Errcode)
	if err2 != nil {
		return err2.Error()
	}
	if Errcode.Errcode != 0 {
		return Errcode.Errmsg
	}
	return "ok"
} //func 企业微信报警系统状态推送1() {
func Get外网IP() string {
	//str := "https://2024.ipchaxun.com/"
	str := "https://www.ip.cn/api/index?ip&type=0"
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", str, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "www.pdlei.cn")
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	input, err := io.ReadAll(resp.Body) //读取流数据
	if err != nil {
		return ""
	}
	//{"ret":"ok","ip":"222.216.41.53","data":["中国","广西","南宁","青秀","电信","530000","0771"]}
	//{"rs":1,"code":0,"address":"中国  广西 南宁市 电信","ip":"222.216.41.53","isDomain":0}
	var result interface{}
	err = json.Unmarshal(input, &result) //解析json数据
	if err != nil {
		return ""
	}
	if data, ok := result.(map[string]interface{}); ok {
		if ip, ok := data["ip"].(string); ok {
			return ip
		}
	}
	return ""
} //func Get外网IP()string {
func 报警发邮件函数(自定义发送人 string, 自定义发邮件token string, 自定义邮件接收者 string, 发邮件主题 string, 邮件正文 string) {
	if 自定义发送人 == "" {
		return
	}
	if 自定义发邮件token == "" {
		return
	}
	if 自定义邮件接收者 == "" {
		return
	}
	if 发邮件主题 == "" {
		return
	}
	if 邮件正文 == "" {
		return
	}
	for {
		if atomic.LoadUint32(&能上网) == 0 {
			time.Sleep(time.Second * 1)
			continue
		}
		break
	}
	for {
		if time.Now().Unix()-atomic.LoadInt64(&上次发邮件时刻) < 发邮件时间间隔秒数 {
			time.Sleep(time.Second * 1)
			continue
		}
		atomic.StoreInt64(&上次发邮件时刻, time.Now().Unix())
		break
	}
	邮件头 := make(map[string][]string)
	邮件头["To"] = strings.Split(strings.Replace(自定义邮件接收者, " ", "", -1), ",")
	邮件头["From"] = []string{自定义发送人}
	邮件头["Subject"] = []string{发邮件主题}
	m := gomail.NewMessage()
	m.SetHeaders(邮件头)
	邮件正文 = strings.Replace(邮件正文, "\r\n", "<br>", -1)
	邮件正文 = strings.Replace(邮件正文, "\r", "<br>", -1)
	邮件正文 = strings.Replace(邮件正文, "\n", "<br>", -1)
	m.SetBody("text/html", 邮件正文)
	d := gomail.NewDialer("smtp.qq.com", 587, 自定义发送人, 自定义发邮件token)
	//发邮件错误信息 = "\r\n上次发邮件错误信息： "
	if err := d.DialAndSend(m); err != nil {
		//发邮件错误信息 = "\r\n上次发邮件错误信息： " + err.Error()
		fmt.Println(err)
	}
} //func 报警发邮件函数(){
func 复制文件(src, dst string) error {
	// 打开源文件
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	// 创建目标文件
	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()
	// 复制文件内容
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}
	// 确保所有缓存的数据被写入到文件中
	err = destinationFile.Sync()
	return err
} //func 复制文件(src, dst string) error {
func 检查token是否失效(id string) string {
	access_token锁.Lock()
	url := "https://api.weixin.qq.com/cgi-bin/user/info?access_token=" + token3.Access_token + "&openid=" + id + "&lang=zh_CN"
	access_token锁.Unlock()
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err.Error()
	}
	req.Header.Set("User-Agent", "www.pdlei.cn")
	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	input, err := io.ReadAll(resp.Body) //读取流数据
	if err != nil {
		return err.Error()
	}
	var result interface{}
	err = json.Unmarshal(input, &result) //解析json数据
	if err != nil {
		return err.Error()
	}
	if data, ok := result.(map[string]interface{}); ok {
		if errmsg, ok := data["errmsg"].(string); ok {
			if strings.Contains(errmsg, "access_token") {
				access_token锁.Lock()
				token3.Access_token = "access_token失效了"
				token3.Expires_in = -1
				access_token锁.Unlock()
				fmt.Println("检查token失效了")
				return "access_token"
			}
		} //if errmsg, ok := data["errmsg"].(string); ok {
	}
	return "检查token是否失效结束"
} //func 检查token是否失效(p *Row, id string) string {
func 检查token是否失效1() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if atomic.LoadUint32(&能上网) == 0 {
			continue
		}
		access_token锁.Lock()
		j := token3.Expires_in
		access_token锁.Unlock()
		if j <= 0 {
			continue
		}
		//fmt.Println(检查token是否失效(检查token是否失效使用的默认openid))
		检查token是否失效(检查token是否失效使用的默认openid)
	}
} //func 检查token是否失效1() {
func init() {
	// 禁用快速编辑模式
	inHandle := windows.Handle(os.Stdin.Fd())
	var inMode uint32
	if err := windows.GetConsoleMode(inHandle, &inMode); err != nil {
		fmt.Println("Error getting console mode:", err)
		return
	}
	inMode &^= windows.ENABLE_QUICK_EDIT_MODE
	if err := windows.SetConsoleMode(inHandle, inMode); err != nil {
		fmt.Println("Error setting console mode:", err)
	}
	fmt.Println("setting console mode:", inMode)
}

type 字符串互斥锁访问结构体 struct {
	内容 string
	锁  sync.Mutex
}

func (s *字符串互斥锁访问结构体) Set(str string) {
	s.锁.Lock()
	defer s.锁.Unlock()
	s.内容 = str
}
func (s *字符串互斥锁访问结构体) Load() string {
	s.锁.Lock()
	defer s.锁.Unlock()
	return s.内容
}

type 字符串累加互斥锁访问结构体 struct {
	内容 string
	锁  sync.Mutex
}

func (s *字符串累加互斥锁访问结构体) Set(str string) {
	s.锁.Lock()
	defer s.锁.Unlock()
	if s.内容 == "" {
		s.内容 = str
		return
	}
	s.内容 = str + ";" + s.内容
}
func (s *字符串累加互斥锁访问结构体) Load() string {
	s.锁.Lock()
	defer s.锁.Unlock()
	return s.内容
}

type 字符串不重复累加互斥锁访问结构体 struct {
	内容 string
	锁  sync.Mutex
}

func (s *字符串不重复累加互斥锁访问结构体) Set(str string) {
	s.锁.Lock()
	defer s.锁.Unlock()
	if s.内容 == "" {
		s.内容 = str
		return
	}
	if !strings.Contains(s.内容, str) {
		s.内容 = str + ";" + s.内容
	}
}
func (s *字符串不重复累加互斥锁访问结构体) Load() string {
	s.锁.Lock()
	defer s.锁.Unlock()
	return s.内容
}
func 修正邮件接收者(设定值 string) string {
	设定值 = strings.Replace(设定值, "，", ",", -1)
	设定值 = strings.Replace(设定值, ";", ",", -1)
	设定值 = strings.Replace(设定值, "；", ",", -1)
	邮箱们 := strings.Split(设定值, ",")
	不重复的邮箱们 := make(map[string]bool)
	合法邮箱们 := make([]string, 0)
	for _, 邮箱 := range 邮箱们 {
		_, err := mail.ParseAddress(strings.TrimSpace(邮箱))
		if err != nil {
			continue
		}
		if _, ok := 不重复的邮箱们[邮箱]; !ok {
			不重复的邮箱们[邮箱] = true
			合法邮箱们 = append(合法邮箱们, 邮箱)
		}
	}
	设定值 = strings.Join(合法邮箱们, ",")
	return 设定值
} //func 修正邮件接收者(设定值 string) string {
func isValidUUIDv4(uuid string) bool {
	// UUID version 4 regex pattern
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	re := regexp.MustCompile(pattern)
	// Check if the string matches the regex pattern
	return re.MatchString(strings.ToLower(uuid))
}
func 企业微信信息推送机器人检查(设定值 string) string {
	str := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="
	if !strings.Contains(设定值, str) {
		return ""
	}
	uuid1 := strings.Split(设定值, "=")
	uuid := uuid1[len(uuid1)-1]
	if !isValidUUIDv4(uuid) {
		return ""
	}
	return 设定值
} //func 企业微信信息推送机器人检查(设定值 string) string {
func 微信报警推送需要的token获得来源检查(设定值 string) string {
	// if 设定值!= "" {
	// //if len(设定值) != 16 {
	// 	return 设定值
	// }
	return 设定值
} //func 微信报警推送需要的token获得来源检查(设定值 string) string {
func 发邮件token检查(设定值 string) string {
	// if 设定值!= "" {
	// //if len(设定值) != 16 {
	// 	return 设定值
	// }
	return 设定值
} //func 发邮件token检查(设定值 string) string {
var SetTagValue被访问次数 uint64

func SetTagValue(w http.ResponseWriter, r *http.Request) {
	go 获取服务器域名(r.Host, r.RemoteAddr)
	ip := 设定的连接信息处理(w, r, &SetTagValue被访问次数, "SetTagValue")
	if ip == "" {
		return
	}
	type ErrCode struct {
		Code    int    `json:"Code"`
		Message string `json:"Message"`
	}
	query := r.URL.Query()
	user := query.Get("user")
	password := query.Get("password")
	var C错误码 ErrCode
	if user != 本网关登录名之MD5 || password != 本网关登录密码之MD5 {
		C错误码.Code = 1
		C错误码.Message = "用户名或密码错误！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	strTagName := query.Get("strTagName")
	if strTagName == "" {
		C错误码.Code = 2
		C错误码.Message = "strTagName不能为空" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	strSetTagValue := query.Get("strSetTagValue")
	if strSetTagValue == "" {
		C错误码.Code = 3
		C错误码.Message = "strSetTagValue不能为空" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	检查结果 := 变量名检查(strTagName)
	if 检查结果 != "ok" {
		C错误码.Code = 4
		C错误码.Message = strTagName + " 请求的变量名中有错误，不符合变量名命名规范！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if _, ok := 变量名所在行[strTagName]; !ok {
		C错误码.Code = 5
		C错误码.Message = "请求的变量名中有错误，系统中不存在这些变量名！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	判断结果 := 写值是否合法(strTagName, strSetTagValue)
	if 判断结果 != "ok" {
		C错误码.Code = 6
		C错误码.Message = 判断结果 + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	var 执行结果 字符串累加互斥锁访问结构体
	将写操作记录到相关map中(nil, &执行结果, strTagName, strSetTagValue)
	C错误码.Code = 0
	C错误码.Message = 执行结果.Load() + "_您的连接: " + r.RemoteAddr
	fmt.Println(C错误码.Message)
	jsonStr, err3 := json.Marshal(C错误码)
	if err3 != nil {
		w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
		return
	}
	w.Write([]byte(jsonStr))
}

var GetTagValue被访问次数 uint64

func GetTagValue(w http.ResponseWriter, r *http.Request) {
	go 获取服务器域名(r.Host, r.RemoteAddr)
	ip := 设定的连接信息处理(w, r, &GetTagValue被访问次数, "GetTagValue")
	if ip == "" {
		return
	}
	type ErrCode struct {
		Code    int    `json:"Code"`
		Message string `json:"Message"`
	}
	query := r.URL.Query()
	user := query.Get("user")
	password := query.Get("password")
	var C错误码 ErrCode
	if user != 本网关登录名之MD5 || password != 本网关登录密码之MD5 {
		C错误码.Code = 1
		C错误码.Message = "用户名或密码错误！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	忽略变量详情 := query.Get("忽略变量详情")
	strTagNameS := query.Get("strTagNameS")
	if strTagNameS == "" {
		C错误码.Code = 2
		C错误码.Message = "strTagNameS不能为空" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	strTagNameS = strings.Replace(strTagNameS, "，", ",", -1)
	请求的变量组 := strings.Split(strTagNameS, ",")
	长度 := len(请求的变量组)
	if 长度 < 1 {
		var C错误码 ErrCode
		C错误码.Code = 4
		C错误码.Message = "J没有请求的变量名" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if 长度 > 没有错误信息行数 {
		w.Write([]byte("请求的变量个数超过系统变量个数"))
		return
	}
	if 长度 > 最大读请求变量个数 {
		w.Write([]byte("请求的变量个数超过最大读请求变量个数(" + 最大读请求变量个数_s + ")"))
		return
	}
	for i := 0; i < 长度; i++ {
		if 请求的变量组[i] == "" {
			var C错误码 ErrCode
			C错误码.Code = 5
			C错误码.Message = "有空的请求变量名！" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	for i := 0; i < 长度; i++ {
		检查结果 := 变量名检查(请求的变量组[i])
		if 检查结果 != "ok" {
			var C错误码 ErrCode
			C错误码.Code = 6
			C错误码.Message = 请求的变量组[i] + " 请求的变量名中有错误，不符合变量名命名规范！" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	var 变量名们 = make(map[string]bool, 0)
	for i := 0; i < 长度; i++ {
		if !变量名们[请求的变量组[i]] {
			变量名们[请求的变量组[i]] = true
			continue
		}
		var C错误码 ErrCode
		C错误码.Code = 7
		C错误码.Message = "请求的变量名中有重复！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	} //for i := 0; i < 长度; i++ {
	for i := 0; i < 长度; i++ {
		if _, ok := 变量名所在行[请求的变量组[i]]; !ok {
			var C错误码 ErrCode
			C错误码.Code = 8
			C错误码.Message = "请求的变量名中有错误，系统中不存在这些变量名！" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	startTime := time.Now()
	变量数据 := ""
	for i := 0; i < 长度; i++ {
		变量数据 += 根据表行获得变量数据(变量名所在行[请求的变量组[i]], 忽略变量详情)
	} //for i := 0; i < 长度; i++ {
	str1 := ""
	str := 编译时间 + "\r\n"
	str1 += str
	str = 至今多少天时分秒(网关启动时刻)
	str = "启动至今: " + str + "\r\n"
	str1 += str
	str = 网关启动时间
	str1 += str
	变量个数 := fmt.Sprintf("%d", 长度)
	变量数据 = str1 + "\r\n" + time.Now().Format("2006-01-02 15:04:05") + "\r\n第几次采集： " +
		strconv.FormatUint(atomic.LoadUint64(&第几次采集), 10) + "\r\n变量个数：" + 变量个数 + "\r\n" + 变量数据
	elapsedTime := time.Since(startTime)
	最大耗时 := fmt.Sprintf("编码耗时：%v", elapsedTime)
	变量数据 = 最大耗时 + "\r\n" + 变量数据
	w.Write([]byte(变量数据))
}
func 给定字符串计算时间差是否满足要求(开始时间2, 结束时间2 string) (time.Duration, error) {
	// 定义时间格式
	const layout = "2006-01-02_15:04:05"

	// 解析时间字符串，UTC时间，中国是CTC
	开始时间, err := time.Parse(layout, 开始时间2)
	if err != nil {
		return 0, errors.New("无法解析开始时间字符串>>" + 开始时间2)
	}
	结束时间, err := time.Parse(layout, 结束时间2)
	if err != nil {
		return 0, errors.New("无法解析结束时间字符串>>" + 结束时间2)
	}
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return 0, errors.New("error loading local time zone")
	}
	当前时间 := time.Now()
	// 获取系统时区与 UTC 的时差（以秒为单位）
	_, offset := 当前时间.In(loc).Zone()
	//结束时间字符串转成字面本地时间，
	结束时间 = 结束时间.Add(-time.Duration(offset) * time.Second)
	diff := 结束时间.Sub(当前时间)
	if diff > 0 {
		return 0, errors.New("结束时间大于当前时间" + 结束时间.In(loc).String() + "_" + 当前时间.String())
	}
	开始时间 = 开始时间.Add(-time.Duration(offset) * time.Second)
	diff = 开始时间.Sub(当前时间)
	if diff >= 0 {
		return 0, errors.New("开始时间大于或等于当前时间" + 开始时间.In(loc).String() + "_" + 当前时间.String())
	}
	diff = 结束时间.Sub(开始时间)
	// 进行条件检查
	if diff > 365*24*time.Hour {
		return 0, errors.New("时间差大于一年")
	}
	if diff < 0 {
		return 0, errors.New("时间差是负数")
	}
	if diff < time.Second {
		return 0, errors.New("时间差小于1秒")
	}

	// 如果所有条件都满足，则返回nil表示成功
	return diff, nil
}

var SQLiteDB查询被访问次数 uint64

type 变量历史数据结构体 struct {
	VarValue string    `json:"V"`
	Time     time.Time `json:"T"`
}

func SQLiteDB查询(w http.ResponseWriter, r *http.Request) {
	访问开始时间 := time.Now()
	go 获取服务器域名(r.Host, r.RemoteAddr)
	ip := 设定的连接信息处理(w, r, &SQLiteDB查询被访问次数, "SQLiteDB查询")
	if ip == "" {
		return
	}
	type ErrCode struct {
		Code    int    `json:"Code"`
		Message string `json:"Message"`
	}
	query := r.URL.Query()
	user := query.Get("user")
	password := query.Get("password")
	var C错误码 ErrCode
	if user != 本网关登录名之MD5 || password != 本网关登录密码之MD5 {
		C错误码.Code = 1
		C错误码.Message = "用户名或密码错误！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	回复格式 := query.Get("回复格式")
	if 回复格式 == "" {
		C错误码.Code = 9
		C错误码.Message = "回复格式不能为空" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	回复格式 = strings.ToUpper(回复格式)
	if 回复格式 != "TEXT" && 回复格式 != "XLSX" && 回复格式 != "JSON" && 回复格式 != "GRAPH" {
		C错误码.Code = 9
		C错误码.Message = "回复格式!=TEXT&&回复格式!=XLSX&&回复格式!=JSON&&回复格式!=GRAPH(" + 回复格式 + ")_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	曲线图Y值使用实际值 := query.Get("曲线图Y值使用实际值")
	忽略特定值们 := query.Get("忽略特定值们")
	忽略特定值们 = strings.Replace(忽略特定值们, "，", ",", -1)
	忽略特定值们数组 := strings.Split(忽略特定值们, ",")
	if 忽略特定值们 != "" {
		for _, 特定值 := range 忽略特定值们数组 {
			if 特定值 == "" {
				C错误码.Code = 9
				C错误码.Message = "即然您已经尝试(url中有 &忽略特定值们=，)输入特定值，就不能有空" + "_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			_, err := strconv.ParseFloat(特定值, 64)
			if err != nil {
				C错误码.Code = 9
				C错误码.Message = "f, err := strconv.ParseFloat(特定值, 64)发生错误：" + err.Error() + "_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
		} //for _, 特定值 := range 忽略特定值们数组 {
	} //if 忽略特定值们!="" {
	忽略特定值们数组有值 := false
	if len(忽略特定值们数组) > 0 {
		忽略特定值们数组有值 = true
	}
	表格历史空数据自动填充 := query.Get("表格历史空数据自动填充")
	计算差值 := query.Get("计算差值")
	忽略曲线说明 := query.Get("忽略曲线说明")
	忽略异常值 := query.Get("忽略异常值")
	开始时间 := query.Get("开始时间")
	if 开始时间 == "" {
		C错误码.Code = 9
		C错误码.Message = "开始时间不能为空" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	结束时间 := query.Get("结束时间")
	if 结束时间 == "" {
		C错误码.Code = 10
		C错误码.Message = "结束时间不能为空" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if atomic.LoadUint32(&已经触发重启程序了) == 1 {
		C错误码.Code = 1
		C错误码.Message = "已经触发重启程序了" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	时间差, err := 给定字符串计算时间差是否满足要求(开始时间, 结束时间)
	if err != nil {
		C错误码.Code = 11
		C错误码.Message = err.Error() + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	间隔时间 := query.Get("间隔时间")
	if 间隔时间 == "" {
		C错误码.Code = 10
		C错误码.Message = "间隔时间不能为空" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	计算间隔时间 := func(s string) (time.Duration, error) {
		// parseTimeString 解析包含时间单位的字符串，并返回对应的 time.Duration
		//func parseTimeString(s string) (time.Duration, error) {
		// 定义正则表达式来匹配时间单位及其整数值
		re := regexp.MustCompile(`(\d+)(天|小时|分钟|秒)`)
		match := re.FindStringSubmatch(s)

		if len(match) != 3 {
			return 0, fmt.Errorf("字符串(%s)中没有匹配到有效的时间单位(天|小时|分钟|秒)", s)
		}

		// 提取整数值和时间单位
		numStr := match[1]
		unit := match[2]
		// fmt.Printf("数值：%s\r\n", numStr)
		// fmt.Printf("单位：%s\r\n", unit)
		// 将整数值转换为整数类型
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("无法将字符串(%s)转换为整数: %v", numStr, err)
		}
		if num == 0 {
			return 0, fmt.Errorf("间隔时间(%s)不能为零！", s)
		}
		// 根据时间单位将整数值转换为 time.Duration
		var duration time.Duration
		switch unit {
		case "天":
			duration = time.Duration(num) * 24 * time.Hour
		case "小时":
			duration = time.Duration(num) * time.Hour
		case "分钟":
			duration = time.Duration(num) * time.Minute
		case "秒":
			duration = time.Duration(num) * time.Second
		default:
			return 0, fmt.Errorf("不支持的时间单位: %s", unit)
		}

		return duration, nil
	}
	间隔时间t, err1 := 计算间隔时间(间隔时间)
	if err1 != nil {
		C错误码.Code = 2
		C错误码.Message = "计算间隔时间(间隔时间)发生错误：" + err1.Error() + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	//fmt.Printf("间隔时间t%v\r\n", 间隔时间t)
	if 间隔时间t > 时间差 {
		C错误码.Code = 2
		str := fmt.Sprintf("间隔时间(%s)(%v)大于开始时间(%s)结束时间(%s)的时间差(%v)", 间隔时间, 间隔时间t, 开始时间, 结束时间, 时间差)
		C错误码.Message = str + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if 时间差%间隔时间t != 0 {
		C错误码.Code = 2
		str := fmt.Sprintf("开始时间(%s)结束时间(%s)的时间差(%v)除以间隔时间(%s)(%v)余数不为0", 开始时间, 结束时间, 时间差, 间隔时间, 间隔时间t)
		C错误码.Message = str + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	每个变量目标记录数 := 时间差/间隔时间t + 1
	fmt.Printf("每个变量目标记录数：%d\r\n", 每个变量目标记录数)
	要查询的变量名们 := query.Get("要查询的变量名们")
	if 要查询的变量名们 == "" {
		C错误码.Code = 2
		C错误码.Message = "要查询的变量名们不能为空" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	要查询的变量名们 = strings.Replace(要查询的变量名们, "|", "|", -1)
	要查询的变量名们 = strings.Replace(要查询的变量名们, "：", ":", -1)
	要查询的变量名们 = strings.Replace(要查询的变量名们, "，", ",", -1)
	请求的变量组 := strings.Split(要查询的变量名们, "|")
	长度 := len(请求的变量组)
	if 长度 < 1 {
		var C错误码 ErrCode
		C错误码.Code = 4
		C错误码.Message = "J没有请求的变量名" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if 长度 > 没有错误信息行数 {
		w.Write([]byte("请求的变量个数超过系统变量个数"))
		return
	}
	if 长度 > 最大读请求变量个数 {
		w.Write([]byte("请求的变量个数超过最大读请求变量个数(" + 最大读请求变量个数_s + ")"))
		return
	}
	for i := 0; i < 长度; i++ {
		if 请求的变量组[i] == "" {
			var C错误码 ErrCode
			C错误码.Code = 5
			C错误码.Message = "有空的请求变量名！" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	var temp = make([]string, len(请求的变量组))
	copy(temp, 请求的变量组)
	var 各变量要忽略的特征值 = make(map[string][]string, 0)
	var 变量是累积值 = make(map[string]bool, 0)
	for i, 变量 := range temp {
		变量名和特征值们 := strings.Split(变量, ":")
		if len(变量名和特征值们) == 0 {
			var C错误码 ErrCode
			C错误码.Code = 6
			C错误码.Message = "变量名和特征值们 := strings.Split(变量, :),len(变量名和特征值们)==0" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
		长度 := len(变量名和特征值们)
		if 长度 > 2 {
			var C错误码 ErrCode
			C错误码.Code = 6
			C错误码.Message = 变量 + " 变量表达错误，应该是 变量名:1,2 1,2表示要忽略的非法历史数据" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
		请求的变量组[i] = 变量名和特征值们[0]
		变量是累积值[请求的变量组[i]] = false
		if 长度 == 1 {
			continue
		}
		变量的类型 := 项目变量信息表组.Rows[变量名所在行[请求的变量组[i]]].B变量类型
		// if strings.Contains(变量的类型, "离散") {//注释后可以计算开关次数
		// 	continue
		// }
		变量的特征值 := strings.Split(变量名和特征值们[1], ",")
		长度 = len(变量的特征值)
		if 长度 == 0 {
			var C错误码 ErrCode
			C错误码.Code = 6
			C错误码.Message = 变量 + " 变量的特征值表达错误，应该是 变量名:1,2 1,2表示要忽略的非法历史数据" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
		for _, 值 := range 变量的特征值 {
			if 值 == "" {
				var C错误码 ErrCode
				C错误码.Code = 6
				C错误码.Message = 变量 + " 变量的特征值中有空，表达错误，应该是 变量名:1,2 1,2表示要忽略的非法历史数据" + "_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			if 值 == "累积" {
				变量是累积值[请求的变量组[i]] = true
				continue
			}
			if 计算差值 == "是" || !strings.Contains(变量的类型, "字符串") {
				_, err := strconv.ParseFloat(值, 64)
				if err != nil {
					var C错误码 ErrCode
					C错误码.Code = 6
					C错误码.Message = 变量 + " 在计算差值或变量类型不是字符串类型时，变量的特征值中有非数值，表达错误，应该是 变量名:1,2 1,2表示要忽略的非法历史数据" + "_您的连接: " + r.RemoteAddr
					fmt.Println(C错误码.Message)
					jsonStr, err3 := json.Marshal(C错误码)
					if err3 != nil {
						w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
						return
					}
					w.Write([]byte(jsonStr))
					return
				}
			}
			各变量要忽略的特征值[请求的变量组[i]] = append(各变量要忽略的特征值[请求的变量组[i]], 值)
		} //for _, 值 := range 变量的特征值 {
	} //for i, 变量 := range temp {
	for i := 0; i < 长度; i++ {
		检查结果 := 变量名检查(请求的变量组[i])
		if 检查结果 != "ok" {
			var C错误码 ErrCode
			C错误码.Code = 6
			C错误码.Message = 请求的变量组[i] + " 请求的变量名中有错误，不符合变量名命名规范！" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	var 变量名们 = make(map[string]bool, 0)
	for i := 0; i < 长度; i++ {
		if !变量名们[请求的变量组[i]] {
			变量名们[请求的变量组[i]] = true
			continue
		}
		var C错误码 ErrCode
		C错误码.Code = 7
		C错误码.Message = "请求的变量名中有重复！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	} //for i := 0; i < 长度; i++ {
	for i := 0; i < 长度; i++ {
		if _, ok := 变量名所在行[请求的变量组[i]]; !ok {
			var C错误码 ErrCode
			C错误码.Code = 8
			C错误码.Message = "请求的变量名中有错误，系统中不存在这些变量名！" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	if 回复格式 == "GRAPH" || 计算差值 == "是" {
		for _, 变量名 := range 请求的变量组 {
			变量类型 := 项目变量信息表组.Rows[变量名所在行[变量名]].B变量类型
			if strings.Contains(变量类型, "字符串") {
				C错误码.Code = 8
				C错误码.Message = 变量名 + "的变量类型是" + 变量类型 + "，是无法生成曲线或计算差值的，请修改请求条件再试" + "_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
		}
	}
	loc, err := time.LoadLocation("Local")
	if err != nil {
		C错误码.Code = 19
		C错误码.Message = err.Error() + "(time.LoadLocation(Local))_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	SQLiteDB文件 := SQLiteDB文件名及路径信息.Load()
	SQLiteDB访问锁.Lock()
	defer SQLiteDB访问锁.Unlock()
	db, err := sql.Open("sqlite3", SQLiteDB文件)
	if err != nil {
		C错误码.Code = 12
		C错误码.Message = err.Error() + "(" + SQLiteDB文件 + ")_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	defer db.Close()
	// 确保数据库连接可用
	err = db.Ping()
	if err != nil {
		C错误码.Code = 16
		C错误码.Message = err.Error() + "(" + SQLiteDB文件 + ")_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	type 多个开始时间结束时间结构体 struct {
		开始时间 string
		结束时间 string
		//时间基准 time.Time
	}
	//var 多个开始时间结束时间 []多个开始时间结束时间结构体
	开始时间去掉_ := strings.Replace(开始时间, "_", " ", -1)
	结束时间去掉_ := strings.Replace(结束时间, "_", " ", -1)
	根据间隔时间获得多个开始时间结束时间 := func() []多个开始时间结束时间结构体 {
		多个开始时间结束时间 := make([]多个开始时间结束时间结构体, 0)
		// 定义时间格式
		const layout = "2006-01-02_15:04:05"
		// 解析时间字符串，UTC时间，中国是CTC
		开始时间2, err := time.Parse(layout, 开始时间)
		if err != nil {
			return 多个开始时间结束时间
		}
		// 获取系统时区与 UTC 的时差（以秒为单位）
		_, offset := time.Now().In(loc).Zone()
		//开始时间转成UTC时间，
		时间基准 := 开始时间2.Add(-time.Duration(offset) * time.Second)
		for i := 0; i < int(每个变量目标记录数); i++ {
			//开始时间3s := 时间基准.Add(-间隔时间t).In(loc).String()
			开始时间3s := 时间基准.In(loc).String()
			结束时间3s := 时间基准.Add(间隔时间t).In(loc).String()
			// if i >= (int(每个变量目标记录数) - 1) {
			// 	开始时间3s = 时间基准.Add(-间隔时间t).In(loc).String()
			// 	结束时间3s = 时间基准.In(loc).String()
			// }
			多个开始时间结束时间 = append(多个开始时间结束时间, 多个开始时间结束时间结构体{开始时间: 开始时间3s, 结束时间: 结束时间3s})
			时间基准 = 时间基准.Add(间隔时间t)
		}
		return 多个开始时间结束时间
	} //根据开始时间结束时间和间隔时间获得多个开始时间结束时间 := func(开始时间, 结束时间 string, 间隔时间 time.Duration) []多个开始时间结束时间结构体 {
	// fmt.Printf("开始时间%s,结束时间%s,间隔时间%s\r\n", 开始时间, 结束时间, 间隔时间)
	// for i, j := range 多个开始时间结束时间 {
	// 	fmt.Printf("第%d个时间段：开始时间%s,结束时间%s\r\n", i+1, j.开始时间, j.结束时间)
	// } //for i, j := range 多个开始时间结束时间 {
	var 单个开始结束时间 []多个开始时间结束时间结构体
	var 多个开始时间结束时间2 []多个开始时间结束时间结构体
	单个开始结束时间 = append(单个开始结束时间, 多个开始时间结束时间结构体{开始时间: 开始时间去掉_, 结束时间: 结束时间去掉_})
	查询参数 := make([]interface{}, 3)
	获取记录参数查询语句 := "SELECT * FROM " + SQLite表格名 + " WHERE 变量名 =? AND 时间 >= ? AND 时间 <= ?"
	问条件记录数参数查询语句 := "SELECT COUNT(*) FROM " + SQLite表格名 + " WHERE 变量名 =? AND 时间 >= ? AND 时间 <= ?"
	插入语句 := "INSERT INTO " + SQLite表格名 + " (变量名, 变量值, 时间) VALUES (?, ?, ?)"
	查询结果记录总数 := 0
	var 变量们的历史数据集合 = make(map[string][]变量历史数据结构体, 0)
	var 变量数据2 strings.Builder
	switch 回复格式 {
	case "TEXT":
		变量数据2.WriteString("id")
		变量数据2.WriteByte(',')
		变量数据2.WriteString("变量名")
		变量数据2.WriteByte(',')
		if 计算差值 == "是" {
			变量数据2.WriteString("差值")
		} else {
			变量数据2.WriteString("变量值")
		}
		变量数据2.WriteByte(',')
		变量数据2.WriteString("时间")
		变量数据2.WriteByte(',')
		变量数据2.WriteString("记录源")
		变量数据2.WriteString("\r\n")
	case "JSON":
	case "XLSX":
	case "GRAPH":
	}
	var 批量处理的数据切片 []interface{}
	要写记录数 := 0
	var 检查重复记录预编译的语句 *sql.Stmt
	var 磁盘检查重复记录预编译的语句 *sql.Stmt
	func() {
		检查重复语句 := "SELECT 1 FROM " + SQLite表格名 + " WHERE 变量名 = ? AND 时间= ?"
		检查重复记录预编译的语句, err1 = SQLiteDB内存数据库连接.Prepare(检查重复语句)
		if err1 != nil {
			fmt.Printf("内存库无法准备检查重复记录的语句: %v\r\n", err1)
			return
		}
		磁盘检查重复记录预编译的语句, err1 = SQLiteDB磁盘查询历史数据库连接.Prepare(检查重复语句)
		if err1 != nil {
			fmt.Printf("磁盘库无法准备检查重复记录的语句: %v\r\n", err1)
			return
		}
	}()
	if 检查重复记录预编译的语句 == nil || 磁盘检查重复记录预编译的语句 == nil {
		C错误码.Code = 20
		C错误码.Message = "检查重复记录预编译的语句 == nil || 磁盘检查重复记录预编译的语句 == nil_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	startTime := time.Now()
	变量是离散 := false
	for _, 变量名 := range 请求的变量组 {
		if 变量名 == "" {
			continue
		}
		前值s := ""
		变量是离散 = false
		变量特征值组, 变量有特征值 := 各变量要忽略的特征值[变量名]
		var 多个开始时间结束时间 []多个开始时间结束时间结构体
		查询参数[0] = 变量名
		var 源数据库条件记录数 int
		var 内存数据库条件记录数 int
		var 获取记录的库源头 *sql.DB
		var 使用的库文件 string
		var 记录数 int
		var 获取记录数耗时 time.Duration
		var 获取记录数耗时s string
		直接返回所有记录无需筛选 := false
		查询参数[1] = 开始时间去掉_
		查询参数[2] = 结束时间去掉_
		row := db.QueryRow(问条件记录数参数查询语句, 查询参数...)
		if err := row.Scan(&源数据库条件记录数); err != nil {
			C错误码.Code = 19
			C错误码.Message = err.Error() + "(" + SQLiteDB文件 + ")_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
		if SQLiteDB内存数据库连接 != nil {
			row := SQLiteDB内存数据库连接.QueryRow(问条件记录数参数查询语句, 查询参数...)
			if err := row.Scan(&内存数据库条件记录数); err != nil {
				fmt.Println("SQLiteDB内存数据库连接.QueryRow(问条件记录数参数查询语句, 查询参数...)_" + err.Error())
				内存数据库条件记录数 = 0
			}
		}
		if 源数据库条件记录数 == 0 && 内存数据库条件记录数 == 0 {
			break
		}
		if 源数据库条件记录数 > 内存数据库条件记录数 {
			记录数 = 源数据库条件记录数
		} else {
			记录数 = 内存数据库条件记录数
		} //if 源数据库条件记录数 > 内存数据库条件记录数 {
		var 总耗时 time.Duration
		每个变量目标记录数与变量实际记录数的倍数 := float64(每个变量目标记录数) / float64(记录数)
		if 每个变量目标记录数与变量实际记录数的倍数 > 每个变量目标记录数与变量实际记录数的倍数最大设定 { //如果不这样做，一旦有人估计查询10年1秒间隔，那将会生成海量多个开始时间结束时间，内存爆满
			if 回复格式 == "XLSX" { // || 回复格式 == "GRAPH" { //如果不这样做，一旦有人估计查询10年1秒间隔，那将会生成海量多个开始时间结束时间，电子表格行数爆满
				str := fmt.Sprintf(变量名+"的记录数(%d),每个变量目标记录数(%d),每个变量目标记录数与变量实际记录数的倍数(%.2f)大于每个变量目标记录数与变量实际记录数的倍数最大设定(%s),在回复格式为 "+回复格式+" 时不允许执行，请重新设定开始结束时间期间或间隔时间，如果还是无法解决问题，说明此变量记录到数据库的条件设定不对导致记录数太少", 记录数, int(每个变量目标记录数), 每个变量目标记录数与变量实际记录数的倍数, 每个变量目标记录数与变量实际记录数的倍数最大设定s)
				C错误码.Code = 19
				C错误码.Message = str + "_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			if 源数据库条件记录数 > 内存数据库条件记录数 {
				SQLiteDB使用源库查询情况.锁.Lock()
				平均查询每记录耗时2 := SQLiteDB使用源库查询情况.使用源库平均查询每记录耗时2
				平均查询每记录耗时 := SQLiteDB使用源库查询情况.使用源库平均查询每记录耗时2s
				SQLiteDB使用源库查询情况.锁.Unlock()
				总耗时 = 平均查询每记录耗时2 * time.Duration(源数据库条件记录数)
				if 总耗时 > SQLiteDB查询最大容忍时长秒*time.Second {
					C错误码.Code = 19
					C错误码.Message = 变量名 + "的预查询记录数(" + strconv.Itoa(源数据库条件记录数) + ")SQLiteDB使用源库查询情况.平均查询每记录耗时(" + 平均查询每记录耗时 + ")总耗时" + fmt.Sprintf("(%v)超过SQLiteDB查询最大容忍时长%d秒,请更改查询条件再试！", 总耗时, SQLiteDB查询最大容忍时长秒) + "_您的连接: " + r.RemoteAddr
					fmt.Println(C错误码.Message)
					jsonStr, err3 := json.Marshal(C错误码)
					if err3 != nil {
						w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
						return
					}
					w.Write([]byte(jsonStr))
					return
				}
			} else { //if 源数据库条件记录数 > 内存数据库条件记录数 {
				平均查询每记录耗时2 := time.Duration(SQLiteDB使用内存库平均查询每记录耗时默认us * time.Microsecond)
				总耗时 = 平均查询每记录耗时2 * time.Duration(内存数据库条件记录数)
				if 总耗时 > SQLiteDB查询最大容忍时长秒*time.Second {
					C错误码.Code = 19
					C错误码.Message = 变量名 + "的预查询记录数(" + strconv.Itoa(内存数据库条件记录数) + ")SQLiteDB使用内存库平均查询每记录耗时默认us(" + fmt.Sprintf("%d", SQLiteDB使用内存库平均查询每记录耗时默认us) + ")总耗时" + fmt.Sprintf("(%v)超过SQLiteDB查询最大容忍时长%d秒,请更改查询条件再试！", 总耗时, SQLiteDB查询最大容忍时长秒) + "_您的连接: " + r.RemoteAddr
					fmt.Println(C错误码.Message)
					jsonStr, err3 := json.Marshal(C错误码)
					if err3 != nil {
						w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
						return
					}
					w.Write([]byte(jsonStr))
					return
				}
			} //}else{//if 源数据库条件记录数 > 内存数据库条件记录数 {
			多个开始时间结束时间 = 单个开始结束时间
			直接返回所有记录无需筛选 = true
		} else {
			if 源数据库条件记录数 > 内存数据库条件记录数 {
				SQLiteDB使用源库查询情况.锁.Lock()
				平均查询每记录耗时2 := SQLiteDB使用源库查询情况.使用源库平均查询每记录耗时2
				平均查询每记录耗时 := SQLiteDB使用源库查询情况.使用源库平均查询每记录耗时2s
				SQLiteDB使用源库查询情况.锁.Unlock()
				总耗时 = 平均查询每记录耗时2 * time.Duration(int(每个变量目标记录数))
				if 总耗时 > SQLiteDB查询最大容忍时长秒*time.Second {
					C错误码.Code = 19
					C错误码.Message = 变量名 + "的目标记录数(" + strconv.Itoa(int(每个变量目标记录数)) + ")SQLiteDB使用源库查询情况.平均查询每记录耗时(" + 平均查询每记录耗时 + ")总耗时" + fmt.Sprintf("(%v)超过SQLiteDB查询最大容忍时长%d秒,请更改查询条件再试！", 总耗时, SQLiteDB查询最大容忍时长秒) + "_您的连接: " + r.RemoteAddr
					fmt.Println(C错误码.Message)
					jsonStr, err3 := json.Marshal(C错误码)
					if err3 != nil {
						w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
						return
					}
					w.Write([]byte(jsonStr))
					return
				}
			} else {
				平均查询每记录耗时2 := time.Duration(SQLiteDB使用内存库平均查询每记录耗时默认us * time.Microsecond)
				总耗时 = 平均查询每记录耗时2 * time.Duration(int(每个变量目标记录数))
				if 总耗时 > SQLiteDB查询最大容忍时长秒*time.Second {
					C错误码.Code = 19
					C错误码.Message = 变量名 + "的目标记录数(" + strconv.Itoa(int(每个变量目标记录数)) + ")SQLiteDB使用内存库平均查询每记录耗时默认us(" + fmt.Sprintf("%d", SQLiteDB使用内存库平均查询每记录耗时默认us) + ")总耗时" + fmt.Sprintf("(%v)超过SQLiteDB查询最大容忍时长%d秒,请更改查询条件再试！", 总耗时, SQLiteDB查询最大容忍时长秒) + "_您的连接: " + r.RemoteAddr
					fmt.Println(C错误码.Message)
					jsonStr, err3 := json.Marshal(C错误码)
					if err3 != nil {
						w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
						return
					}
					w.Write([]byte(jsonStr))
					return
				}
			}
			if len(多个开始时间结束时间2) == 0 {
				多个开始时间结束时间2 = 根据间隔时间获得多个开始时间结束时间()
			}
			多个开始时间结束时间 = 多个开始时间结束时间2
		} //if 记录数 < int(每个变量目标记录数) {
		变量分段获取记录情况 := fmt.Sprintf("变量名%s在时间期间%s~%s时间间隔%s获取的磁盘库文件记录数%d,内存库记录数%d,目标记录数%d,直接返回所有记录无需筛选%v,预计获取记录总耗时%v\r\n", 变量名, 开始时间去掉_, 结束时间去掉_, 间隔时间, 源数据库条件记录数, 内存数据库条件记录数, int(每个变量目标记录数), 直接返回所有记录无需筛选, 总耗时)
		fmt.Println(变量分段获取记录情况)
		变量类型 := 项目变量信息表组.Rows[变量名所在行[变量名]].B变量类型
		if strings.Contains(变量类型, "离散") {
			变量是离散 = true
		}
		for i, 开始时间结束时间 := range 多个开始时间结束时间 {
			查询参数[1] = 开始时间结束时间.开始时间
			查询参数[2] = 开始时间结束时间.结束时间
			var 需要将记录写入内存数据库 bool = false
			startTime := time.Now()
			row := db.QueryRow(问条件记录数参数查询语句, 查询参数...)
			if err := row.Scan(&源数据库条件记录数); err != nil {
				C错误码.Code = 19
				C错误码.Message = err.Error() + "(" + SQLiteDB文件 + ")_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			if i == 0 {
				获取记录数耗时 = time.Since(startTime)
				获取记录数耗时s = fmt.Sprintf("变量名:%s,记录数%d,从库文件：%s 获取记录数耗时%v\r\n", 变量名, 源数据库条件记录数, SQLiteDB文件, 获取记录数耗时)
				fmt.Println(获取记录数耗时s)
			}
			startTime = time.Now()
			if SQLiteDB内存数据库连接 != nil {
				row := SQLiteDB内存数据库连接.QueryRow(问条件记录数参数查询语句, 查询参数...)
				if err := row.Scan(&内存数据库条件记录数); err != nil {
					fmt.Println("SQLiteDB内存数据库连接.QueryRow(问条件记录数参数查询语句, 查询参数...)_" + err.Error())
					内存数据库条件记录数 = 0
				}
			}
			if i == 0 {
				获取记录数耗时 = time.Since(startTime)
				获取记录数耗时s = fmt.Sprintf("变量名:%s,记录数%d,从库文件：%s 获取记录数耗时%v\r\n", 变量名, 内存数据库条件记录数, "SQLiteDB内存数据库", 获取记录数耗时)
				fmt.Println(获取记录数耗时s)
			}
			if 源数据库条件记录数 == 0 && 内存数据库条件记录数 == 0 {
				continue
			}
			if 源数据库条件记录数 > 内存数据库条件记录数 {
				需要将记录写入内存数据库 = true
				获取记录的库源头 = db
				使用的库文件 = SQLiteDB文件
				//记录数 = 源数据库条件记录数
			} else {
				获取记录的库源头 = SQLiteDB内存数据库连接
				使用的库文件 = "SQLiteDB内存数据库"
				//记录数 = 内存数据库条件记录数
			}
			rows, err := 获取记录的库源头.Query(获取记录参数查询语句, 查询参数...)
			if err != nil {
				C错误码.Code = 19
				C错误码.Message = err.Error() + "(" + 使用的库文件 + ")_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			var id uint64
			var 变量名 string
			var 变量值 string
			var 时间 time.Time
			var 记录源 string
			if 使用的库文件 == "SQLiteDB内存数据库" {
				记录源 = "内存"
			} else {
				记录源 = "磁盘"
			}
			startTime = time.Now()
			for rows.Next() {
				err := rows.Scan(&id, &变量名, &变量值, &时间)
				if err != nil {
					rows.Close()
					C错误码.Code = 20
					C错误码.Message = err.Error() + "(" + 使用的库文件 + ")_您的连接: " + r.RemoteAddr
					fmt.Println(C错误码.Message)
					jsonStr, err3 := json.Marshal(C错误码)
					if err3 != nil {
						w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
						return
					}
					w.Write([]byte(jsonStr))
					return
				}
				查询结果记录总数++
				if 需要将记录写入内存数据库 {
					var exists bool
					var exists2 bool
					err = 检查重复记录预编译的语句.QueryRow(变量名, 时间).Scan(&exists)
					if err != nil && err != sql.ErrNoRows {
						fmt.Printf("检查重复记录预编译的语句.QueryRow(变量名, 时间).Scan(&exists): %v\r\n", err)
						C错误码.Code = 20
						C错误码.Message = "err = 检查重复记录预编译的语句.QueryRow(变量名, 时间).Scan(&exists):" + err.Error() + "_您的连接: " + r.RemoteAddr
						fmt.Println(C错误码.Message)
						jsonStr, err3 := json.Marshal(C错误码)
						if err3 != nil {
							w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
							return
						}
						w.Write([]byte(jsonStr))
						return
					}
					err = 磁盘检查重复记录预编译的语句.QueryRow(变量名, 时间).Scan(&exists2)
					if err != nil && err != sql.ErrNoRows {
						fmt.Printf("磁盘检查重复记录预编译的语句.QueryRow(变量名, 时间).Scan(&exists2): %v\r\n", err)
						C错误码.Code = 20
						C错误码.Message = "err = 磁盘检查重复记录预编译的语句.QueryRow(变量名, 时间).Scan(&exists2):" + err.Error() + "_您的连接: " + r.RemoteAddr
						fmt.Println(C错误码.Message)
						jsonStr, err3 := json.Marshal(C错误码)
						if err3 != nil {
							w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
							return
						}
						w.Write([]byte(jsonStr))
						return
					}
					if (!exists && exists2) || (exists && !exists2) {
						if !exists {
							stmt, _ := SQLiteDB内存数据库连接.Prepare(插入语句)
							_, err := stmt.Exec(变量名, 变量值, 时间)
							if err != nil {
								C错误码.Code = 20
								C错误码.Message = err.Error() + "(stmt, _ := SQLiteDB内存数据库连接.Prepare(插入语句)" + ")_您的连接: " + r.RemoteAddr
								fmt.Println(C错误码.Message)
								jsonStr, err3 := json.Marshal(C错误码)
								if err3 != nil {
									w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
									return
								}
								w.Write([]byte(jsonStr))
								return
							}
							stmt.Close()
						}
						if !exists2 {
							stmt, _ := SQLiteDB磁盘查询历史数据库连接.Prepare(插入语句)
							_, err := stmt.Exec(变量名, 变量值, 时间)
							if err != nil {
								C错误码.Code = 20
								C错误码.Message = err.Error() + "(stmt, _ := SQLiteDB磁盘查询历史数据库连接.Prepare(插入语句)" + ")_您的连接: " + r.RemoteAddr
								fmt.Println(C错误码.Message)
								jsonStr, err3 := json.Marshal(C错误码)
								if err3 != nil {
									w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
									return
								}
								w.Write([]byte(jsonStr))
								return
							}
							stmt.Close()
						}
						fmt.Printf("变量名：%s,时间：%v,变量值：%s,此条记录在磁盘库中存在%v在内存库中存在%v,不一致！\r\n", 变量名, 时间, 变量值, exists2, exists)
					} else { //if (!exists&&exists2)||(exists&&!exists2) {
						if !exists && !exists2 {
							批量处理的数据切片 = append(批量处理的数据切片, 变量名, 变量值, 时间)
							要写记录数++
						}
					} //}else{ //if (!exists&&exists2)||(exists&&!exists2) {
				} //if 需要将记录写入内存数据库 {
				可以生成数据 := true
				if 忽略异常值 == "是" { //if !直接返回所有记录无需筛选 && 忽略异常值 == "是" {
					if 变量值是否合法(变量名, 变量值) != "ok" {
						可以生成数据 = false
					}
				}
				if 可以生成数据 && 忽略特定值们数组有值 && !变量是离散 {
					for _, 值字符串 := range 忽略特定值们数组 {
						if 值字符串 == 变量值 {
							可以生成数据 = false
							break
						}
					}
				}
				if 可以生成数据 && 变量有特征值 && !变量是离散 {
					for _, 值字符串 := range 变量特征值组 {
						if 值字符串 == 变量值 {
							可以生成数据 = false
							break
						}
					}
				}
				if 可以生成数据 {
					switch 回复格式 {
					case "TEXT":
						变量数据2.WriteString(fmt.Sprint(id))
						变量数据2.WriteByte(',')
						变量数据2.WriteString(变量名)
						变量数据2.WriteByte(',')
						if 计算差值 == "是" {
							if 前值s == "" {
								变量数据2.WriteString("0.00")
							} else {
								前值, err := strconv.ParseFloat(前值s, 64)
								if err != nil {
									变量数据2.WriteString("前值, err := strconv.ParseFloat(前值s, 64)：" + err.Error())
								} else {
									当前值, err1 := strconv.ParseFloat(变量值, 64)
									if err1 != nil {
										变量数据2.WriteString("当前值, err1 := strconv.ParseFloat(变量值, 64)：" + err1.Error())
									} else {
										差值 := 当前值 - 前值
										if 变量是累积值[变量名] && 差值 < 0 {
											差值 = 0
										}
										变量数据2.WriteString(fmt.Sprintf("%.2f", 差值))
									}
								}
							}
							前值s = 变量值
						} else {
							变量数据2.WriteString(变量值)
						}
						变量数据2.WriteByte(',')
						if 当地时区信息 != nil {
							变量数据2.WriteString(时间.In(当地时区信息).Format("2006-01-02 15:04:05"))
						} else {
							变量数据2.WriteString(时间.Format("2006-01-02 15:04:05"))
						}
						变量数据2.WriteByte(',')
						变量数据2.WriteString(记录源)
						变量数据2.WriteString("\r\n")
					case "XLSX", "JSON", "GRAPH":
						变量们的历史数据集合[变量名] = append(变量们的历史数据集合[变量名], 变量历史数据结构体{VarValue: 变量值, Time: 时间})
					}
					if !直接返回所有记录无需筛选 {
						break
					}
				}
			} //for rows.Next() {
			if err = rows.Err(); err != nil {
				C错误码.Code = 20
				C错误码.Message = err.Error() + "(" + 使用的库文件 + ")_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			rows.Close()
			if i == 0 {
				处理记录耗时 := time.Since(startTime)
				处理记录耗时s := fmt.Sprintf("变量名:%s,记录数%d,从库文件：%s 处理记录耗时%v\r\n", 变量名, 记录数, 使用的库文件, 处理记录耗时)
				fmt.Println(处理记录耗时s)
			}
		} //for i,开始时间结束时间:=range 多个开始时间结束时间{
	} //for _,变量名:=range 请求的变量组{
	if 检查重复记录预编译的语句 != nil {
		检查重复记录预编译的语句.Close()
		// 	defer func() { 检查重复记录预编译的语句.Close() }()
	}
	if 磁盘检查重复记录预编译的语句 != nil {
		磁盘检查重复记录预编译的语句.Close()
		// 	defer func() { 磁盘检查重复记录预编译的语句.Close() }()
	}
	if 要写记录数 > 0 {
		var 批量插入预编译的语句 *sql.Stmt
		var 磁盘批量插入预编译的语句 *sql.Stmt
		func() {
			插入语句 := "INSERT INTO " + SQLite表格名 + " (变量名, 变量值, 时间) VALUES (?, ?, ?)"
			批量插入预编译的语句, err1 = SQLiteDB内存数据库连接.Prepare(插入语句)
			if err1 != nil {
				fmt.Printf("内存库无法准备插入语句: %v\r\n", err1)
				return
			}
			磁盘批量插入预编译的语句, err1 = SQLiteDB磁盘查询历史数据库连接.Prepare(插入语句)
			if err1 != nil {
				fmt.Printf("磁盘库无法准备插入语句: %v\r\n", err1)
				return
			}
		}()
		if 批量插入预编译的语句 != nil {
			defer 批量插入预编译的语句.Close()
		}
		if 磁盘批量插入预编译的语句 != nil {
			defer 磁盘批量插入预编译的语句.Close()
		}
		if 批量插入预编译的语句 == nil || 磁盘批量插入预编译的语句 == nil {
			C错误码.Code = 16
			C错误码.Message = "批量插入预编译的语句 == nil || 磁盘批量插入预编译的语句 == nil" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
		var wg sync.WaitGroup
		errChan := make(chan error, 2) // 创建一个带缓冲的错误通道
		startTime2 := time.Now()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := execBatch(批量插入预编译的语句, 批量处理的数据切片); err != nil {
				errChan <- err
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := execBatch(磁盘批量插入预编译的语句, 批量处理的数据切片); err != nil {
				errChan <- err
			}
		}()
		// 等待所有异步操作完成
		go func() {
			wg.Wait()
			close(errChan) // 关闭通道以指示没有更多的错误
		}()
		// 检查和处理错误
		for err := range errChan {
			if err != nil {
				http.Error(w, "批量写内存数据库库和磁盘查询历史数据库发生错误："+err.Error(), http.StatusInternalServerError)
				fmt.Println("批量写内存数据库库和磁盘查询历史数据库发生错误：" + err.Error())
				return
			}
		}
		同时写内存库和磁盘历史库耗时 := time.Since(startTime2)
		同时写内存库和磁盘历史库耗时s := fmt.Sprintf("要写记录数%d,同时写内存库和磁盘历史库耗时%v\r\n", 要写记录数, 同时写内存库和磁盘历史库耗时)
		fmt.Println(同时写内存库和磁盘历史库耗时s)
	} //if 要写记录数 > 0 {
	var SQLiteDB内存数据库记录总数 int
	var SQLiteDB磁盘查询历史数据库记录总数 int
	查询最大id语句 := "SELECT MAX(rowid) FROM " + SQLite表格名
	var 等待组 sync.WaitGroup
	等待组.Add(1)
	go func() {
		defer 等待组.Done()
		if SQLiteDB内存数据库连接 == nil {
			return
		}
		row := SQLiteDB内存数据库连接.QueryRow(查询最大id语句)
		if err := row.Scan(&SQLiteDB内存数据库记录总数); err != nil {
			http.Error(w, "row.Scan(&SQLiteDB内存数据库记录总数)"+err.Error(), http.StatusInternalServerError)
			fmt.Println("row.Scan(&SQLiteDB内存数据库记录总数)" + err.Error())
			return
		}
	}()
	等待组.Add(1)
	go func() {
		defer 等待组.Done()
		if SQLiteDB磁盘查询历史数据库连接 == nil {
			return
		}
		row := SQLiteDB磁盘查询历史数据库连接.QueryRow(查询最大id语句)
		if err := row.Scan(&SQLiteDB磁盘查询历史数据库记录总数); err != nil {
			http.Error(w, "row.Scan(&SQLiteDB磁盘查询历史数据库记录总数)"+err.Error(), http.StatusInternalServerError)
			fmt.Println("row.Scan(&SQLiteDB磁盘查询历史数据库记录总数)" + err.Error())
			return
		}
	}()
	等待组.Wait()
	elapsedTime := time.Since(startTime)
	访问耗时 := time.Since(访问开始时间)
	var 平均查询每记录耗时 time.Duration
	if 查询结果记录总数 == 0 {
		平均查询每记录耗时 = 0
	} else {
		平均查询每记录耗时 = elapsedTime / time.Duration(查询结果记录总数)
	}
	变量数据 := ""
	str1 := ""
	str := 编译时间 + "\r\n"
	str1 += str
	str = 至今多少天时分秒(网关启动时刻)
	str = "启动至今: " + str + "\r\n"
	str1 += str
	str = 网关启动时间
	str1 += str
	变量个数 := fmt.Sprintf("%d", 长度)
	要写记录数s := fmt.Sprintf("%d", 要写记录数)
	查询结果记录总数s := strconv.Itoa(查询结果记录总数)
	当前时间 := time.Now().Format("2006-01-02 15:04:05")
	平均查询每记录耗时s := fmt.Sprintf("%v", 平均查询每记录耗时)
	最大耗时 := fmt.Sprintf("SQLiteDB磁盘查询历史数据库记录总数:%d\r\nSQLiteDB内存数据库记录总数:%d\r\n查询结果记录总数:%s\r\n查询耗时：%v\r\n平均查询每记录耗时：%s\r\n访问总耗时：%v\r\n", SQLiteDB磁盘查询历史数据库记录总数, SQLiteDB内存数据库记录总数, 查询结果记录总数s, elapsedTime, 平均查询每记录耗时s, 访问耗时) + "写内存库记录数:" + 要写记录数s + "\r\n"
	查询的url := "http://" + r.Host + "/SQLiteDB查询?开始时间=" + 开始时间 +
		"&结束时间=" + 结束时间 +
		"&间隔时间=" + 间隔时间 +
		"&user=" + 本网关登录名之MD5 +
		"&password=" + 本网关登录密码之MD5 +
		"&曲线图Y值使用实际值=" + 曲线图Y值使用实际值 +
		"&忽略异常值=" + 忽略异常值 +
		"&表格历史空数据自动填充=" + 表格历史空数据自动填充 +
		"&忽略特定值们=" + 忽略特定值们 +
		"&忽略曲线说明=" + 忽略曲线说明 +
		"&计算差值=" + 计算差值 +
		"&回复格式=" + 回复格式 +
		"&要查询的变量名们=" + 要查询的变量名们
	变量数据 = str1 + "\r\n" + 当前时间 + "\r\n第几次采集： " +
		strconv.FormatUint(atomic.LoadUint64(&第几次采集), 10) + "\r\n" + 最大耗时 + "查询变量个数：" + 变量个数 + "\r\n" +
		"查询的url：\r\n" + 查询的url + "\r\n" +
		"每个变量目标记录数：" + fmt.Sprintf("%d", int(每个变量目标记录数)) + "\r\n" +
		"查询人：" + r.RemoteAddr + "\r\n" + 变量数据2.String()
	//elapsedTime := time.Since(startTime)
	SQLiteDB使用源库查询情况.锁.Lock()
	SQLiteDB使用源库查询情况.平均查询每记录耗时 = 平均查询每记录耗时s
	SQLiteDB使用源库查询情况.平均查询每记录耗时2 = 平均查询每记录耗时
	SQLiteDB使用源库查询情况.使用源库平均查询每记录耗时2 = 平均查询每记录耗时
	SQLiteDB使用源库查询情况.使用源库平均查询每记录耗时2s = 平均查询每记录耗时s
	SQLiteDB使用源库查询情况.查询结果记录总数 = 查询结果记录总数s
	SQLiteDB使用源库查询情况.写内存库记录数 = 要写记录数s
	SQLiteDB使用源库查询情况.查询变量个数 = 变量个数
	SQLiteDB使用源库查询情况.查询的url = 查询的url
	SQLiteDB使用源库查询情况.查询时间 = time.Now().Format("2006-01-02 15:04:05")
	SQLiteDB使用源库查询情况.查询耗时 = fmt.Sprintf("%v", elapsedTime)
	SQLiteDB使用源库查询情况.访问总耗时 = fmt.Sprintf("%v", 访问耗时)
	SQLiteDB使用源库查询情况.查询人 = r.RemoteAddr
	SQLiteDB使用源库查询情况.SQLiteDB内存数据库记录总数 = fmt.Sprintf("%d", SQLiteDB内存数据库记录总数)
	SQLiteDB使用源库查询情况.SQLiteDB磁盘查询历史数据库记录总数 = fmt.Sprintf("%d", SQLiteDB磁盘查询历史数据库记录总数)
	SQLiteDB使用源库查询情况.锁.Unlock()
	switch 回复格式 {
	case "TEXT":
		w.Write([]byte(变量数据))
	case "JSON":
		if 计算差值 == "是" {
			for 变量名, 数据点们 := range 变量们的历史数据集合 {
				func() {
					// 备份原始数值
					temp := make([]float64, len(数据点们))
					for i, 数据点 := range 数据点们 {
						val, err := strconv.ParseFloat(数据点.VarValue, 64)
						if err != nil {
							// 处理错误，例如记录日志后跳过该变量
							return
						}
						temp[i] = val
					}

					// 设置第一个数据点为0.00
					if len(数据点们) > 0 {
						变量们的历史数据集合[变量名][0].VarValue = "0.00"
					}
					// 计算后续差值
					for i := 1; i < len(数据点们); i++ {
						diff := temp[i] - temp[i-1]
						if 变量是累积值[变量名] && diff < 0 {
							diff = 0
						}
						变量们的历史数据集合[变量名][i].VarValue = fmt.Sprintf("%.2f", diff)
					}
				}()
			}
		} //if 计算差值 == "是" {
		jsonData, err := json.Marshal(变量们的历史数据集合)
		if err != nil {
			http.Error(w, "jsonData, err := json.Marshal(变量们的历史数据集合)发生错误："+err.Error(), http.StatusInternalServerError)
			fmt.Println("jsonData, err := json.Marshal(变量们的历史数据集合)发生错误：" + err.Error())
			return
		}
		w.Write([]byte(string(jsonData)))
	case "XLSX":
		数据长度 := len(变量们的历史数据集合)
		if 数据长度 < 1 {
			fmt.Println("数据长度:=len(变量们的历史数据集合)>>数据长度<1")
			C错误码.Code = 16
			C错误码.Message = "数据长度:=len(变量们的历史数据集合),数据长度小于1" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
		startTime := time.Now()
		var rows [][]string
		newSlice := make([]string, len(请求的变量组)+1)
		newSlice[0] = "时间"
		copy(newSlice[1:], 请求的变量组)
		rows = append(rows, newSlice)
		const layout = "2006-01-02_15:04:05"
		// 解析时间字符串，UTC时间，中国是CTC
		开始时间2, err := time.Parse(layout, 开始时间)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		// 获取系统时区与 UTC 的时差（以秒为单位）
		_, offset := time.Now().In(loc).Zone()
		时间基准 := 开始时间2.Add(-time.Duration(offset) * time.Second)
		var 变量们的值被选中是第几个 = make(map[string]int, 0)
		for 变量名 := range 变量们的历史数据集合 {
			变量们的值被选中是第几个[变量名] = -1
		}
		for i := 0; i < int(每个变量目标记录数); i++ {
			var row []string
			row = append(row, 时间基准.In(loc).String())
			时间基准2 := 时间基准.Add(间隔时间t)
			for _, 变量名 := range 请求的变量组 {
				变量在此时间基准没有数据 := true
				for i, 结构体 := range 变量们的历史数据集合[变量名] {
					t := 结构体.Time
					if t.After(时间基准2) {
						break
					}
					if i <= 变量们的值被选中是第几个[变量名] {
						continue
					}
					if (t.After(时间基准) && t.Before(时间基准2)) ||
						t.Equal(时间基准) || t.Equal(时间基准2) {
						row = append(row, 结构体.VarValue)
						变量们的值被选中是第几个[变量名] = i
						变量在此时间基准没有数据 = false
						break
					}
				}
				if 变量在此时间基准没有数据 {
					row = append(row, "")
				}
			}
			rows = append(rows, row)
			时间基准 = 时间基准.Add(间隔时间t)
		}
		elapsedTime := time.Since(startTime)
		变量有非法值 := ""
		if len(各变量要忽略的特征值) > 0 {
			变量有非法值 = "_变量有非法值"
		}
		downloadExcel(变量是累积值, 变量有非法值, &rows, w, r, 变量个数, 开始时间去掉_, 结束时间去掉_, 时间差, 间隔时间t, 忽略异常值, 忽略特定值们, 计算差值, 表格历史空数据自动填充, elapsedTime, 访问耗时)
	case "GRAPH":
		graph(变量是累积值, 各变量要忽略的特征值, 请求的变量组, 变量们的历史数据集合, 曲线图Y值使用实际值, w, r, 变量个数, 开始时间去掉_, 结束时间去掉_, 时间差, 间隔时间t, 忽略异常值, 忽略曲线说明, 计算差值, 忽略特定值们, 访问耗时)
	}
} //func SQLiteDB查询(w http.ResponseWriter, r *http.Request) {
var SQLiteDB使用源库查询情况 SQLiteDB查询情况结构体

type SQLiteDB查询情况结构体 struct {
	锁                     sync.Mutex
	查询结果记录总数              string
	写内存库记录数               string
	查询变量个数                string
	查询的url                string
	查询时间                  string
	查询耗时                  string
	访问总耗时                 string
	平均查询每记录耗时             string
	平均查询每记录耗时2            time.Duration
	使用源库平均查询每记录耗时2        time.Duration
	使用源库平均查询每记录耗时2s       string
	SQLiteDB内存数据库记录总数     string
	SQLiteDB磁盘查询历史数据库记录总数 string
	查询人                   string
}

// TimeTicker 实现自定义时间轴刻度
type TimeTicker struct {
	Format string // 时间格式（如 "2006-01-02 15:04"）
}

// Ticks 返回时间轴的刻度和标签
func (t TimeTicker) Ticks(min, max float64) []plot.Tick {
	// 根据数据范围生成时间间隔
	step := (max - min) / SQLiteDB查询曲线图时间刻度数 // 示例：5 个主刻度
	var ticks []plot.Tick
	loc, err := time.LoadLocation("Local")
	if err != nil {
		fmt.Printf("loc, err := time.LoadLocation(Local)>>%s", err.Error())
		return ticks
	}
	_, offset := time.Now().In(loc).Zone()
	for x := min; x <= max; x += step {
		tTime := time.Unix(int64(x), 0).UTC()
		tTime = tTime.Add(time.Duration(offset) * time.Second)
		label := tTime.Format(t.Format)
		ticks = append(ticks, plot.Tick{
			Value: x,
			Label: label,
		})
	}
	return ticks
}

func graph(变量是累积值 map[string]bool, 各变量要忽略的特征值 map[string][]string, 请求的变量组 []string, 变量们的历史数据集合 map[string][]变量历史数据结构体, 曲线图Y值使用实际值 string, w http.ResponseWriter, r *http.Request, 变量个数, 开始时间, 结束时间 string, 时间差, 间隔时间t time.Duration, 忽略异常值, 忽略曲线说明, 计算差值, 忽略特定值们 string, 访问耗时 time.Duration) {
	type ErrCode struct {
		Code    int    `json:"Code"`
		Message string `json:"Message"`
	}
	var C错误码 ErrCode
	数据长度 := len(变量们的历史数据集合)
	if 数据长度 < 1 {
		fmt.Println("数据长度:=len(变量们的历史数据集合)>>数据长度<1")
		C错误码.Code = 16
		C错误码.Message = "数据长度:=len(变量们的历史数据集合),数据长度小于1" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	当前时间 := time.Now()
	type 变量历史数据结构体2 struct {
		VarValue float64
		Time     float64
	}
	const 数字长度限制 = 5
	浮点数转字符串 := func(值 float64) string {
		值s := fmt.Sprintf("%.1f", 值)
		if len(值s) > 数字长度限制 {
			值s = fmt.Sprintf("%.1e", 值)
		}
		return 值s
	}
	每个变量目标记录数 := 时间差/间隔时间t + 1
	var 变量们的最大值 = make(map[string]float64, 0)
	var 变量们的最小值 = make(map[string]float64, 0)
	var 变量们的差值 = make(map[string]float64, 0)
	var 变量们的最大值s = make(map[string]string, 0)
	var 变量们的最小值s = make(map[string]string, 0)
	var 变量们的差值s = make(map[string]string, 0)
	var 变量们的缩放说明 = make(map[string]string, 0)
	var 各变量要忽略的特征值s = make(map[string]string, 0)
	var 变量们的数据个数 = make(map[string]string, 0)
	var 变量们的历史数据集合2 = make(map[string][]变量历史数据结构体2, 0)
	var 最大时间 float64 = -math.MaxFloat64
	var 最小时间 float64 = math.MaxFloat64
	var 最大Y值 float64 = -math.MaxFloat64
	var 最小Y值 float64 = math.MaxFloat64

	for 变量名, 数据点们 := range 变量们的历史数据集合 {
		个数 := len(数据点们)
		if 个数 < 1 {
			continue
		}
		for _, 数据点 := range 数据点们 {
			num, err := strconv.ParseFloat(数据点.VarValue, 64)
			if err != nil {
				C错误码.Code = 16
				C错误码.Message = "num, err := strconv.ParseFloat(数据点.VarValue, 64)：" + err.Error() + "_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			if num > 最大Y值 {
				最大Y值 = num
			}
			if num < 最小Y值 {
				最小Y值 = num
			}
			时间 := float64(数据点.Time.Unix()) + float64(数据点.Time.Nanosecond())/1e9
			if 时间 > 最大时间 {
				最大时间 = 时间
			}
			if 时间 < 最小时间 {
				最小时间 = 时间
			}
			变量们的历史数据集合2[变量名] = append(变量们的历史数据集合2[变量名], 变量历史数据结构体2{VarValue: num, Time: 时间})
		}
	} //for 变量名, 数据点们 := range 变量们的历史数据集合 {
	if 最大Y值 <= 最小Y值 {
		str := 浮点数转字符串(最大Y值)
		C错误码.Code = 16
		C错误码.Message = "最小Y值 大于或等于 最大Y值(" + str + ")没必要作图也无法作图_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if 最大时间 <= 最小时间 {
		str := 浮点数转字符串(最大时间)
		C错误码.Code = 16
		C错误码.Message = "最小时间 大于或等于 最大时间(" + str + ")没必要作图也无法作图_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if 计算差值 == "是" {
		for 变量名, 数据点们 := range 变量们的历史数据集合2 {
			var 累计值 float64
			// 备份原始数值
			temp := make([]float64, len(数据点们))
			for i, 数据点 := range 数据点们 {
				temp[i] = 数据点.VarValue
			}
			// 设置第一个数据点为0.00
			if len(数据点们) > 0 {
				变量们的历史数据集合2[变量名][0].VarValue = 0
			}
			// 计算后续差值
			for i := 1; i < len(数据点们); i++ {
				diff := temp[i] - temp[i-1]
				if 变量是累积值[变量名] && diff < 0 {
					diff = 0
				}
				变量们的历史数据集合2[变量名][i].VarValue = diff
				累计值 += diff
			}
			变量们的差值s[变量名] = fmt.Sprintf("%.2f", 累计值)
		} //for 变量名, 数据点们 := range 变量们的历史数据集合2 {
	} //if 计算差值 == "是" {
	for 变量名, 数据点们 := range 变量们的历史数据集合2 {
		个数 := len(数据点们)
		if 个数 < 1 {
			continue
		}
		变量们的缩放说明[变量名] = ""
		if 变量是累积值[变量名] {
			各变量要忽略的特征值s[变量名] = "(累积)"
		} else {
			各变量要忽略的特征值s[变量名] = ""
		}

		var Max float64 = -math.MaxFloat64
		var Min float64 = math.MaxFloat64
		百分比 := float64(个数) / float64(每个变量目标记录数) * 100
		百分比s := fmt.Sprintf("%.0f", 百分比) + "%"
		变量们的数据个数[变量名] = fmt.Sprintf("(%d,%s)", 个数, 百分比s)
		for i, 数据点 := range 数据点们 {
			num := 数据点.VarValue
			if num > Max {
				Max = num
				if 计算差值 == "是" {
					变量们的最大值s[变量名] = 浮点数转字符串(num)
				} else {
					变量们的最大值s[变量名] = 变量们的历史数据集合[变量名][i].VarValue
				}
			}
			if num < Min {
				Min = num
				if 计算差值 == "是" {
					变量们的最小值s[变量名] = 浮点数转字符串(num)
				} else {
					变量们的最小值s[变量名] = 变量们的历史数据集合[变量名][i].VarValue
				}
			}
		}
		变量们的最大值[变量名] = Max
		变量们的最小值[变量名] = Min
		变量们的差值[变量名] = Max - Min
		if 计算差值 != "是" {
			if 变量是累积值[变量名] {
				var 累计值 float64
				for i := range 变量们的历史数据集合2[变量名] {
					if i == 0 {
						continue
					}
					diff := 变量们的历史数据集合2[变量名][i].VarValue - 变量们的历史数据集合2[变量名][i-1].VarValue
					if diff < 0 {
						diff = 0
					}
					累计值 += diff
				}
				变量们的差值s[变量名] = fmt.Sprintf("%.2f", 累计值)
			} else {
				变量们的差值s[变量名] = fmt.Sprintf("%.2f", 变量们的差值[变量名])
			}
		}
	} //for 变量名, 数据点们 := range 变量们的历史数据集合 {
	for 变量名, 特征值组 := range 各变量要忽略的特征值 {
		str := ""
		for _, 特征值 := range 特征值组 {
			str += 特征值 + ","
		}
		if str != "" {
			str = str[:len(str)-1]
			str = "(" + str + ")"
			各变量要忽略的特征值s[变量名] += str
		}
	}
	曲线条数 := len(变量们的历史数据集合2)
	// fmt.Printf("曲线条数%d\r\n", 曲线条数)
	// fmt.Println("曲线图Y值使用实际值:" + 曲线图Y值使用实际值)
	p := plot.New()
	目标记录数 := fmt.Sprintf("目标记录数%d", 每个变量目标记录数)
	刻度 := 时间差 / SQLiteDB查询曲线图时间刻度数
	p.X.Label.Text = 开始时间 + "~" + 结束时间 + "_" + 时间差.String() + "_" + 间隔时间t.String() + "_" + 目标记录数 + "_忽略异常值:" + 忽略异常值 + "_计算差值:" + 计算差值 + "_目标曲线数" + 变量个数 + "_刻度" + 刻度.String() + "\r\n忽略特定值们:" + 忽略特定值们
	if 曲线图Y值使用实际值 == "是" {
		p.Y.Label.Text = "实际值"
		if 忽略曲线说明 != "是" {
			for 变量名 := range 变量们的历史数据集合2 {
				// 最小值s := 浮点数转字符串(变量们的最小值[变量名])
				// 最大值s := 浮点数转字符串(变量们的最大值[变量名])
				变量们的缩放说明[变量名] = fmt.Sprintf("(%s~%s,%s)", 变量们的最小值s[变量名], 变量们的最大值s[变量名], 变量们的差值s[变量名])
			}
		}
	} else { //if 曲线图Y值使用实际值 == "是" {
		p.Y.Label.Text = "相对值：实际值乘以倍率再加上或减去一个值"
		var 最大差值 float64 = -math.MaxFloat64
		var 最大差值变量名 string
		const epsilon = 1e-9 // 根据实际需求调整阈值
		for 变量名, 差值 := range 变量们的差值 {
			if 差值 > 最大差值 {
				最大差值 = 差值
				最大差值变量名 = 变量名
			}
		}
		func() {
			if 最大差值 < epsilon {
				p.Y.Label.Text = "实际值(由于最大变化率曲线的变化率太小，所以无法用相对值画图)"
				if 忽略曲线说明 != "是" {
					for 变量名 := range 变量们的历史数据集合2 {
						// 最小值s := 浮点数转字符串(变量们的最小值[变量名])
						// 最大值s := 浮点数转字符串(变量们的最大值[变量名])
						变量们的缩放说明[变量名] = fmt.Sprintf("(%s~%s,%s)", 变量们的最小值s[变量名], 变量们的最大值s[变量名], 变量们的差值s[变量名])
					}
				}
				return
			}
			const 最小变化率倍数 = 4
			最小变化率 := 最大差值 / 最小变化率倍数
			放大到变化率 := 最大差值 / 2
			for 变量名 := range 变量们的历史数据集合2 {
				if 变量们的差值[变量名] >= 最小变化率 &&
					变量们的最小值[变量名] >= 变量们的最小值[最大差值变量名] &&
					变量们的最大值[变量名] <= 变量们的最大值[最大差值变量名] {
					if 忽略曲线说明 != "是" {
						// 最小值s := 浮点数转字符串(变量们的最小值[变量名])
						// 最大值s := 浮点数转字符串(变量们的最大值[变量名])
						变量们的缩放说明[变量名] = fmt.Sprintf("(%s~%s,%s)", 变量们的最小值s[变量名], 变量们的最大值s[变量名], 变量们的差值s[变量名])
					}
					continue
				}
				var 都加上一个值 float64 = 0
				var 都减去一个值 float64 = 0
				var 都乘以一个值 float64 = 1
				if 变量们的差值[变量名] >= 最小变化率 || math.Abs(变量们的差值[变量名]) < epsilon {
					if 变量们的最小值[变量名] < 变量们的最小值[最大差值变量名] {
						都加上一个值 = 变量们的最小值[最大差值变量名] - 变量们的最小值[变量名]
					} else { //if 变量们的最小值[变量名] < 变量们的最小值[最大差值变量名] {
						if 变量们的最大值[变量名] > 变量们的最大值[最大差值变量名] {
							都减去一个值 = 变量们的最大值[变量名] - 变量们的最大值[最大差值变量名]
						} //if 变量们的最大值[变量名] > 变量们的最大值[最大差值变量名] {
					} //} else {//if 变量们的最小值[变量名] < 变量们的最小值[最大差值变量名] {
					if 忽略曲线说明 != "是" {
						// 最小值s := 浮点数转字符串(变量们的最小值[变量名])
						// 最大值s := 浮点数转字符串(变量们的最大值[变量名])
						值 := 变量们的最小值[变量名] - 都减去一个值 + 都加上一个值
						最小值2s := 浮点数转字符串(值)
						值 = 变量们的最大值[变量名] - 都减去一个值 + 都加上一个值
						最大值2s := 浮点数转字符串(值)
						变量们的缩放说明[变量名] = fmt.Sprintf("(%s~%s,%s,%s~%s)", 变量们的最小值s[变量名], 变量们的最大值s[变量名], 变量们的差值s[变量名], 最小值2s, 最大值2s)
					}
				} else { //if 变量们的差值[变量名] >= 最小变化率 || math.Abs(变量们的差值[变量名]) < epsilon {
					都乘以一个值 = 放大到变化率 / 变量们的差值[变量名]
					//fmt.Printf(变量名+"_都乘以一个值(%.1f) = 放大到变化率(%.1f) / 变量们的差值[变量名](%.1f)\r\n", 都乘以一个值, 放大到变化率, 变量们的差值[变量名])
					乘后最小值 := 都乘以一个值 * 变量们的最小值[变量名]
					乘后最大值 := 都乘以一个值 * 变量们的最大值[变量名]
					if 乘后最小值 < 变量们的最小值[最大差值变量名] {
						都加上一个值 = 变量们的最小值[最大差值变量名] - 乘后最小值
					} else {
						if 乘后最大值 > 变量们的最大值[最大差值变量名] {
							都减去一个值 = 乘后最大值 - 变量们的最大值[最大差值变量名]
						}
					}
					if 忽略曲线说明 != "是" {
						// 最小值s := 浮点数转字符串(变量们的最小值[变量名])
						// 最大值s := 浮点数转字符串(变量们的最大值[变量名])
						都乘以一个值s := 浮点数转字符串(都乘以一个值)
						//fmt.Println("都乘以一个值s：" + 都乘以一个值s)
						值 := 变量们的最小值[变量名]*都乘以一个值 - 都减去一个值 + 都加上一个值
						最小值2s := 浮点数转字符串(值)
						值2 := 变量们的最大值[变量名]*都乘以一个值 - 都减去一个值 + 都加上一个值
						最大值2s := 浮点数转字符串(值2)
						放大后差值s := 浮点数转字符串(值2 - 值)
						变量们的缩放说明[变量名] = fmt.Sprintf("(%s~%s,%s,%s~%s,%s,倍率%s)", 变量们的最小值s[变量名], 变量们的最大值s[变量名], 变量们的差值s[变量名], 最小值2s, 最大值2s, 放大后差值s, 都乘以一个值s)
					}
				} //}else{//if 变量们的差值[变量名] >= 最小变化率 || math.Abs(变量们的差值[变量名]) < epsilon {
				for i := range 变量们的历史数据集合2[变量名] {
					乘后值 := 变量们的历史数据集合2[变量名][i].VarValue
					if 都乘以一个值 != 1 {
						乘后值 *= 都乘以一个值
					}
					变量们的历史数据集合2[变量名][i].VarValue = 乘后值 - 都减去一个值 + 都加上一个值
				}
			} //for 变量名, 数据点们 := range 变量们的历史数据集合2 {
		}() //func (){
	} //}else{//if 曲线图Y值使用实际值 == "是" {
	colors := func(n int) []color.Color {
		var colors []color.Color
		for i := 0; i < n; i++ {
			hue := float64(i) * 360.0 / float64(n) // 均匀分布色相
			c := colorful.Hsl(hue, 1.0, 0.5)       // 固定饱和度和亮度
			colors = append(colors, c)
		}
		return colors
	}(曲线条数)
	//rand.Seed(time.Now().UnixNano())
	// 定义切片
	//numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	// Fisher-Yates 洗牌算法
	for i := len(colors) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		colors[i], colors[j] = colors[j], colors[i]
	}
	p.Legend = plot.NewLegend()
	p.Legend.Top = true // 设置图例位置（如顶部）‌:ml-citation{ref="3,4" data="citationList"}
	第几个变量 := 0
	for _, 变量名 := range 请求的变量组 {
		数据点们 := 变量们的历史数据集合2[变量名]
		//} //for _, 变量名 := range 请求的变量组 {
		//for 变量名, 数据点们 := range 变量们的历史数据集合2 {
		数据个数 := len(数据点们)
		if 数据个数 < 1 {
			continue
		}
		pts := make(plotter.XYs, 数据个数)
		指针 := &项目变量信息表组.Rows[变量名所在行[变量名]]
		变量类型 := 指针.B变量类型
		开关量 := ""
		if strings.Contains(变量类型, "离散") {
			开关量 = "(开关量)"
		}
		单位 := 指针.D单位
		if 单位 != "" {
			单位 = "(" + 单位 + ")"
		}
		for j, 数据点 := range 数据点们 {
			pts[j].X = 数据点.Time
			pts[j].Y = 数据点.VarValue
		} //for _, 数据点 := range 数据点们 {
		line, _ := plotter.NewLine(pts)
		line.Color = colors[第几个变量] // 设置颜色
		if 忽略曲线说明 != "是" {
			p.Legend.Add(fmt.Sprintf("%d:%s%s%s%s%s%s", 第几个变量+1, 变量名, 单位, 开关量, 变量们的缩放说明[变量名], 变量们的数据个数[变量名], 各变量要忽略的特征值s[变量名]), line)
		}
		p.Add(line)
		第几个变量++
	} //for 变量名, 数据点们 := range 变量们的历史数据集合2 {
	生成人 := strings.Replace(r.RemoteAddr, ":", " ", -1)
	// 保存为PNG
	宽 := SQLiteDB查询曲线图长度cm * vg.Centimeter
	高 := 9.0 / 16.0 * 宽
	忽略异常值2 := ""
	if 忽略异常值 == "是" {
		忽略异常值2 = "_忽略异常值"
	}
	忽略特定值们2 := ""
	if 忽略特定值们 != "" {
		忽略特定值们2 = "_忽略值" + 忽略特定值们
	}
	忽略曲线说明2 := ""
	if 忽略曲线说明 == "是" {
		忽略曲线说明2 = "_忽略曲线说明"
	}
	计算差值2 := ""
	if 计算差值 == "是" {
		计算差值2 = "_计算差值"
	}
	有累积变量 := ""
	if len(变量是累积值) > 0 {
		有累积变量 = "_有累积变量"
	}
	变量有非法值 := ""
	if len(各变量要忽略的特征值) > 0 {
		变量有非法值 = "_变量有非法值"
	}
	p.X.Tick.Marker = TimeTicker{Format: "2006-01-02\r\n15:04:04"}
	p.X.Tick.Label.XAlign = text.XRight   // 水平对齐到右侧
	p.X.Tick.Label.YAlign = text.YCenter  // 垂直居中对齐
	p.X.Tick.Label.Rotation = math.Pi / 2 // 旋转-90度（逆时针90度）
	耗时 := time.Since(当前时间) + 访问耗时
	fmt.Printf("p.X.Tick.Label.Rotation后耗时%v\r\n", 耗时)
	当前时间s := time.Now().Format("2006-01-02 15:04:05")
	图片名称2 := 目录2 + "SQLiteDB查询_" + strings.Replace(当前时间s, ":", "", -1) + "_" + 生成人 + "_" + strings.Replace(开始时间, ":", "", -1) + "~" + strings.Replace(结束时间, ":", "", -1) + "_" + 时间差.String() + "_" + 间隔时间t.String() + "_" + 目标记录数 + "_目标曲线数" + 变量个数 + 忽略异常值2 + 变量有非法值 + 忽略曲线说明2 + 计算差值2 + 有累积变量 + "_" + 耗时.String() + 忽略特定值们2
	图片名称 := 图片名称2 + ".png"
	// 安全截断中文字符串（按字符数）
	长度 := utf8.RuneCountInString(图片名称)
	if 长度 > 255 {
		fmt.Printf(图片名称+" 长度%d\r\n", 长度)
		图片名称 = 安全截取中文字符串(图片名称2, 251) + ".png"
		fmt.Printf("修改为：%s,长度%d\r\n", 图片名称, utf8.RuneCountInString(图片名称))
	}
	耗时1 := time.Since(当前时间) + 访问耗时
	fmt.Printf("安全截取中文字符串后耗时%v\r\n", 耗时1)
	//图片名称 = 图片名称 + "_" + 耗时.String() + ".png"
	p.Title.Text = 当前时间s + "_" + r.RemoteAddr + "_" + 耗时.String()
	//图片名称 = "\"" + 图片名称 + "\""
	imgPath := filepath.Join(图片名称)
	if err := p.Save(宽, 高, imgPath); err != nil {
		fmt.Printf("err := p.Save(曲线图长度cm, 宽, imgPath); err != nil: %+v\r\n", err)
		return
	}
	耗时2 := time.Since(当前时间) + 访问耗时
	fmt.Printf("p.Save(宽, 高, imgPath)后耗时%v\r\n", 耗时2)
	http.ServeFile(w, r, imgPath)
	耗时3 := time.Since(当前时间) + 访问耗时
	fmt.Printf("http.ServeFile(w, r, imgPath)后耗时%v\r\n", 耗时3)
} //func graph(变量们的历史数据集合 map[string][]变量历史数据结构体, 曲线图Y值使用实际值 string, w http.ResponseWriter, r *http.Request, 变量个数, 开始时间 string, 结束时间 string, 间隔时间 string, 访问耗时 time.Duration) {
func 安全截取中文字符串(s string, maxLen int) string {
	// 转换为rune数组确保正确处理Unicode
	runes := []rune(s)
	// 判断实际字符长度
	if utf8.RuneCountInString(s) > maxLen {
		// 截取前maxLen个字符
		runes = runes[:maxLen]
		// 添加可选后缀（如...）
		// runes = append(runes, []rune("...")...)
	}
	return string(runes)
}

func init() {
	ttfBytes1, err := os.ReadFile("msyh.TTF")
	if err != nil {
		fmt.Printf("ttfBytes1, err := os.ReadFile(\"msyh.TTF\"): %+v\r\n", err)
		ttfBytes1, err = os.ReadFile("C:/Windows/Fonts/simhei.TTF")
		if err != nil {
			fmt.Printf("ttfBytes1, err := os.ReadFile(\"C:/Windows/Fonts/simhei.TTF\"): %+v\r\n", err)
			return
		}
	}
	fontTTF, err := opentype.Parse(ttfBytes1)
	if err != nil {
		fmt.Printf("fontTTF, err := opentype.Parse(ttf): %+v\r\n", err)
		return
	}
	mincho := font.Font{Typeface: "Mincho"}
	font.DefaultCache.Add([]font.Face{
		{
			Font: mincho,
			Face: fontTTF,
		},
	})
	if !font.DefaultCache.Has(mincho) {
		fmt.Printf("!font.DefaultCache.Has(mincho) %q!\r\n", mincho.Typeface)
		return
	}
	plot.DefaultFont = mincho
	plotter.DefaultFont = mincho
}
func downloadExcel(变量是累积值 map[string]bool, 变量有非法值 string, rows *[][]string, w http.ResponseWriter, r *http.Request, 变量个数, 开始时间, 结束时间 string, 时间差, 间隔时间t time.Duration, 忽略异常值, 忽略特定值们, 计算差值, 表格历史空数据自动填充 string, 耗时, 访问耗时 time.Duration) {
	数据行数 := len(*rows)
	if 数据行数 < 1 {
		http.Error(w, "数据行数 := len(*rows),数据行数 < 1"+"_您的连接："+r.RemoteAddr, http.StatusInternalServerError)
		return
	}
	startTime := time.Now()
	当前时间 := time.Now().Format("2006-01-02 15:04:05")
	开始时间 = strings.Replace(开始时间, ":", "", -1)
	结束时间 = strings.Replace(结束时间, ":", "", -1)
	当前时间 = strings.Replace(当前时间, ":", "", -1)
	生成人 := strings.Replace(r.RemoteAddr, ":", " ", -1)
	每个变量目标记录数 := 时间差/间隔时间t + 1
	目标记录数 := fmt.Sprintf("目标记录数%d", 每个变量目标记录数)
	表格历史空数据自动填充2 := ""
	if 表格历史空数据自动填充 == "是" {
		表格历史空数据自动填充2 = "_填充空数据"
		for 行号, row := range *rows {
			if 行号 == 0 {
				continue
			}
			for 列号, val := range row {
				if 列号 == 0 {
					continue
				}
				if val == "" {
					if 行号 == 1 {
						for 行号2, row2 := range *rows {
							if 行号2 <= 1 {
								continue
							}
							if row2[列号] != "" {
								(*rows)[行号][列号] = (*rows)[行号2][列号]
								break
							}
						}
					} else { //if 行号==0{
						(*rows)[行号][列号] = (*rows)[行号-1][列号]
					} //}else{//if 行号==0{
				} //if val == "" {
			} //for 列号, val := range row {
		} //for 行号, row := range *rows {
	} //if 表格历史空数据自动填充{
	忽略异常值2 := ""
	if 忽略异常值 == "是" {
		忽略异常值2 = "_忽略异常值"
	}
	计算差值2 := ""
	if 计算差值 == "是" {
		计算差值2 = "_计算差值"
	}
	f := excelize.NewFile()
	index, err := f.NewSheet("Sheet1")
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Failed to index, err := f.NewSheet(Sheet1)"+"_您的连接："+r.RemoteAddr, http.StatusInternalServerError)
		return
	}
	f.SetActiveSheet(index)
	// rows := [][]string{
	// 	{"Name", "Age", "City"},
	// 	{"Alice", "30", "New York"},
	// 	{"Bob", "25", "Los Angeles"},
	// 	{"Charlie", "35", "Chicago"},
	// }
	for 行号, row := range *rows {
		for 列号, val := range row {
			cell, err := excelize.CoordinatesToCellName(列号+1, 行号+1) // 注意行列转换
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, "Failed to cell, err := excelize.CoordinatesToCellName(列号+1, 行号+1)"+"_您的连接："+r.RemoteAddr, http.StatusInternalServerError)
				return
			}
			f.SetCellValue("Sheet1", cell, val)
		}
	}
	有累积变量 := ""
	if 计算差值 == "是" {
		for 行号, row := range *rows {
			if 行号 == 0 {
				continue
			}
			for 列号, val := range row {
				if 列号 == 0 {
					continue
				}
				cell, err := excelize.CoordinatesToCellName(列号+1, 行号+1) // 注意行列转换
				if err != nil {
					fmt.Println(err.Error())
					http.Error(w, "Failed to cell, err := excelize.CoordinatesToCellName(列号+1, 行号+1)"+"_您的连接："+r.RemoteAddr, http.StatusInternalServerError)
					return
				}
				if 行号 == 1 {
					f.SetCellValue("Sheet1", cell, "0.00")
					continue
				}
				前值s := (*rows)[行号-1][列号]
				前值, err := strconv.ParseFloat(前值s, 64)
				if err != nil {
					f.SetCellValue("Sheet1", cell, "0.00")
					continue
				}
				当前值, err1 := strconv.ParseFloat(val, 64)
				if err1 != nil {
					f.SetCellValue("Sheet1", cell, "0.00")
					continue
				}
				差值 := 当前值 - 前值
				if 变量是累积值[(*rows)[0][列号]] {
					if 差值 < 0 {
						差值 = 0
					}
					有累积变量 = "_有累积变量"
				}
				f.SetCellValue("Sheet1", cell, fmt.Sprintf("%.2f", 差值))
			} //for 列号, val := range row {
		} //for 行号, row := range *rows {
	} //if 计算差值 == "是" {
	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Failed to _, err := f.WriteTo(&buf)"+"_您的连接："+r.RemoteAddr, http.StatusInternalServerError)
		return
	}
	忽略特定值们2 := ""
	if 忽略特定值们 != "" {
		忽略特定值们2 = "_忽略值" + 忽略特定值们
	}
	elapsedTime := time.Since(startTime)
	耗时2 := fmt.Sprintf("%v", elapsedTime+耗时)
	耗时3 := fmt.Sprintf("%v", elapsedTime+访问耗时)
	filename := "SQLiteDB查询_" + 当前时间 + "_" + 生成人 + "_" + 开始时间 + "~" + 结束时间 + "_" + 时间差.String() + "_" + 间隔时间t.String() + "_" + 目标记录数 + "_变量个数" + 变量个数 + 忽略异常值2 + 变量有非法值 + 计算差值2 + 有累积变量 + 表格历史空数据自动填充2 + "_" + 耗时2 + "_" + 耗时3 + 忽略特定值们2
	filename2 := filename + ".xlsx"
	w.Header().Set("Content-Disposition", "attachment; filename="+"\""+filename2+"\"")
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))
	filename2 = 目录2 + filename2
	长度 := len(filename2)
	const 长度限制 = 207
	const 扩展名长度 = 5
	if 长度 > 长度限制 {
		fmt.Printf(filename2+" 长度%d\r\n", 长度)
		filename2 = 目录2 + filename
		i := 长度限制 - 扩展名长度
		文件名及路径 := ""
		迭代次数 := 1
		for i > 0 {
			文件名及路径 = 安全截取中文字符串(filename2, i) + ".xlsx"
			if len(文件名及路径) > 长度限制 {
				i--
				迭代次数++
			} else {
				filename2 = 文件名及路径
				fmt.Printf("安全截取中文字符串迭代次数%d\r\n", 迭代次数)
				break
			}
		}
		fmt.Printf("修改为：%s,中文长度%d,字节长度%d\r\n", filename2, utf8.RuneCountInString(filename2), len(filename2))
	}
	err = f.SaveAs(filename2)
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = w.Write(buf.Bytes())
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Failed to _, err = w.Write(buf.Bytes())"+"_您的连接："+r.RemoteAddr, http.StatusInternalServerError)
		return
	}
}

var SQLiteDB复制被访问次数 uint64
var SQLiteDB复制情况 字符串不重复累加互斥锁访问结构体
var 上次SQLiteDB复制时刻 int64

const SQLiteDB复制间隔秒数 = 60
const SQLiteDB复制间隔秒数s = "60"

func SQLiteDB复制(w http.ResponseWriter, r *http.Request) {
	go 获取服务器域名(r.Host, r.RemoteAddr)
	ip := 设定的连接信息处理(w, r, &SQLiteDB复制被访问次数, "SQLiteDB复制")
	if ip == "" {
		return
	}
	if time.Now().Unix()-atomic.LoadInt64(&上次SQLiteDB复制时刻) < SQLiteDB复制间隔秒数 {
		atomic.StoreInt64(&上次SQLiteDB复制时刻, time.Now().Unix())
		w.Write([]byte("距离上次SQLiteDB复制小于SQLiteDB复制间隔秒数(" + SQLiteDB复制间隔秒数s + ")设定，本次触发无效！如果有多个触发源，那么必须在SQLiteDB复制间隔秒数内没有触发后再才有效！"))
		return
	}
	atomic.StoreInt64(&上次SQLiteDB复制时刻, time.Now().Unix())

	type ErrCode struct {
		Code    int    `json:"Code"`
		Message string `json:"Message"`
	}
	query := r.URL.Query()
	user := query.Get("user")
	password := query.Get("password")
	var C错误码 ErrCode
	if user != 本网关登录名之MD5 || password != 本网关登录密码之MD5 {
		C错误码.Code = 1
		C错误码.Message = "用户名或密码错误！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	if atomic.LoadUint32(&已经触发重启程序了) == 1 {
		C错误码.Code = 1
		C错误码.Message = "已经触发重启程序了" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	被复制的库文件们 := query.Get("被复制的库文件们")
	if 被复制的库文件们 == "" {
		C错误码.Code = 2
		C错误码.Message = "被复制的库文件们不能为空" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	被复制的库文件们 = strings.Replace(被复制的库文件们, "，", ",", -1)
	被复制的库文件组 := strings.Split(被复制的库文件们, ",")
	长度 := len(被复制的库文件组)
	if 长度 < 1 {
		var C错误码 ErrCode
		C错误码.Code = 4
		C错误码.Message = "J没有被复制的库文件" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	for i := 0; i < 长度; i++ {
		if 被复制的库文件组[i] == "" {
			var C错误码 ErrCode
			C错误码.Code = 5
			C错误码.Message = "有空的被复制的库文件名！" + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	for i := 0; i < 长度; i++ {
		错误信息, _ := 获取文件信息(被复制的库文件组[i])
		if 错误信息 != "" {
			var C错误码 ErrCode
			C错误码.Code = 6
			C错误码.Message = "错误的库文件：" + 被复制的库文件组[i] + "_" + 错误信息 + "_您的连接: " + r.RemoteAddr
			fmt.Println(C错误码.Message)
			jsonStr, err3 := json.Marshal(C错误码)
			if err3 != nil {
				w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
				return
			}
			w.Write([]byte(jsonStr))
			return
		}
	} //for i := 0; i < 长度; i++ {
	var 库文件名们 = make(map[string]bool, 0)
	for i := 0; i < 长度; i++ {
		if !库文件名们[被复制的库文件组[i]] {
			库文件名们[被复制的库文件组[i]] = true
			continue
		}
		var C错误码 ErrCode
		C错误码.Code = 7
		C错误码.Message = "库文件名中有重复！" + "_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	} //for i := 0; i < 长度; i++ {
	SQLiteDB文件 := SQLiteDB文件名及路径信息.Load()
	SQLiteDB访问锁.Lock()
	defer SQLiteDB访问锁.Unlock()
	db, err := sql.Open("sqlite3", SQLiteDB文件)
	if err != nil {
		C错误码.Code = 12
		C错误码.Message = err.Error() + "(" + SQLiteDB文件 + ")_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	defer db.Close()
	// 确保数据库连接可用
	err = db.Ping()
	if err != nil {
		C错误码.Code = 16
		C错误码.Message = err.Error() + "(" + SQLiteDB文件 + ")_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	var 批量插入预编译的语句 *sql.Stmt
	var 检查重复记录预编译的语句 *sql.Stmt
	var err1 error
	func() {
		插入语句 := "INSERT INTO " + SQLite表格名 + " (变量名, 变量值, 时间) VALUES (?, ?, ?)"
		批量插入预编译的语句, err1 = db.Prepare(插入语句)
		if err1 != nil {
			fmt.Printf("目标库无法准备插入语句: %v\r\n", err1)
			return
		}
		检查重复语句 := "SELECT 1 FROM " + SQLite表格名 + " WHERE 变量名 = ? AND 时间= ?"
		检查重复记录预编译的语句, err1 = db.Prepare(检查重复语句)
		if err1 != nil {
			fmt.Printf("目标库无法准备检查重复记录的语句: %v\r\n", err1)
			return
		}
	}()
	if 检查重复记录预编译的语句 != nil {
		defer 检查重复记录预编译的语句.Close()
	}
	if 批量插入预编译的语句 != nil {
		defer 批量插入预编译的语句.Close()
	}
	if 检查重复记录预编译的语句 == nil || 批量插入预编译的语句 == nil {
		C错误码.Code = 12
		C错误码.Message = "检查重复记录预编译的语句 == nil || 批量插入预编译的语句 == nil(" + SQLiteDB文件 + ")_您的连接: " + r.RemoteAddr
		fmt.Println(C错误码.Message)
		jsonStr, err3 := json.Marshal(C错误码)
		if err3 != nil {
			w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
			return
		}
		w.Write([]byte(jsonStr))
		return
	}
	获取记录查询语句 := "SELECT * FROM " + SQLite表格名 + " WHERE id > 0"
	复制情况 := ""
	for _, 库文件名 := range 被复制的库文件组 {
		if 库文件名 == "" {
			continue
		}
		需要将记录写入目标数据库 := true
		var 批量处理的数据切片 []interface{}
		要写记录数 := 0
		func() {
			db1, err := sql.Open("sqlite3", 库文件名)
			if err != nil {
				C错误码.Code = 12
				C错误码.Message = err.Error() + "(" + 库文件名 + ")_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			defer db1.Close()
			err = db1.Ping()
			if err != nil {
				C错误码.Code = 16
				C错误码.Message = err.Error() + "(" + 库文件名 + ")_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			var 记录数 int
			获取记录数查询语句 := "SELECT COUNT(*) FROM " + SQLite表格名 + " WHERE id > 0"
			startTime := time.Now()
			row := db1.QueryRow(获取记录数查询语句)
			if err := row.Scan(&记录数); err != nil {
				C错误码.Code = 19
				C错误码.Message = err.Error() + "_row.Scan(&记录数)(" + 库文件名 + ")_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			获取记录数耗时 := time.Since(startTime)
			获取记录数耗时s := fmt.Sprintf("库文件名:%s,记录数%d,获取记录数耗时%v\r\n", 库文件名, 记录数, 获取记录数耗时)
			fmt.Println(获取记录数耗时s)
			if 记录数 == 0 {
				return
			}
			rows, err := db1.Query(获取记录查询语句)
			if err != nil {
				C错误码.Code = 19
				C错误码.Message = err.Error() + "_db1.Query(获取记录查询语句)(" + 库文件名 + ")_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			defer rows.Close()
			var id uint64
			var 变量名 string
			var 变量值 string
			var 时间 time.Time
			for rows.Next() {
				err := rows.Scan(&id, &变量名, &变量值, &时间)
				if err != nil {
					rows.Close()
					C错误码.Code = 20
					C错误码.Message = err.Error() + "(" + 库文件名 + ")_您的连接: " + r.RemoteAddr
					fmt.Println(C错误码.Message)
					jsonStr, err3 := json.Marshal(C错误码)
					if err3 != nil {
						w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
						return
					}
					w.Write([]byte(jsonStr))
					return
				}
				if 需要将记录写入目标数据库 {
					var exists bool
					err = 检查重复记录预编译的语句.QueryRow(变量名, 时间).Scan(&exists)
					if err != nil && err != sql.ErrNoRows {
						str := fmt.Sprintf("目标库检查重复记录失败: %v\r\n", err)
						复制情况 += str
						需要将记录写入目标数据库 = false
						fmt.Println(str)
						break
					}
					if exists {
						continue
					} //if (!exists&&exists2)||(exists&&!exists2) {
					批量处理的数据切片 = append(批量处理的数据切片, 变量名, 变量值, 时间)
					要写记录数++
				}
			} //for rows.Next() {
			if err = rows.Err(); err != nil {
				C错误码.Code = 20
				C错误码.Message = err.Error() + "(" + 库文件名 + ")_您的连接: " + r.RemoteAddr
				fmt.Println(C错误码.Message)
				jsonStr, err3 := json.Marshal(C错误码)
				if err3 != nil {
					w.Write([]byte("\r\n服务器执行 json Marshal(C错误码) error !"))
					return
				}
				w.Write([]byte(jsonStr))
				return
			}
			startTime2 := time.Now()
			if 要写记录数 > 0 && 需要将记录写入目标数据库 {
				if err := execBatch(批量插入预编译的语句, 批量处理的数据切片); err != nil {
					str := fmt.Sprintf("execBatch(批量插入预编译的语句, 批量处理的数据切片)失败: %v\r\n", err)
					复制情况 += str
					需要将记录写入目标数据库 = false
					fmt.Println(str)
					return
				}
			}
			处理记录耗时 := time.Since(startTime2)
			处理记录耗时s := fmt.Sprintf("库文件名:%s,总记录数%d,写目标库记录数%d,写耗时%v\r\n", 库文件名, 记录数, 要写记录数, 处理记录耗时)
			fmt.Println(处理记录耗时s)
			复制情况 += 处理记录耗时s
		}() //func() {
	} //for _,变量名:=range 被复制的库文件{
	变量数据 := ""
	str1 := ""
	str := 编译时间 + "\r\n"
	str1 += str
	str = 至今多少天时分秒(网关启动时刻)
	str = "启动至今: " + str + "\r\n"
	str1 += str
	str = 网关启动时间
	str1 += str
	当前时间 := time.Now().Format("2006-01-02 15:04:05")
	变量数据 = str1 + "\r\n" + 当前时间 + "\r\n" + 复制情况
	w.Write([]byte(变量数据))
	str = time.Now().Format("2006-01-02 15:04:05") + "_" + r.RemoteAddr + "\r\n" + 复制情况
	SQLiteDB复制情况.Set(str)
} //func SQLiteDB复制(w http.ResponseWriter, r *http.Request) {

// 定义数据结构
type 哪些表行要写保存结构体 struct {
	mu      sync.RWMutex
	Y要写的表行们 []int
}

// 初始化 哪些表行要写保存结构体
func New哪些表行要写保存结构体() *哪些表行要写保存结构体 {
	return &哪些表行要写保存结构体{
		Y要写的表行们: make([]int, 0),
	}
}
func (nm *哪些表行要写保存结构体) 获取数组对象() ([]int, bool) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	if len(nm.Y要写的表行们) < 1 {
		return nil, false
	}
	return nm.Y要写的表行们, true
}
func (nm *哪些表行要写保存结构体) AddOrUpdate(表行 int) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	for _, v := range nm.Y要写的表行们 {
		if v == 表行 {
			return
		}
	}
	nm.Y要写的表行们 = append(nm.Y要写的表行们, 表行)
}
func (nm *哪些表行要写保存结构体) Delete(表行 int) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	var slice = make([]int, 0)
	更新了 := false
	for i, v := range nm.Y要写的表行们 {
		if v == 表行 {
			slice = append(nm.Y要写的表行们[:i], nm.Y要写的表行们[i+1:]...)
			更新了 = true
			break
		}
	}
	if 更新了 {
		nm.Y要写的表行们 = slice
	}
}

var 哪些表行要写保存 = New哪些表行要写保存结构体()

// 定义数据结构
type 哪个串口有哪些变量要写操作结构体 struct {
	mu       sync.RWMutex
	outerMap map[string][]map[string]interface{}
}

// 初始化 哪个串口有哪些变量要写操作结构体
func New哪个串口有哪些变量要写操作结构体() *哪个串口有哪些变量要写操作结构体 {
	return &哪个串口有哪些变量要写操作结构体{
		outerMap: make(map[string][]map[string]interface{}),
	}
}

var 哪个串口有哪些变量要写操作 = New哪个串口有哪些变量要写操作结构体()

func (nm *哪个串口有哪些变量要写操作结构体) 获取数组对象(key string) ([]map[string]interface{}, bool) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	var 对象数组 []map[string]interface{}
	var exists bool
	if 对象数组, exists = nm.outerMap[key]; !exists {
		return nil, false
	}
	if len(对象数组) < 1 {
		return nil, false
	}
	return nm.outerMap[key], true
}

// 向数组中添加或更新元素
func (nm *哪个串口有哪些变量要写操作结构体) AddOrUpdate(key string, innerKey string, innerValue interface{}) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	// fmt.Println(key)
	// fmt.Println(innerKey)
	// fmt.Println(innerValue)
	// 检查 outerMap 中是否存在 key
	if _, exists := nm.outerMap[key]; !exists {
		nm.outerMap[key] = []map[string]interface{}{{innerKey: innerValue}}
		return
	}
	// 查找 innerKey 并更新值，或添加新元素
	innerMaps := nm.outerMap[key]
	for i, innerMap := range innerMaps {
		if _, exists := innerMap[innerKey]; exists {
			innerMaps[i][innerKey] = innerValue
			return
		}
	}
	// 如果找不到 innerKey，添加新元素
	newInnerMap := make(map[string]interface{})
	newInnerMap[innerKey] = innerValue
	nm.outerMap[key] = append(innerMaps, newInnerMap)
}

// 删除数组元素
func (nm *哪个串口有哪些变量要写操作结构体) Delete(key string, innerKey string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	if innerMaps, exists := nm.outerMap[key]; exists {
		for i, innerMap := range innerMaps {
			if _, exists := innerMap[innerKey]; exists {
				delete(innerMap, innerKey)
				if len(innerMap) == 0 {
					nm.outerMap[key] = append(innerMaps[:i], innerMaps[i+1:]...)
				}
				return
			}
		}
	}
}

// 读取数据
func (nm *哪个串口有哪些变量要写操作结构体) Get(key string, innerKey string) (interface{}, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	if innerMaps, exists := nm.outerMap[key]; exists {
		for _, innerMap := range innerMaps {
			if value, exists := innerMap[innerKey]; exists {
				return value, true
			}
		}
	}
	return "", false
}
func 项目变量信息表组当前值和初始值写更新(要写的变量名 string, 要写的值 string) {
	项目变量信息表组.Rows[变量名所在行[要写的变量名]].D当前值锁.Lock()
	defer 项目变量信息表组.Rows[变量名所在行[要写的变量名]].D当前值锁.Unlock()
	项目变量信息表组.Rows[变量名所在行[要写的变量名]].D当前值 = 要写的值
	if 项目变量信息表组.Rows[变量名所在行[要写的变量名]].S是否保存值 != "是" {
		return
	}
	if 项目变量信息表组.Rows[变量名所在行[要写的变量名]].C初始值 == 要写的值 {
		return
	}
	项目变量信息表组.Rows[变量名所在行[要写的变量名]].C初始值 = 要写的值
	哪些表行要写保存.AddOrUpdate(变量名所在行[要写的变量名])
} //func 项目变量信息表组当前值和初始值写更新(要写的变量名 string, 要写的值 string) {
func areFloatsEqual(a, b float64) bool {
	const epsilon = 1e-9 // 定义一个小的容差值
	return math.Abs(a-b) < epsilon
}
func 将写操作记录到相关map中(w *sync.WaitGroup, z *字符串累加互斥锁访问结构体, 要写的变量名 string, 要写的值 string) {
	if w != nil {
		w.Done()
	}
	项目变量信息表组.Rows[变量名所在行[要写的变量名]].D当前值锁.Lock()
	当前值 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].D当前值
	项目变量信息表组.Rows[变量名所在行[要写的变量名]].D当前值锁.Unlock()
	变量类型 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].B变量类型
	数据类型 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].S数据类型
	串口号 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].C串口号
	连续通讯失败次数 := atomic.LoadUint32(&项目变量信息表组.Rows[变量名所在行[要写的变量名]].L连续通讯失败次数)
	执行结果 := ""
	// str := fmt.Sprintf(要写的变量名+"变量名所在行：%d", 变量名所在行[要写的变量名])
	// fmt.Println(str)
	// str = fmt.Sprintf(要写的变量名+"变量名所在表行：%d", 变量名所在表行[要写的变量名])
	// fmt.Println(str)
	if 连续通讯失败次数 > 0 {
		执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为" + "此变量连续通讯失败次数大于零：" + fmt.Sprintf("%d", 连续通讯失败次数)
		z.Set(执行结果)
		return
	}
	var 要写的值3 interface{}
	要写 := false
	switch 数据类型 {
	case "FLOAT":
		要写的值2, err := strconv.ParseFloat(要写的值, 32)
		if err != nil {
			执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值非法：" + err.Error()
			z.Set(执行结果)
			return
		}
		当前值2, err := strconv.ParseFloat(当前值, 32)
		if err != nil {
			执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为当前值非法：" + err.Error()
			z.Set(执行结果)
			return
		}
		if !areFloatsEqual(要写的值2, 当前值2) {
			要写 = true
			要写的值3 = float32(要写的值2)
			执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
			z.Set(执行结果)
		} else {
			执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值和当前值相等"
			z.Set(执行结果)
			return
		}
	case "LONGBCD", "ULONG", "BCD", "USHORT":
		switch 变量类型 {
		case "IO整型":
			要写的值2, err := strconv.ParseUint(要写的值, 10, 32)
			if err != nil {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值非法：" + err.Error()
				z.Set(执行结果)
				return
			}
			当前值2, err := strconv.ParseUint(当前值, 10, 32)
			if err != nil {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为当前值非法：" + err.Error()
				z.Set(执行结果)
				return
			}
			if 要写的值2 != 当前值2 {
				要写 = true
				switch 数据类型 {
				case "LONGBCD":
					要写的值3 = uint32(要写的值2)
				case "ULONG":
					要写的值3 = uint32(要写的值2)
				case "BCD":
					要写的值3 = uint16(要写的值2)
				case "USHORT":
					要写的值3 = uint16(要写的值2)
				}
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
			} else {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值和当前值相等"
				z.Set(执行结果)
				return
			}
		case "IO实型":
			要写的值2, err := strconv.ParseFloat(要写的值, 32)
			if err != nil {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值非法：" + err.Error()
				z.Set(执行结果)
				return
			}
			当前值2, err := strconv.ParseFloat(当前值, 32)
			if err != nil {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为当前值非法：" + err.Error()
				z.Set(执行结果)
				return
			}
			if !areFloatsEqual(要写的值2, 当前值2) {
				要写 = true
				要写的值3 = float32(要写的值2)
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
			} else {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值和当前值相等"
				z.Set(执行结果)
				return
			}
		default:
			执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为未知数据类型或变量类型"
			z.Set(执行结果)
			return
		}
	case "LONG", "SHORT":
		switch 变量类型 {
		case "IO整型":
			要写的值2, err := strconv.ParseInt(要写的值, 10, 32)
			if err != nil {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值非法：" + err.Error()
				z.Set(执行结果)
				return
			}
			当前值2, err := strconv.ParseInt(当前值, 10, 32)
			if err != nil {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为当前值非法：" + err.Error()
				z.Set(执行结果)
				return
			}
			if 要写的值2 != 当前值2 {
				要写 = true
				switch 数据类型 {
				case "LONG":
					要写的值3 = int32(要写的值2)
				case "SHORT":
					要写的值3 = int16(要写的值2)
				}
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
			} else {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值和当前值相等"
				z.Set(执行结果)
				return
			}
		case "IO实型":
			要写的值2, err := strconv.ParseFloat(要写的值, 32)
			if err != nil {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值非法：" + err.Error()
				z.Set(执行结果)
				return
			}
			当前值2, err := strconv.ParseFloat(当前值, 32)
			if err != nil {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为当前值非法：" + err.Error()
				z.Set(执行结果)
				return
			}
			if !areFloatsEqual(要写的值2, 当前值2) {
				要写 = true
				要写的值3 = float32(要写的值2)
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
			} else {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值和当前值相等"
				z.Set(执行结果)
				return
			}
		default:
			执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为未知数据类型或变量类型"
			z.Set(执行结果)
			return
		}
	default:
		switch 变量类型 {
		case "IO离散":
			if 要写的值 != 当前值 {
				要写 = true
				switch 要写的值 {
				case "0":
					要写的值3 = false
				case "1":
					要写的值3 = true
				default:
					执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值非法" + "不是1或0"
					z.Set(执行结果)
					return
				}
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
			} else {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值和当前值相等"
				z.Set(执行结果)
				return
			}
		case "IO字符串":
			if 要写的值 != 当前值 {
				要写 = true
				要写的值3 = 要写的值
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
			} else {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值和当前值相等"
				z.Set(执行结果)
				return
			}
		case "内存离散":
			if 要写的值 != "0" && 要写的值 != "1" {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值非法" + "不是1或0"
				z.Set(执行结果)
				return
			}
			if 要写的值 != 当前值 {
				项目变量信息表组当前值和初始值写更新(要写的变量名, 要写的值)
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
				return
			} else {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值和当前值相等"
				z.Set(执行结果)
				return
			}
		case "内存字符串":
			if 要写的值 != 当前值 {
				项目变量信息表组当前值和初始值写更新(要写的变量名, 要写的值)
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
				return
			} else {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值和当前值相等"
				z.Set(执行结果)
				return
			}
		case "内存整型":
			要写的值2, err := strconv.ParseInt(要写的值, 10, 64)
			if err != nil {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值非法：" + err.Error()
				z.Set(执行结果)
				return
			}
			当前值2, err := strconv.ParseInt(当前值, 10, 64)
			不用比较 := false
			if err != nil {
				不用比较 = true
			}
			if 不用比较 {
				项目变量信息表组当前值和初始值写更新(要写的变量名, 要写的值)
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
				return
			}
			if 要写的值2 != 当前值2 {
				项目变量信息表组当前值和初始值写更新(要写的变量名, 要写的值)
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
				return
			} else {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值和当前值相等"
				z.Set(执行结果)
				return
			}
		case "内存实型":
			要写的值2, err := strconv.ParseFloat(要写的值, 64)
			if err != nil {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值非法：" + err.Error()
				z.Set(执行结果)
				return
			}
			当前值2, err := strconv.ParseFloat(当前值, 64)
			不用比较 := false
			if err != nil {
				不用比较 = true
			}
			if 不用比较 {
				项目变量信息表组当前值和初始值写更新(要写的变量名, 要写的值)
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
				return
			}
			if !areFloatsEqual(要写的值2, 当前值2) {
				项目变量信息表组当前值和初始值写更新(要写的变量名, 要写的值)
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写受理"
				z.Set(执行结果)
				return
			} else {
				执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为要写的值和当前值相等"
				z.Set(执行结果)
				return
			}
		default:
			执行结果 = 要写的变量名 + "=" + 要写的值 + "_写失败，因为未知数据类型或变量类型"
			z.Set(执行结果)
			return
		} //switch 变量类型 {
	} //switch 数据类型 {
	if 要写 {
		哪个串口有哪些变量要写操作.AddOrUpdate(串口号, 要写的变量名, 要写的值3)
	}
} //func 将写操作记录到相关map中(要写的变量名 string, 要写的值 string) {
func 写值是否合法(要写的变量名 string, 要写的值 string) string {
	//项目变量信息表组[变量名所在行[要写的变量名]]
	读写属性 := 项目变量信息表组.Rows[变量名所在行[要写的变量名]].D读写属性
	if 读写属性 == "只读" {
		return 要写的变量名 + "_只读"
	}
	return 变量值是否合法(要写的变量名, 要写的值)
} //func 写值是否合法(要写的变量名 string, 要写的值 string) string {
func 变量值是否合法(变量名 string, 变量值 string) string {
	//项目变量信息表组[变量名所在行[变量名]]
	变量名所在行指针 := &项目变量信息表组.Rows[变量名所在行[变量名]]
	变量类型 := 变量名所在行指针.B变量类型
	数据类型 := 变量名所在行指针.S数据类型

	switch 数据类型 {
	case "FLOAT":
		值, err := strconv.ParseFloat(变量值, 32)
		if err != nil {
			return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
		}
		if 值 < 最小浮点数 || 值 > 最大浮点数 {
			return 变量名 + "的值：" + 变量值 + ">>不在数据类型FLOAT正常范围内(" + 最小浮点数_s + "~" + 最大浮点数_s + ")"
		}
	case "LONGBCD":
		if 变量类型 == "IO实型" {
			值, err := strconv.ParseFloat(变量值, 32)
			if err != nil {
				return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
			}
			最大值 := 变量名所在行指针.J计算值除原始值 * 99999999
			if 值 > 最大值 {
				return 变量名 + "的值：" + 变量值 + ">>不在正常范围内(0~" + fmt.Sprintf("%f", 最大值) + ")"
			}
		} else {
			numUint64, err := strconv.ParseUint(变量值, 10, 32) // 第三个参数指定解析结果的位数为 16
			if err != nil {
				return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
			}
			if numUint64 > 99999999 {
				return 变量名 + "的值：" + 变量值 + ">>不在数据类型LONGBCD正常范围内(0~99999999)"
			}
		}
	case "LONG":
		值, err := strconv.ParseInt(变量值, 10, 32)
		if err != nil {
			return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
		}
		if 值 < int32最小值 || 值 > int32最大值 {
			return 变量名 + "的值：" + 变量值 + ">>不在数据类型LONG正常范围内(" + int32最小值_s + "~" + int32最大值_s + ")"
		}
	case "ULONG":
		值, err := strconv.ParseUint(变量值, 10, 32)
		if err != nil {
			return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
		}
		if 值 > 4294967295 {
			return 变量名 + "的值：" + 变量值 + ">>不在数据类型ULONG正常范围内(" + "0" + "~" + "4294967295" + ")"
		}
	case "BCD":
		if 变量类型 == "IO实型" {
			值, err := strconv.ParseFloat(变量值, 32)
			if err != nil {
				return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
			}
			最大值 := 变量名所在行指针.J计算值除原始值 * 9999
			if 值 > 最大值 {
				return 变量名 + "的值：" + 变量值 + ">>不在正常范围内(0~" + fmt.Sprintf("%f", 最大值) + ")"
			}
		} else {
			numUint64, err := strconv.ParseUint(变量值, 10, 16) // 第三个参数指定解析结果的位数为 16
			if err != nil {
				return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
			}
			if numUint64 > 9999 {
				return 变量名 + "的值：" + 变量值 + ">>不在数据类型BCD正常范围内(0~9999)"
			}
		}
	case "SHORT":
		if 变量类型 == "IO实型" {
			值, err := strconv.ParseFloat(变量值, 32)
			if err != nil {
				return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
			}
			最大值 := 变量名所在行指针.J计算值除原始值 * 32767
			最小值 := 变量名所在行指针.J计算值除原始值 * -32768
			if 值 > 最大值 || 值 < 最小值 {
				return 变量名 + "的值：" + 变量值 + ">>不在正常范围内(" + fmt.Sprintf("%f", 最小值) + "~" + fmt.Sprintf("%f", 最大值) + ")"
			}
		} else {
			值, err := strconv.ParseInt(变量值, 10, 16) // 第三个参数指定解析结果的位数为 16
			if err != nil {
				return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
			}
			if 值 > 32767 || 值 < -32768 {
				return 变量名 + "的值：" + 变量值 + ">>不在数据类型SHORT正常范围内(-32768~32767)"
			}
		}
	case "USHORT":
		if 变量类型 == "IO实型" {
			值, err := strconv.ParseFloat(变量值, 32)
			if err != nil {
				return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
			}
			最大值 := 变量名所在行指针.J计算值除原始值 * 65535
			if 值 > 最大值 {
				return 变量名 + "的值：" + 变量值 + ">>不在正常范围内(0~" + fmt.Sprintf("%f", 最大值) + ")"
			}
		} else {
			numUint64, err := strconv.ParseUint(变量值, 10, 16) // 第三个参数指定解析结果的位数为 16
			if err != nil {
				return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
			}
			if numUint64 > 65535 {
				return 变量名 + "的值：" + 变量值 + ">>不在数据类型USHORT正常范围内(0~65535)"
			}
		}
	default:
		switch 变量类型 {
		case "IO离散", "内存离散":
			if 变量值 != "1" && 变量值 != "0" {
				return 变量名 + "的值：" + 变量值 + ">>非法：不是1又不是0"
			}
		case "IO字符串":
			if 变量值 == 变量名所在行指针.T通讯异常值 {
				return 变量名 + "的值：" + 变量值 + ">>非法：是T通讯异常值"
			}
			字符串长度 := len(变量值)
			if 字符串长度 > 最大字符串字节长度 {
				return 变量名 + "的值：" + 变量值 + ">>非法：字符串长度 > IO字符串最多占用字节数" + 最大字符串字节长度_s
			}
		case "内存字符串":
			字符串长度 := len(变量值)
			if 字符串长度 > 内存字符串最多占用字节数 {
				return 变量名 + "的值：" + 变量值 + ">>非法：字符串长度 > 内存字符串最多占用字节数" + 内存字符串最多占用字节数_s
			}
		case "内存整型":
			_, err := strconv.ParseInt(变量值, 10, 64)
			if err != nil {
				return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
			}
		case "内存实型":
			_, err := strconv.ParseFloat(变量值, 64)
			if err != nil {
				return 变量名 + "的值：" + 变量值 + ">>非法：" + err.Error()
			}
		default:
			return 变量名 + "的类型(" + 变量类型 + ")是未知类型"
		}
	} //switch 数据类型 {
	return "ok"
} //func 变量值是否合法(变量名 string, 变量值 string) string {
func 设定的连接信息处理(w http.ResponseWriter, r *http.Request, 被访问次数 *uint64, fieldName string) string {
	const 同一个ip最小设置间隔秒数 = 1
	ip := 获取访问连接中的ip(r.RemoteAddr)
	连接信息锁.Lock()
	defer 连接信息锁.Unlock()
	if 连接信息[ip] == nil {
		p := make([]连接信息2, 1)
		连接信息[ip] = &p[0]
		if ok, v := 外网ip访问判断(r.RemoteAddr); ok {
			go ips属地输入(v)
		}
	}
	s := 连接信息[ip]
	fmt.Println(ip)
	err := incrementField(s, fieldName)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	连接信息[ip].L连接时间 = time.Now().Format("2006-01-02 15:04:05")
	atomic.AddUint64(被访问次数, 1)
	if time.Now().Unix()-连接信息[ip].L连接时刻秒 < 同一个ip最小设置间隔秒数 {
		str := "设置过于频繁，请在同一个ip最小设置间隔秒数后再操作！" + "\r\n您的连接: " + r.RemoteAddr
		w.Write([]byte(str))
		return ""
	}
	连接信息[ip].L连接时刻秒 = time.Now().Unix()
	if fieldName == "SetTagValue" {
		连接信息[ip].SetTagValue请求信息 = r.RequestURI
	}
	return ip
} //func 连接信息处理(RemoteAddr string, 被访问次数 *uint32) {

// GetDiskInfoWindows 获取Windows系统中的硬盘信息
func GetDiskInfoWindows() string {
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	procGetLogicalDrives := kernel32.MustFindProc("GetLogicalDrives")
	procGetDiskFreeSpaceEx := kernel32.MustFindProc("GetDiskFreeSpaceExW")

	drives, _, _ := procGetLogicalDrives.Call()
	结果 := ""
	for drive := 'A'; drive <= 'Z'; drive++ {
		if drives&(1<<(drive-'A')) != 0 {
			rootPath := string(drive) + `:\`
			fi, err := os.Stat(rootPath)
			if err != nil {
				continue
			}
			if !fi.IsDir() {
				continue
			}

			var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
			ptr, err := syscall.UTF16PtrFromString(rootPath)
			if err != nil {
				continue
			}
			procGetDiskFreeSpaceEx.Call(
				uintptr(unsafe.Pointer(ptr)),
				uintptr(unsafe.Pointer(&freeBytesAvailable)),
				uintptr(unsafe.Pointer(&totalNumberOfBytes)),
				uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
			)
			结果 += fmt.Sprintf("驱动器: %s\n  磁盘空间: %d bytes (%s)\n  可用空间: %d bytes (%s)\n", rootPath, totalNumberOfBytes, humanReadableSize(totalNumberOfBytes), freeBytesAvailable, humanReadableSize(freeBytesAvailable))
		}
	}
	fmt.Println(结果)
	return 结果
}
func humanReadableSize(size uint64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	_, exp := uint64(unit), 0
	for size >= unit && exp < 4 {
		size /= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(size)/unit, "KMGTPEZY"[exp])
}

// 检查SQLite数据库文件格式是否正确
// 这里简单地检查文件大小是否大于0，实际应用中可能需要更复杂的检查
// func isSQLiteDBFormatCorrect(filePath string) (bool, error) {
// 	fileInfo, err := os.Stat(filePath)
// 	if err != nil {
// 		return false, err
// 	}
// 	return fileInfo.Size() > 0, nil
// }

var SQLite错误信息 字符串不重复累加互斥锁访问结构体

func 创建复合索引函数(参数们 ...string) string {
	参数个数 := len(参数们)
	if 参数个数 < 2 || 参数个数 > 数据库表格结构列数 {
		return ""
	}
	表名 := ""
	索引名 := ""
	索引结构 := ""
	for i, 参数 := range 参数们 {
		if i == 0 {
			表名 = 参数
			continue
		}
		索引名 += 参数 + "_"
		索引结构 += 参数 + ","
	}
	索引名 = 索引名[:len(索引名)-1]
	索引结构 = 索引结构[:len(索引结构)-1]
	索引结构 = strings.Replace(索引结构, ",", ", ", -1)
	return "CREATE INDEX IF NOT EXISTS idx_" + 索引名 + " ON " + 表名 + " (" + 索引结构 + ")"
} //func 创建复合索引函数(参数们 ...string) string {
const 自增主键名 = "id"

var 预期SQLite表格结构 = []ExpectedTableSchema{
	{ColumnName: 自增主键名, DataType: "INTEGER", NotNull: false, PrimaryKey: true, Autoincr: true}, //主键没必要再声明NOT NULL
	{ColumnName: "变量名", DataType: "TEXT", NotNull: true, PrimaryKey: false, Autoincr: false},
	{ColumnName: "变量值", DataType: "TEXT", NotNull: true, PrimaryKey: false, Autoincr: false},
	{ColumnName: "时间", DataType: "DATETIME", NotNull: true, PrimaryKey: false, Autoincr: false},
}

const 数据库表格结构列数 = 3

var 创建SQLite表格结构语句 = ""

func 创建SQLite表格结构函数(表名 string, 是内存数据库 bool) string {
	if 创建SQLite表格结构语句 != "" {
		return 创建SQLite表格结构语句
	}
	结果1 := `
	CREATE TABLE IF NOT EXISTS `
	结果3 := "("
	for _, 结构 := range 预期SQLite表格结构 {
		NotNull := ""
		if 结构.NotNull {
			NotNull = "NOT NULL"
		}
		PrimaryKey := ""
		if 结构.PrimaryKey {
			PrimaryKey = "PRIMARY KEY"
		}
		Autoincr := ""
		if 结构.Autoincr && !是内存数据库 {
			Autoincr = "AUTOINCREMENT"
		}

		结果3 += 结构.ColumnName + " " + 结构.DataType
		if NotNull != "" {
			结果3 += " " + NotNull
		}
		if PrimaryKey != "" {
			结果3 += " " + PrimaryKey
		}
		if Autoincr != "" {
			结果3 += " " + Autoincr
		}
		结果3 += ","
	}
	// 将字符串转换为rune切片，以便处理多字节字符
	runes := []rune(结果3)
	// 替换最后一个字符
	runes[len(runes)-1] = ')'
	// 将rune切片转换回字符串
	结果3 = string(runes)
	创建SQLite表格结构语句 = 结果1 + 表名 + 结果3
	return 创建SQLite表格结构语句
} //func 创建数据库表格结构(表名 string)string{
//	func 创建MYSQL表格结构函数(表名 string) string {
//		结果1 := `
//			CREATE TABLE IF NOT EXISTS `
//		结果2 := ` (
//				id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY
//				变量名 VARCHAR(255) NOT NULL,
//				变量值 VARCHAR(255) NOT NULL,
//				时间 DATETIME NOT NULL
//			)
//			`
//		return 结果1 + 表名 + 结果2
//	} //func 创建数据库表格结构(表名 string)string{
var 等待插入SQLiteDB的记录们 sync.Map

func 周期生成SQLite记录() {
	ticker := time.NewTicker(time.Millisecond * 1000)
	defer ticker.Stop()
	for range ticker.C {
		// 创建一个待插入的记录
		当前时刻毫秒 := time.Now().UnixMilli()
		当前时刻秒 := time.Now().Unix()
		当前时间字符串 := time.Now().Format("2006-01-02 15:04:05")
		for 项目代号, 表行们 := range 项目代号的表行们 {
			项目最小采集频率毫秒 := 各项目最小采集频率毫秒[项目代号]
			for _, 表行 := range 表行们 {
				z := &项目变量信息表组.Rows[表行]
				需要生成 := false
				if strings.Contains(z.B变量类型, "内存") {
					需要生成 = true
				} else {
					if 当前时刻毫秒-z.S上次生成记录时刻毫秒 > 项目最小采集频率毫秒 {
						需要生成 = true
					}
				}
				if 需要生成 {
					指针 := 生成SQLite记录(z)
					if 指针 != nil {
						等待插入SQLiteDB的记录们.Store(z.B变量名称+"_"+当前时间字符串, 指针)
						z.S上次生成记录时刻秒 = 当前时刻秒
						z.S上次生成记录时刻毫秒 = 当前时刻毫秒
						z.S上次生成记录的变量值 = 指针.变量值
					}
				}
			}
		}
	}
} //func 周期生成SQLite记录() {
func 生成SQLite记录(z *Row) *SQLite表结构 {
	if z.D多少变化率才记录到数据库 == 0 && z.T记录到数据库间隔_秒 == 0 {
		return nil
	}
	if z.T记录到数据库间隔_秒 > 0 && time.Now().Unix()-z.S上次生成记录时刻秒 > z.T记录到数据库间隔_秒 {
		z.D当前值锁.Lock()
		当前值 := z.D当前值
		z.D当前值锁.Unlock()
		return &SQLite表结构{变量名: z.B变量名称, 变量值: 当前值, 时间: time.Now()}
	}
	if z.D多少变化率才记录到数据库 == 0 {
		return nil
	}
	z.D当前值锁.Lock()
	当前值 := z.D当前值
	z.D当前值锁.Unlock()
	if 当前值 == z.S上次生成记录的变量值 {
		return nil
	}
	if 当前值 == z.T通讯异常值 {
		return &SQLite表结构{变量名: z.B变量名称, 变量值: 当前值, 时间: time.Now()}
	}
	if z.S上次生成记录的变量值 == z.T通讯异常值 {
		return &SQLite表结构{变量名: z.B变量名称, 变量值: 当前值, 时间: time.Now()}
	}
	if !strings.Contains(z.B变量类型, "实型") && !strings.Contains(z.B变量类型, "整型") {
		return &SQLite表结构{变量名: z.B变量名称, 变量值: 当前值, 时间: time.Now()}
	}
	上次生成记录的变量值, err := strconv.ParseFloat(z.S上次生成记录的变量值, 64)
	if err != nil {
		fmt.Println("转换失败:", err)
		return nil
	}
	当前值2, err := strconv.ParseFloat(当前值, 64)
	if err != nil {
		fmt.Println("转换失败:", err)
		return nil
	}
	min := 上次生成记录的变量值 - z.D多少变化率才记录到数据库
	max := 上次生成记录的变量值 + z.D多少变化率才记录到数据库
	if 当前值2 > min && 当前值2 < max {
		return nil
	}
	return &SQLite表结构{变量名: z.B变量名称, 变量值: 当前值, 时间: time.Now()}
} //func 生成SQLite记录()*[]SQLite表结构{

// SQLite表结构 表示一条数据记录
type SQLite表结构 struct {
	变量名 string
	变量值 string
	时间  time.Time
}

func SQLite() {
	SQLite周期毫秒 := 最大采集频率毫秒 * 2
	var 生成SQLite记录间隔秒数_毫秒 int64 = 生成SQLite记录间隔秒数 * 1000
	if SQLite周期毫秒 < 生成SQLite记录间隔秒数_毫秒 {
		SQLite周期毫秒 = 生成SQLite记录间隔秒数_毫秒
	}
	if SQLite周期毫秒 > 最大生成SQLite记录间隔秒数*1000 {
		SQLite周期毫秒 = 最大生成SQLite记录间隔秒数 * 1000
	}
	//var SQLite周期毫秒 int64 = 10000
	fmt.Printf("最大采集频率毫秒：%d\r\n", 最大采集频率毫秒)
	fmt.Printf("SQLite周期毫秒：%d\r\n", SQLite周期毫秒)
	ticker := time.NewTicker(time.Millisecond * time.Duration(SQLite周期毫秒))
	defer ticker.Stop()
	for range ticker.C {
		ok, SQLitedbPath := SQLite2()
		if !ok {
			continue
		}
		if !SQLiteDB启动后没有备份过 {
			SQLiteDB启动后没有备份过 = true
			备份目录 := SQLiteDB库文件备份目录信息.Load()
			db文件名 := filepath.Base(SQLitedbPath)
			err := 复制文件(SQLitedbPath, 备份目录+db文件名+SQLiteDB启动后备份后缀名)
			if err != nil {
				str := err.Error()
				SQLite错误信息.Set(str)
				fmt.Println(str)
				SQLiteDB启动后没有备份过 = false
			}
		}
		当前时刻秒 := time.Now().Unix()
		if 当前时刻秒-SQLiteDB上次备份时刻 > SQLiteDB运行时备份间隔秒数 {
			SQLiteDB上次备份时刻 = 当前时刻秒
			备份目录 := SQLiteDB库文件备份目录信息.Load()
			db文件名 := filepath.Base(SQLitedbPath)
			err := 复制文件(SQLitedbPath, 备份目录+db文件名+SQLiteDB运行时间隔备份后缀名)
			if err != nil {
				str := err.Error()
				SQLite错误信息.Set(str)
				fmt.Println(str)
			}
		}
	}
} //func SQLite() {
var SQLiteDB访问锁 sync.Mutex

func SQLite2() (bool, string) {
	var 记录数 uint64 = 0
	等待插入SQLiteDB的记录们.Range(func(key, value interface{}) bool {
		记录数++
		return true // 继续遍历
	})
	if 记录数 == 0 {
		//fmt.Printf("记录数：%d\r\n", 记录数)
		return false, ""
	}
	检查修正SQLitedbPath()
	// 打开（或创建）数据库文件
	SQLitedbPath := SQLiteDB文件名及路径信息.Load()
	// 初始化错误 := 初始化SQLiteDB(SQLitedbPath)
	// if 初始化错误 != "" {
	// 	SQLite错误信息.Set(初始化错误)
	// 	fmt.Println(初始化错误)
	// 	return false,""
	// }
	SQLiteDB访问锁.Lock()
	defer SQLiteDB访问锁.Unlock()
	SQLiteDB连接, err := sql.Open("sqlite3", SQLitedbPath)
	if err != nil {
		str := fmt.Sprintf("无法打开（或创建）数据库文件 %s: %v", SQLitedbPath, err)
		SQLite错误信息.Set(str)
		fmt.Println(str)
		return false, ""
	}
	defer SQLiteDB连接.Close()
	// 创建一个示例表（如果不存在）
	_, err = SQLiteDB连接.Exec(创建SQLite表格结构函数(SQLite表格名, false))
	if err != nil {
		str := fmt.Sprintf("创建表失败: %v", err)
		SQLite错误信息.Set(str)
		fmt.Println(str)
		return false, ""
	}
	// 创建复合索引（如果不存在）
	创建复合索引 := 创建复合索引函数(SQLite表格名, "变量名", "时间")
	_, err = SQLiteDB连接.Exec(创建复合索引)
	if err != nil {
		str := fmt.Sprintf("创建索引失败: %v", err)
		SQLite错误信息.Set(str)
		fmt.Println(str)
		return false, ""
	}

	startTime := time.Now()
	批量插入预编译的语句, err := SQLiteDB连接.Prepare(`
	INSERT INTO SQLite (变量名, 变量值, 时间)
	VALUES (?, ?, ?)
	`)
	if err != nil {
		str := fmt.Sprintf("准备插入语句失败: %v", err)
		SQLite错误信息.Set(str)
		fmt.Println(str)
		return false, ""
	}
	defer 批量插入预编译的语句.Close()

	// 批量插入记录
	tx, err := SQLiteDB连接.Begin()
	if err != nil {
		str := fmt.Sprintf("开始事务失败: %v", err)
		SQLite错误信息.Set(str)
		fmt.Println(str)
		return false, ""
	}
	已插入的记录头 := make([]string, 0)
	记录数 = 0
	等待插入SQLiteDB的记录们.Range(func(key, value interface{}) bool {
		//fmt.Printf("Key: %v, Value: %+v\n", key, value.(*SQLite表结构))
		record := value.(*SQLite表结构)
		_, err := 批量插入预编译的语句.Exec(record.变量名, record.变量值, record.时间)
		if err != nil {
			tx.Rollback()
			str := fmt.Sprintf("插入记录失败: %v", err)
			SQLite错误信息.Set(str)
			fmt.Println(str)
			return false
		}
		已插入的记录头 = append(已插入的记录头, key.(string))
		记录数++
		return true // 返回 true 继续遍历，返回 false 提前终止
	})
	if err := tx.Commit(); err != nil {
		str := fmt.Sprintf("提交事务失败: %v", err)
		SQLite错误信息.Set(str)
		fmt.Println(str)
		return false, ""
	}
	elapsedTime := time.Since(startTime)
	for _, 记录头 := range 已插入的记录头 {
		等待插入SQLiteDB的记录们.Delete(记录头)
	}
	atomic.AddUint64(&SQLiteDB第几次批量插入, 1)
	atomic.AddUint64(&SQLiteDB插入记录总数, 记录数)
	SQLiteDB批量插入结果.锁.Lock()
	最多记录数 := SQLiteDB批量插入结果.SQLiteDB批量插入最小耗时记录数
	原最小耗时 := SQLiteDB批量插入结果.SQLiteDB批量插入最小耗时
	SQLiteDB批量插入结果.锁.Unlock()
	需要更新 := false
	记录数增加了 := false
	if 记录数 > 最多记录数 { //软件每次重启批量插入最多，所以没有参考意义
		需要更新 = true
		记录数增加了 = true
	} else {
		if 记录数 == 最多记录数 {
			if elapsedTime < 原最小耗时 {
				需要更新 = true
			}
		}
	}
	if atomic.LoadUint64(&SQLiteDB第几次批量插入) == 2 { // //软件每次重启批量插入最多，所以没有参考意义
		需要更新 = true
		记录数增加了 = true
	}
	if 需要更新 {
		SQLiteDB批量插入结果.锁.Lock()
		if SQLiteDB批量插入结果.SQLiteDB批量插入最小耗时 > elapsedTime || 记录数增加了 {
			SQLiteDB批量插入结果.SQLiteDB批量插入最小耗时 = elapsedTime
			SQLiteDB批量插入结果.平均记录次最小耗时 = formatNanoseconds(int64(uint64(elapsedTime) / 记录数))
			SQLiteDB批量插入结果.最小耗时发生时间 = time.Now().Format("2006-01-02 15:04:05")
			SQLiteDB批量插入结果.SQLiteDB批量插入最小耗时记录数 = 记录数
		}
		SQLiteDB批量插入结果.锁.Unlock()
	}
	最大耗时 := fmt.Sprintf("成功插入数据库("+SQLitedbPath+")记录条目%d 耗时：%v", 记录数, elapsedTime)
	fmt.Println(最大耗时)
	startTime = time.Now()
	// 查询记录数
	查询最大id语句 := "SELECT MAX(rowid) FROM " + SQLite表格名
	row := SQLiteDB连接.QueryRow(查询最大id语句)
	var count uint64
	if err := row.Scan(&count); err != nil {
		str := err.Error()
		SQLite错误信息.Set(str)
		fmt.Println(str)
	} else {
		atomic.StoreUint64(&SQLiteDB表格记录总数, count)
		elapsedTime := time.Since(startTime)
		最大耗时 := fmt.Sprintf("库表格记录总数查询耗时：%v", elapsedTime)
		fmt.Println(最大耗时)
		SQLiteDB表格记录总数查询耗时 = 最大耗时
	}
	return true, SQLitedbPath
} //func SQLite() {
var SQLiteDB批量插入结果 = SQLiteDB批量插入结果结构{SQLiteDB批量插入最小耗时: time.Duration(math.MaxInt64)}
var SQLiteDB第几次批量插入 uint64
var SQLiteDB插入记录总数 uint64
var SQLiteDB表格记录总数 uint64
var SQLiteDB表格记录总数查询耗时 string = "库表格记录总数查询耗时："
var SQLiteDB启动后没有备份过 bool = false
var SQLiteDB上次备份时刻 int64

const SQLiteDB启动后备份后缀名 = "_启动备份"
const SQLiteDB运行时间隔备份后缀名 = "_运行时间隔备份"

type SQLiteDB批量插入结果结构 struct {
	锁                   sync.Mutex
	SQLiteDB批量插入最小耗时记录数 uint64
	SQLiteDB批量插入最小耗时    time.Duration
	平均记录次最小耗时           string
	最小耗时发生时间            string
}

func formatNanoseconds(ns int64) string {
	// 定义各个时间单位的常量
	const (
		nanosecondsPerSecond      = int64(1e9)
		nanosecondsPerMillisecond = int64(1e6)
		nanosecondsPerMicrosecond = int64(1000)
	)
	// 计算秒、毫秒、微秒和剩余的纳秒
	seconds := ns / nanosecondsPerSecond
	ns %= nanosecondsPerSecond
	milliseconds := ns / nanosecondsPerMillisecond
	ns %= nanosecondsPerMillisecond
	microseconds := ns / nanosecondsPerMicrosecond
	ns %= nanosecondsPerMicrosecond
	// 构建结果字符串
	var result string
	if seconds > 0 {
		result += strconv.FormatInt(seconds, 10) + "s "
	}
	if milliseconds > 0 {
		result += strconv.FormatInt(milliseconds, 10) + "ms "
	}
	if microseconds > 0 {
		result += strconv.FormatInt(microseconds, 10) + "μs "
	}
	if ns > 0 {
		result += strconv.FormatInt(ns, 10) + "ns"
	}
	return result
} //func formatNanoseconds(ns int64) string {

func 获取文件大小(filePath string) string {
	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "-9.99TiB"
	}

	// 获取文件大小（以字节为单位）
	fileSize := fileInfo.Size()
	return humanReadableSize(uint64(fileSize))
}

const SQLiteDB文件名及路径信息文件名 = "SQLiteDB文件名及路径信息"
const SQLiteDB库文件备份目录信息文件名 = "SQLiteDB库文件备份目录信息"

// 保存全局字符串到文件
func 将SQLiteDB库文件备份目录信息写入文件() {
	str := SQLiteDB库文件备份目录信息.Load()
	if str == "" {
		return
	}
	fmt.Println(目录 + SQLiteDB库文件备份目录信息文件名)
	f, err := os.Create(目录 + SQLiteDB库文件备份目录信息文件名)
	if err != nil {
		启动软件碰到的问题.Set("创建文件(" + SQLiteDB库文件备份目录信息文件名 + ")发生了错误(" + err.Error() + ")")
		fmt.Println(err.Error())
		return
	} else {
		f.Write([]byte(str))
	}
	defer f.Close()
}
func 将SQLiteDB文件名及路径信息写入文件() {
	str := SQLiteDB文件名及路径信息.Load()
	if str == "" {
		return
	}
	fmt.Println(目录 + SQLiteDB文件名及路径信息文件名)
	f, err := os.Create(目录 + SQLiteDB文件名及路径信息文件名)
	if err != nil {
		启动软件碰到的问题.Set("创建文件(" + SQLiteDB文件名及路径信息文件名 + ")发生了错误(" + err.Error() + ")")
		fmt.Println(err.Error())
		return
	} else {
		f.Write([]byte(str))
	}
	defer f.Close()
}

// 从文件中读取全局字符串
func 从文件中读入SQLiteDB库文件备份目录信息() {
	// 从文件中读取数据
	data, err := os.ReadFile(目录 + SQLiteDB库文件备份目录信息文件名)
	if err != nil {
		fmt.Println(err)
		return
	}
	str := string(data)
	if str == "" {
		return
	}
	str = strings.Replace(str, "\r", "", -1)
	str = strings.Replace(str, "\n", "", -1)

	SQLiteDB库文件备份目录信息.Set(str)
}
func 从文件中读入SQLiteDB文件名及路径信息() {
	// 从文件中读取数据
	data, err := os.ReadFile(目录 + SQLiteDB文件名及路径信息文件名)
	if err != nil {
		fmt.Println(err)
		return
	}
	str := string(data)
	if str == "" {
		return
	}
	str = strings.Replace(str, "\r", "", -1)
	str = strings.Replace(str, "\n", "", -1)

	SQLiteDB文件名及路径信息.Set(str)
}

var SQLiteDB文件名及路径信息 字符串互斥锁访问结构体
var SQLiteDB库文件备份目录信息 字符串互斥锁访问结构体

func 获取文件信息(文件名及路径 string) (错误信息 string, 自定义路径 string) {
	错误信息 = ""
	自定义路径 = ""
	if 文件名及路径 == "" {
		错误信息 = "文件名及路径信息为空"
		return 错误信息, 自定义路径
	}
	fileInfo, err := os.Stat(文件名及路径)
	if err != nil {
		错误信息 = fmt.Sprintf("指定的文件不存在>>%s", 文件名及路径)
		return 错误信息, 自定义路径
	}
	// 判断是否是目录
	if fileInfo.IsDir() {
		错误信息 = fmt.Sprintf("这是一个目录>>%s", 文件名及路径)
		自定义路径 = 文件名及路径
		return 错误信息, 自定义路径
	}
	db, err := sql.Open("sqlite3", 文件名及路径)
	if err != nil {
		错误信息 = fmt.Sprintf("无法打开（或创建）数据库文件 %s: %v", 文件名及路径, err)
		自定义路径 = filepath.Dir(文件名及路径)
		return 错误信息, 自定义路径
	}
	defer db.Close()

	if correct := 判断SQLiteDB表结构是否正确(db); !correct {
		错误信息 = fmt.Sprintf("SQLiteDB文件格式错误>>%s，创建SQLite表格结构SQL语句是>>%s", 文件名及路径, 创建SQLite表格结构函数(SQLite表格名, false))
		自定义路径 = filepath.Dir(文件名及路径)
	}
	return 错误信息, 自定义路径
} //func 获取文件信息(文件名及路径 string) (错误信息 string, 自定义路径 string) {
func 检查修正SQLitedbPath备份() {
	SQLitedbPath := SQLiteDB库文件备份目录信息.Load()
	自定义路径 := ""
	错误信息 := ""
	if SQLitedbPath == "" {
		错误信息 = "从" + SQLiteDB库文件备份目录信息文件名 + "中读入SQLiteDB库文件备份目录信息为空将使用默认目录：" + 目录
		自定义路径 = 目录
	} else { //if SQLitedbPath == "" {
		//尝试创建目录
		err := os.MkdirAll(SQLitedbPath, os.ModePerm)
		if err == nil {
			// 获取绝对路径
			absPath, err := filepath.Abs(SQLitedbPath)
			if err == nil {
				自定义路径 = absPath
				if 自定义路径[len(自定义路径)-1] != '\\' {
					自定义路径 += "\\"
				}
			} else {
				错误信息 = "absPath, err := filepath.Abs(SQLitedbPath)发生错误：" + err.Error() + "将使用默认目录：" + 目录
				自定义路径 = 目录
			}
		} else {
			错误信息 = "err := os.MkdirAll(SQLitedbPath, os.ModePerm)发生错误：" + err.Error() + "将使用默认目录：" + 目录
			自定义路径 = 目录
		}
	} //} else {//if SQLitedbPath == "" {
	if 错误信息 != "" {
		SQLite错误信息.Set(错误信息)
		fmt.Println(错误信息)
	}
	SQLiteDB库文件备份目录信息.Set(自定义路径)
	将SQLiteDB库文件备份目录信息写入文件()
} //func 检查修正SQLitedbPath(){
func 检查修正SQLitedbPath() {
	检查修正SQLitedbPath备份()
	备份目录 := SQLiteDB库文件备份目录信息.Load()
	SQLitedbPath := SQLiteDB文件名及路径信息.Load()
	错误信息, 自定义路径 := 获取文件信息(SQLitedbPath)
	使用备份1 := false
	使用备份2 := false
	db文件名 := filepath.Base(SQLitedbPath)
	if SQLitedbPath == "" {
		错误信息 = "从" + SQLiteDB文件名及路径信息文件名 + "中读入SQLiteDB文件名及路径信息为空"
	} else { //if SQLitedbPath == "" {
		if 错误信息 != "" {
			错误信息2, _ := 获取文件信息(备份目录 + db文件名 + SQLiteDB运行时间隔备份后缀名)
			if 错误信息2 == "" {
				使用备份1 = true
			} else {
				错误信息3, _ := 获取文件信息(备份目录 + db文件名 + SQLiteDB启动后备份后缀名)
				if 错误信息3 == "" {
					使用备份2 = true
				}
			}
		}
		// 如果当前路径不是目录也不是文件
		if 错误信息 != "" && 自定义路径 == "" {
			//尝试创建目录
			err := os.MkdirAll(SQLitedbPath, os.ModePerm)
			if err == nil {
				// 获取绝对路径
				absPath, err := filepath.Abs(SQLitedbPath)
				if err == nil {
					自定义路径 = absPath
				}
			}
		}
	} //} else {//if SQLitedbPath == "" {
	if 错误信息 != "" {
		SQLite错误信息.Set(错误信息)
		fmt.Println(错误信息)
		SQLitedbPath2 := ""
		if 自定义路径 == "" {
			SQLitedbPath2 = 目录 + SQLiteDb + "_" + time.Now().Format("20060102150405")
		} else {
			if 自定义路径[len(自定义路径)-1] != '\\' {
				自定义路径 += "\\"
			}
			SQLitedbPath2 = 自定义路径 + SQLiteDb + "_" + time.Now().Format("20060102150405")
		}
		错误信息 = "将创建:" + SQLitedbPath2
		fmt.Println(错误信息)
		SQLite错误信息.Set(错误信息)
		SQLiteDB文件名及路径信息.Set(SQLitedbPath2)
		//db文件名 = filepath.Base(SQLitedbPath2)
		将SQLiteDB文件名及路径信息写入文件()
		if 使用备份1 {
			err := 复制文件(备份目录+db文件名+SQLiteDB运行时间隔备份后缀名, SQLitedbPath2)
			if err != nil {
				错误信息 = "复制文件发生错误，需人工介入:" + err.Error()
				fmt.Println(错误信息)
				SQLite错误信息.Set(错误信息)
			} else {
				错误信息 = "成功复制>>" + 备份目录 + db文件名 + SQLiteDB运行时间隔备份后缀名 + ">>" + SQLitedbPath2
				fmt.Println(错误信息)
				SQLite错误信息.Set(错误信息)
			}
		}
		if 使用备份2 {
			err := 复制文件(备份目录+db文件名+SQLiteDB启动后备份后缀名, SQLitedbPath2)
			if err != nil {
				错误信息 = "复制文件发生错误，需人工介入:" + err.Error()
				fmt.Println(错误信息)
				SQLite错误信息.Set(错误信息)
			} else {
				错误信息 = "成功复制>>" + 备份目录 + db文件名 + SQLiteDB启动后备份后缀名 + ">>" + SQLitedbPath2
				fmt.Println(错误信息)
				SQLite错误信息.Set(错误信息)
			}
		}
	} //if 错误信息 != "" {
} //func 检查修正SQLitedbPath(){

// ExpectedTableSchema 定义预期的表结构
type ExpectedTableSchema struct {
	ColumnName string
	DataType   string
	NotNull    bool
	PrimaryKey bool
	Autoincr   bool
}

// isSQLiteDBFormatCorrect 判断SQLite表结构是否为预期的结构
func isSQLiteDBFormatCorrect(db *sql.DB, tableName string, 预期SQLite表格结构 []ExpectedTableSchema) (bool, error) {
	// 获取实际的表结构
	rows, err := db.Query("PRAGMA table_info(" + tableName + ");")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var actualSchema []ExpectedTableSchema
	for rows.Next() {
		var cid, name, dtype, notnull, pk, autoincr *string
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &pk, &autoincr); err != nil {
			return false, err
		}
		name2 := ""
		dtype2 := ""
		notnull2 := ""
		pk2 := ""
		autoincr2 := ""
		if name != nil {
			name2 = *name
		}
		if dtype != nil {
			dtype2 = *dtype
		}
		if notnull != nil {
			notnull2 = *notnull
		}
		if pk != nil {
			pk2 = *pk
		}
		if name2 == 自增主键名 {
			pk2 = "1"
		}
		if autoincr != nil {
			autoincr2 = *autoincr
		}
		actualSchema = append(actualSchema, ExpectedTableSchema{
			ColumnName: name2,
			DataType:   dtype2,
			NotNull:    notnull2 == "1",
			PrimaryKey: pk2 == "1",
			Autoincr:   autoincr2 == "1",
		})
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	// 比较实际表结构与预期表结构
	if len(actualSchema) != len(预期SQLite表格结构) {
		return false, errors.New("len(actualSchema) != len(预期SQLite表格结构)")
	}
	for i := range actualSchema {
		if actualSchema[i] != 预期SQLite表格结构[i] {
			错误内容 := ""
			错误内容 += actualSchema[i].ColumnName + "<>" + 预期SQLite表格结构[i].ColumnName + "\n"
			错误内容 += actualSchema[i].DataType + "<>" + 预期SQLite表格结构[i].DataType + "\n"
			错误内容 += strconv.FormatBool(actualSchema[i].NotNull) + "<>" + strconv.FormatBool(预期SQLite表格结构[i].NotNull) + "\n"
			错误内容 += strconv.FormatBool(actualSchema[i].PrimaryKey) + "<>" + strconv.FormatBool(预期SQLite表格结构[i].PrimaryKey) + "\n"
			错误内容 += strconv.FormatBool(actualSchema[i].Autoincr) + "<>" + strconv.FormatBool(预期SQLite表格结构[i].Autoincr) + "\n"
			return false, errors.New(错误内容)
		}
	}
	return true, nil
}

func 判断SQLiteDB表结构是否正确(db *sql.DB) bool {
	// 检查表结构
	isCorrect, err := isSQLiteDBFormatCorrect(db, SQLite表格名, 预期SQLite表格结构)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return isCorrect
}

var SQLiteDB内存数据库连接 *sql.DB
var SQLiteDB磁盘查询历史数据库连接 *sql.DB

const SQLiteDB磁盘查询历史数据库文件名 = "SQLiteDB磁盘查询历史数据库"

func 创建SQLiteDB内存数据库() {
	// 连接到内存数据库
	db, err := sql.Open("sqlite3", ":memory:?_journal_mode=WAL&_sync=NORMAL")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	//defer db.Close()
	创建表结构 := 创建SQLite表格结构函数(SQLite表格名, false)
	_, err = db.Exec(创建表结构)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return
	}
	创建复合索引 := 创建复合索引函数(SQLite表格名, "变量名", "时间")
	_, err = db.Exec(创建复合索引)
	if err != nil {
		fmt.Println("创建复合索引失败:", err)
		return
	}
	SQLiteDB内存数据库连接 = db
	db1, err1 := sql.Open("sqlite3", 目录+SQLiteDB磁盘查询历史数据库文件名+"?_journal_mode=WAL&_sync=NORMAL")
	if err1 != nil {
		fmt.Println("Error opening database:", err1)
		return
	}
	_, err = db1.Exec(创建表结构)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return
	}
	_, err = db1.Exec(创建复合索引)
	if err != nil {
		fmt.Println("创建复合索引失败:", err)
		return
	}
	var 记录总数 int
	row := db1.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", SQLite表格名))
	if err := row.Scan(&记录总数); err != nil {
		fmt.Println("row.Scan(&SQLiteDB磁盘查询历史数据库记录总数)" + err.Error())
		return
	}
	SQLiteDB磁盘查询历史数据库连接 = db1
	if 记录总数 == 0 {
		fmt.Println("SQLiteDB磁盘查询历史数据库记录总数 == 0")
		return
	}
	// 获取源数据库中的所有列名
	columns, err := getColumnNames(SQLiteDB磁盘查询历史数据库连接, SQLite表格名)
	if err != nil {
		fmt.Println("getColumnNames(SQLiteDB磁盘查询历史数据库连接, SQLite表格名)" + err.Error())
		return
	}

	// 构建 SELECT 查询语句
	selectQuery := fmt.Sprintf("SELECT %s FROM %s;", strings.Join(columns, ", "), SQLite表格名)

	// 执行查询并复制数据到目标数据库
	rows, err := SQLiteDB磁盘查询历史数据库连接.Query(selectQuery)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer rows.Close()

	// 构建 INSERT 语句的占位符
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "?"
	}
	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", SQLite表格名, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	// 遍历查询结果并插入到目标数据库
	for rows.Next() {
		// 为每一列创建一个指向值的指针
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// 扫描行数据到值指针中
		if err := rows.Scan(valuePtrs...); err != nil {
			fmt.Println(err.Error())
			return
		}

		// 执行 INSERT 语句将数据插入到目标数据库
		if _, err := SQLiteDB内存数据库连接.Exec(insertQuery, values...); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
	// 检查是否有任何错误发生
	if err := rows.Err(); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("SQLiteDB磁盘查询历史数据库数据复制到内存数据库完成，记录数%d\r\n", 记录总数)
} //func 创建SQLiteDB内存数据库() {

// 获取表的列名
func getColumnNames(db *sql.DB, tableName string) ([]string, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s);", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var cid int
		var name string
		var typ string
		var notnull int
		var dfltValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			fmt.Println("rows.Scan error:" + err.Error())
			return nil, err
		}
		columns = append(columns, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

// 执行批量插入操作
func execBatch(stmt *sql.Stmt, data []interface{}) error {
	// 将数据转换为二维切片，因为stmt.Exec需要每个参数作为单独的切片元素
	params := make([][]interface{}, 0, len(data)/3)
	for i := 0; i < len(data); i += 3 {
		params = append(params, data[i:i+3])
	}
	// 执行批量插入，使用...操作符展开params中的每个子切片
	for _, param := range params {
		_, err := stmt.Exec(param...)
		if err != nil {
			fmt.Printf("批量插入数据失败: %v\r\n", err)
			return err
		}
	}
	return nil
}

var 当地时区信息 *time.Location

func init() {
	// 加载指定的时区信息，例如 "Asia/Shanghai" 表示东八区（北京时间）
	loc, err := time.LoadLocation("Local")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return
	}
	当地时区信息 = loc
	//.In(loc).Format("2006-01-02 15:04:05")
}
func 获取文件绝对路径及文件名(文件名 string) string {
	// 获取文件的绝对路径
	absPath, err := filepath.Abs(文件名)
	if err != nil {
		fmt.Printf("无法获取绝对路径: %v\r\n", err)
		return ""
	}
	// // 获取文件名（包括扩展名）
	// fileName := filepath.Base(absPath)
	// fmt.Println("文件名:", fileName)
	// // 如果你只需要文件名而不包括扩展名，可以使用 filepath.Ext 来移除扩展名
	// ext := filepath.Ext(fileName)
	// fileNameWithoutExt := fileName[:len(fileName)-len(ext)]
	// fmt.Println("文件名（不包括扩展名）:", fileNameWithoutExt)
	return absPath
}

// /////////////////文件服务器代码开始

var (
	username = "admin"
	password = "password"
)

type FileInfo struct {
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
	Type    string
}

type FileListPage struct {
	Title    string
	Files    []FileInfo
	SortBy   string
	SortDesc bool
}

var 根目录被访问次数 uint64

// 登录处理
func loginHandler(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &根目录被访问次数, "G根目录")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	if r.Method == http.MethodPost {
		user := r.FormValue("username")
		pass := r.FormValue("password")
		if user == username && pass == password {
			http.SetCookie(w, &http.Cookie{Name: 登录验证Cookie名称, HttpOnly: true, Value: "true"})
			http.Redirect(w, r, "/files", http.StatusFound)
			return
		}
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 嵌入的登录页面模板
	tmpl := template.Must(template.New("login").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Login</title>
</head>
<body>
    <h1>Login</h1>
    <form method="POST">
        <label>Username:</label>
        <input type="text" name="username" required><br>
        <label>Password:</label>
        <input type="password" name="password" required><br>
        <button type="submit">Login</button>
    </form>
</body>
</html>
	`))
	tmpl.Execute(w, nil)
}

var 文件浏览被访问次数 uint64

// 文件浏览处理
func fileHandler(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &文件浏览被访问次数, "W文件浏览")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	if !开启文件服务器 {
		http.Error(w, "嗨,"+r.RemoteAddr+"您好!文件服务器没有开启", http.StatusInternalServerError)
		return
	}
	dir := 目录2 // 文件目录
	sortBy := r.URL.Query().Get("sort")
	sortDesc := r.URL.Query().Get("desc") == "true"

	files, err := getFileList(dir)
	if err != nil {
		http.Error(w, "Unable to read directory", http.StatusInternalServerError)
		return
	}

	// 排序
	switch sortBy {
	case "name":
		sort.Slice(files, func(i, j int) bool {
			if sortDesc {
				return files[i].Name > files[j].Name
			}
			return files[i].Name < files[j].Name
		})
	case "size":
		sort.Slice(files, func(i, j int) bool {
			if sortDesc {
				return files[i].Size > files[j].Size
			}
			return files[i].Size < files[j].Size
		})
	case "date":
		sort.Slice(files, func(i, j int) bool {
			if sortDesc {
				return files[i].ModTime.After(files[j].ModTime)
			}
			return files[i].ModTime.Before(files[j].ModTime)
		})
	case "type":
		sort.Slice(files, func(i, j int) bool {
			if sortDesc {
				return files[i].Type > files[j].Type
			}
			return files[i].Type < files[j].Type
		})
	}

	page := FileListPage{
		Title:    目录2,
		Files:    files,
		SortBy:   sortBy,
		SortDesc: sortDesc,
	}

	// 嵌入的文件浏览页面模板
	tmpl := template.Must(template.New("files").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    <h1>{{.Title}}</h1>
    <table border="1">
        <tr>
            <th><a href="/files?sort=name&desc={{not .SortDesc}}">Name</a></th>
            <th><a href="/files?sort=size&desc={{not .SortDesc}}">Size</a></th>
            <th><a href="/files?sort=date&desc={{not .SortDesc}}">Date</a></th>
            <th><a href="/files?sort=type&desc={{not .SortDesc}}">Type</a></th>
            <th>Action</th>
        </tr>
        {{range .Files}}
        <tr>
            <td>{{.Name}}</td>
            <td>{{if .IsDir}}-{{else}}{{.Size}}{{end}}</td>
            <td>{{.ModTime.Format "2006-01-02 15:04:05"}}</td>
            <td>{{.Type}}</td>
            <td>
                {{if not .IsDir}}
                <a href="/download?file=` + 目录2 + `{{.Name}}">Download</a>
                {{end}}
            </td>
        </tr>
        {{end}}
    </table>
</body>
</html>
	`))
	tmpl.Execute(w, page)
}

var 下载文件被访问次数 uint64

// 文件下载处理
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	go 连接信息处理(r.RemoteAddr, &下载文件被访问次数, "X下载文件")
	go 获取服务器域名(r.Host, r.RemoteAddr)
	if !开启文件服务器 {
		http.Error(w, "嗨,"+r.RemoteAddr+"您好!文件服务器没有开启", http.StatusInternalServerError)
		return
	}
	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		http.Error(w, "File not specified", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// 设置响应头，确保文件名与原始文件一致
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(filePath)))
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to download file", http.StatusInternalServerError)
		return
	}
}

// 获取文件列表
func getFileList(dir string) ([]FileInfo, error) {
	var files []FileInfo
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fileType := "Folder"
		if !entry.IsDir() {
			fileType = strings.ToUpper(strings.TrimPrefix(filepath.Ext(entry.Name()), "."))
			if fileType == "" {
				fileType = "File"
			}
		}
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
			Type:    fileType,
		})
	}
	return files, nil
}

var 登录验证Cookie名称 = "auth"

// 登录验证中间件
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(登录验证Cookie名称)
		if err != nil || cookie.Value != "true" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		next(w, r)
	}
}

///////////////////文件服务器代码结束
