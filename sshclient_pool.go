package dssh

import (
	//"fmt"
	"golang.org/x/crypto/ssh"
	"sync"
	"time"
)

type Client struct {
	ip        string
	sshclient *ssh.Client
	lock      *sync.Mutex
}

func (self *Client) AllocSession(conn_timeout time.Duration) (*ssh.Session, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if self.sshclient != nil {
		return self.sshclient.NewSession()
	}

	var sshclient, err = conn(self.ip, conn_timeout)
	if err != nil {
		return nil, err
	}
	self.sshclient = sshclient

	return self.sshclient.NewSession()
}

type SSHClientPool struct {
	map_ip_client map[string]*Client
	lock          *sync.Mutex
}

func new_ssh_client_pool() *SSHClientPool {
	return &SSHClientPool{
		map_ip_client: map[string]*Client{},
		lock:          &sync.Mutex{},
	}
}

/*
	如果已经存在连接, 获取一个session
	如果不存在, 新加连接并保存, 然后获取一个session

	如果多个routine同时获取相同ip的连接, 连接又恰好未建立, 则某个获得锁的routine会尝试建立连接, 如果连接建立成功, 其他routine直接在该连接上创建session.
	如果连接建立失败, 其他routine拿到锁后会再次尝试建立连接
*/
func (self *SSHClientPool) AllocSession(ip string, conn_timeout time.Duration) (*ssh.Session, error) {
	self.lock.Lock()

	var client = self.map_ip_client[ip]
	if client == nil {
		client = &Client{
			ip:   ip,
			lock: &sync.Mutex{},
		}
		self.map_ip_client[ip] = client
	}
	self.lock.Unlock()
	return client.AllocSession(conn_timeout)
}
