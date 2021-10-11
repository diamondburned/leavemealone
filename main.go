package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/state"
)

func init() {
	log.SetFlags(0)

	flag.Usage = func() {
		log.Println("usage:", filepath.Base(os.Args[0]), "folder_name")
		flag.PrintDefaults()
	}
}

func main() {
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatalln("missing $TOKEN")
	}

	flag.Parse()

	folderName := flag.Arg(0)
	if folderName == "" {
		flag.Usage()
		os.Exit(1)
	}

	s, err := state.New(token)
	if err != nil {
		log.Fatalln("failed to create state:", err)
	}

	ready, cancel := s.ChanFor(func(v interface{}) bool {
		_, ok := v.(*gateway.ReadyEvent)
		return ok
	})
	defer cancel()

	if err := s.Open(); err != nil {
		log.Fatalln("failed to open:", err)
	}

	defer s.CloseGracefully()

	<-ready
	cancel()

	for _, folder := range s.Ready().UserSettings.GuildFolders {
		if folder.Name != folderName {
			continue
		}

		fmt.Println("found folder with ID", folder.ID)
		fmt.Println("found these guilds:")

		for _, gID := range folder.GuildIDs {
			g, err := s.Guild(gID)
			if err != nil {
				fmt.Printf("  - %d (error: %s)\n", gID, err)
				continue
			}
			fmt.Printf("  - %s (%d)\n", g.Name, gID)
		}

		if !ask("continue?", 'Y', 'y') {
			fmt.Println()
			log.Fatalln("cancelled")
		}

		for _, gID := range folder.GuildIDs {
			if err := s.LeaveGuild(gID); err != nil {
				log.Printf("failed to leave guild %d: %v", gID, err)
			}
		}
	}
}

func ask(prompt string, expect ...byte) bool {
	fmt.Print(prompt, " ")

	var c [1]byte

	_, err := os.Stdin.Read(c[:])
	if err != nil {
		return false
	}

	for _, b := range expect {
		if b == c[0] {
			return true
		}
	}

	return false
}
