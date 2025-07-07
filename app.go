package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/ZenLiuCN/fn"
	"github.com/orestonce/m3u8d"
	"github.com/orestonce/m3u8d/m3u8dcpp"
	. "modernc.org/tk9.0"
	_ "modernc.org/tk9.0/themes/azure"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//go:embed favicon.ico
var ico []byte

var iconData = Data(ico)

type TKApp struct {
	GridLayout
	Req                                                                                              m3u8d.StartDownload_Req
	uri, folder, file, skipScript, proxy, tempFolder                                                 *TEntryWidget
	threadVal                                                                                        *LabelWidget
	msg                                                                                              *TLabelWidget
	progress, threads, Insecure, SkipRemoveTs, UseServerSideTime, SkipMergeTs, WithSkipLog, DebugLog *VariableOpt
	merge, download                                                                                  *TButtonWidget
	locator                                                                                          *TButtonWidget
	ctx                                                                                              context.Context
	ctxCC                                                                                            context.CancelFunc
	render                                                                                           []func()
	mux                                                                                              sync.RWMutex
	rows                                                                                             int
	state                                                                                            atomic.Int32
	ticker                                                                                           *Ticker
	Language
	qux         sync.RWMutex
	Queue       []m3u8d.StartDownload_Req
	btnAddQueue *TButtonWidget
	listQueue   *ListboxWidget
	inQueue     bool
	qw, qh      int
	note        *TNotebookWidget
	tab0, tab1  *TFrameWidget
	msgFile     *TLabelWidget
}

func (s *TKApp) queueWidth() int {
	if s.qw > 0 {
		return s.qw
	}
	var t = s.Get(numQueueWidth)
	if t == "" {
		s.qw = 130
	} else {
		s.qw, _ = strconv.Atoi(t)
		if s.qw == 0 {
			s.qw = 130
		}
	}
	return s.qw
}
func (s *TKApp) queueHeight() int {
	if s.qh > 0 {
		return s.qh
	}
	var t = s.Get(numQueueHeight)
	if t == "" {
		s.qh = 30
	} else {
		s.qh, _ = strconv.Atoi(t)
		if s.qh == 0 {
			s.qh = 30
		}
	}
	return s.qh
}

func (s *TKApp) loadLang() {
	s.Language = make(map[Lang]string)
	if s.LoadLanguage() {
		return
	}
	s.InitializeDefault()
}

var stateNormal = State("normal")
var stateDisabled = State("disabled")

func (s *TKApp) init() *TKApp {
	defer func() {
		switch x := recover().(type) {
		case nil:
		case error:
			s.error((x.Error()))
		default:
			s.error(fmt.Sprintf("%+v", x))
		}
	}()
	s.Padding = Opts{Padx(3), Pady(3)}
	s.loadLang()

	s.load()
	note := TNotebook()
	s.note = note

	//region main table
	T0 := note.TFrame()
	s.tab0 = T0
	note.Add(T0, Txt(s.Get(tabMain)))
	s.uri = T0.TEntry(Textvariable(s.Req.M3u8Url), Placeholder(""))
	s.folder = T0.TEntry(Textvariable(s.Req.SaveDir), Placeholder(s.Get(phtFolder)))
	s.file = T0.TEntry(Textvariable(s.Req.FileName), Placeholder(s.Get(phtFileName)))
	s.skipScript = T0.TEntry(Textvariable(s.Req.SkipTsExpr), Placeholder(s.Get(phtSkipCode)))
	s.proxy = T0.TEntry(Textvariable(s.Req.SetProxy), Placeholder(s.Get(phtProxy)))
	n := runtime.NumCPU()
	if s.Req.ThreadCount <= 0 {
		s.Req.ThreadCount = n
	}
	s.threads = Variable(s.Req.ThreadCount)
	s.threadVal = T0.Label(Textvariable(strconv.Itoa(s.Req.ThreadCount)))
	s.tempFolder = T0.TEntry(Textvariable(s.Req.TsTempDir), Placeholder(s.Get(phtTemp)))
	s.Insecure = Variable(0)
	s.SkipRemoveTs = Variable(0)
	s.UseServerSideTime = Variable(0)
	s.SkipMergeTs = Variable(0)
	s.WithSkipLog = Variable(0)
	s.DebugLog = Variable(0)
	s.msg = T0.TLabel(Textvariable(s.Get(msgIdle)))
	s.msgFile = T0.TLabel(Textvariable(""))
	s.progress = Variable(0)
	s.merge = T0.TButton(Txt(s.Get(btnMerge)), Command(func() {
		switch s.state.Load() {
		case 0:
			if !s.sync(true) {
				return
			}
			if s.Req.TsTempDir == "" {
				s.error(s.Get(errTSFolder))

				return
			}
			if s.Req.FileName == "" || s.Req.SaveDir == "" {
				s.error(s.Get(errSaveFolder))
				return
			}
			var out = filepath.Join(s.Req.SaveDir, s.Req.FileName)
			if !strings.HasSuffix(out, ".mp4") {
				out += ".mp4"
			}
			r := m3u8dcpp.MergeTsDir(s.Req.TsTempDir, out, s.Req.UseServerSideTime, true)
			if r.ErrMsg != "" {
				s.error(fmt.Sprintf(s.Get(errMerge), r.ErrMsg))
				return
			}
			s.ctx, s.ctxCC = context.WithCancel(context.Background())
			waitingMerge(s.ctx, time.Millisecond*800, s.handleMerge)
			s.merge.Configure(Txt(s.Get(btnStopMerge)))
			s.state.Store(1)
			s.download.Configure(stateDisabled)
			s.capMerge()
			s.msgFile.Configure(Textvariable(s.Req.FileName))
		case 1:
			m3u8dcpp.MergeStop()
			s.ctxCC()
			s.ctx = nil
			s.ctxCC = nil
			s.merge.Configure(Txt(s.Get(btnMerge)))
			s.state.Store(0)
			s.download.Configure(stateNormal)
			s.msgFile.Configure(Textvariable(""))
			s.capIdle()
		}
	}))
	Tooltip(s.merge, s.Get(tipMerge))
	s.download = T0.TButton(Txt(s.Get(btnDownload)), Command(func() {
		switch s.state.Load() {
		case 0:
			if !s.sync(true) {
				if len(s.Queue) > 0 {
					s.ctx, s.ctxCC = context.WithCancel(context.Background())
					v, _ := s.nextQueue()
					s.inQueue = true
					s.msgFile.Configure(Textvariable(v.FileName))
					m3u8dcpp.StartDownload(v)
				} else {
					s.sync(false)
					return
				}
			} else {
				s.ctx, s.ctxCC = context.WithCancel(context.Background())
				s.msgFile.Configure(Textvariable(s.Req.FileName))
				m3u8dcpp.StartDownload(s.Req)
			}
			waitingDownload(s.ctx, time.Millisecond*800, s.handleDownload)
			s.download.Configure(Txt(s.Get(btnCancelDownload)))
			s.state.Store(2)
			s.merge.Configure(stateDisabled)
			s.capDownloading()
		case 2:
			s.ctxCC()
			s.ctx = nil
			s.ctxCC = nil
			m3u8dcpp.CloseOldEnv()
			s.msg.Configure(Txt(s.Get(msgIdle)))
			s.progress.Set(0)
			s.download.Configure(Txt(s.Get(btnDownload)))
			s.state.Store(0)
			s.merge.Configure(stateNormal)
			s.capIdle()
			s.msgFile.Configure(Textvariable(""))
		}
	}))
	Tooltip(s.download, s.Get(tipDownload))
	s.btnAddQueue = T0.TButton(Txt(s.Get(btnAddQueueTxt)), Command(func() {
		s.addQueue(false)
	}))

	s.Row(
		T0.Label(Txt(s.Get(lblM3u8URL))), Sticky("nw"),
		s.uri, ColSpan(9, Sticky("new")),
		T0.TButton(Txt(s.Get(btnCurl)), Command(func() {
			s.sync(true)
			if s.Req.M3u8Url == "" {
				s.error(s.Get(errInvalidM3U8))
				return
			}
			var d = newCurlDialog(m3u8d.RunDownload_Req_ToCurlStr(s.Req), s)
			d.ShowModal()
			if d.Ok {
				v := m3u8d.ParseCurlStr(d.Text)
				if v.ErrMsg != "" {
					s.error(fmt.Sprintf(s.Get(errInvalidCURL), v.ErrMsg))
					return
				} else {
					s.Req = v.DownloadReq
				}
			}
			s.syncUI(nil)
		})), Padx(10),
	)
	s.Row(
		T0.Label(Txt(s.Get(lblSaveFolder))), Sticky("nw"),
		s.folder, ColSpan(9, Sticky("new")),
		T0.TButton(Txt(s.Get(btnChoose)), Command(func() {
			dir := ChooseDirectory(
				Initialdir(s.folder.Variable()),
				Mustexist(true),
				Parent(App),
				Title(s.Get(capSaveFolder)))
			if dir != "" {
				s.folder.Configure(Textvariable(dir))
			}
		})),
	)

	s.Row(
		T0.Label(Txt(s.Get(lblSaveName))), Sticky("nw"),
		s.file, ColSpan(9, Sticky("new")),
	)

	s.Row(
		T0.Label(Txt(s.Get(lblSkipCode))), Sticky("nw"),
		s.skipScript, ColSpan(9, Sticky("new")),
	)

	s.Row(
		T0.Label(Txt(s.Get(lblProxy))), Sticky("nw"),
		s.proxy, ColSpan(9, Sticky("new")),
	)

	s.Row(
		T0.Label(Txt(s.Get(lblBatch))), Sticky("nw"),
		s.threadVal, Sticky("nw"),
		T0.TScale(From(1), To(4*n), s.threads, Command(func() {
			s.threadVal.Configure(Textvariable(truncate(s.threads.Get())))
		})), ColSpan(8, Sticky("new")),
	)

	s.Row(
		T0.Label(Txt(s.Get(lblTemp))), Sticky("nw"),
		s.tempFolder, ColSpan(9, Sticky("new")),
		T0.TButton(Txt(s.Get(btnChoose)), Command(func() {
			dir := ChooseDirectory(Initialdir(s.tempFolder.Variable()), Mustexist(true), Parent(App), Title(s.Get(capTemp)))
			if dir != "" {
				s.tempFolder.Configure(Textvariable(dir))
			}
		})),
	)

	s.Row(
		T0.TCheckbutton(Txt(s.Get(lblInsecure)), s.Insecure, Onvalue(1), Offvalue(-1), Command(func() {
			s.Req.Insecure = s.Insecure.Get() == "1"
		})), ColSpan(5, Sticky("nw")),
		T0.TCheckbutton(Txt(s.Get(lblKeepTS)), s.SkipRemoveTs, Onvalue(1), Offvalue(-1), Command(func() {
			s.Req.SkipRemoveTs = s.SkipRemoveTs.Get() == "1"
		})), ColSpan(5, Sticky("nw")),
	)

	s.Row(
		T0.TCheckbutton(Txt(s.Get(lblUseServerTime)), s.UseServerSideTime, Onvalue(1), Offvalue(-1), Command(func() {
			s.Req.UseServerSideTime = s.UseServerSideTime.Get() == "1"
		})), ColSpan(5, Sticky("nw")),
		T0.TCheckbutton(Txt(s.Get(lblSkipMerge)), s.SkipMergeTs, Onvalue(1), Offvalue(-1), Command(func() {
			s.Req.SkipMergeTs = s.SkipMergeTs.Get() == "1"
		})), ColSpan(5, Sticky("nw")),
	)

	s.Row(
		T0.TCheckbutton(Txt(s.Get(lblDebugLog)), s.DebugLog, Onvalue(1), Offvalue(-1), Command(func() {
			s.Req.DebugLog = s.DebugLog.Get() == "1"
		})), ColSpan(5, Sticky("nw")),
		T0.TCheckbutton(Txt(s.Get(lblSkipLog)), s.WithSkipLog, Onvalue(1), Offvalue(-1), Command(func() {
			s.Req.WithSkipLog = s.WithSkipLog.Get() == "1"
		})), ColSpan(5, Sticky("nw")),
	)

	s.Row(s.msgFile, ColSpan(11, Sticky("nw")))
	s.Row(s.msg, ColSpan(11, Sticky("nw")))
	s.Row(
		T0.TProgressbar(Value(0), s.progress, Length(850)), ColSpan(10, Sticky("new")),
		s.btnAddQueue,
	)
	if s.Get(btnFindIDMWindow) != "" && IDMConfig(s) {
		s.Row(
			ColSpan(2),
			s.merge, ColSpan(2),
			s.download, ColSpan(1),
			s.locator, ColSpan(1),
		)
	} else {
		s.Row(
			ColSpan(2),
			s.merge, ColSpan(2),
			s.download, ColSpan(1),
		)
	}

	//endregion

	//region queue table

	s.ResetRow()
	T1 := note.TFrame()
	note.Add(T1, Txt(s.Get(tabQueue)))
	s.tab1 = T1
	s.listQueue = T1.Listbox(Height(s.queueHeight()), Setgrid(true), Width(s.queueWidth()))
	s.syncQueue()
	s.Row(
		s.listQueue, ColSpan(10, Sticky("news")),
	)
	s.Row(
		T1.TButton(Txt(s.Get(btnRemoveQueue)), Command(func() {
			if len(s.listQueue.Curselection()) == 0 {
				return
			}
			var i = s.listQueue.Curselection()[0]
			if i == 0 && s.inQueue {
				s.error(s.Get(msgQueueDownloading))
				return
			}
			s.qux.Lock()
			defer s.qux.Unlock()
			s.Queue = slices.Delete(s.Queue, i, i+1)
			s.listQueue.Delete(i)
		})),
		T1.TButton(Txt(s.Get(btnQueueTask)), Command(func() {
			if len(s.listQueue.Curselection()) == 0 {
				return
			}
			var i = s.listQueue.Curselection()[0]
			var r = s.Queue[i]
			var m = printf(
				s.Get(lblM3u8URL), r.M3u8Url,
				s.Get(lblProxy), r.SetProxy,
				s.Get(lblSaveFolder), r.SaveDir,
				s.Get(lblSaveName), r.FileName,
				s.Get(lblSkipCode), r.SkipTsExpr,
				s.Get(lblTemp), r.TsTempDir,
				s.Get(lblBatch), r.ThreadCount,
				s.Get(lblInsecure), r.Insecure,
				s.Get(lblKeepTS), r.SkipRemoveTs,
				s.Get(lblUseServerTime), r.UseServerSideTime,
				s.Get(lblSkipMerge), r.SkipMergeTs,
				s.Get(lblDebugLog), r.DebugLog,
				s.Get(lblSkipLog), r.WithSkipLog,
			)
			s.infoClip(m)
		})),
		T1.TButton(Txt(s.Get(btnQueueModify)), Command(func() {
			if len(s.listQueue.Curselection()) == 0 {
				return
			}
			var i = s.listQueue.Curselection()[0]
			if i == 0 && s.inQueue {
				s.error(s.Get(msgQueueDownloading))
				return
			}
			s.qux.Lock()
			defer s.qux.Unlock()
			var r = s.Queue[i]
			s.syncUI(&r)
			s.Queue = slices.Delete(s.Queue, i, i+1)
			s.listQueue.Delete(i)
			s.note.Select(0)
		})),
	)
	//endregion
	Grid(note)

	if s.Get(themeColor) == "dark" {
		fn.Panic(ActivateTheme("azure dark"))
	} else {
		fn.Panic(ActivateTheme("azure light"))
	}
	WmProtocol(App, "WM_DELETE_WINDOW", Command(func(e *Event) {
		if s.ctxCC != nil {
			s.ctxCC()
			s.ctx = nil
			s.ctxCC = nil
			switch s.state.Load() {
			case 1:
				m3u8dcpp.MergeStop()
			case 2:
				m3u8dcpp.CloseOldEnv()
			}
		}
		s.save()
		os.Exit(0)
	}))
	App.SetResizable(false, false)
	s.capIdle()
	return s
}

var ml = 0

func printf(v ...any) string {
	var s = strings.Builder{}
	if ml == 0 {
		for i := 0; i < len(v); i += 2 {
			var l = v[i]
			var xl = len(l.(string))
			if xl > ml {
				ml = xl
			}
		}
		ml += 1
	}
	for i := 0; i < len(v); i += 2 {
		if i > 0 {
			s.WriteByte('\n')
		}
		var lbl = v[i].(string)
		s.WriteString(lbl)
		s.WriteString(strings.Repeat(" ", (ml-len(lbl))/2+2))
		if x, ok := v[i+1].(string); ok {
			s.WriteString(x)
		} else {
			s.WriteString(fmt.Sprintf("%v", v[i+1]))
		}
	}
	return s.String()
}
func (s *TKApp) capIdle() {
	App.WmTitle("m3u8d")
}
func (s *TKApp) capDownloading() {
	App.WmTitle("m3u8d (downloading)")
}
func (s *TKApp) capMerge() {
	App.WmTitle("m3u8d (merging)")
}
func (s *TKApp) stage(v int) string {
	switch v {
	case 0:
		return s.Get(stageIdle)
	case 1:
		return s.Get(stageFetchM3U8)
	case 2:
		return s.Get(stageM3U8Analysis)
	case 3:
		return s.Get(stageDownloading)
	case 4:
		return s.Get(stageMergeAnalysis)
	case 5:
		return s.Get(stageMergeFiles)
	default:
		return fmt.Sprintf("%d", v)
	}
}
func (s *TKApp) speed(v m3u8d.SpeedInfo) string {
	return fmt.Sprintf("%s %s", v.BytePerSecondText, v.RemainTimeText)
}
func (s *TKApp) handleMerge(r m3u8dcpp.MergeGetProgressPercent_Resp) bool {
	if r.IsRunning {
		s.task(func() {
			s.progress.Set(r.Percent)
			s.msg.Configure(Txt(fmt.Sprintf("(%s) %s", r.Title, r.SpeedText))) //TODO i18n supports need modify source repository
		})
		return false
	}
	s.mustTask(func() {
		s.msg.Configure(Txt(s.Get(msgIdle)))
		s.progress.Set(0)
		s.merge.Configure(Txt(s.Get(btnMerge)))
		s.state.Store(0)
		s.download.Configure(stateNormal)
		s.ctxCC()
		s.ctx = nil
		s.ctxCC = nil
		s.capIdle()
		s.msgFile.Configure(Textvariable(""))
	})
	return true
}
func (s *TKApp) handleDownload(r m3u8d.GetStatus_Resp) bool {
	if r.IsSkipped {
		s.doDownloadFinished(r.SaveFileTo)
		return true
	} else if r.ErrMsg != "" {
		s.doDownloadError(r.ErrMsg)
		return true
	} else if r.IsCancel {
		s.doDownloadCancel()
		return true
	} else if r.SaveFileTo != "" {
		s.doDownloadDone(r.SaveFileTo)
		return true
	} else {
		s.doDownloadUpdate(r.Percent, "("+r.Title+")"+r.StatusBar) //TODO i18n supports need modify underling codes
		return false
	}
}
func (s *TKApp) Await() {
	s.ticker, _ = NewTicker(time.Millisecond*300, s.tick)
	App.IconPhoto(NewPhoto(iconData))
	App.Center().Wait()
}
func (s *TKApp) doDownloadUpdate(percent int, message string) {
	s.task(func() {
		s.progress.Set(percent)
		s.msg.Configure(Textvariable(message))
	})
}
func (s *TKApp) doDownloadFinished(file string) {
	s.mustTask(func() {
		s.msg.Configure(Textvariable(s.Get(msgIdle)))
		s.progress.Set(0)
		s.download.Configure(Txt(s.Get(btnDownload)))
		s.state.Store(0)
		s.merge.Configure(stateNormal)
		s.capIdle()
		s.msgFile.Configure(Textvariable(""))
		s.error(fmt.Sprintf(s.Get(errDownloaded), file))
		s.inQueue = false
	})
}
func (s *TKApp) doDownloadDone(file string) {
	s.mustTask(func() {
		s.msg.Configure(Textvariable(s.Get(msgIdle)))
		s.progress.Set(0)
		s.download.Configure(Txt(s.Get(btnDownload)))
		if v, ok := s.nextQueue(); ok {
			if s.ctxCC != nil {
				s.ctxCC()
			}
			s.inQueue = true
			s.ctx, s.ctxCC = context.WithCancel(context.Background())
			s.msgFile.Configure(Textvariable(v.FileName))
			m3u8dcpp.StartDownload(v)
			waitingDownload(s.ctx, time.Millisecond*800, s.handleDownload)
			s.download.Configure(Txt(s.Get(btnCancelDownload)))
			s.state.Store(2)
			s.merge.Configure(stateDisabled)
			return
		}
		s.state.Store(0)
		s.merge.Configure(stateNormal)
		s.capIdle()
		if s.inQueue {
			s.info(s.Get(infQueueDownloadDone))
		} else {
			s.info(fmt.Sprintf(s.Get(infDownloadDone), file))
		}
		s.msgFile.Configure(Textvariable(""))
		s.inQueue = false
	})
}
func (s *TKApp) doDownloadError(errMsg string) {
	s.mustTask(func() {
		s.msg.Configure(Textvariable(s.Get(msgIdle)))
		s.progress.Set(0)
		s.download.Configure(Txt(s.Get(btnDownload)))
		s.state.Store(0)
		s.merge.Configure(stateNormal)
		s.capIdle()
		//s.msgFile.Configure(Textvariable(""))
		if strings.Contains(errMsg, "合并") {
			var v m3u8d.StartDownload_Req
			if s.inQueue && len(s.Queue) > 0 {
				v = s.Queue[0]
			} else {
				v = s.Req
			}
			s.syncUI(&v)
		}
		s.inQueue = false
		s.error(fmt.Sprintf(s.Get(errDownload), errMsg))
	})
}
func (s *TKApp) doDownloadCancel() {
	s.ctxCC()
	s.ctx = nil
	s.ctxCC = nil
	m3u8dcpp.CloseOldEnv()
	s.mustTask(func() {
		s.inQueue = false
		s.msg.Configure(Textvariable(s.Get(msgIdle)))
		s.progress.Set(0)
		s.download.Configure(Txt(s.Get(btnDownload)))
		s.state.Store(0)
		s.merge.Configure(stateNormal)
		s.msgFile.Configure(Textvariable(""))
		s.capIdle()
	})
}
func (s *TKApp) task(act ...func()) {
	if s.mux.TryLock() {
		s.render = append(s.render, act...)
		s.mux.Unlock()
	}
}
func (s *TKApp) mustTask(act ...func()) {
	for !s.mux.TryLock() {
	}
	defer s.mux.Unlock()
	if len(s.render) > 10 {
		s.render = s.render[len(s.render)-10:]
	}
	s.render = append(s.render, act...)
}
func (s *TKApp) syncUI(r *m3u8d.StartDownload_Req) {
	if r == nil {
		r = &s.Req
	}
	if r.Insecure {
		s.Insecure.Set(1)
	}
	if r.SkipRemoveTs {
		s.SkipRemoveTs.Set(1)
	}
	if r.UseServerSideTime {
		s.UseServerSideTime.Set(1)
	}
	if r.SkipMergeTs {
		s.SkipMergeTs.Set(1)
	}
	if r.ThreadCount <= 0 {
		r.ThreadCount = runtime.NumCPU()
	}
	if r.WithSkipLog {
		s.WithSkipLog.Set(1)
	}
	if r.DebugLog {
		s.DebugLog.Set(1)
	}
	s.uri.Configure(Textvariable(r.M3u8Url))
	s.file.Configure(Textvariable(r.FileName))
	s.folder.Configure(Textvariable(r.SaveDir))
	s.skipScript.Configure(Textvariable(r.SkipTsExpr))
	s.tempFolder.Configure(Textvariable(r.TsTempDir))
	s.proxy.Configure(Textvariable(r.SetProxy))
	s.threads.Set(r.ThreadCount)
}
func (s *TKApp) sync(mute bool) bool {
	s.Req.M3u8Url = s.uri.Textvariable()
	s.Req.SetProxy = s.proxy.Textvariable()
	s.Req.SaveDir = s.folder.Textvariable()
	s.Req.FileName = s.file.Textvariable()
	s.Req.SkipTsExpr = s.skipScript.Textvariable()
	s.Req.TsTempDir = s.tempFolder.Textvariable()
	//s.Req.ProgressBarShow = false
	var err error
	s.Req.ThreadCount, err = strconv.Atoi(truncate(s.threads.Get()))
	if err != nil || s.Req.ThreadCount < 1 {
		if !mute {
			s.error(fmt.Sprintf(s.Get(errBatchNum), truncate(s.threads.Get())))
		}
		return false
	}
	if s.Req.M3u8Url == "" {
		if !mute {
			s.error(s.Get(errInvalidM3U8))
		}
		return false
	}

	s.save()
	return true
}

func (s *TKApp) syncQueue() {
	s.qux.RLock()
	defer s.qux.RUnlock()
	s.listQueue.Delete(0, "end")
	for i, req := range s.Queue {
		s.listQueue.Insert(i, fmt.Sprintf("%s(%s)", req.FileName, req.M3u8Url))
	}
	s.updateQueueCaption()
}
func (s *TKApp) updateQueueCaption() {

}
func (s *TKApp) addQueue(mute bool) bool {
	if !s.sync(mute) {
		return false
	}
	var req m3u8d.StartDownload_Req
	req = s.Req
	for !s.qux.TryLock() {
	}
	s.Queue = append(s.Queue, req)
	s.qux.Unlock()
	s.listQueue.Insert(len(s.Queue)-1, fmt.Sprintf("%s(%s)", req.FileName, req.M3u8Url))
	s.save()
	s.uri.Configure(Textvariable(""))
	s.Req.M3u8Url = ""
	s.updateQueueCaption()
	return true
}
func (s *TKApp) nextQueue() (v m3u8d.StartDownload_Req, ok bool) {
	for s.qux.TryLock() {
	}
	defer s.qux.Unlock()
	var n = len(s.Queue)
	if n == 0 {
		ok = false
		return
	}
	if s.inQueue {
		s.Queue = s.Queue[1:]
		s.listQueue.Delete(0)
		s.save()
		s.updateQueueCaption()
	}
	if len(s.Queue) == 0 {
		ok = false
		return
	}
	v = s.Queue[0]
	ok = true
	return
}

type Save struct {
	Req   m3u8d.StartDownload_Req
	Queue []m3u8d.StartDownload_Req
}

func (s *TKApp) save() {
	f, _ := os.OpenFile("config.json", os.O_CREATE|os.O_TRUNC, os.ModePerm)
	_ = json.NewEncoder(f).Encode(Save{
		s.Req,
		s.Queue,
	})
	_ = f.Close()
}
func (s *TKApp) load() {
	f, err := os.Open("config.json")
	if err == nil {
		var srv Save
		fn.Panic(json.NewDecoder(f).Decode(&srv))
		s.Req = srv.Req
		s.Queue = srv.Queue
		_ = f.Close()
	}
}
func (s *TKApp) tick() {
	if len(s.render) > 0 {
		s.mux.RLock()
		var r = s.render
		s.render = nil
		s.mux.RUnlock()
		for _, f := range r {
			f()
		}
	}
}
func (s *TKApp) error(msg string) {
	MessageBox(Default("ok"), Title(s.Get(capError)), Detail(msg), Parent(App), Type("ok"), Icon("error"))
}
func (s *TKApp) info(msg string) {
	MessageBox(Default("ok"), Title(s.Get(capInfo)), Detail(msg), Parent(App), Type("ok"))
}
func (s *TKApp) infoClip(msg string) {
	if MessageBox(Default("ok"), Title(s.Get(capInfo)), Detail(msg), Parent(App), Type("okcancel")) == "ok" {
		ClipboardAppend(msg)
	}
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
					break f
				}
			case <-ctx.Done():
				break f
			}
		}
		tk.Stop()
	}()
}
func waitingMerge(ctx context.Context, d time.Duration, act func(m3u8dcpp.MergeGetProgressPercent_Resp) bool) {
	go func() {
		tk := time.NewTicker(d)
	f:
		for {
			select {
			case <-tk.C:
				if act(m3u8dcpp.MergeGetProgressPercent()) {
					break
				}
			case <-ctx.Done():
				break f
			}
		}
		tk.Stop()
	}()
}
