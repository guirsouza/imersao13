package entity

type OrderQueue struct {
	Orders []*Order
}

// Less -> comprar valores i e j
func (oq *OrderQueue) Less(i, j int) bool {
	return oq.Orders[i].Price < oq.Orders[j].Price
}

// Swap -> inverter i <-> j
func (oq *OrderQueue) Swap(i, j int) {
	oq.Orders[i], oq.Orders[j] = oq.Orders[j], oq.Orders[i]
}

// Len -> tamanho dos dados
func (oq *OrderQueue) Len() int {
	return len(oq.Orders)
}

// push -> adiciona novos objetos
func (oq *OrderQueue) Push(x interface{}) {
	oq.Orders = append(oq.Orders, x.(*Order))
}

// pop -> remove de uma posicao
func (oq *OrderQueue) Pop() interface{} {
	old := oq.Orders
	n := len(old)
	item := old[n-1]
	oq.Orders = old[0 : n-1] // sublist 0 até n-1
	return item
}

func NewOrderQueue() *OrderQueue {
	return &OrderQueue{}
}
