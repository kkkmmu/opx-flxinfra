//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//	 Unless required by applicable law or agreed to in writing, software
//	 distributed under the License is distributed on an "AS IS" BASIS,
//	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	 See the License for the specific language governing permissions and
//	 limitations under the License.
//
//	_______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

package publisher

import (
	"github.com/op/go-nanomsg"
	"syscall"
	"utils/logging"
)

const (
	NOTIFICATION_BUFFER_SIZE    = 100
	PUB_SOCKET_SEND_BUFFER_SIZE = 1024 * 1024
)

type PublisherInfo struct {
	Logger     logging.LoggerIntf
	PubChan    chan []byte
	Socket     *nanomsg.PubSocket
	SocketAddr string
}

func (p *PublisherInfo) CreateAndBindPubSock(socketAddr string, sockBufSize int64) *nanomsg.PubSocket {
	pubSock, err := nanomsg.NewPubSocket()
	if err != nil {
		p.Logger.Err("Failed to open publisher socket")
	}
	_, err = pubSock.Bind(socketAddr)
	if err != nil {
		p.Logger.Err("Failed to bind publisher socket")
	}
	err = pubSock.SetSendBuffer(sockBufSize)
	if err != nil {
		p.Logger.Err("Failed to set send buffer size for pub socket")
	}
	return pubSock
}

func (p *PublisherInfo) PublishEvents() {
	for {
		var msg []byte
		//Drain notification channels and publish event
		select {
		case msg = <-p.PubChan:
			_, rv := p.Socket.Send(msg, nanomsg.DontWait)
			if rv == syscall.EAGAIN {
				p.Logger.Err("Failed to publish event to all clients")
			}
		}
	}
}

func (p *PublisherInfo) DeinitPublisher() {
	//Close nanomsg sockets
	if err := p.Socket.Close(); err != nil {
		p.Logger.Err("Failed to close nano msg publisher socket:", p.SocketAddr)
	}
	return
}

func NewPublisher(logger logging.LoggerIntf, socketAddr string) *PublisherInfo {
	pub := new(PublisherInfo)
	pub.Logger = logger
	pub.SocketAddr = socketAddr
	pub.PubChan = make(chan []byte, NOTIFICATION_BUFFER_SIZE)
	pub.Socket = pub.CreateAndBindPubSock(socketAddr, PUB_SOCKET_SEND_BUFFER_SIZE)
	go pub.PublishEvents()
	return pub
}
