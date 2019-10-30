package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"strings"
	"flag"
	"fmt"
	"os"
	"github.com/gotk3/gotk3/gtk"
	"./forms"
)

// The GnoteApp struct which holds common global object being used across gui and handler objects
type GnoteApp struct {
	builder *gtk.Builder
	signals map[string]interface{}
	mainWindow *gtk.Window
	model *gtk.ListStore
	treeView *gtk.TreeView
	selectedID *[]int
}

//RowActivated - Process when a treeview list row activated. Pop up a note window with the id
func (app *GnoteApp) RowActivated(treeView *gtk.TreeView,
	path *gtk.TreePath,
	column *gtk.TreeViewColumn) {
	// note_id = model.get_value(model.get_iter(path), 0)
	model, _ := treeView.GetModel()
	iter, _ := model.GetIter(path)
	_v, _ := model.GetValue( iter, 0 )
	v, _ := _v.GoValue()
	forms.NewNotePad(v.(int))
}

//ResultListKeyPress - evt
func (app *GnoteApp) ResultListKeyPress(w *gtk.TreeView, ev *gdk.Event) {
	keyEvent := &gdk.EventKey{ev}
	// fmt.Printf("DEBUG KEY %v\n", keyEvent.KeyVal() )
	if keyEvent.KeyVal() == 65535 {//Delete key
		for _, id := range(*app.selectedID) {
			fmt.Printf("ID %v\n",id)
			sql := fmt.Sprintf("DELETE FROM notes WHERE ID = '%d';", id)
			if e := forms.DbConn.Unscoped().Exec(sql).Error; e != nil {
				fmt.Printf("ERROR %v\n",e)
			}
		}
		app.doSearch()
	}
}

//TreeSelectionChanged - evt
func (app *GnoteApp) TreeSelectionChanged(s *gtk.TreeSelection) {
	// Returns glib.List of gtk.TreePath pointers
	ListStore := app.model

	rows := s.GetSelectedRows(ListStore)
	items := make([]int, 0, rows.Length())

	for l := rows; l != nil; l = l.Next() {
		path := l.Data().(*gtk.TreePath)
		iter, _ := ListStore.GetIter(path)
		value, _ := ListStore.GetValue(iter, 0)
		str, _ := value.GoValue()
		items = append(items, str.(int))
	}
	app.selectedID = &items
}

//NewNoteFromFile -
func (app *GnoteApp) NewNoteFromFile(o *gtk.FileChooserButton) {
	forms.NewNoteFromFile(o.GetFilename())
}

func (app *GnoteApp) initApp() {
	builder := app.builder

	signals := map[string]interface{} {
		"ShowAbout": app.showAbout,
		"OpenPref": app.openPref,
		"NewNote": app.newNote,
		"OpenDbfile": app.openDBFile,
		"DoExit": app.doExit,
		"DoClearSearchbox": app.doClearSearchbox,
		"DoSearch": app.doSearch,
		"RowActivated": app.RowActivated,
		"ResultListKeyPress": app.ResultListKeyPress,
		"TreeSelectionChanged": app.TreeSelectionChanged,
		"NewNoteFromFile": app.NewNoteFromFile,
	}

	builder.ConnectSignals(signals)

	_obj, err := builder.GetObject("window")
	if err != nil {
		panic(err)
	}
	window := _obj.(*gtk.Window)

	window.SetTitle("gnote")
	window.SetDefaultSize(300, 250)
	_, err = window.Connect("destroy", func() {
		gtk.MainQuit()
	})
	if err != nil {
		panic(err)
	}

	_obj, err = builder.GetObject("status_bar")
	statusBar := _obj.(*gtk.Statusbar)
	statusBar.Push(1, "Welcome to gnote")

	app.mainWindow = window

	_w, _ := builder.GetObject("treeView")
	wT := _w.(*gtk.TreeView)
	app.treeView = wT
	app.model, _ = gtk.ListStoreNew(glib.TYPE_INT, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)
	wT.SetModel(app.model)
	wT.SetHeadersVisible(true)
	renderer, _ := gtk.CellRendererTextNew()
    col1, _ := gtk.TreeViewColumnNewWithAttribute("Title", renderer, "text", 1)
	col2, _ := gtk.TreeViewColumnNewWithAttribute("Date Log", renderer,"text", 2)
	col3, _ := gtk.TreeViewColumnNewWithAttribute("Last update", renderer,"text", 3)
    col1.SetResizable(true)
	col2.SetResizable(true)
	col3.SetResizable(true)
    col1.SetMaxWidth(200)
    col1.SetMinWidth(20)
    col2.SetMinWidth(20)
	wT.AppendColumn(col1)
	wT.AppendColumn(col2)
	wT.AppendColumn(col3)
    selection, _ := wT.GetSelection()
	selection.SetMode(gtk.SELECTION_MULTIPLE)
	// wT.SetSearchColumn(0)

	window.ShowAll()
}

// looks like handlers can literally be any function or method
func (app *GnoteApp) openPref() {
	fmt.Println("Open Pref")
	builder, err := gtk.BuilderNewFromFile("glade/gnote-editpref.glade")
	if err != nil {
		panic(err)
	}
	_obj, err := builder.GetObject("edit_pref")
	if err != nil {
		panic(err)
	}
	dialogWindow := _obj.(*gtk.Dialog)
	dialogWindow.Show()
}

func (app *GnoteApp) openDBFile() {
	fmt.Println("Open DB File")
}

func (app *GnoteApp) newNote() {
	forms.NewNotePad(-1)
}

func (app *GnoteApp) doExit() {
	fmt.Println("Exit called")
}

func (app *GnoteApp) showAbout() {
	fmt.Println("show about")
}

func (app *GnoteApp) doClearSearchbox() {
	fmt.Println("doClearSearchbox")
}

func (app *GnoteApp) doSearch() {
	fmt.Println("doSearch")
	b := app.builder
	_w, _ := b.GetObject("searchBox")
	w := _w.(*gtk.SearchEntry)
	searchFlags := false
	keyword, _ := w.GetText()
	q := ""
	tokens := []string{}
	if strings.HasPrefix(keyword, "F:") {
		tokens = strings.Split(keyword[2:], ":")
		searchFlags = true
	} else if strings.HasPrefix(keyword, "FLAGS:"){
		tokens = strings.Split(keyword[6:], ":")
		searchFlags = true
	}
	if searchFlags {
		_l := len(tokens)
		for i, t := range(tokens) {
			if i == _l - 1 {
				q = fmt.Sprintf("%v (flags LIKE '%%%v%%') ORDER BY datelog DESC;", q, t)
			} else {
				q = fmt.Sprintf("%v (flags LIKE '%%%v%%') AND ", q, t)
			}
		}
		q = fmt.Sprintf("SELECT id, title, datelog, timestamp from notes WHERE %v", q)
	} else {
		tokens := strings.Split(keyword, " & ")
		_l := len(tokens)

		for i, t := range(tokens) {
			if i == _l - 1 {
				q = fmt.Sprintf("%v (title LIKE '%%%v%%' OR content LIKE '%%%v%%') ORDER BY datelog DESC;", q, t, t)
			} else {
				q = fmt.Sprintf("%v (title LIKE '%%%v%%' OR content LIKE '%%%v%%') AND ", q, t, t)
			}
		}
		q = fmt.Sprintf("SELECT id, title, datelog, timestamp from notes WHERE %v", q)
	}
	fmt.Println(q)
	rows, e := forms.DbConn.Raw(q).Rows()
	if e != nil {
		fmt.Printf("ERROR - exec sql\n")
	}
	defer rows.Close()

	app.model.Clear()

	id, count := 0, 0
	var title, datelog, lastUpdate string
	for rows.Next() {
		rows.Scan(&id, &title, &datelog, &lastUpdate)
		// fmt.Printf("row: %v - %v %v\n", id, title, datelog)
		iter := app.model.Append()
		if e := app.model.Set(iter,
			[]int{0, 1, 2, 3},
			[]interface{}{id, title, datelog, lastUpdate}); e != nil {
				fmt.Printf("ERROR appending data to model %v\n", e)
			}
		count = count + 1
	}
	_o, _ := app.builder.GetObject("status_bar")
	s := _o.(*gtk.Statusbar)
	s.Pop(1)
	s.Push(1, fmt.Sprintf( "Found %d notes", count))
}

func main() {
	dbPath := flag.String("db","","Path to the database file")
	flag.Parse()
	homeDir, e := os.UserHomeDir()
	if e != nil {
		fmt.Printf("ERROR %v\n", e)
	}
	if *dbPath == "" {
		*dbPath =  fmt.Sprintf("%s%s%s", homeDir, string(os.PathSeparator), ".gnote.db")
		fmt.Printf("Use the database file %s\n", *dbPath)
	}
	os.Setenv("DBPATH", *dbPath)
	forms.SetupConfigDB()
	if _, e := forms.GetConfig("config_created"); e != nil {
		fmt.Println("Setup default config ....")
		forms.SetupDefaultConfig()
	}
	gtk.Init(&os.Args)
	builder, err := gtk.BuilderNewFromFile("glade/gnote.glade")
	if err != nil {
		panic(err)
	}
	gnoteApp := GnoteApp {
		builder: builder,
	}
	gnoteApp.initApp()
	gtk.Main()
}