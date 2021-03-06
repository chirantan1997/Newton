package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"Newton/db"
	"Newton/helpers"
	model "Newton/models"
	"Newton/query"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	otp     string = "0000"
	trials         = 0
	testobj primitive.ObjectID
	teststr string
)

func Check(url string, method string, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/"+url {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != method {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
}

func otpauth() {
	accountSid := ""
	authToken := ""
	urlStr := ""

	max := 9999
	min := 1000
	rand.Seed(time.Now().UnixNano())
	otp = strconv.Itoa(rand.Intn(max-min+1) + min)

	msgData := url.Values{}
	msgData.Set("To", "+918338905321")
	msgData.Set("From", "+12014489733")
	msgData.Set("Body", otp)
	msgDataReader := *strings.NewReader(msgData.Encode())

	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlStr, &msgDataReader)
	req.SetBasicAuth(accountSid, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {

		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)

		if err == nil {
			fmt.Println(data["sid"])
		}
	} else {
		fmt.Println(resp.Status)
	}
}

//signup/update handler
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	Check("signup", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")
	var data model.User
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &data)
	var res model.ResponseResult
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)

	}
	filter := bson.M{"_id": data.ID}
	update := bson.M{"$set": data}
	query.UpdateOne("user", filter, update)

	res.Result = "Phone Authentication Required!"
	json.NewEncoder(w).Encode(res)
	otpauth()

}

// AccountHandler ...
func AccountHandler(w http.ResponseWriter, r *http.Request) {

	Check("account", "POST", w, r)

	w.Header().Set("Content-Type", "application/json")
	var data model.Account
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &data)
	var res model.ResponseResult
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)

	}
	var user model.User

	_ = query.FindoneID("user", data.ID, "_id").Decode(&user)
	match, err := regexp.MatchString("[0-9]{10}", user.Phone)
	//fmt.Println(match)
	if data.Exist == false && user.Phone == "" {
		res.Result = "Not registered"
		json.NewEncoder(w).Encode(res)
	} else if data.Exist == false && match {
		res.Result = "Login required"
		json.NewEncoder(w).Encode(res)
	} else if data.Exist == true {
		json.NewEncoder(w).Encode(user)

	}
}

//SignUpAuthHandler ...
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	Check("auth", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")
	var userOtp model.OtpContainer
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &userOtp)
	var res model.ResponseResult

	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	if userOtp.OtpEntered == otp && userOtp.From == "signup" {
		fmt.Println("The signUp authentication is successful!")
		res.Result = "The signUp authentication is successful!"
		json.NewEncoder(w).Encode(res)

	} else if userOtp.OtpEntered != otp && userOtp.From == "login" {
		res.Error = "OTP Did not Match!"
		json.NewEncoder(w).Encode(res)

	} else if userOtp.OtpEntered == otp && userOtp.From == "login" {
		fmt.Println("The Login authentication is successful!")
		res.Result = "The Login authentication is successful!"
		json.NewEncoder(w).Encode(res)

	} else if userOtp.OtpEntered != otp && userOtp.From == "signup" {
		if trials < 5 {
			res.Error = "OTP Did not Match!"
			json.NewEncoder(w).Encode(res)
			trials++

		} else if trials == 5 {
			res.Error = "Data Deleted, Signup again!"
			json.NewEncoder(w).Encode(res)
			collection, client, err := db.GetDBCollection("user")
			_, err = collection.DeleteOne(context.TODO(), bson.M{"phone": userOtp.Number})
			if err != nil {
				log.Fatal(err)
			}
			trials = 0
			err = client.Disconnect(context.TODO())
			if err != nil {
				log.Fatal(err)
			}
		}

	}

}

//login handler

//NewLoginHandler ...
func NewLoginHandler(w http.ResponseWriter, r *http.Request) {

	Check("loginNew", "POST", w, r)
	var login model.Newlogin
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &login)
	var res model.ResponseResult
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

	var guestcart model.Cart
	var guestwishlist model.Wishlist

	collection, client, err := db.GetDBCollection("user")
	var result model.User
	err = collection.FindOne(context.TODO(), bson.D{{Key: "phone", Value: "+91" + login.Contact}}).Decode(&result)
	fmt.Println(result)
	if err != nil {
		login.Name = "guest"
		json.NewEncoder(w).Encode(login)
		return
	}
	if err == nil {
		//res.Result = "Welcome Buddy,Enter Otp!"
		//json.NewEncoder(w).Encode(res)
		if login.Userid != result.ID {

			collection1, client1, err := db.GetDBCollection("cart")
			err = collection1.FindOne(context.TODO(), bson.D{{Key: "userid", Value: login.Userid}}).Decode(&guestcart)
			if err != nil {
				res.Error = err.Error()
				json.NewEncoder(w).Encode(res)
				return
			}
			//fmt.Println(guestcart)

			for i := 0; i < len(guestcart.Product); i++ {
				filter := bson.M{"userid": result.ID}
				update := bson.M{"$push": bson.M{"product": guestcart.Product[i]}}
				query.UpdateOne("cart", filter, update)
				if err != nil {
					res.Error = err.Error()
					json.NewEncoder(w).Encode(res)
					return
				}
			}

			//deleting the guest cart
			//_, err = collection1.DeleteOne(context.TODO(), bson.D{{Key: "userid", Value: login.Userid}})

			collection2, client2, err := db.GetDBCollection("wishlist")
			err = collection2.FindOne(context.TODO(), bson.D{{Key: "userid", Value: login.Userid}}).Decode(&guestwishlist)
			if err != nil {
				res.Error = err.Error()
				json.NewEncoder(w).Encode(res)
				return
			}

			for i := 0; i < len(guestwishlist.ItemsId); i++ {
				err1 := collection2.FindOne(context.TODO(), bson.M{"userid": result.ID, "itemsId": guestwishlist.ItemsId[i]}).Err()
				if err1 != nil {
					filter := bson.M{"userid": result.ID}
					update := bson.M{"$push": bson.M{"itemsId": guestwishlist.ItemsId[i]}}
					query.UpdateOne("wishlist", filter, update)
					if err != nil {
						res.Error = err.Error()
						json.NewEncoder(w).Encode(res)
						return
					}
				}
			}

			//deleting the guest wishlist
			//_, err = collection2.DeleteOne(context.TODO(), bson.D{{Key: "userid", Value: login.Userid}})

			//deleting the guest user

			collection3, client3, err := db.GetDBCollection("user")
			_, err = collection3.DeleteOne(context.TODO(), bson.D{{Key: "userid", Value: login.Userid}})

			err = client1.Disconnect(context.TODO())
			if err != nil {
				log.Fatal(err)
			}

			err = client2.Disconnect(context.TODO())
			if err != nil {
				log.Fatal(err)
			}

			err = client3.Disconnect(context.TODO())
			if err != nil {
				log.Fatal(err)
			}

			otpauth()

			login.Userid = result.ID
			login.Name = result.Name
			login.Contact = result.Phone
			login.Address = result.Address
			login.Email = result.Email
			json.NewEncoder(w).Encode(login)
			return

		} else {

			otpauth()
			login.Userid = result.ID
			login.Name = result.Name
			login.Contact = result.Phone
			login.Address = result.Address
			login.Email = result.Email
			json.NewEncoder(w).Encode(login)

		}

	}
	err = client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

}

//resend otp

func Resendotp(w http.ResponseWriter, r *http.Request) {

	Check("resend", "GET", w, r)

	otpauth()
}

//carousel

func Carousel(w http.ResponseWriter, r *http.Request) {

	Check("carousel", "GET", w, r)

	var picture model.Carousel

	picture.Carousel = [7]string{"https://rht007.s3.amazonaws.com/carousel/1.jpg", "https://rht007.s3.amazonaws.com/carousel/2.jpg", "https://rht007.s3.amazonaws.com/carousel/3.jpg", "https://rht007.s3.amazonaws.com/carousel/4.jpg", "https://rht007.s3.amazonaws.com/carousel/5.jpg", "https://rht007.s3.amazonaws.com/carousel/6.jpg", "https://rht007.s3.amazonaws.com/carousel/7.jpg"}

	json.NewEncoder(w).Encode(picture)

}

//productslist APi
func ProductsList(w http.ResponseWriter, r *http.Request) {

	Check("productslist", "POST", w, r)

	w.Header().Set("Content-Type", "application/json")
	var items model.Items
	var id model.Id
	var list []model.Items
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &id)

	if err != nil {
		log.Fatal(err)
	}
	filter := bson.M{"locationid": id.ID1, "subcategoryid": id.Sub}

	cursor := query.FindAll("products", filter)

	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		items.Img = nil
		items.Itemsid = nil
		if err = cursor.Decode(&items); err != nil {
			log.Fatal(err)
		}
		list = append(list, items)

	}
	json.NewEncoder(w).Encode(list)
}

//user creation

func UserCreationHandler(w http.ResponseWriter, r *http.Request) {

	Check("usercreation", "GET", w, r)

	var user model.User

	var id model.Id
	result := query.InsertOne("user", user)

	oid, _ := result.InsertedID.(primitive.ObjectID)
	id.ID1 = oid

	var wish model.Wishlist
	wish.Userid = oid

	result1 := query.InsertOne("wishlist", wish)
	oidw, _ := result1.InsertedID.(primitive.ObjectID)

	fmt.Println(oidw)

	result2 := query.InsertOne("cart", wish)
	oidc, _ := result2.InsertedID.(primitive.ObjectID)

	fmt.Println(oidc)
	json.NewEncoder(w).Encode(id)
}

// wishlist api

func WishlistHandler(w http.ResponseWriter, r *http.Request) {

	Check("wishlist", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")
	var wishlist model.Cartproduct
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &wishlist)

	if err != nil {
		log.Fatal(err)
	}

	filter := bson.M{"userid": wishlist.Userid}

	if wishlist.Status == true {
		update := bson.M{"$push": bson.M{"itemsId": wishlist.Productid}}
		query.UpdateOne("wishlist", filter, update)
		response := true
		json.NewEncoder(w).Encode(response)

	} else if wishlist.Status == false {
		update := bson.M{"$pull": bson.M{"itemsId": wishlist.Productid}}
		query.UpdateOne("wishlist", filter, update)
		response := false
		json.NewEncoder(w).Encode(response)
	}
}

//wishlist products showing api

func WishlistProductsHandler(w http.ResponseWriter, r *http.Request) {

	Check("wishlistproducts", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")
	var res model.ResponseResult
	var id model.Id
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &id)

	if err != nil {
		log.Fatal(err)

	}
	var product model.Wishlistarray
	var list []model.Items
	var item model.Items
	err = query.FindoneID("wishlist", id.ID1, "userid").Decode(&product)
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(list)

	}

	collection, client := query.Connection("products")
	for i := 0; i < len(product.Wisharr); i++ {
		item.Img = nil
		item.Itemsid = nil
		_ = collection.FindOne(context.TODO(), bson.M{"_id": product.Wisharr[i]}).Decode(&item)
		list = append(list, item)
	}
	defer query.Endconn(client)
	json.NewEncoder(w).Encode(list)
}

//product details showing api
func ProductDetailsHandler(w http.ResponseWriter, r *http.Request) {

	Check("productdetails", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")

	var id model.Id
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &id)
	var res model.ResponseResult

	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	var item model.Items
	err = query.FindoneID("products", id.ID1, "_id").Decode(&item)
	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(item)

}

//checkout api

func CheckoutHandler(w http.ResponseWriter, id primitive.ObjectID) {

	var res model.ResponseResult
	filter := bson.M{"userid": id}
	update := bson.M{"$set": bson.M{"product": bson.A{}}}

	var products model.Cart

	err := query.FindoneID("cart", id, "userid").Decode(&products)
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

	query.UpdateOne("cart", filter, update)

	var count []int
	var productid []primitive.ObjectID
	count = nil
	productid = nil
	for i := 0; i < len(products.Product); i++ {
		count = append(count, products.Product[i].Count)
		productid = append(productid, products.Product[i].P_id)
		products.Product[i].Date = time.Now()
	}

	filter1 := bson.M{"_id": id}
	collection, client := query.Connection("user")
	for i := 0; i < len(products.Product); i++ {
		update1 := bson.M{"$push": bson.M{"intransit": products.Product[i]}}
		_, err := collection.UpdateOne(context.TODO(), filter1, update1)
		if err != nil {
			log.Fatal(err)
		}
	}
	query.Endconn(client)

	//fmt.Println(productid)
	//fmt.Println(count)
	collection2, client2 := query.Connection("products")
	for i := 0; i < len(products.Product); i++ {
		filter2 := bson.M{"_id": productid[i]}
		update2 := bson.M{"$inc": bson.M{"stock": -count[i]}}
		update3 := bson.M{"$inc": bson.M{"demand": 1}}
		_, err := collection2.UpdateOne(context.TODO(), filter2, update2, options.Update().SetUpsert(true))
		if err != nil {
			log.Fatal(err)
		}
		_, err1 := collection2.UpdateOne(context.TODO(), filter2, update3, options.Update().SetUpsert(true))
		if err1 != nil {
			log.Fatal(err1)
		}

	}
	query.Endconn(client2)

}

//update cart api
func UpdateCart(w http.ResponseWriter, r *http.Request) {

	Check("updatecart", "PUT", w, r)

	w.Header().Set("Content-Type", "application/json")
	var cart model.CartContainer
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &cart)
	var res model.ResponseResult

	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

	filter := bson.M{"userid": cart.UserID}

	if bool(cart.Status) == true {
		update := bson.M{"$push": bson.M{"itemsId": cart.ItemID}}
		query.UpdateOne("cart", filter, update)

		response := true
		json.NewEncoder(w).Encode(response)

	} else if bool(cart.Status) == false {
		update1 := bson.M{"$push": bson.M{"itemsId": cart.ItemID}}

		query.UpdateOne("cart", filter, update1)

		response := false
		json.NewEncoder(w).Encode(response)
	}
}

//SearcEngine api
func SearchEngine(w http.ResponseWriter, r *http.Request) {

	Check("searchengine", "POST", w, r)

	w.Header().Set("Content-Type", "application/json")

	body, _ := ioutil.ReadAll(r.Body)

	var srch model.SearchProduct

	var res model.ResponseResult

	err := json.Unmarshal(body, &srch)
	if err != nil {
		log.Fatal(w, "error occured while unmarshling")
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
	}

	search := bson.M{"$text": bson.M{"$search": srch.Search}}

	cursor := query.FindAll("products", search)

	var show []model.Items
	var product model.Items
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		product.Img = nil
		product.Itemsid = nil
		if err = cursor.Decode(&product); err != nil {
			log.Fatal(err)
		}
		show = append(show, product)
	}

	json.NewEncoder(w).Encode(show)

}

//cart products api
func CartProducts(w http.ResponseWriter, r *http.Request) {

	Check("cartproducts", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")

	var id model.Id
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &id)
	var res model.ResponseResult

	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
	}

	var item model.Cart

	err = query.FindoneID("cart", id.ID1, "userid").Decode(&item)
	if err != nil {
		log.Fatal(err)
	}
	var prod []model.Product
	for i := 0; i < len(item.Product); i++ {
		prod = append(prod, item.Product[i])
	}

	json.NewEncoder(w).Encode(prod)

}

//Cart First Time
func CartFirstTime(w http.ResponseWriter, r *http.Request) {

	Check("cartfirsttime", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")

	var ct model.CartInput
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &ct)
	var res model.ResponseResult
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	collection, client := query.Connection("cart")

	var doc model.Cart

	err = collection.FindOne(context.TODO(), bson.M{"userid": ct.Userid, "product.p_id": ct.Product.P_id, "product.duration": ct.Product.Duration}).Decode(&doc)

	if err != nil {

		_, err = collection.UpdateOne(context.TODO(), bson.M{"userid": ct.Userid}, bson.M{"$push": bson.M{"product": ct.Product}})

		res1 := "New product added"
		json.NewEncoder(w).Encode(res1)

	} else {

		update := bson.M{"$set": bson.M{"product.$.count": ct.Product.Count + 1}}

		_, err = collection.UpdateOne(context.TODO(), bson.M{"userid": ct.Userid, "product.p_id": ct.Product.P_id, "product.duration": ct.Product.Duration}, update)
		if err != nil {
			log.Fatal(err)
		}
		res2 := "Count of product increased"
		json.NewEncoder(w).Encode(res2)

	}
	query.Endconn(client)

}

//CartInput

func CartInput(w http.ResponseWriter, r *http.Request) {

	Check("cartinput", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")

	var ct model.CartInput
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &ct)
	var res model.ResponseResult

	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
	}
	collection, client := query.Connection("cart")

	var doc model.Cart

	err = collection.FindOne(context.TODO(), bson.M{"userid": ct.Userid, "product.p_id": ct.Product.P_id, "product.count": ct.Product.Count, "product.duration": ct.Product.Duration}).Decode(&doc)
	if err != nil {
		//didn't found any match
		_, err = collection.UpdateOne(context.TODO(), bson.M{"userid": ct.Userid}, bson.M{"$push": bson.M{"product": ct.Product}})

		respn := "New Product Added"
		json.NewEncoder(w).Encode(respn)

	} else {
		// if found the match
		_, err = collection.UpdateOne(context.TODO(), bson.M{"userid": ct.Userid, "product.p_id": ct.Product.P_id}, bson.M{"$set": bson.M{"product.count": ct.Product.Count, "product.duration": ct.Product.Duration, "product._rent": ct.Product.Rent}})

		respm := "Existing Product Updated"
		json.NewEncoder(w).Encode(respm)

	}
	query.Endconn(client)
}

//Remove Cart Products

func RemoveCartProduct(w http.ResponseWriter, r *http.Request) {

	Check("removecartproduct", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")

	var rem model.RemoveCartProduct

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &rem)

	filter := bson.M{"userid": rem.UserId}
	update := bson.M{"$pull": bson.M{"product": bson.M{"p_id": rem.ProductId, "count": rem.Count, "duration": rem.Duration}}}
	query.UpdateOne("cart", filter, update)

	if err != nil {
		log.Fatal(err)
	}

	respn := "Data Removed"
	json.NewEncoder(w).Encode(respn)
}

//to be changed
func ProductStock(w http.ResponseWriter, r *http.Request) {

	Check("stock", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")

	var s model.StockId
	var sd model.StockData

	body, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal(body, &s)

	var res model.ResponseResult

	err = query.FindoneID("products", s.ProductId, "_id").Decode(&sd)
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)

	}

	json.NewEncoder(w).Encode(sd.Stock)
}

func CartUpdate(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/api/cartupdate" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	//	userid, value, productid,count,duration.

	var ct model.CartInput
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &ct)
	var res model.ResponseResult

	fmt.Println("Unmarshall body", ct)
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	collection, client, err := db.GetDBCollection("cart")
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)

	}

	var doc model.Cart
	err = collection.FindOne(context.TODO(), bson.M{"userid": ct.Userid, "product.p_id": ct.Product.P_id, "product.count": ct.Product.Count, "product.duration": ct.Product.Duration}).Decode(&doc)
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
	} else {

		if ct.Status == true {

			update1 := bson.M{"$set": bson.M{"product.$.Deposit": ct.Product.Deposit, "product.$._rent": ct.Product.Rent}, "$inc": bson.M{"product.$.count": ct.Value}} //, "$inc": bson.M{"product.$.count": ct.Value}}
			_, err := collection.UpdateOne(context.TODO(), bson.M{"userid": ct.Userid, "product.p_id": ct.Product.P_id, "product.duration": ct.Product.Duration}, update1)
			if err != nil {
				res.Error = err.Error()
				json.NewEncoder(w).Encode(res)
			}
			res2 := "Count of product increased"
			json.NewEncoder(w).Encode(res2)

		} else if ct.Status == false {

			update2 := bson.M{"$set": bson.M{"product.$.Deposit": ct.Product.Deposit, "product.$._rent": ct.Product.Rent}, "$inc": bson.M{"product.$.count": -ct.Value}}
			_, err := collection.UpdateOne(context.TODO(), bson.M{"userid": ct.Userid, "product.p_id": ct.Product.P_id, "product.duration": ct.Product.Duration}, update2)
			if err != nil {
				res.Error = err.Error()
				json.NewEncoder(w).Encode(res)
			}
			res2 := "Count of product decreased"
			json.NewEncoder(w).Encode(res2)
		}
	}

	err = client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}

func StockCheckHandler(w http.ResponseWriter, id primitive.ObjectID) {
	var res model.ResponseResult
	var products model.Cart

	err := query.FindoneID("cart", id, "userid").Decode(&products)
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	var count []int
	var productid []primitive.ObjectID
	count = nil
	productid = nil
	for i := 0; i < len(products.Product); i++ {
		count = append(count, products.Product[i].Count)
		productid = append(productid, products.Product[i].P_id)
	}
	var stock []int
	var item model.Items
	stock = nil
	collection, client := query.Connection("products")
	for i := 0; i < len(products.Product); i++ {
		err := collection.FindOne(context.TODO(), bson.M{"_id": productid[i]}).Decode(&item)
		if err != nil {
			res.Error = err.Error()
			json.NewEncoder(w).Encode(res)
			return
		}
		stock = append(stock, item.Stock)
	}

	for i := 0; i < len(products.Product); i++ {
		if count[i] > stock[i] {

			json.NewEncoder(w).Encode("fail")
			return
		}
	}
	if len(count) == 0 {
		json.NewEncoder(w).Encode("No Item in Cart")
	} else {
		json.NewEncoder(w).Encode("success")
	}

	query.Endconn(client)
}

//ProfileHandler ...
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	Check("editprofile", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")
	var data model.User
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &data)
	var res model.ResponseResult
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	filter := bson.M{"_id": data.ID}
	update := bson.M{"$set": data}
	query.UpdateOne("user", filter, update)

	res.Result = "Details successfully updated!"
	res.Error = ""
	json.NewEncoder(w).Encode(res)

}

// ValueHandler ...
func ValueHandler(w http.ResponseWriter, r *http.Request) {

	Check("values", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")

	name := r.FormValue("name")
	price := r.FormValue("price")
	details := r.FormValue("details")
	rent := r.FormValue("rent")
	deposit := r.FormValue("deposit")
	stock := r.FormValue("stock")
	subcategoryid := r.FormValue("subcategoryid")
	locationid := r.FormValue("locationid")
	subname := r.FormValue("subname")
	catname := r.FormValue("catname")

	fmt.Println(subname, catname)

	jsonData := map[string]string{"name": name, "price": price, "details": details, "rent": rent, "deposit": deposit, "subcategoryid": subcategoryid, "locationid": locationid}
	jsonValue, _ := json.Marshal(jsonData)

	detailsResponse, err := http.Post("http://localhost:8080/api/details", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		insertedID, _ := ioutil.ReadAll(detailsResponse.Body)
		fmt.Println(string(insertedID))
		fmt.Println(insertedID)

	}

	jsonData1 := map[string]string{"quantity": stock, "productid": teststr}
	jsonValue1, _ := json.Marshal(jsonData1)

	stockResponse, err := http.Post("http://localhost:8080/api/stocker", "application/json", bytes.NewBuffer(jsonValue1))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		fmt.Println("hella")
		response, _ := ioutil.ReadAll(stockResponse.Body)
		fmt.Println(string(response))
	}

	helpers.ProductImageHandler(w, r, testobj, catname, subname)

}

// DetailsHandler ...
func DetailsHandler(w http.ResponseWriter, r *http.Request) {

	Check("details", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")

	var details model.ProductDetails
	var convDetails model.ProductUpload

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &details)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
	}

	convDetails.Subcategoryid, _ = primitive.ObjectIDFromHex(details.Subcategoryid)
	convDetails.LocationID, _ = primitive.ObjectIDFromHex(details.LocationID)
	convDetails.Name = details.Name
	convDetails.Details = details.Details
	convDetails.Price, _ = strconv.Atoi(details.Price)
	convDetails.Rent, _ = strconv.Atoi(details.Rent)
	convDetails.Deposit, _ = strconv.Atoi(details.Deposit)
	convDetails.Itemsid = []primitive.ObjectID{}
	convDetails.Img = []string{}
	convDetails.Stock = 0
	convDetails.Demand = 0
	//convDetails.Createdat = time.Now()

	res := query.InsertOne("products", convDetails)
	json.NewEncoder(w).Encode(res.InsertedID)
	testobj = res.InsertedID.(primitive.ObjectID)
	teststr = testobj.Hex()
}

// StockHiandler ...
func StockHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("hellao")

	Check("stocker", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")

	var stock model.ProductStock
	var items model.ProductItems
	var itemarr []primitive.ObjectID

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &stock)
	if err != nil {
		json.NewEncoder(w).Encode("Error in Unmarshalling")
		return
	}

	quantity, _ := strconv.Atoi(stock.Quantity)
	fmt.Println("hello")
	items.Productid, _ = primitive.ObjectIDFromHex(stock.Productid)
	items.Createdat = time.Now()

	for i := 0; i < quantity; i++ {
		res := query.InsertOne("items", items)
		if err != nil {
			json.NewEncoder(w).Encode("Error in creating items")
			return
		}
		itemarr = append(itemarr, res.InsertedID.(primitive.ObjectID))
	}

	for i := 0; i < quantity; i++ {
		filter := bson.M{"_id": items.Productid}
		update := bson.M{"$push": bson.M{"itemsid": itemarr[i]}}
		query.UpdateOne("products", filter, update)
	}

	filter := bson.M{"_id": items.Productid}
	update := bson.M{"$inc": bson.M{"stock": quantity}}
	query.UpdateOne("products", filter, update)

}

// DeleteHandler ...
func DeleteHandler(w http.ResponseWriter, r *http.Request) {

	Check("delete", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")

	var delete model.Delete
	var deleted model.Deleteditems

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &delete)
	if err != nil {
		json.NewEncoder(w).Encode("Error in Unmarshalling")
		return
	}

	fmt.Println(query.DocId(delete.Productid))

	var itemsarr []primitive.ObjectID

	result := query.FindoneID("products", query.DocId(delete.Productid), "_id")
	if err = result.Decode(&deleted); err != nil {
		log.Fatal(err)
	}
	itemsarr = deleted.Itemsid
	fmt.Println(itemsarr)

	collection1, client1 := query.Connection("items")

	for i := 0; i < len(itemsarr); i++ {
		_, err = collection1.DeleteOne(context.TODO(), bson.M{"_id": itemsarr[i]})
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	defer query.Endconn(client1)

	collection2, client2 := query.Connection("products")
	_, err = collection2.DeleteOne(context.TODO(), bson.M{"_id": query.DocId(delete.Productid)})
	if err != nil {
		fmt.Println(err)
		return
	}

	defer query.Endconn(client2)

	// collection3, client3 := query.Connection("wishlist")

	// filter := bson.M{"userid": }
	// result, err = collection3.UpdateMany(context.Background(), filter, update)

}

//AdminStockHandler ...
func AdminStockHandler(w http.ResponseWriter, r *http.Request) {

	Check("adminstock", "POST", w, r)
	w.Header().Set("Content-Type", "application/json")

	var stock model.ProductStock
	var items model.ProductItems
	var product model.Items
	var itemarr []primitive.ObjectID

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &stock)
	if err != nil {
		json.NewEncoder(w).Encode("Error in Unmarshalling")
		return
	}

	quantity, _ := strconv.Atoi(stock.Quantity)
	items.Productid, _ = primitive.ObjectIDFromHex(stock.Productid)
	items.Createdat = time.Now()

	collection1, client1, err := db.GetDBCollection("products")
	err = collection1.FindOne(context.TODO(), bson.D{{Key: "_id", Value: items.Productid}}).Decode(&product)

	if product.Stock < quantity {

		for i := 0; i < (quantity - product.Stock); i++ {
			res := query.InsertOne("items", items)
			if err != nil {
				json.NewEncoder(w).Encode("Error in creating items")
				return
			}
			itemarr = append(itemarr, res.InsertedID.(primitive.ObjectID))
		}

		for i := 0; i < (quantity - product.Stock); i++ {
			filter := bson.M{"_id": items.Productid}
			update := bson.M{"$push": bson.M{"itemsid": itemarr[i]}}
			query.UpdateOne("products", filter, update)
		}

		filter := bson.M{"_id": items.Productid}
		update := bson.M{"$inc": bson.M{"stock": (quantity - product.Stock)}}
		query.UpdateOne("products", filter, update)
	} else {

		for i := 0; i < (product.Stock - quantity); i++ {
			filter := bson.M{"_id": items.Productid}
			update := bson.M{"$pull": bson.M{"itemsid": product.Itemsid[i]}}
			query.UpdateOne("products", filter, update)
		}

		filter := bson.M{"_id": items.Productid}
		update := bson.M{"$inc": bson.M{"stock": (quantity - product.Stock)}}
		query.UpdateOne("products", filter, update)

	}
	err = client1.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

}

//Intransit showing api
func IntransitHandler(w http.ResponseWriter, id primitive.ObjectID) {
	var res model.ResponseResult
	var user model.User
	collection, client := query.Connection("user")
	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	var response []model.Product
	response = nil
	for i := 0; i < len(user.InTransit); i++ {
		response = append(response, user.InTransit[i])

	}

	filter := bson.M{"_id": id}

	for i := 0; i < len(response); i++ {
		if time.Now().Day()-response[i].Date.Day() > 1 {

			update := bson.M{"$push": bson.M{"currentorder": response[i]}}
			_, err1 := collection.UpdateOne(context.TODO(), filter, update)
			if err1 != nil {
				log.Fatal(err1)
			}
			update1 := bson.M{"$pull": bson.M{"intransit": bson.M{"checkoutdate": response[i].Date}}}
			_, err := collection.UpdateOne(context.TODO(), filter, update1)
			if err != nil {
				log.Fatal(err)
			}

		}
	}
	err4 := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
	if err4 != nil {
		res.Error = err4.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	response = nil
	for i := 0; i < len(user.InTransit); i++ {
		response = append(response, user.InTransit[i])

	}
	query.Endconn(client)
	json.NewEncoder(w).Encode(response)
}

//current order showing api
func CurrentOrderHandler(w http.ResponseWriter, id primitive.ObjectID) {
	var res model.ResponseResult
	var user model.User
	collection, client := query.Connection("user")
	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	var response []model.Product
	response = nil
	for i := 0; i < len(user.CurrentOrder); i++ {
		response = append(response, user.CurrentOrder[i])

	}

	day := time.Now().Day()
	month := int(time.Now().Month())
	year := time.Now().Year()

	for i := 0; i < len(response); i++ {
		if response[i].Duration == 12 {
			d := response[i].Date.AddDate(1, 0, 0).Day()
			m := int(response[i].Date.AddDate(1, 0, 0).Month())
			y := response[i].Date.AddDate(1, 0, 0).Year()
			if year >= y && month >= m && day > d {
				query.CurrentUpdate(response[i], id, collection)
			}

		} else if response[i].Duration == 24 {
			d := response[i].Date.AddDate(2, 0, 0).Day()
			m := int(response[i].Date.AddDate(2, 0, 0).Month())
			y := response[i].Date.AddDate(2, 0, 0).Year()
			if year >= y && month >= m && d < day {
				query.CurrentUpdate(response[i], id, collection)
			}
		} else if response[i].Duration == 6 {
			d := response[i].Date.AddDate(0, 6, 0).Day()
			m := int(response[i].Date.AddDate(0, 6, 0).Month())
			y := response[i].Date.AddDate(0, 6, 0).Year()
			if year >= y && month >= m && d < day {
				query.CurrentUpdate(response[i], id, collection)
			}
		} else if response[i].Duration == 3 {
			d := response[i].Date.AddDate(0, 3, 0).Day()
			m := int(response[i].Date.AddDate(0, 3, 0).Month())
			y := response[i].Date.AddDate(0, 3, 0).Year()
			if year >= y && month >= m && d < day {
				query.CurrentUpdate(response[i], id, collection)
			}
		}
	}
	err4 := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
	if err4 != nil {
		res.Error = err4.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

	response = nil
	for i := 0; i < len(user.CurrentOrder); i++ {
		response = append(response, user.CurrentOrder[i])

	}
	json.NewEncoder(w).Encode(response)

	query.Endconn(client)
}

func PastOrderHandler(w http.ResponseWriter, id primitive.ObjectID) {
	var res model.ResponseResult
	var user model.User
	collection, client := query.Connection("user")
	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	var response []model.Product
	response = nil
	for i := 0; i < len(user.PastOrder); i++ {
		response = append(response, user.PastOrder[i])

	}
	json.NewEncoder(w).Encode(response)

	query.Endconn(client)
}
