/**
 * Copyright 2014 @ z3q.net.
 * name :
 * author : jarryliu
 * date : 2014-01-08 21:35
 * description :
 * history :
 */

package daemon

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"go2o/core"
	"go2o/core/service/rsi"
	"go2o/core/variable"
	"log"
	"strconv"
	"strings"
	"time"
)

// 监视新订单
func superviseOrder(ss []Service) {
	sv := rsi.ShoppingService
	notify := func(id int, ss []Service) {
		o, _ := sv.GetSubOrder(int32(id))
		if o != nil {
			for _, v := range ss {
				if !v.OrderObs(o) {
					break
				}
			}
		}
	}
	// 监听队列
	id := 0
	conn := core.GetRedisConn()
	defer conn.Close()
	for {
		arr, err := redis.ByteSlices(conn.Do("BLPOP",
			variable.KvOrderBusinessQueue, 0)) //取出队列的一个元素
		if err == nil {
			//通知订单更新
			id, err = strconv.Atoi(string(arr[1]))
			if err == nil {
				go notify(id, ss)
			}
		} else {
			appCtx.Log().Println("[ Daemon][ OrderQueue][ Error]:",
				err.Error(), "; retry after 10 seconds.")
			time.Sleep(time.Second * 10)
		}

	}
}

// 监视新会员
func superviseMemberUpdate(ss []Service) {
	sv := rsi.MemberService
	notify := func(id int32, action string, ss []Service) {
		m, _ := sv.GetMember(id)
		if m != nil {
			for _, v := range ss {
				if !v.MemberObs(m, action == "create") {
					break
				}
			}
		}
	}
	id := 0
	conn := core.GetRedisConn()
	defer conn.Close()
	for {
		arr, err := redis.ByteSlices(conn.Do("BLPOP",
			variable.KvMemberUpdateQueue, 0))
		if err == nil {
			//通知会员修改,格式如: 1-[create|update]
			s := string(arr[1])
			mArr := strings.Split(s, "-")
			id, err = strconv.Atoi(mArr[0])
			if err == nil {
				go notify(int32(id), mArr[1], ss)
			}
		} else {
			appCtx.Log().Println("[ Daemon][ MemberQueue][ Error]:",
				err.Error(), "; retry after 10 seconds.")
			time.Sleep(time.Second * 10)
		}
	}
}

// 监视支付单完成
func supervisePaymentOrderFinish(ss []Service) {
	sv := rsi.PaymentService
	notify := func(id int, ss []Service) {
		order, _ := sv.GetPaymentOrderById(int32(id))
		if order != nil {
			for _, v := range ss {
				if !v.PaymentOrderObs(order) {
					break
				}
			}
		}
	}
	id := 0
	conn := core.GetRedisConn()
	defer conn.Close()
	for {
		arr, err := redis.ByteSlices(conn.Do("BLPOP",
			variable.KvPaymentOrderFinishQueue, 0))
		if err == nil {
			//通知服务
			s := string(arr[1])
			id, err = strconv.Atoi(s)
			if err == nil {
				go notify(id, ss)
			}
		} else {
			appCtx.Log().Println("[ Daemon][ PaymentOrderQueue][ Error]:",
				err.Error(), "; retry after 10 seconds.")
			time.Sleep(time.Second * 10)
		}
	}
}

// 检测已过期的订单并标记
func detectOrderExpires() {
	if appCtx.Debug() {
		log.Println("[ Order]: detect order time out...")
	}
	conn := core.GetRedisConn()
	defer conn.Close()
	t := time.Now()
	key := fmt.Sprintf("%s:%s:*", variable.KvOrderExpiresTime, getTick(t))
	//key = "go2o:order:timeout:11-0-2:*"
	//获取标记为等待过期的订单
	orderId := 0
	ss := rsi.ShoppingService
	list, err := redis.Strings(conn.Do("KEYS", key))
	if err == nil {
		for _, oKey := range list {
			//订单号
			i := strings.LastIndex(oKey, ":")
			orderId, err = strconv.Atoi(oKey[i+1:])
			if err == nil && orderId > 0 {
				err = ss.CancelOrder(int32(orderId), "订单超时,自动取消")
				//清除待取消记录
				conn.Do("DEL", oKey)
				//log.Println("---",orderId,"---",unix, "--", time.Now().Unix(), v, err)
			}
		}
	} else {
		log.Println("[ Daemon][ Order][ Cancel][ Error]:",
			err.Error(), "; retry after 10 seconds.")
		time.Sleep(time.Second * 10)
	}
}

// 订单自动收货
func orderAutoRecive() {
	if appCtx.Debug() {
		log.Println("[ Order]: order auto receive ...")
	}
	conn := core.GetRedisConn()
	defer conn.Close()
	t := time.Now()
	key := fmt.Sprintf("%s:%s:*", variable.KvOrderAutoReceive, getTick(t))
	//key = "go2o:order:autoreceive:11-0-2:*"
	//获取标记为自动收货的订单
	orderId := 0
	ss := rsi.ShoppingService
	list, err := redis.Strings(conn.Do("KEYS", key))
	if err == nil {
		for _, oKey := range list {
			//订单号
			i := strings.LastIndex(oKey, ":")
			orderId, err = strconv.Atoi(oKey[i+1:])
			if err == nil && orderId > 0 {
				err = ss.BuyerReceived(int32(orderId))
			}
		}
	} else {
		log.Println("[ Daemon][ Order][ Receive][ Error]:",
			err.Error(), "; retry after 10 seconds.")
		time.Sleep(time.Second * 10)
	}
}
