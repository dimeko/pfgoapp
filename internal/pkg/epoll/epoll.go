package epoll

import (
	"log"
	"reflect"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
	"golang.org/x/sys/unix"
)

type Epollt struct {
	Fd          int
	Connections map[int]*websocket.Conn
	Lock        *sync.RWMutex
}

func MkEpoll() (*Epollt, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &Epollt{
		Fd:          fd,
		Lock:        &sync.RWMutex{},
		Connections: make(map[int]*websocket.Conn),
	}, nil
}

func (e *Epollt) Add(conn *websocket.Conn) error {
	fd := websocketFD(conn)
	err := unix.EpollCtl(e.Fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}
	e.Lock.Lock()
	defer e.Lock.Unlock()
	e.Connections[fd] = conn
	if len(e.Connections)%100 == 0 {
		log.Printf("Total number of connections: %v", len(e.Connections))
	}
	return nil
}

func (e *Epollt) Remove(conn *websocket.Conn) error {
	fd := websocketFD(conn)
	err := unix.EpollCtl(e.Fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	e.Lock.Lock()
	defer e.Lock.Unlock()
	delete(e.Connections, fd)
	if len(e.Connections)%100 == 0 {
		log.Printf("Total number of connections: %v", len(e.Connections))
	}
	return nil
}

func (e *Epollt) Wait() ([]*websocket.Conn, error) {
	events := make([]unix.EpollEvent, 1000)
	n, err := unix.EpollWait(e.Fd, events, 1000)
	if err != nil {
		return nil, err
	}
	e.Lock.RLock()
	defer e.Lock.RUnlock()
	var connections []*websocket.Conn
	for i := 0; i < n; i++ {
		conn := e.Connections[int(events[i].Fd)]
		connections = append(connections, conn)
	}
	return connections, nil
}

func websocketFD(conn *websocket.Conn) int {
	connVal := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn").Elem()
	tcpConn := reflect.Indirect(connVal).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
