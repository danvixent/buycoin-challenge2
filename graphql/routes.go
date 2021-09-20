package graphql

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/danvixent/buycoin-challenge2/handlers/account"
	log "github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/gqlerror"
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

	g := handler.GraphQL(NewExecutableSchema(c),
		handler.ErrorPresenter(
			func(ctx context.Context, err error) *gqlerror.Error {
				log.Errorf("issue carrying out graphql operation: %v", err)
				return graphql.DefaultErrorPresenter(ctx, err)
			},
		),
	)

	return g.ServeHTTP
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
