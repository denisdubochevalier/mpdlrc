// Package main
package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/fhs/gompd/v2/mpd"

	"github.com/env25/mpdlrc/internal/client"
	"github.com/env25/mpdlrc/internal/config"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	cfg := config.DefaultConfig()
	cfg.Expand()

	func() {
		c, err := client.NewMPDClient(&cfg.MPD.Connection, &cfg.MPD.Address, &cfg.MPD.Password, &cfg.LyricsDir)
		check(err)
		defer func() {
			check(c.Close())
		}()

		attrs, err := c.Data()
		ret, errr := json.MarshalIndent(attrs, "", "  ")
		fmt.Printf("%s %v %v\n", ret, err, errr)
	}()

	func() {
		m, err := mpd.DialAuthenticated(cfg.MPD.Connection, cfg.MPD.Address, cfg.MPD.Password)
		check(err)
		defer func() {
			check(m.Close())
		}()

		mattrss, err := m.Command("listmounts").AttrsList("mount")
		ret, errr := json.MarshalIndent(mattrss, "", "  ")
		fmt.Printf("%s %v %v\n", ret, err, errr)

		mattrs, err := m.Command("config").Attrs()
		ret, errr = json.MarshalIndent(mattrs, "", "  ")
		fmt.Printf("%s %v %v\n", ret, err, errr)
	}()
}
