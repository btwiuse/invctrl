package agent

import (
	"context"
	"sync"
	"io"
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/btwiuse/conntroll/pkg/uuid"
	"github.com/btwiuse/conntroll/pkg/api"
	rpcimpl "github.com/btwiuse/conntroll/pkg/api/rpc/impl"
	"github.com/btwiuse/conntroll/pkg/hub"
	"github.com/btwiuse/conntroll/pkg/hub/session"
	"github.com/btwiuse/conntroll/pkg/wrap"
	"google.golang.org/grpc"
)

var (
	_ hub.Agent = (*agent)(nil)
)

func NewAgent(conn net.Conn, opts ...AgentOpt) hub.Agent {
	ag := &agent{
		sch:            make(chan hub.Session),
		SessionManager: NewSessionManager(),
		RPCManager:     NewRPCManager(),
		created:        time.Now(),
		done:		make(chan struct{}),
		once:		&sync.Once{},
	}
	for _, opt := range opts {
		opt(ag)
	}
	ag.AddRPCConn(conn)
	for i := 0; i < 3; i ++ {
		ag.NewRPC()
	}
	return ag
}

type AgentOpt func(*agent)

func SetID(id string) AgentOpt {
	return func(ag *agent) {
		ag.Id = id
	}
}

func SetUsername(u string) AgentOpt {
	return func(ag *agent) {
		ag.Username = u
	}
}

func SetIP(ip string) AgentOpt {
	return func(ag *agent) {
		ag.IP = ip
	}
}

func SetPwd(p string) AgentOpt {
	return func(ag *agent) {
		ag.Pwd = p
	}
}

func SetUser(u string) AgentOpt {
	return func(ag *agent) {
		ag.User = u
		ag.Whoami = u
	}
}

func SetHostname(h string) AgentOpt {
	return func(ag *agent) {
		ag.Hostname = h
	}
}

func SetConnected(t int64) AgentOpt {
	return func(ag *agent) {
		ag.Connected = t
	}
}

func SetOS(o string) AgentOpt {
	return func(ag *agent) {
		ag.OS = o
	}
}

func SetARCH(a string) AgentOpt {
	return func(ag *agent) {
		ag.ARCH = a
	}
}

// make server call AddRPCConn
// this func is asynchronuous, we don't care the result
func (ag *agent) NewRPC() {
	req := rpcimpl.RPCRequest{}
	resp := &rpcimpl.RPCResponse{}
	rc := ag.RPCManager.Last()
	if rc == nil {
		log.Println("client dead:", ag.ID())
		ag.Close()
		return
	}
	err := rc.Call("RPC.New", req, resp)
	if err != nil {
		log.Println("RPC.New", err)
	}
}

func (ag *agent) NewSession() hub.Session {
	req := rpcimpl.SessionRequest{}
	resp := &rpcimpl.SessionResponse{}

	done := make(chan *rpc.Call, 1)
	rpcClient := ag.RPCManager.Last()

	rpcClient.Go("Session.New", req, resp, done)
	<-done

	return <-ag.sch
}

func (ag *agent) Close() {
	ag.once.Do(func(){
		close(ag.done)
	})
}

func (ag *agent) Done() <-chan struct{} {
	return ag.done
}

type agent struct {
	hub.SessionManager `json:"-"`
	RPCManager         hub.RPCManager   `json:"-"`
	sch                chan hub.Session `json:"-"`
	done chan struct{}
	once *sync.Once

	created time.Time

	// Metadata
	Id        string `json:"id"`
	Connected int64  `json:"connected"`
	Hostname  string `json:"hostname"`
	Whoami    string `json:"whoami"`
	User      string `json:"user"`     // compat
	Username  string `json:"username"` // compat
	Pwd       string `json:"pwd"`
	OS        string `json:"os"`
	ARCH      string `json:"arch"`
	IP        string `json:"ip"`
}

func (ag *agent) Time() time.Time {
	return ag.created
}

func (ag *agent) ID() string {
	return ag.Id
}

// we use NewRPCClient over rpc.NewClient(conn)
// so we can remove agent from pool immediately when it is disconnected

// when we have multiple RPC clients, things man change a bit
// we may assume the agent is always connected now
// we should probably spawn another RPCClient onClose, instead of simply letting it die
func (ag *agent) AddRPCConn(c net.Conn) {
	id := uuid.New()
	pr, pw := io.Pipe()
	go func() {
		defer ag.onRPCClose(id)
		if _, err := io.Copy(pw, c); err != nil {
			log.Println(err)
		}
	}()
	rpcClient := rpc.NewClient(&wrap.ReadWriteCloser{
		Reader: pr,
		Writer: c,
		Closer: c,
	})
	rc := ToRPC(id)(rpcClient)
	ag.RPCManager.AddRPC(rc)
}

// onclose is called when rpc connection is lost
func (ag *agent) onRPCClose(id string) {
	log.Println("disconnected:", ag.Id, id)
	ag.RPCManager.Del(id)
	// assuming there are other rpc conns left
	ag.NewRPC()
}

// blocks until agent.NewSession reads the channel
func (ag *agent) AddSessionConn(c net.Conn) {
	toGRPCClientConn := func(c net.Conn) *grpc.ClientConn {
		options := []grpc.DialOption{
			// grpc.WithTransportCredentials(creds),
			grpc.WithInsecure(),
			grpc.WithContextDialer(
				func(ctx context.Context, s string) (net.Conn, error) {
					return c, nil
				},
			),
		}

		// TODO: handle this
		cc, err := grpc.Dial("", options...)
		if err != nil {
			log.Fatal(err.Error())
		}
		return cc
	}
	ag.sch <- session.NewSession(api.NewSessionClient(toGRPCClientConn(c)))
}
