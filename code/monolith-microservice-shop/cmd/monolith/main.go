package monolith

// да, количество импортов велико и код из-за этого выглядит не очень красиво :(
import (
	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/shop"
	"log"
	"net/http"
	"os"

	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/common/cmd"
	orders_app "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/application"
	orders_infra_orders "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/infrastructure/orders"
	orders_infra_payments "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/infrastructure/payments"
	orders_infra_product "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/infrastructure/shop"
	orders_interfaces_intraprocess "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/interfaces/private/intraprocess"
	orders_interfaces_http "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/interfaces/public/http"
	payments_app "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/payments/application"
	payments_infra_orders "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/payments/infrastructure/orders"
	payments_interfaces_intraprocess "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/payments/interfaces/intraprocess"
	shop_app "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/shop/application"
	shop_infra_product "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/shop/infrastructure/products"
	shop_interfaces_intraprocess "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/shop/interfaces/private/intraprocess"
	shop_interfaces_http "github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/shop/interfaces/public/http"
	"github.com/go-chi/chi/v5"
)

func main() {
	log.Println("Starting monolith")
	ctx := cmd.Context()

	ordersToPay := make(chan payments_interfaces_intraprocess.OrderToProcess)
	router, paymentsInterface := createMonolith(ordersToPay)
	go paymentsInterface.Run()

	server := &http.Server{Addr: os.Getenv("SHOP_MONOLITH_BIND_ADDR"), Handler: router}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()
	log.Printf("Monolith is listening on %s", server.Addr)

	<-ctx.Done()
	log.Println("Closing monolith")

	if err := server.Close(); err != nil {
		panic(err)
	}

	close(ordersToPay)
	paymentsInterface.Close()
}

func createMonolith(ordersToPay chan payments_interfaces_intraprocess.OrderToProcess) (*chi.Mux, payments_interfaces_intraprocess.PaymentsInterface) {
	shopProductRepo := shop_infra_product.NewMemoryRepository()
	shopProductsService := shop_app.NewProductsService(shopProductRepo, shopProductRepo)
	shopProductIntraprocessInterface := shop_interfaces_intraprocess.NewProductInterface(shopProductRepo)

	ordersRepo := orders_infra_orders.NewMemoryRepository()
	orderService := orders_app.NewOrdersService(
		orders_infra_product.NewIntraprocessService(shopProductIntraprocessInterface),
		orders_infra_payments.NewIntraprocessService(ordersToPay),
		ordersRepo,
	)
	ordersIntraprocessInterface := orders_interfaces_intraprocess.NewOrdersInterface(orderService)

	paymentsService := payments_app.NewPaymentsService(
		payments_infra_orders.NewIntraprocessService(ordersIntraprocessInterface),
	)
	paymentsIntraprocessInterface := payments_interfaces_intraprocess.NewPaymentsInterface(ordersToPay, paymentsService)

	if err := shop.LoadShopFixtures(shopProductsService); err != nil {
		panic(err)
	}

	r := cmd.CreateRouter()
	shop_interfaces_http.AddRoutes(r, shopProductRepo)
	orders_interfaces_http.AddRoutes(r, orderService, ordersRepo)

	return r, paymentsIntraprocessInterface
}
