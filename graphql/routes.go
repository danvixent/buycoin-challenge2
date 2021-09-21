package graphql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/danvixent/buycoin-challenge2/handlers/account"
	"net/http"
)

type Handler struct {
	accountHandler *account.Handler
}

const graphqlEndpoint = "/graphql"

func NewHandler(accountHandler *account.Handler) *Handler {
	return &Handler{accountHandler: accountHandler}
}

func (h *Handler) graphqlHandler() http.HandlerFunc {
	c := Config{
		Resolvers: &Resolver{accountHandler: h.accountHandler},
	}

	s := handler.NewDefaultServer(NewExecutableSchema(c))

	return s.ServeHTTP
}

func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	graphqlHandlerFunc := h.graphqlHandler()
	mux.HandleFunc(graphqlEndpoint, handleMethod(http.MethodPost, graphqlHandlerFunc))
}

func handleMethod(method string, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "method not allowed", http.StatusBadGateway)
			return
		}
		handlerFunc(w, r)
	}
}
