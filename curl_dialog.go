package main

import (
	. "modernc.org/tk9.0"
)

type CurlDialog struct {
	Ok           bool
	Text         string
	win          *ToplevelWidget
	code         *TextWidget
	buttonFrame  *TFrameWidget
	okButton     *TButtonWidget
	cancelButton *TButtonWidget
}

func (s *CurlDialog) onOk() {
	s.Text = s.code.Text()
	s.Ok = true
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

func newCurlDialog(data string, s *TKApp) *CurlDialog {
	dlg := &CurlDialog{Text: data}
	dlg.win = App.Toplevel()
	dlg.win.SetResizable(false, false)
	dlg.win.IconPhoto(NewPhoto(iconData))
	dlg.win.WmTitle(s.Get(capCURL))
	WmProtocol(dlg.win.Window, WM_DELETE_WINDOW, dlg.onCancel)
	dlg.code = dlg.win.Text()
	dlg.code.Insert("1.0", data)
	dlg.buttonFrame = dlg.win.TFrame()
	dlg.okButton = dlg.buttonFrame.TButton(Txt(s.Get(btnOk)),
		Command(dlg.onOk))
	dlg.cancelButton = dlg.buttonFrame.TButton(Txt(s.Get(btnCancel)),
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
