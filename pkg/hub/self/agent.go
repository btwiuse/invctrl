package self

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"k0s.io/pkg/agent/agent"
	"k0s.io/pkg/agent/config"
	"k0s.io/pkg/agent/info"
	"k0s.io/pkg/uuid"
	"k0s.io/pkg/version"
)

func Agent(hub string) {
	c := &config.Config{
		ID:  uuid.New(),
		Hub: hub,
		/*
			Htpasswd: map[string]string{
				"aaron": "$2a$10$WbZm/thAZI/f/QrcJn6V4OS.I61V2cLnOV.z7uXxtjHY8tZkTacLm",
			},
		*/
		Tags: []string{
			"os.Args = " + strings.Join(os.Args, " "),
			// "os.Env = " + strings.Join(os.Environ(), ":"),
		},
		Version: version.GetVersion(),
		Name:    "self",
		Info:    info.CollectInfo(),
	}

	// println(c.GetHost())
	config.SetURI()(c)
	// println(c.GetHost())

	ag := agent.NewAgent(c)

	for {
		time.Sleep(time.Second)

		log.Println(fmt.Sprintf("connecting self agent to %s", hub))
		// _ = agent.Run([]string{"-name", "embedded-client", "-c", "/dev/null", "-tags", "embedded-client"})

		log.Println(ag.ConnectAndServe())
	}
}
