package payments

import (
	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/common/price"
	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/domain/orders"
	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/payments/interfaces/intraprocess"
)

type IntraprocessService struct {
	orders chan<- intraprocess.OrderToProcess
}

func NewIntraprocessService(ordersChannel chan<- intraprocess.OrderToProcess) IntraprocessService {
	return IntraprocessService{ordersChannel}
}

func (i IntraprocessService) InitializeOrderPayment(id orders.ID, price price.Price) error {
	i.orders <- intraprocess.OrderToProcess{ID: string(id), Price: price}
	return nil
}
