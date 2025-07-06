package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/ZenLiuCN/fn"
	"github.com/orestonce/m3u8d"
	"github.com/orestonce/m3u8d/m3u8dcpp"
	. "modernc.org/tk9.0"
	_ "modernc.org/tk9.0/themes/azure"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed favicon.ico
var ico []byte

type conf = func(col int) ([]Opt, int)

var rows = 0
var pad = Opts{Padx(3), Pady(3)}

func row(val ...any) {
	r := Row(rows)
	i := 0
	xi := 0
	var w Widget
	var s []Opt = append([]Opt{}, pad...)
	for _, v := range val {
		switch x := v.(type) {
		case Widget:
			if w != nil {
				Grid(w, append(s, r, Column(i))...)
				//println(n, i)
				i += xi + 1
			}
			w = x
			s = s[:0]
			s = append(s, pad...)
		case Opt:
			s = append(s, x)
		case conf:
			o, d := x(i)
			xi += d
			s = append(s, o...)
		}
	}
	Grid(w, append(s, r, Column(i))...)
	rows++
}

func colSpan(n int, opts ...Opt) conf {
	return func(i int) ([]Opt, int) { return append(opts, Columnspan(n)), n }
}

var req m3u8d.StartDownload_Req
var (
	uri, folder, file, skipScript, proxy, tempFolder                                       *TEntryWidget
	threadVal                                                                              *LabelWidget
	threads, Insecure, SkipRemoveTs, UseServerSideTime, SkipMergeTs, WithSkipLog, DebugLog *VariableOpt
	merge, download                                                                        *TButtonWidget
	ctx                                                                                    context.Context
	ctxCC                                                                                  context.CancelFunc
)

func syncGUI() {
	if req.Insecure {
		Insecure.Set(1)
	}
	if req.SkipRemoveTs {
		SkipRemoveTs.Set(1)
	}
	if req.UseServerSideTime {
		UseServerSideTime.Set(1)
	}
	if req.SkipMergeTs {
		SkipMergeTs.Set(1)
	}
	if req.ThreadCount == 0 {
		req.ThreadCount = runtime.NumCPU()
	}
	if req.WithSkipLog {
		WithSkipLog.Set(1)
	}
	if req.DebugLog {
		DebugLog.Set(1)
	}
	uri.Configure(Textvariable(req.M3u8Url))
	folder.Configure(Textvariable(req.SaveDir))
	skipScript.Configure(Textvariable(req.SkipTsExpr))
	tempFolder.Configure(Textvariable(req.TsTempDir))
	proxy.Configure(Textvariable(req.SetProxy))
	threads.Set(req.ThreadCount)
}
func syncRequest(mute bool) bool {
	req.M3u8Url = uri.Textvariable()
	req.SetProxy = proxy.Textvariable()
	req.SaveDir = folder.Textvariable()
	req.FileName = file.Textvariable()
	req.SkipTsExpr = skipScript.Textvariable()
	req.TsTempDir = tempFolder.Textvariable()
	//req.ProgressBarShow = false
	var err error
	req.ThreadCount, err = strconv.Atoi(truncate(threads.Get()))
	if err != nil || req.ThreadCount < 1 {
		if !mute {
			MessageBox(Default("ok"), Title("错误"), Detail("线程数只能为正整数:"+threads.Get()), Parent(App), Type("ok"))
		}
		return false
	}
	if !mute {
		if req.M3u8Url == "" {
			MessageBox(Default("ok"), Title("错误"), Detail("无效的M3U8地址:"), Parent(App), Type("ok"))
			return false
		}
	}
	saveConfig()
	return true
}
func loadConfig() {
	f, err := os.Open("config.json")
	if err == nil {
		fn.Panic(json.NewDecoder(f).Decode(&req))
		_ = f.Close()
	}
}
func saveConfig() {
	f, _ := os.OpenFile("config.json", os.O_CREATE|os.O_TRUNC, os.ModePerm)
	_ = json.NewEncoder(f).Encode(&req)
	_ = f.Close()
}
func main() {
	loadConfig()
	var render []func()
	renderLock := &sync.RWMutex{}
	uri = TEntry(Textvariable(req.M3u8Url))
	row(
		Label(Txt("m3u8地址:")), Sticky("nw"),
		uri, colSpan(9, Sticky("new")),
		TButton(Txt("curl模式"), Command(func() {
			syncRequest(true)
			if req.M3u8Url == "" {
				MessageBox(Default("ok"), Title("错误"), Detail("无效的M3U8地址:"), Parent(App), Type("ok"))
				return
			}
			var d = CurlData{
				Text: m3u8d.RunDownload_Req_ToCurlStr(req),
			}
			curlDialog(&d).ShowModal()
			if d.Ok {
				v := m3u8d.ParseCurlStr(d.Text)
				if v.ErrMsg != "" {
					MessageBox(Default("ok"), Title("错误"), Detail("无效的Curl:"+v.ErrMsg), Parent(App), Type("ok"))
				} else {
					req = v.DownloadReq
				}
			}
			syncGUI()
		})), Padx(10),
	)
	folder = TEntry(Textvariable(req.SaveDir))
	row(
		Label(Txt("保存位置:")), Sticky("nw"),
		folder, colSpan(9, Sticky("new")),
		TButton(Txt("选择"), Command(func() {
			dir := ChooseDirectory(Initialdir(folder.Variable()), Mustexist(true), Parent(App), Title("保存位置"))
			if dir != "" {
				folder.Configure(Textvariable(dir))
			}
		})),
	)
	file = TEntry(Textvariable(req.FileName))
	row(
		Label(Txt("保存文件名:")), Sticky("nw"),
		file, colSpan(9, Sticky("new")),
	)
	skipScript = TEntry(Textvariable(req.SkipTsExpr))
	row(
		Label(Txt("跳过TS信息:")), Sticky("nw"),
		skipScript, colSpan(9, Sticky("new")),
	)
	proxy = TEntry(Textvariable(req.SetProxy))
	row(
		Label(Txt("代理设置:")), Sticky("nw"),
		proxy, colSpan(9, Sticky("new")),
	)
	n := runtime.NumCPU()
	if req.ThreadCount <= 0 {
		req.ThreadCount = n
	}
	threads = Variable(req.ThreadCount)
	threadVal = Label(Textvariable(strconv.Itoa(req.ThreadCount)))
	row(
		Label(Txt("线程数:")), Sticky("nw"),
		threadVal, Sticky("nw"),
		TScale(From(1), To(4*n), threads, Command(func() {
			threadVal.Configure(Textvariable(truncate(threads.Get())))
		})), colSpan(8, Sticky("new")),
	)
	tempFolder = TEntry(Textvariable(req.TsTempDir))
	row(
		Label(Txt("临时文件夹:")), Sticky("nw"),
		tempFolder, colSpan(9, Sticky("new")),
		TButton(Txt("选择"), Command(func() {
			dir := ChooseDirectory(Initialdir(tempFolder.Variable()), Mustexist(true), Parent(App), Title("下载临时保存位置"))
			if dir != "" {
				tempFolder.Configure(Textvariable(dir))
			}
		})),
	)
	Insecure = Variable(0)
	SkipRemoveTs = Variable(0)
	row(
		TCheckbutton(Txt("允许不安全的https请求"), Insecure, Onvalue(1), Offvalue(-1), Command(func() {
			req.Insecure = Insecure.Get() == "1"
		})), colSpan(2, Sticky("nw")),
		TCheckbutton(Txt("不删除下载的ts文件"), SkipRemoveTs, Onvalue(1), Offvalue(-1), Command(func() {
			req.SkipRemoveTs = SkipRemoveTs.Get() == "1"
		})), colSpan(2, Sticky("nw")),
	)
	UseServerSideTime = Variable(0)
	SkipMergeTs = Variable(0)
	row(
		TCheckbutton(Txt("使用务端提供的文件时间"), UseServerSideTime, Onvalue(1), Offvalue(-1), Command(func() {
			req.UseServerSideTime = UseServerSideTime.Get() == "1"
		})), colSpan(2, Sticky("nw")),
		TCheckbutton(Txt("不合并TS为MP4"), SkipMergeTs, Onvalue(1), Offvalue(-1), Command(func() {
			req.SkipMergeTs = SkipMergeTs.Get() == "1"
		})), colSpan(2, Sticky("nw")),
	)
	WithSkipLog = Variable(0)
	DebugLog = Variable(0)
	row(
		TCheckbutton(Txt("调试日志"), DebugLog, Onvalue(1), Offvalue(-1), Command(func() {
			req.DebugLog = DebugLog.Get() == "1"
		})), colSpan(2, Sticky("nw")),
		TCheckbutton(Txt("记录跳过日志"), WithSkipLog, Onvalue(1), Offvalue(-1), Command(func() {
			req.WithSkipLog = WithSkipLog.Get() == "1"
		})), colSpan(2, Sticky("nw")),
	)
	msg := TLabel(Textvariable("空闲"))
	sv := Variable(0)
	row(msg, colSpan(6, Sticky("nw")))
	row(TProgressbar(Value(0), sv, Length(800)), colSpan(6, Sticky("new")))
	onMerge := false
	onDownload := false
	merge = TButton(Txt("合并TS"), Command(func() {
		msg.Configure(Textvariable("合并点击"))
		if onMerge {
			merge.Configure(Textvariable("合并TS"))
		} else {
			merge.Configure(Textvariable("取消合并"))
		}
		onMerge = !onMerge
	}))
	download = TButton(Txt("开始下载"), Command(func() {
		if onDownload {
			if ctxCC != nil {
				ctxCC()
				ctx = nil
				ctxCC = nil
			}
			m3u8dcpp.CloseOldEnv()
			msg.Configure(Textvariable("空闲"))
			download.Configure(Textvariable("开始下载"))
		} else {
			if ctxCC != nil {
				ctxCC()
				ctx = nil
				ctxCC = nil
			}
			if !syncRequest(false) {
				return
			}
			ctx, ctxCC = context.WithCancel(context.Background())
			m3u8dcpp.StartDownload(req)
			waitingDownload(ctx, time.Millisecond*800, func(s m3u8d.GetStatus_Resp) bool {
				if renderLock.TryLock() {
					render = append(render, func() {
						sv.Set(s.Percent)
						msg.Configure(Textvariable("(" + s.Title + ")" + s.StatusBar))
					})
					renderLock.Unlock()
				}
				if s.ErrMsg != "" {
					ctxCC()
					ctx = nil
					ctxCC = nil
					onDownload = false
					m3u8dcpp.CloseOldEnv()
					defer func() {
						for !renderLock.TryLock() {
						}
						defer renderLock.Unlock()
						if len(render) > 10 {
							render = render[len(render)-10:]
						}
						render = append(render, func() {
							download.Configure(Textvariable("开始下载"))
							msg.Configure(Textvariable("空闲"))
							MessageBox(Default("ok"), Title("错误"), Detail("下载出错:"+s.ErrMsg), Parent(App), Type("ok"))
						})
					}()
					return true
				} else if s.IsCancel || !s.IsDownloading {
					ctxCC()
					ctx = nil
					ctxCC = nil
					onDownload = false
					m3u8dcpp.CloseOldEnv()
					defer func() {
						for !renderLock.TryLock() {

						}
						defer renderLock.Unlock()
						if len(render) > 10 {
							render = render[len(render)-10:]
						}
						render = append(render, func() {
							download.Configure(Textvariable("开始下载"))
							msg.Configure(Textvariable("空闲"))
						})
					}()
					return true
				}
				return false
			})
			download.Configure(Textvariable("取消下载"))
		}
		onDownload = !onDownload
	}))
	row(
		colSpan(2),
		merge, colSpan(2),
		download, colSpan(2),
	)
	syncGUI()
	fn.Panic(ActivateTheme("azure light"))
	WmProtocol(App, "WM_DELETE_WINDOW", Command(func(e *Event) {
		if ctxCC != nil {
			ctxCC()
			ctx = nil
			ctxCC = nil
			if onDownload {
				m3u8dcpp.CloseOldEnv()
			}
		}
		saveConfig()
		os.Exit(0)
	}))
	App.SetResizable(false, false)
	_, _ = NewTicker(time.Millisecond*500, func() {
		if len(render) > 0 {
			renderLock.RLock()
			defer renderLock.RUnlock()
			for _, f := range render {
				f()
			}
		}
	})
	App.IconPhoto(NewPhoto(iconData))
	App.Center().Wait()
}
func truncate(v string) string {
	var i = strings.IndexByte(v, '.')
	if i < 0 {
		return v
	}
	return v[0:i]
}
func waitingDownload(ctx context.Context, d time.Duration, act func(m3u8d.GetStatus_Resp) bool) {
	go func() {
		tk := time.NewTicker(d)
	f:
		for {
			select {
			case <-tk.C:
				if act(m3u8dcpp.GetStatus()) {
					break
				}
			case <-ctx.Done():
				break f
			}
		}
		tk.Stop()
	}()
}

type CurlDialog struct {
	data         *CurlData
	win          *ToplevelWidget
	code         *TextWidget
	buttonFrame  *TFrameWidget
	okButton     *TButtonWidget
	cancelButton *TButtonWidget
}

func (s *CurlDialog) onOk() {
	s.data.Text = s.code.Text()
	s.data.Ok = true
	Destroy(s.win)
}

func (s *CurlDialog) onCancel() { Destroy(s.win) }

func (s *CurlDialog) ShowModal() {
	s.win.Raise(App)
	Focus(s.win)
	Focus(s.code)
	GrabSet(s.win)
	s.win.Center().Wait()
}

var iconData = Data(ico)

type CurlData struct {
	Ok   bool
	Text string
}

func curlDialog(data *CurlData) *CurlDialog {
	dlg := &CurlDialog{data: data}
	dlg.win = App.Toplevel()
	dlg.win.SetResizable(false, false)
	dlg.win.IconPhoto(NewPhoto(iconData))
	dlg.win.WmTitle("curl模式")
	WmProtocol(dlg.win.Window, WM_DELETE_WINDOW, dlg.onCancel)
	dlg.code = dlg.win.Text()
	dlg.code.Insert("1.0", data.Text)
	dlg.buttonFrame = dlg.win.TFrame()
	dlg.okButton = dlg.buttonFrame.TButton(Txt("确定"),
		Command(dlg.onOk))
	dlg.cancelButton = dlg.buttonFrame.TButton(Txt("取消"),
		Command(dlg.onCancel))
	opts := Opts{Padx(3), Pady(3)}
	Grid(dlg.code, Row(0), Column(1), Sticky(WE), opts)
	Grid(dlg.buttonFrame, Row(1), Column(0), Columnspan(2),
		opts)
	Grid(dlg.okButton, Row(0), Column(0), Sticky(E), opts)
	Grid(dlg.cancelButton, Row(0), Column(1), Sticky(E),
		opts)
	GridColumnConfigure(dlg.win, 1, Weight(1))
	return dlg

}
