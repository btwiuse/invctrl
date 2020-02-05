// +build freebsd openbsd netbsd

package tun

import (
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
)

const (
	IFF_TUN   = 0x0001
	IFF_TAP   = 0x0002
	IFF_NO_PI = 0x1000
)

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

func OpenTunDevice(name, addr, gw, mask string, dns []string) (io.ReadWriteCloser, error) {
	file, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	var req ifReq
	copy(req.Name[:], name)
	req.Flags = IFF_TUN | IFF_NO_PI
	//log.Printf("openning tun device")
	/*
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&req)))
		if errno != 0 {
			err = errno
			return nil, err
		}
	*/

	// config address
	//log.Printf("configuring tun device address")
	cmd := exec.Command("ip", "link", "set", name, "up")
	err = cmd.Run()
	if err != nil {
		file.Close()
		//log.Printf("failed to configure tun device address")
		return nil, err
	}
	cmd = exec.Command("ip", "link", "set", name, "mtu", "1500")
	err = cmd.Run()
	if err != nil {
		file.Close()
		//log.Printf("failed to configure tun device address")
		return nil, err
	}
	cmd = exec.Command("ip", "addr", "add", addr+"/"+mask, "dev", name)
	err = cmd.Run()
	if err != nil {
		file.Close()
		//log.Printf("failed to configure tun device address")
		return nil, err
	}
	// syscall.SetNonblock(int(file.Fd()), false)
	return &tunDev{
		f:      file,
		addr:   addr,
		addrIP: net.ParseIP(addr).To4(),
		gw:     gw,
		gwIP:   net.ParseIP(gw).To4(),
	}, nil
}

func NewTunDev(fd uintptr, name string, addr string, gw string) io.ReadWriteCloser {
	// syscall.SetNonblock(int(fd), false)
	return &tunDev{
		f:      os.NewFile(fd, name),
		addr:   addr,
		addrIP: net.ParseIP(addr).To4(),
		gw:     gw,
		gwIP:   net.ParseIP(gw).To4(),
	}
}

type tunDev struct {
	name   string
	addr   string
	addrIP net.IP
	gw     string
	gwIP   net.IP
	marker []byte
	f      *os.File
}

func (dev *tunDev) Read(data []byte) (int, error) {
	n, e := dev.f.Read(data)
	if e == nil && isStopMarker(data[:n], dev.addrIP, dev.gwIP) {
		return 0, errors.New("received stop marker")
	}
	return n, e
}

func (dev *tunDev) Write(data []byte) (int, error) {
	return dev.f.Write(data)
}

func (dev *tunDev) Close() error {
	//log.Printf("send stop marker")
	sendStopMarker(dev.addr, dev.gw)
	return dev.f.Close()
}
