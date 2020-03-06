package impl

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

type Connection struct {
	// 存放websocket连接
	wsConn *websocket.Conn
	// 用于存放数据
	inChan chan []byte // 进入的管道
	// 用于读取数据
	outChan chan []byte // 读取的管道
	closeChan chan byte // 关闭管道
	mutex sync.Mutex // ？？同步？
	// chan是否被关闭
	isClosed bool
}

// 读取Api 实现接收信息的处理
func (conn *Connection) ReadMessage() (data []byte, err error) { // 读取信息，传入连接，返回数据和错误
	// select是Go中的一个控制结构，类似于用于通信的switch语句。
	// 每个case必须是一个通信操作，要么是发送要么是接收。
	// select随机执行一个可运行的case。
	// 如果没有case可运行，它将阻塞，直到有case可运行。
	// 一个默认的子句应该总是可运行的。
	select {
	case data = <- conn.inChan: // 把输入管道的数据传到data里
	case <- conn.closeChan: // 如果关闭管道则不传入
		err = errors.New("connection is closed") // 返回一个err
	}
	return
}

// 发送Api
func (conn *Connection) WriteMessage(data []byte) (err error)  { // 传入一个连接，返回写入数据结果
	select {
	case conn.outChan <- data: // 把数据传到暑促胡管道
	case <- conn.closeChan: // 管道如果关闭，就返回错误
		err = errors.New("connection is closed")
	}
	return
}

// 关闭连接的Api
func (conn *Connection) Close()  {
	// 线程安全的Close，可以并发多次调用也叫做可重入的Close
	conn.wsConn.Close() // 连接关闭
	conn.mutex.Lock() // 启动互斥锁
	if !conn.isClosed { // 如果连接没关闭
		// 关闭chan,但是chan只能关闭一次
		close(conn.closeChan) // 关闭这个连接
		conn.isClosed = true // 重置标志位为true
	}
	conn.mutex.Unlock() // 解除互斥锁
}

// 初始化长连接
// 传入一个ws连接，返回一个connect连接结构体
func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error)  {
	conn = &Connection{
		wsConn: wsConn, // 初始化连接
		inChan: make(chan []byte, 1000), // 初始化读取管道
		outChan: make(chan []byte, 1000), // 初始化输出管道
		closeChan: make(chan byte, 1), // 初始化关闭管道
	}

	// 启动读协程
	go conn.readLoop()

	// 启动写协程
	go conn.writeLoop()

	return
}

// 内部实现 读取信息
func (conn *Connection) readLoop()  {
	var (
		data []byte
		err error
	)
	for { //
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			goto ERR
		}
		// 容易阻塞到这里，等待inChan有空闲的位置
		select {
		case conn.inChan <- data:
		case <- conn.closeChan: // closeChan关闭的时候执行
			goto ERR
		}
	}

ERR:
	conn.Close()
}

func (conn *Connection) writeLoop()  {
	var (
		data []byte
		err error
	)
	for {
		select {
		case data = <- conn.outChan:
		case <- conn.closeChan:
			goto ERR
		}
		data = <- conn.outChan
		if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()
}