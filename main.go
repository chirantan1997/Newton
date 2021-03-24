package main

import (
	check "Newton/check"
	controller "Newton/controllers"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	http.Handle("/", r)
	fmt.Println("Starting Server.....")
	fmt.Println("Listening on Port 8080......")

	r.HandleFunc("/api/account", controller.AccountHandler).Methods("POST")
	r.HandleFunc("/api/auth", controller.AuthHandler).Methods("POST")
	//r.HandleFunc("/api/login", controller.LoginHandler).Methods("POST")
	r.HandleFunc("/api/resend", controller.Resendotp).Methods("GET")
	r.HandleFunc("/api/carousel", controller.Carousel).Methods("GET")

	r.HandleFunc("/api/signup", controller.SignupHandler).Methods("POST")

	r.HandleFunc("/api/productslist", controller.ProductsList).Methods("POST")
	r.HandleFunc("/api/usercreation", controller.UserCreationHandler).Methods("GET")
	r.HandleFunc("/api/wishlist", controller.WishlistHandler).Methods("POST")
	r.HandleFunc("/api/wishlistproducts", controller.WishlistProductsHandler).Methods("POST")

	r.HandleFunc("/api/productdetails", controller.ProductDetailsHandler).Methods("POST")

	r.HandleFunc("/api/checkout", check.Checkout).Methods("POST")
	r.HandleFunc("/api/updatecart", controller.UpdateCart).Methods("PUT")
	r.HandleFunc("/api/searchengine", controller.SearchEngine).Methods("POST")

	r.HandleFunc("/api/cartproducts", controller.CartProducts).Methods("POST")
	r.HandleFunc("/api/cartinput", controller.CartInput).Methods("POST")
	r.HandleFunc("/api/removecartproduct", controller.RemoveCartProduct).Methods("POST")
	r.HandleFunc("/api/cartfirsttime", controller.CartFirstTime).Methods("POST")
	r.HandleFunc("/api/stock", controller.ProductStock).Methods("POST")
	r.HandleFunc("/api/cartupdate", controller.CartUpdate).Methods("POST")

	r.HandleFunc("/api/stockcheck", check.StockCheck).Methods("POST")
	r.HandleFunc("/api/currentorder", check.CurrentOrder).Methods("POST")
	r.HandleFunc("/api/intransit", check.InTransit).Methods("POST")
	r.HandleFunc("/api/currentorder", check.CurrentOrder).Methods("POST")
	r.HandleFunc("/api/pastorder", check.PastOrder).Methods("POST")

	r.HandleFunc("/api/loginNew", controller.NewLoginHandler).Methods("POST")
	r.HandleFunc("/api/editprofile", controller.ProfileHandler).Methods("POST")
	r.HandleFunc("/api/values", controller.ValueHandler).Methods("POST")
	r.HandleFunc("/api/details", controller.DetailsHandler).Methods("POST")
	r.HandleFunc("/api/stocker", controller.StockHandler).Methods("POST")
	r.HandleFunc("/api/delete", controller.DeleteHandler).Methods("POST")
	r.HandleFunc("/api/adminstock", controller.AdminStockHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", r))
}
