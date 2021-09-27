package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type serverCmd struct {
	Port string `help:"listen port" default:":8080"`
}

func (c *serverCmd) Run(ctx *runctx) error {

	r := mux.NewRouter()

	showSummary := func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if len(q) == 0 {
			status, err := getStatus(r.Context(), ctx.cctx, ctx.denom, ctx.locked)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			buf, err := json.Marshal(status)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Add("Content-Type", "application/javascript")
			w.Write(buf)
		} else {
			denom := r.URL.Query().Get("denom")
			if len(denom) == 0 {
				http.Error(w, "`denom` value is required", http.StatusBadRequest)
				return
			}

			status, err := getStatus(r.Context(), ctx.cctx, denom, ctx.locked)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			metric := sdk.Int{}
			switch q {
			case "total":
				metric = status.Total.Amount
			case "circulating":
				metric = status.Circulating.Amount
			case "bonded":
				metric = status.Bonded.Amount
			default:
				http.Error(w, "wrong query value, allowed values are `total`, `circulating`, `bonded`", http.StatusInternalServerError)
				return
			}

			decString := r.URL.Query().Get("decimals")
			if len(decString) == 0 {
				http.Error(w, "`decimals` value is required", http.StatusBadRequest)
				return
			}

			decimals, err := strconv.ParseUint(decString, 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Add("Content-Type", "text/plain")
			fmt.Fprint(w, formatAmount(metric, int64(math.Pow(10, float64(decimals)))))
		}
	}

	r.HandleFunc("/supply/summary", showSummary)

	r.HandleFunc("/", showSummary)

	server := handlers.LoggingHandler(os.Stdout, r)

	fmt.Printf("running server on port %v\n\n", c.Port)

	return http.ListenAndServe(c.Port, server)
}

func formatAmount(amount sdk.Int, dec int64) string {
	whole := amount.QuoRaw(dec).Uint64()
	frac := amount.ModRaw(dec).Uint64()
	return fmt.Sprintf("%d.%06d", whole, frac)
}
