package main

import (
	"fmt"
	"log"
	m "msngr"
	t "msngr/taxi"
	i "msngr/taxi/infinity"
	sh "msngr/shop"
	d "msngr/db"
	n "msngr/notify"
	s "msngr/structs"
	"net/http"
	"time"
	"errors"
)

func startAfter(check s.CheckFunc, what func()) {
	for {
		if message, ok := check(); ok {
			break
		}else {
			log.Printf("wait %v", message)
			time.Sleep(5 * time.Second)
		}
	}
	go what()
}

func GetAPIInstruments(params t.ApiParams) (t.TaxiInterface, t.AddressSupplier, error) {
	switch api_name := params.Name; api_name{
	case "infinity":
		return i.GetInfinityAPI(params), i.GetInfinityAddressSupplier(params), nil
	case "fake":
		return t.GetFakeInfinityAPI(), i.GetInfinityAddressSupplier(params), nil
	}
	return nil, nil, errors.New("Not imply name of api")
}


func main() {
	conf := m.ReadConfig()

	for _, taxi_conf := range conf.Taxis {
		external_api, address_supplier, err := GetAPIInstruments(taxi_conf.Api)

		if err != nil {
			log.Printf("Skip this taxi api [%+v]\nBecause: %v", taxi_conf.Api, err)
			continue
		}

		apiMixin := t.ExternalApiMixin{API: external_api}
		db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name, taxi_conf.Name)
		carsCache := t.NewCarsCache(external_api)
		notifier := n.NewNotifier(conf.Main.CallbackAddr, taxi_conf.Key)

		botContext := t.FormTaxiCommands(&apiMixin, db, taxi_conf.DictUrl, taxi_conf.Name, taxi_conf.Information)
		taxiContext := t.TaxiContext{API:external_api, DataBase:db, Cars:carsCache, Notifier:notifier}

		controller := m.FormBotController(botContext)

		http.HandleFunc(fmt.Sprintf("/taxi/%v", taxi_conf.Name), controller)
		startAfter(botContext.Check, func() {
			t.TaxiOrderWatch(&taxiContext, botContext)
		})


		http.HandleFunc(fmt.Sprintf("/taxi/%v/streets", taxi_conf.Name), func(w http.ResponseWriter, r *http.Request) {
			t.StreetsSearchController(w, r, address_supplier)
		})
	}

	for _, shop_conf := range conf.Shops {
		db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name, shop_conf.Name)
		bot_context := sh.FormShopCommands(db)
		shop_controller := m.FormBotController(bot_context)
		http.HandleFunc(fmt.Sprintf("/shop/%v", shop_conf.Name), shop_controller)

	}

	db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name, "")
	db.Users.SetUserPassword("test", "123")

	server_address := fmt.Sprintf(":%v", conf.Main.Port)
	log.Printf("\nStart listen and serving at: %v\n", server_address)
	server := &http.Server{
		Addr: server_address,
	}

	log.Fatal(server.ListenAndServe())
}
