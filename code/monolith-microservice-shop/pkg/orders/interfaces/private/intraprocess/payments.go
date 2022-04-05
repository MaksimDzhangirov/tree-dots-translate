package intraprocess

import (
	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/application"
	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/domain/orders"
)

type OrdersInterface struct {
	service application.OrdersService
}

func NewOrdersInterface(service application.OrdersService) OrdersInterface {
	return OrdersInterface{service}
}

func (p OrdersInterface) MarkOrderAsPaid(orderID string) error {
	return p.service.MarkOrderAsPaid(application.MarkOrderAsPaidCommand{OrderID: orders.ID(orderID)})
}
