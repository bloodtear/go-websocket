package main

import ( // 导入对应的包
	"github.com/bloodtear/go-websocket/impl"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var ( // 初始化握手升级ws的处理
	upgrader = websocket.Upgrader {
		// 读取存储空间大小
		ReadBufferSize:1024,
		// 写入存储空间大小
		WriteBufferSize:1024,
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool { // 加上这个就是允许跨域了
			return true
		},
	}
)

// ws处理器
func wsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		wsConn *websocket.Conn // 定义ws连接，这是真实的连接
		err error // 错误
		// data []byte
		conn *impl.Connection // 走impl下的Connection的结构体，连接类
		data []byte // 数据
	)
	// 完成http应答，在httpheader中放下如下参数
	if wsConn, err = upgrader.Upgrade(w, r, nil);err != nil {
		return // 获取连接失败直接返回
	}

	if conn, err = impl.InitConnection(wsConn); err != nil { // 初始化conn连接
		goto ERR
	}

	go func() { // 拉起一个线程，因为有定时器，走并发
		var (
			err error
		)
		for {
			// 每隔一秒发送一次心跳
			if err = conn.WriteMessage([]byte("heartbeat")); err != nil { // 启动connect的发送消息
				return
			}
			time.Sleep(1 * time.Second)
		}

	}()

	for {
		if data, err = conn.ReadMessage(); err != nil { // 持续接收信息，信息为data，走conn的写入
			goto ERR
		}
		if err = conn.WriteMessage(data); err != nil { // 持续写信息, 把data传入， 走conn的写入
			goto ERR
		}
	}

ERR:
	// 关闭当前连接

}

func main()  {
	// 当有请求访问ws时，执行此回调方法
	http.HandleFunc("/ws",wsHandler)
	// 监听127.0.0.1:7777
	err := http.ListenAndServe("0.0.0.0:7777", nil)
	if err != nil {
		log.Fatal("ListenAndServe", err.Error())
	}
}