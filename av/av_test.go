package av

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type mockClient struct {
	mock.Mock
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

/*
func TestGetTsDailyAdj(t *testing.T) {
	// Overwrite query limit since the calls are mocked
	go limitQueryRate(1 * time.Millisecond)

	// Test erroneous http request setup
	symbol := "No\r"
	req, _ := http.NewRequest(http.MethodGet, avURL(symbol, avAPIKey), nil)
	retResp := &http.Response{}

	testClient := new(mockClient)
	testClient.On("Do", req).Return(retResp, nil)

	_, err := getTsDailyAdj(symbol, testClient)
	assert.NotNil(t, err, "Expected a parsing error")

	// Test failing client call
	symbol = "IBM"
	req, _ = http.NewRequest(http.MethodGet, avURL(symbol, avAPIKey), nil)
	retResp = &http.Response{}
	retErr := errors.New("Client call failed")

	testClient = new(mockClient)
	testClient.On("Do", req).Return(retResp, retErr)

	_, err = getTsDailyAdj(symbol, testClient)
	assert.Equal(t, retErr, err, "Expected call to return an error from the client call")

}
*/
