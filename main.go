package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unicode"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	Configuration string `short:"c" long:"configuration" description:"the configuration file"`
	Daemon        bool   `short:"d" long:"daemon" description:"run as daemon"`
	EnvFile       string `long:"env-file" description:"the environment file"`
}

func init() {
	log.SetOutput(os.Stdout)
	if runtime.GOOS == "windows" {
		log.SetFormatter(&log.TextFormatter{DisableColors: true, FullTimestamp: true})
	} else {
		log.SetFormatter(&log.TextFormatter{DisableColors: false, FullTimestamp: true})
	}
	log.SetLevel(log.DebugLevel)
}

func initSignals(s *Supervisor) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.WithFields(log.Fields{"reason": "Signal received", "signal": sig}).Warn("Daemon is SHUTTING DOWN")
		s.procMgr.StopAllProcesses()
		os.Exit(-1)
	}()

}

var options Options
var parser = flags.NewParser(&options, flags.Default & ^flags.PrintErrors)

func LoadEnvFile() {
	if len(options.EnvFile) <= 0 {
		return
	}
	//try to open the environment file
	f, err := os.Open(options.EnvFile)
	if err != nil {
		log.WithFields(log.Fields{"file": options.EnvFile}).Error("Could not open the specified ENV file")
		return
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for {
		//for each line
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		//if line starts with '#', it is a comment line, ignore it
		line = strings.TrimSpace(line)
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		//if environment variable is exported with "export"
		if strings.HasPrefix(line, "export") && len(line) > len("export") && unicode.IsSpace(rune(line[len("export")])) {
			line = strings.TrimSpace(line[len("export"):])
		}
		//split the environment variable with "="
		pos := strings.Index(line, "=")
		if pos != -1 {
			k := strings.TrimSpace(line[0:pos])
			v := strings.TrimSpace(line[pos+1:])
			//if key and value are not empty, put it into the environment
			if len(k) > 0 && len(v) > 0 {
				os.Setenv(k, v)
			}
		}
	}
}

// find the supervisord.conf in following order:
//
// 1. $CWD/supervisord.conf
// 2. $CWD/etc/supervisord.conf
// 3. /etc/supervisord.conf
// 4. /etc/supervisor/supervisord.conf (since Supervisor 3.3.0)
// 5. ../etc/supervisord.conf (Relative to the executable)
// 6. ../supervisord.conf (Relative to the executable)
func findSupervisordConf() (string, error) {
	possibleSupervisordConf := []string{options.Configuration,
		"./supervisord.conf",
		"./etc/supervisord.conf",
		"/etc/supervisord.conf",
		"/etc/supervisor/supervisord.conf",
		"../etc/supervisord.conf",
		"../supervisord.conf"}

	for _, file := range possibleSupervisordConf {
		if _, err := os.Stat(file); err == nil {
			abs_file, err := filepath.Abs(file)
			if err == nil {
				return abs_file, nil
			} else {
				return file, nil
			}
		}
	}

	return "", fmt.Errorf("fail to find supervisord.conf")
}

func RunServer() {
	// infinite loop for handling Restart ('reload' command)
	LoadEnvFile()
	for true {
		options.Configuration, _ = findSupervisordConf()
		s := NewSupervisor(options.Configuration)
		initSignals(s)
		if sErr, _, _, _ := s.Reload(); sErr != nil {
			panic(sErr)
		}
		s.WaitForExit()
	}
}

func main() {
	ReapZombie()

	if _, err := parser.Parse(); err != nil {
		flagsErr, ok := err.(*flags.Error)
		if ok {
			switch flagsErr.Type {
			case flags.ErrHelp:
				fmt.Fprintln(os.Stdout, err)
				os.Exit(0)
			case flags.ErrCommandRequired:
				if options.Daemon {
					Deamonize(RunServer)
				} else {
					RunServer()
				}
			default:
				fmt.Fprintf(os.Stderr, "Invalid arguments: %s\n", err)
				os.Exit(1)
			}
		}
	}
}
