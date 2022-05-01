package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/pflag"

	"github.com/env25/mpdlrc/internal"
	"github.com/env25/mpdlrc/internal/config"
)

const PROGNAME = "mpdlrc"

func main() {
	exitCode := 0

	defer func() {
		os.Exit(exitCode)
	}()

	var (
		args        = os.Args[1:]
		cfg         = config.DefaultConfig()
		configFiles = config.ConfigFiles()
	)

	flags_cfg := pflag.NewFlagSet(PROGNAME, pflag.ContinueOnError)
	flags_cfg.SortFlags = false
	flags_cfg.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: true}

	flags_cfg.StringVar(&cfg.MusicDir, `musicdir`, cfg.MusicDir, `override cfg.MusicDir`)
	flags_cfg.StringVar(&cfg.LyricsDir, `lyricsdir`, cfg.LyricsDir, `override cfg.LyricsDir`)
	flags_cfg.StringVar(&cfg.MPD.Connection, `mpd-connection`, cfg.MPD.Connection, `override cfg.MPD.Connection ("unix" or "tcp")`)
	flags_cfg.StringVar(&cfg.MPD.Address, `mpd-address`, cfg.MPD.Address, `override cfg.MPD.Address ("socket" or "host:port")`)
	flags_cfg.StringVar(&cfg.MPD.Password, `mpd-password`, cfg.MPD.Password, `override cfg.MPD.Password`)

	var (
		flag_dumpcfg = false
		flag_usage   = false
	)

	flags := pflag.NewFlagSet(PROGNAME, pflag.ContinueOnError)
	flags.SortFlags = false

	flags.BoolVar(&flag_dumpcfg, `dump-config`, false, `dump final config`)
	flags.BoolVarP(&flag_usage, `help`, `h`, false, `display this help and exit`)
	flags.StringArrayVar(&configFiles, `config`, configFiles, `use config file`)

	flags_cfg.VisitAll(func(f *pflag.Flag) {
		flags.Var((*fakeValue)(f), f.Name, f.Usage)
	})

	if err := flags.Parse(args); err != nil {
		fmt.Println(err)
		exitCode = 1
		return
	}

	if flag_usage {
		fmt.Println("Usage of " + PROGNAME + ":")
		fmt.Print(flags.FlagUsages())
		exitCode = 0
		return
	}

	for _, fpath := range configFiles {
		f, err := os.Open(fpath)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintln(os.Stderr, "open config file:", err)
			}
			continue
		}
		if err := toml.NewDecoder(f).Decode(cfg); err != nil {
			fmt.Fprintln(os.Stderr, "decode config file:", err)
		}
		f.Close()
	}

	_ = flags_cfg.Parse(args)

	cfg.Expand()

	var logBuilder strings.Builder
	defer fmt.Fprint(os.Stderr, &logBuilder)

	if flag_dumpcfg {
		_ = toml.NewEncoder(&logBuilder).Encode(cfg)
		exitCode = 0
		return
	}

	if config.Debug {
		log.SetOutput(&logBuilder)
	} else {
		log.SetOutput(io.Discard)
	}

	log.SetFlags(0)

	if err := internal.NewApplication(cfg).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		exitCode = 1
		return
	}
}

type fakeValue pflag.Flag

func (*fakeValue) Set(string) error { return nil }
func (v *fakeValue) Type() string   { return v.Value.Type() }
func (v *fakeValue) String() string { return v.DefValue }
