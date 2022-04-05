package intraprocess

import (
	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/common/price"
	"github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop/pkg/shop/domain/products"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProductFromDomainProduct(t *testing.T) {
	productPrice := price.NewPriceP(42, "USD")
	domainProduct, err := products.NewProduct("123", "name", "desc", productPrice)
	assert.NoError(t, err)

	p := ProductFromDomainProduct(*domainProduct)

	assert.EqualValues(t, Product{
		"123",
		"name",
		"desc",
		productPrice,
	}, p)
}
