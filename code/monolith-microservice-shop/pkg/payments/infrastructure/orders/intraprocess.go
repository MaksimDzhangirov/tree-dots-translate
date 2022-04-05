package orders

import "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/interfaces/private/intraprocess"

type IntraprocessService struct {
	paymentsInterface intraprocess.OrdersInterface
}

func NewIntraprocessService(paymentsInterface intraprocess.OrdersInterface) IntraprocessService {
	return IntraprocessService{paymentsInterface: paymentsInterface}
}

func (o IntraprocessService) MarkOrderAsPaid(orderID string) error {
	return o.paymentsInterface.MarkOrderAsPaid(orderID)
}
