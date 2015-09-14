package infinity

import (
	"log"
	"math/rand"
	"time"
)

//////////////////////////////////////////////////////////////////////////
///////THIS IS FAKE API FOR TEST//////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////

type FakeInfinity struct {
	orders []Order
}

func send_states(order_id int64, inf *FakeInfinity) {
	log.Printf("FA will send fake states for order %v", order_id)
	for _, i := range []int{2, 3, 4, 5, 6, 7} {
		log.Println("FA send state: ", i)
		inf.set_order_state(order_id, i)
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

	}
}

func (inf *FakeInfinity) set_order_state(order_id int64, new_state int) {
	for i, order := range inf.orders {
		if order.ID == order_id {
			inf.orders[i].State = new_state
		}
	}
}

func (inf *FakeInfinity) NewOrder(order NewOrder) (ans Answer, e error) {
	saved_order := Order{
		ID:    int64(len(inf.orders) + 1),
		State: 1,
		Cost:  150,
		IDCar: 5033615557,
	}

	inf.orders = append(inf.orders, saved_order)

	ans = Answer{
		IsSuccess: true,
		Message:   "test order was formed",
	}
	ans.Content.Id = saved_order.ID
	log.Println("FA now i have orders: ", len(inf.orders))

	go send_states(saved_order.ID, inf)

	return
}

func (inf *FakeInfinity) Orders() []Order {
	return inf.orders
}

func (inf *FakeInfinity) CancelOrder(order_id int64) (bool, string) {
	log.Println("FA order was canceled", order_id)
	for i, order := range inf.orders {
		if order.ID == order_id {
			inf.orders[i].State = 7
			return true, "test order was cancelled"
		}
	}
	return true, "Test order not found :( "
}

func (p *FakeInfinity) CalcOrderCost(order NewOrder) (int, string) {
	log.Println("FA calulate cost for order: ", order)
	return 100500, "Good cost!"
}

func (p *FakeInfinity) Feedback(f Feedback) (bool, string) {
	return true, "Test feedback was received! Thanks!"
}