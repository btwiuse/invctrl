package webproc

import (
	"log"
	"os"

	"github.com/jpillora/opts"
	"github.com/jpillora/webproc/agent"
)

var version = "0.0.0-src"

func Run(args []string) error {
	os.Args = append([]string{"webproc"}, args...)
	//prepare config!
	c := agent.Config{}
	//parse cli
	opts.New(&c).Name("webproc").PkgRepo().Version(version).Parse()
	//if args contains has one non-executable file, treat as webproc file
	//TODO: allow cli to override config file
	args = c.ProgramArgs
	if len(args) == 1 {
		path := args[0]
		if info, err := os.Stat(path); err == nil && info.Mode()&0111 == 0 {
			c.ProgramArgs = nil
			if err := agent.LoadConfig(path, &c); err != nil {
				log.Printf("[webproc] load config error: %s", err)
				return err
			}
		}
	}
	//validate and apply defaults
	if err := agent.ValidateConfig(&c); err != nil {
		log.Printf("[webproc] load config error: %s", err)
		return err
	}
	//server listener
	if err := agent.Run(version, c); err != nil {
		log.Printf("[webproc] agent error: %s", err)
		return err
	}
	return nil
}
