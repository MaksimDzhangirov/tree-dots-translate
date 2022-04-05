package application

import (
	"log"
	"time"

	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/common/price"
)

type orderService interface {
	MarkOrderAsPaid(orderID string) error
}

type PaymentsService struct {
	orderService orderService
}

func NewPaymentsService(orderService orderService) PaymentsService {
	return PaymentsService{orderService: orderService}
}

func (s PaymentsService) InitializeOrderPayment(orderID string, price price.Price) error {
	// ...
	log.Printf("initializing payment for order %s", orderID)

	go func() {
		time.Sleep(time.Millisecond * 500)
		if err := s.PostOrderPayment(orderID); err != nil {
			log.Printf("cannot post order payment: %s", err)
		}
	}()

	// имитируем задержку провайдера проведения платежей
	// time.Sleep(time.Second)

	return nil
}

func (s PaymentsService) PostOrderPayment(orderID string) error {
	log.Printf("payment for order %s done, marking order as paid", orderID)

	return s.orderService.MarkOrderAsPaid(orderID)
}
