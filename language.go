package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:generate stringer -type=Lang
type Lang int

func (i *Lang) OfString(v string) error {
	x := strings.Index(_Lang_name, v)
	if x < 0 {
		return fmt.Errorf("invalid language name: %s", v)
	}
	xi := -1
	for i2, index := range _Lang_index {
		if index == uint16(x) {
			xi = i2
			break
		}
	}
	if xi < 0 {
		return fmt.Errorf("invalid language name: %s", v)
	}
	*i = Lang(xi)
	return nil
}

const (
	lblM3u8URL Lang = iota
	capCURL
	capError
	errInvalidM3U8
	errInvalidCURL
	lblSaveFolder
	btnChoose
	capSaveFolder
	lblSaveName
	lblSkipCode
	lblProxy
	lblBatch
	lblTemp
	capTemp
	lblInsecure
	lblKeepTS
	lblUseServerTime
	lblSkipMerge
	lblDebugLog
	lblSkipLog
	msgIdle
	btnMerge
	tipMerge
	btnStopMerge
	btnDownload
	tipDownload
	btnCancelDownload
	errBatchNum
	btnCurl
	btnOk
	btnCancel
	errDownload
	errDownloaded
	infDownloadDone
	infQueueDownloadDone
	errMerge
	errTSFolder
	errSaveFolder
	capInfo
	phtFolder
	phtSkipCode
	phtTemp
	phtProxy
	phtFileName

	btnFindIDMWindow
	txtIDMCancel
	tipIDMConfig
	txtIDMPromptTitle
	txtIDMPromptClass
	regIDMUrl
	regIDMName

	themeColor

	tabMain
	tabQueue
	btnAddQueueTxt
	btnRemoveQueue
	btnQueueTask
	btnQueueModify
	msgQueueDownloading
	numQueueHeight
	numQueueWidth

	stageIdle
	stageFetchM3U8
	stageM3U8Analysis
	stageDownloading
	stageMergeAnalysis
	stageMergeFiles
)

type Language map[Lang]string

func (s Language) UnmarshalJSON(bytes []byte) error {
	var m = make(map[string]string)
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}
	var i Lang
	for k, v := range m {
		if err = (&i).OfString(k); err != nil {
			return err
		}
		s[i] = v
	}
	return nil
}

func (s Language) MarshalJSON() ([]byte, error) {
	var m = make(map[string]string, len(s))
	for k, v := range s {
		m[k.String()] = v
	}
	return json.Marshal(m)
}

func (s Language) LoadLanguage() bool {
	f, err := os.Open("language.json")
	if err != nil {
		return false
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&s)
	if err == nil {
		return false
	}
	return true
}
func (s Language) SaveLanguage() {
	if _, err := os.Stat("language.template.json"); os.IsNotExist(err) {
		f, err := os.OpenFile("language.template.json", os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return
		}
		_ = json.NewEncoder(f).Encode(s)
	}
}
func (s Language) Get(key Lang) string {
	if v, ok := s[key]; ok {
		return v
	}
	return key.String()
}
func (s Language) InitializeDefault() {
	s[lblM3u8URL] = "m3u8地址:"
	s[capCURL] = "curl模式"
	s[capError] = "错误"
	s[errInvalidM3U8] = "无效的M3U8地址:"
	s[errInvalidCURL] = "无效的curl:%s"
	s[lblSaveFolder] = "保存位置:"
	s[btnChoose] = "选择"
	s[capSaveFolder] = "保存位置"
	s[phtFolder] = "下载后MP4保存目录,默认为当前目录"
	s[lblSaveName] = "保存文件名:"
	s[lblSkipCode] = "跳过TS信息:"
	s[lblProxy] = "代理设置:"
	s[lblBatch] = "线程数:"
	s[lblTemp] = "临时文件夹:"
	s[capTemp] = "下载临时保存位置"
	s[lblInsecure] = "允许不安全的https请求"
	s[lblKeepTS] = "不删除下载的ts文件"
	s[lblUseServerTime] = "使用服务器文件时间"
	s[lblSkipMerge] = "不合并TS为MP4"
	s[lblDebugLog] = "调试日志"
	s[lblSkipLog] = "记录跳过日志"
	s[msgIdle] = "空闲"
	s[btnMerge] = "合并TS"
	s[tipMerge] = "合并临时目录中已经下载的ts片段并保存到输出目录"
	s[btnStopMerge] = "停止合并"
	s[btnDownload] = "下载"
	s[tipDownload] = "下载当前URL,保存到指定的目录"
	s[btnCancelDownload] = "取消下载"
	s[errBatchNum] = "线程数只能为正整数:%s"
	s[btnCurl] = "curl模式"
	s[btnOk] = "确定"
	s[btnCancel] = "取消"
	s[errDownload] = "下载出错: %s"
	s[errDownloaded] = "已经下载: %s"
	s[infDownloadDone] = "下载完成: %s"
	s[infQueueDownloadDone] = "队列下载完毕"
	s[errMerge] = "合并错误: %s"
	s[errTSFolder] = "请选择临时目录"
	s[errSaveFolder] = "请选择保存文件名和保存位置"
	s[capInfo] = "提示"
	s[phtSkipCode] = "1,92-100,http.code=403,if-http.code-merge_ts,time:00:05:12-00:07:20"
	s[phtTemp] = "TS文件保存目录,默认为保存位置"
	s[phtProxy] = "http://127.0.0.1:8080 socks5://127.0.0.1:1089"
	s[phtFileName] = "all"

	s[btnFindIDMWindow] = "IDM加载"
	s[txtIDMCancel] = "取消"
	s[tipIDMConfig] = "IDM需要配置:language.json中的'*IDM*'属性"
	s[txtIDMPromptTitle] = "下载文件信息"
	s[txtIDMPromptClass] = "#32770"
	s[regIDMUrl] = "(?i)^(http|https)://"
	s[regIDMName] = "([^\\\\]+)\\.ts$"

	s[themeColor] = "dark"

	s[tabMain] = "任务"
	s[tabQueue] = "队列"
	s[msgQueueDownloading] = "下载中的任务不能删除"
	s[btnAddQueueTxt] = "加入队列"
	s[btnRemoveQueue] = "删除"
	s[btnQueueModify] = "修改"
	s[btnQueueTask] = "详情"
	s[numQueueHeight] = "24"
	s[numQueueWidth] = "136"
	s.SaveLanguage()
}
