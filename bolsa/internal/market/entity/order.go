package entity

const (
	Sell = "Sell"
	Buy  = "Buy"
)

const (
	Open    = "Open"
	Closed  = "Closed"
	Pending = "Pending"
)

type Order struct {
	ID            string
	Investor      *Investor
	Asset         *Asset
	Shares        int
	PendingShares int
	Price         float64
	OrderType     string
	Status        string
	Transactions  []*Transaction
}

func NewOrder(id string, investor *Investor, asset *Asset, shares int, price float64, orderType string) *Order {
	return &Order{
		ID:            id,
		Investor:      investor,
		Asset:         asset,
		Shares:        shares,
		PendingShares: shares,
		Price:         price,
		OrderType:     orderType,
		Status:        Open,
		Transactions:  []*Transaction{},
	}
}

func (o *Order) CloseIfNoPendingShares() {
	if o.PendingShares == 0 {
		o.Status = Closed
	}
}

func (o *Order) AddPendingShares(shares int) {
	o.PendingShares += shares
}
