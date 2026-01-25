package integration

import (
	"net/http"
	"regexp"
	"strconv"

	accdomain "github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/onsi/gomega/ghttp"
)

type TestAccrualOrder struct {
	Status  accdomain.OrderStatus
	Accrual float64
}

type TestAccrualData map[accdomain.OrderID]TestAccrualOrder

type TestAccrualServer struct {
	server        *ghttp.Server
	regexGetOrder *regexp.Regexp
	testData      TestAccrualData
}

func NewTestAccrualServer() *TestAccrualServer {
	return &TestAccrualServer{
		server:        ghttp.NewServer(),
		regexGetOrder: regexp.MustCompile(`^/api/orders/(\d+)$`),
		testData:      TestAccrualData{},
	}
}

func (s *TestAccrualServer) URL() string {
	return s.server.URL()
}

func (s *TestAccrualServer) Close() {
	s.server.Close()
}

func (s *TestAccrualServer) SetupData(data TestAccrualData) {
	s.testData = data

	s.server.RouteToHandler(http.MethodGet, s.regexGetOrder, s.handleRequest)
}

func (s *TestAccrualServer) handleRequest(out http.ResponseWriter, req *http.Request) {
	order, ok := s.getOrder(req)
	if !ok {
		ghttp.RespondWith(http.StatusNoContent, nil)(out, req)

		return
	}

	ghttp.RespondWithJSONEncoded(http.StatusOK, order)(out, req)
}

func (s *TestAccrualServer) getOrder(
	req *http.Request,
) (accdomain.Order, bool) {
	var order accdomain.Order

	orderID, valid := s.extractOrderID(req)
	if !valid {
		return order, false
	}

	orderData, ok := s.testData[orderID]
	if !ok {
		return order, false
	}

	order = accdomain.Order{
		ID:            orderID,
		Status:        orderData.Status,
		AccrualPoints: orderData.Accrual,
	}

	return order, true
}

func (s *TestAccrualServer) extractOrderID(req *http.Request) (accdomain.OrderID, bool) {
	match := s.regexGetOrder.FindStringSubmatch(req.URL.Path)
	if len(match) < 2 {
		return 0, false
	}

	orderNumber, err := strconv.ParseUint(match[1], 10, 64)
	if err != nil {
		return 0, false
	}

	return accdomain.OrderID(orderNumber), true
}
