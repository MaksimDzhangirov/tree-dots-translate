package shop_test

import (
	"testing"

	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/common/price"
	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/domain/orders"
	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/orders/infrastructure/shop"
	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/shop/interfaces/private/intraprocess"
	"github.com/stretchr/testify/assert"
)

func TestOrderProductFromShopProduct(t *testing.T) {
	shopProduct := intraprocess.Product{
		ID:          "123",
		Name:        "name",
		Description: "desc",
		Price:       price.NewPriceP(42, "EUR"),
	}
	orderProduct, err := shop.OrderProductFromIntraprocess(shopProduct)
	assert.NoError(t, err)

	expectedOrderProduct, err := orders.NewProduct("123", "name", price.NewPriceP(42, "EUR"))
	assert.NoError(t, err)

	assert.EqualValues(t, expectedOrderProduct, orderProduct)
}
