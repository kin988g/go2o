/**
 * Copyright 2015 @ S1N1 Team.
 * name : kv_manager
 * author : jarryliu
 * date : 2015-07-26 22:44
 * description :
 * history :
 */
package partner

import (
	"go2o/src/core/domain/interface/partner"
	"strconv"
)

var _ partner.IKvManager = new(KvManager)

type KvManager struct {
	_partner   *Partner
	_partnerId int
}

func newKvManager(p *Partner) partner.IKvManager {
	return &KvManager{
		_partner:   p,
		_partnerId: p.GetAggregateRootId(),
	}
}

// 获取键值
func (this *KvManager) Get(k string) string {
	return this._partner._rep.GetKeyValue(this._partnerId, k)
}

// 获取int类型的键值
func (this *KvManager) GetInt(k string) int {
	i, _ := strconv.Atoi(this.Get(k))
	return i
}

// 设置
func (this *KvManager) Set(k, v string) {
	this._partner._rep.SaveKeyValue(this._partnerId, k, v)
}

// 获取多项
func (this *KvManager) Gets(k []string) map[string]string {
	return this._partner._rep.GetKeyMap(this._partnerId, k)
}

// 设置多项
func (this *KvManager) Sets(v map[string]string) {
	for k, v := range v {
		this.Set(k, v)
	}
}
