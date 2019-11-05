package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/gotk3/gotk3/gtk"
	"github.com/sunshine69/gnote/forms"
)

func main() {
	dbPath := flag.String("db","","Path to the database file")
	doMigrate := flag.Bool("mig",false,"Migrate")
	flag.Parse()
	if *doMigrate {
		forms.DoMigration()
		os.Exit(0)
	}

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
	gnoteApp := forms.GnoteApp {
		Builder: builder,
	}

	gnoteApp.InitApp()
	gtk.Main()
}