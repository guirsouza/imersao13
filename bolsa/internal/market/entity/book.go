package entity

import (
	"container/heap"
	"math"
	"sync"
)

type Book struct {
	Order            []*Order
	Transactions     []*Transaction
	OrdersChannel    chan *Order // input
	OrdersChannelOut chan *Order // output
	WaitGroup        *sync.WaitGroup
}

func NewBook(orderChannel chan *Order, orderChannelOut chan *Order, waitGroup *sync.WaitGroup) *Book {
	return &Book{
		Order:            []*Order{},
		Transactions:     []*Transaction{},
		OrdersChannel:    orderChannel,
		OrdersChannelOut: orderChannelOut,
		WaitGroup:        waitGroup,
	}
}

func (b *Book) Trade() {
	buyOrders := make(map[string]*OrderQueue)
	sellOrders := make(map[string]*OrderQueue)
	// buyOrders := NewOrderQueue()
	// sellOrders := NewOrderQueue()

	// heap.Init(buyOrders)
	// heap.Init(sellOrders)

	for order := range b.OrdersChannel {
		asset := order.Asset.ID

		if buyOrders[asset] == nil {
			buyOrders[asset] = NewOrderQueue()
			heap.Init(buyOrders[asset])
		}

		if sellOrders[asset] == nil {
			sellOrders[asset] = NewOrderQueue()
			heap.Init(sellOrders[asset])
		}

		if order.OrderType == Buy {
			buyOrders[asset].Push(order)
			if sellOrders[asset].Len() > 0 && sellOrders[asset].Orders[0].Price <= order.Price {
				sellOrder := sellOrders[asset].Pop().(*Order)
				if sellOrder.PendingShares > 0 {
					transaction := NewTransaction(sellOrder, order, order.Shares, sellOrder.Price)
					b.AddTransaction(transaction, b.WaitGroup)
					sellOrder.Transactions = append(order.Transactions, transaction)
					order.Transactions = append(order.Transactions, transaction)
					b.OrdersChannelOut <- sellOrder
					b.OrdersChannelOut <- order
					if sellOrder.PendingShares > 0 {
						sellOrders[asset].Push(sellOrder)
					}
				}
			}
		} else if order.OrderType == Sell {
			sellOrders[asset].Push(order)
			if buyOrders[asset].Len() > 0 && buyOrders[asset].Orders[0].Price >= order.Price {
				buyOrder := buyOrders[asset].Pop().(*Order)
				if buyOrder.PendingShares > 0 {
					transaction := NewTransaction(order, buyOrder, order.Shares, buyOrder.Price)
					b.AddTransaction(transaction, b.WaitGroup)
					buyOrder.Transactions = append(order.Transactions, transaction)
					order.Transactions = append(order.Transactions, transaction)
					b.OrdersChannelOut <- buyOrder
					b.OrdersChannelOut <- order
					if buyOrder.PendingShares > 0 {
						buyOrders[asset].Push(buyOrder)
					}
				}
			}
		}
	}
}

func (b *Book) AddTransaction(transaction *Transaction, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	sellingShares := transaction.SellingOrder.PendingShares
	buyingShares := transaction.BuyingOrder.PendingShares
	minShares := int(math.Min(float64(sellingShares), float64(buyingShares)))

	transaction.SellingOrder.Investor.UpdateAssetPosition(transaction.SellingOrder.Asset.ID, -minShares)
	transaction.SellingOrder.AddPendingShares(-minShares)
	transaction.BuyingOrder.Investor.UpdateAssetPosition(transaction.BuyingOrder.Asset.ID, minShares)
	transaction.BuyingOrder.AddPendingShares(-minShares)

	transaction.CalculateTotal(transaction.Shares, transaction.BuyingOrder.Price)

	transaction.BuyingOrder.CloseIfNoPendingShares()
	transaction.SellingOrder.CloseIfNoPendingShares()

	b.Transactions = append(b.Transactions, transaction)

}
