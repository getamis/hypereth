// Copyright 2018 AMIS Technologies
// This file is part of the hypereth library.
//
// The hypereth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The hypereth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the hypereth library. If not, see <http://www.gnu.org/licenses/>.

package multiclient

import (
	"sync"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/getamis/sirius/log"
)

type Map struct {
	// the mapping form url to client
	clientMap map[string]*client
	// the mapping subscription id to url
	idMap              map[uint64]string
	subscripionCounter uint64
	newClientCh        chan<- string

	lock sync.RWMutex
}

type client struct {
	*rpc.Client
	Id uint64
}

func NewMap(newClientCh chan<- string) *Map {
	return &Map{
		clientMap:          make(map[string]*client),
		idMap:              make(map[uint64]string),
		subscripionCounter: 0,
		newClientCh:        newClientCh,
	}
}

func (m *Map) Delete(key string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	c, ok := m.clientMap[key]
	if !ok {
		return
	}
	if c.Client != nil {
		c.Client.Close()
	}
	delete(m.idMap, c.Id)
	delete(m.clientMap, key)
	log.Trace("Eth client removed", "c.Id", c.Id, "url", key)
}

func (m *Map) Add(key string, value *rpc.Client) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.subscripionCounter++
	m.clientMap[key] = &client{
		Id:     m.subscripionCounter,
		Client: value,
	}
	m.idMap[m.subscripionCounter] = key

	if m.newClientCh != nil {
		select {
		case m.newClientCh <- key:
		default:
		}
	}
	log.Trace("Eth client added", "id", m.subscripionCounter, "url", key)
}

func (m *Map) Replace(key string, value *rpc.Client) uint64 {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.clientMap[key]; ok {
		m.clientMap[key].Client = value
	}

	return m.clientMap[key].Id
}

func (m *Map) Get(key string) *rpc.Client {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.clientMap[key].Client
}

func (m *Map) GetById(id uint64) (*rpc.Client, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	key, ok := m.idMap[id]
	if !ok {
		return nil, false
	}
	c, ok := m.clientMap[key]
	if !ok {
		return nil, false
	}
	return c.Client, true
}

func (m *Map) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return len(m.clientMap)
}

// List returns a deep copy of client list
func (m *Map) List() []*rpc.Client {
	m.lock.RLock()
	defer m.lock.RUnlock()

	l := []*rpc.Client{}
	for _, v := range m.clientMap {
		if v.Client != nil {
			l = append(l, v.Client)
		}
	}
	return l
}

// Map returns a deep copy of client map
func (m *Map) Map() map[string]*rpc.Client {
	m.lock.RLock()
	defer m.lock.RUnlock()

	newMap := map[string]*rpc.Client{}
	for k, v := range m.clientMap {
		if v.Client != nil {
			newMap[k] = v.Client
		}
	}
	return newMap
}

func (m *Map) Ids() []uint64 {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ids := make([]uint64, len(m.idMap))
	i := 0
	for k := range m.idMap {
		ids[i] = k
		i++
	}
	return ids
}

func (m *Map) Keys() []string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	urls := make([]string, len(m.clientMap))
	index := 0
	for k := range m.clientMap {
		urls[index] = k
		index++
	}
	return urls
}

func (m *Map) NilClients() []string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	urls := make([]string, 0)
	for k, v := range m.clientMap {
		if v.Client == nil {
			urls = append(urls, k)
		}
	}
	return urls
}
