package wetty

import (
	"log"
	"net"
	"net/http"

	"k0s.io/third_party/pkg/localcmd"
)

// Server provides a wetty HTTP endpoint.
type Server struct {
	factory *localcmd.Factory
}

// New creates a new instance of Server.
// Server will use the New() of the factory provided to handle each request.
func New(args []string) *Server {
	return &Server{
		factory: &localcmd.Factory{
			Args: args,
		},
	}
}

// Run starts the main process of the Server.
// The cancelation of ctx will shutdown the server immediately with aborting
// existing connections. Use WithGracefullContext() to support graceful shutdown.
func (server *Server) Run() error {
	scheme, host, port, path := "http", "0.0.0.0", "45080", "/"
	hostPort := host + ":" + port
	url := scheme + "://" + hostPort + path

	srv := &http.Server{Handler: server.setupHandlers(path)}

	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		return err // ors.Wrapf(err, "failed to listen at `%s`", hostPort)
	}

	log.Printf("HTTP server is listening at: %s", url)
	return srv.Serve(listener)
}
